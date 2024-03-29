package indexer

import (
	"github.com/idena-network/idena-go/blockchain"
	"github.com/idena-network/idena-go/blockchain/types"
	"github.com/idena-network/idena-go/common"
	math2 "github.com/idena-network/idena-go/common/math"
	"github.com/idena-network/idena-go/config"
	"github.com/idena-network/idena-go/core/state"
	statsTypes "github.com/idena-network/idena-go/stats/types"
	"github.com/idena-network/idena-indexer/core/conversion"
	"github.com/idena-network/idena-indexer/core/stats"
	"github.com/idena-network/idena-indexer/db"
	"github.com/idena-network/idena-indexer/log"
	"github.com/ipfs/go-cid"
	"github.com/shopspring/decimal"
	"math"
	"math/big"
)

const (
	missedRewardReasonEmpty            = byte(0)
	missedRewardReasonPenalty          = byte(1)
	missedRewardReasonNotValidated     = byte(2)
	missedRewardReasonMissedValidation = byte(3)
	missedRewardReasonNotAllFlips      = byte(4)
	missedRewardReasonNotAllReports    = byte(5)
	missedRewardReasonIgnoredGrades    = byte(10)

	requiredFlips = 3
)

type validationRewardSummariesCalculator struct {
	initialized bool

	rewardsStats    *stats.RewardsStats
	validationStats *statsTypes.ValidationStats

	rewardsByAddrAndType                 map[common.Address]map[stats.RewardType]*big.Int
	rewardedReportedFlipsByAddr          map[string]map[string]struct{}
	flipsWithReportConsensusByRespondent map[common.Address][]string
	rewardedFlips                        map[string]struct{}
	rewardedExtraFlips                   map[string]struct{}

	consensusConfig *config.ConsensusConf
}

func newValidationRewardSummariesCalculator(
	rewardsStats *stats.RewardsStats,
	validationStats *statsTypes.ValidationStats,
	consensusConfig *config.ConsensusConf,
) *validationRewardSummariesCalculator {
	res := &validationRewardSummariesCalculator{
		rewardsStats:    rewardsStats,
		validationStats: validationStats,
		consensusConfig: consensusConfig,
	}
	return res
}

func (c *validationRewardSummariesCalculator) init() {
	rewardsByAddrAndType := make(map[common.Address]map[stats.RewardType]*big.Int)
	for _, reward := range c.rewardsStats.Rewards {
		addressRewardsByType, ok := rewardsByAddrAndType[reward.Address]
		if !ok {
			addressRewardsByType = make(map[stats.RewardType]*big.Int)
			rewardsByAddrAndType[reward.Address] = addressRewardsByType
		}
		addressRewardsByType[reward.Type] = new(big.Int).Add(zeroIfNil(reward.Balance), zeroIfNil(reward.Stake))
	}
	c.rewardsByAddrAndType = rewardsByAddrAndType

	rewardedReportedFlipsByAddr := make(map[string]map[string]struct{})
	for _, reportedFlipReward := range c.rewardsStats.ReportedFlipRewards {
		addressRewardedReportedFlips, ok := rewardedReportedFlipsByAddr[reportedFlipReward.Address]
		if !ok {
			addressRewardedReportedFlips = make(map[string]struct{})
			rewardedReportedFlipsByAddr[reportedFlipReward.Address] = addressRewardedReportedFlips
		}
		addressRewardedReportedFlips[reportedFlipReward.Cid] = struct{}{}
	}
	c.rewardedReportedFlipsByAddr = rewardedReportedFlipsByAddr

	flipsWithReportConsensusByRespondent := make(map[common.Address][]string)
	if c.validationStats != nil {
		for _, shardValidationStats := range c.validationStats.Shards {
			for address, identityStats := range shardValidationStats.IdentitiesPerAddr {
				var flipsWithReportConsensus []string
				for _, flipToSolve := range identityStats.LongFlipsToSolve {
					if flipStats, ok := shardValidationStats.FlipsPerIdx[flipToSolve]; ok && flipStats.Grade == types.GradeReported {
						flipCidBytes := shardValidationStats.FlipCids[flipToSolve]
						flipCid, err := cid.Parse(flipCidBytes)
						if err != nil {
							log.Error("Unable to parse flip cid. Skipped.", "bytes", flipCidBytes, "err", err)
							continue
						}
						flipsWithReportConsensus = append(flipsWithReportConsensus, convertCid(flipCid))
					}
				}
				flipsWithReportConsensusByRespondent[address] = flipsWithReportConsensus
			}
		}
	}
	c.flipsWithReportConsensusByRespondent = flipsWithReportConsensusByRespondent

	convertRewardedFlips := func(cids []string) map[string]struct{} {
		res := make(map[string]struct{}, len(cids))
		for _, flipCid := range cids {
			res[flipCid] = struct{}{}
		}
		return res
	}
	c.rewardedFlips = convertRewardedFlips(c.rewardsStats.RewardedFlipCids)
	c.rewardedExtraFlips = convertRewardedFlips(c.rewardsStats.RewardedExtraFlipCids)
}

func (c *validationRewardSummariesCalculator) calculateValidationRewardSummaries(
	address common.Address,
	shardId common.ShardId,
	age uint16,
	identityFlips []state.IdentityFlip,
	prevState state.IdentityState,
	newState state.IdentityState,
	availableFlips uint8,
	prevStake *big.Int,
	ignoredGrades bool,
	reportedFlips map[string]struct{},
) db.ValidationRewardSummaries {
	if !c.initialized {
		c.init()
		c.initialized = true
	}
	rewardsByType := c.rewardsByAddrAndType[address]
	validationResults := c.rewardsStats.ValidationResults
	shardValidationResults, shardValidationResultsOk := validationResults[shardId]
	var penalized bool
	if shardValidationResultsOk {
		_, penalized = shardValidationResults.BadAuthors[address]
	}

	validation := calculateValidationRewardSummary(
		rewardsByType[stats.Validation],
		newState,
		penalized,
		age,
		c.rewardsStats.ValidationShare,
	)

	candidate := calculateCandidateRewardSummary(
		rewardsByType[stats.Candidate],
		prevState,
		newState,
		penalized,
		c.rewardsStats.CandidateShare,
	)

	staking := calculateStakingRewardSummary(
		rewardsByType[stats.Staking],
		newState,
		penalized,
		c.rewardsStats.StakingShare,
		prevStake,
	)

	var missedValidation bool
	if shardValidationResultsOk {
		if goodAuthor, goodAuthorOk := shardValidationResults.GoodAuthors[address]; goodAuthorOk {
			missedValidation = goodAuthor.Missed
		}
	}

	flips := calculateFlipsRewardSummary(
		rewardsByType[stats.Flips],
		c.rewardsStats.FlipsShare,
		availableFlips,
		identityFlips,
		c.rewardedFlips,
		c.rewardedExtraFlips,
		penalized,
		missedValidation,
		c.consensusConfig.EnableUpgrade10,
		c.consensusConfig.EnableUpgrade11,
	)

	extraFlips := calculateExtraFlipsRewardSummary(
		rewardsByType[stats.ExtraFlips],
		c.rewardsStats.FlipsExtraShare,
		availableFlips,
		identityFlips,
		c.rewardedExtraFlips,
		penalized,
		missedValidation,
		prevStake,
	)

	invitations := calculateInvitationsRewardSummary(rewardsByType, penalized)
	invitee := calculateInviteeRewardSummary(rewardsByType, penalized, newState)

	convertedAddress := conversion.ConvertAddress(address)
	reportsShare := c.rewardsStats.ReportsShare
	if reportsShare == nil {
		var flipsShare *big.Int
		if c.rewardsStats.FlipsShare != nil {
			flipsShare = c.rewardsStats.FlipsShare
		} else {
			flipsShare = new(big.Int)
		}
		reportsShare = new(big.Int).Div(flipsShare, new(big.Int).SetInt64(5))
	}
	reports := calculateReportsRewardSummary(
		rewardsByType[stats.ReportedFlips],
		reportsShare,
		penalized,
		ignoredGrades,
		reportedFlips,
		c.rewardedReportedFlipsByAddr[convertedAddress],
		c.flipsWithReportConsensusByRespondent[address],
	)

	return db.ValidationRewardSummaries{
		Address:     convertedAddress,
		PrevStake:   prevStake,
		Validation:  validation,
		Candidate:   candidate,
		Staking:     staking,
		Flips:       flips,
		ExtraFlips:  extraFlips,
		Invitations: invitations,
		Invitee:     invitee,
		Reports:     reports,
	}
}

func getRewardedFlips(identityFlips []state.IdentityFlip, rewardedFlips map[string]struct{}) uint8 {
	var res uint8
	for _, identityFlip := range identityFlips {
		identityFlipCid, err := cid.Parse(identityFlip.Cid)
		if err != nil {
			log.Error("Unable to parse flip cid. Skipped.", "bytes", identityFlip.Cid, "err", err)
			continue
		}
		cidStr := convertCid(identityFlipCid)
		if _, ok := rewardedFlips[cidStr]; ok {
			res++
		}
	}
	return res
}

func calculateValidationRewardSummary(
	reward *big.Int,
	newState state.IdentityState,
	penalized bool,
	age uint16,
	share *big.Int,
) db.ValidationRewardSummary {
	var earned *big.Int
	if reward != nil {
		earned = new(big.Int).Set(reward)
	}
	var missed *big.Int
	var missedReason byte
	if share != nil && (!newState.NewbieOrBetter() || penalized) {
		ageCoef := float32(math.Pow(float64(age), float64(1)/3))
		missed = math2.ToInt(decimal.NewFromBigInt(share, 0).Mul(decimal.NewFromFloat32(ageCoef)))
		if missed.Sign() > 0 {
			if penalized {
				missedReason = missedRewardReasonPenalty
			} else {
				missedReason = missedRewardReasonNotValidated
			}
		}
	}
	return db.ValidationRewardSummary{
		Earned:       earned,
		Missed:       missed,
		MissedReason: missedRewardReasonOrNil(missedReason),
	}
}

func calculateCandidateRewardSummary(
	reward *big.Int,
	prevState state.IdentityState,
	newState state.IdentityState,
	penalized bool,
	share *big.Int,
) db.ValidationRewardSummary {
	var earned *big.Int
	if reward != nil {
		earned = new(big.Int).Set(reward)
	}
	var missed *big.Int
	var missedReason byte
	if share != nil && prevState == state.Candidate && (!newState.NewbieOrBetter() || penalized) {
		missed = new(big.Int).Set(share)
		if missed.Sign() > 0 {
			if penalized {
				missedReason = missedRewardReasonPenalty
			} else {
				missedReason = missedRewardReasonNotValidated
			}
		}
	}
	return db.ValidationRewardSummary{
		Earned:       earned,
		Missed:       missed,
		MissedReason: missedRewardReasonOrNil(missedReason),
	}
}

func stakeWeight(amount *big.Int) float32 {
	stakeF, _ := blockchain.ConvertToFloat(amount).Float64()
	return float32(math.Pow(stakeF, 0.9))
}

func calculateStakingRewardSummary(
	reward *big.Int,
	newState state.IdentityState,
	penalized bool,
	share *big.Int,
	stake *big.Int,
) db.ValidationRewardSummary {
	var earned *big.Int
	if reward != nil {
		earned = new(big.Int).Set(reward)
	}
	var missed *big.Int
	var missedReason byte
	if share != nil && stake != nil && stake.Sign() > 0 && (!newState.NewbieOrBetter() || penalized) {
		weight := stakeWeight(stake)
		missed = math2.ToInt(decimal.NewFromBigInt(share, 0).Mul(decimal.NewFromFloat32(weight)))
		if missed.Sign() > 0 {
			if penalized {
				missedReason = missedRewardReasonPenalty
			} else {
				missedReason = missedRewardReasonNotValidated
			}
		}
	}
	return db.ValidationRewardSummary{
		Earned:       earned,
		Missed:       missed,
		MissedReason: missedRewardReasonOrNil(missedReason),
	}
}

func calculateFlipsRewardSummary(
	reward *big.Int,
	share *big.Int,
	availableFlips uint8,
	identityFlips []state.IdentityFlip,
	rewardedFlips map[string]struct{},
	rewardedExtraFlips map[string]struct{},
	penalized bool,
	missedValidation bool,
	enableUpgrade10 bool,
	enableUpgrade11 bool,
) db.ValidationRewardSummary {
	rewardedFlipsCnt := getRewardedFlips(identityFlips, rewardedFlips)
	var missed *big.Int
	if share != nil && availableFlips > 0 {
		if enableUpgrade10 && !enableUpgrade11 {
			missed = new(big.Int).Mul(share, new(big.Int).SetUint64(uint64(requiredFlips-rewardedFlipsCnt)))
		} else {
			rewardedFlipsCnt += getRewardedFlips(identityFlips, rewardedExtraFlips)
			missed = new(big.Int).Mul(share, new(big.Int).SetUint64(uint64(availableFlips-rewardedFlipsCnt)))
		}
	}
	var missedReason byte
	if missed != nil && missed.Sign() > 0 {
		if penalized {
			missedReason = missedRewardReasonPenalty
		} else if missedValidation {
			missedReason = missedRewardReasonMissedValidation
		} else {
			missedReason = missedRewardReasonNotAllFlips
		}
	}
	var earned *big.Int
	if reward != nil {
		earned = new(big.Int).Set(reward)
	}
	return db.ValidationRewardSummary{
		Earned:       earned,
		Missed:       missed,
		MissedReason: missedRewardReasonOrNil(missedReason),
	}
}

func calculateExtraFlipsRewardSummary(
	reward *big.Int,
	share *big.Int,
	availableFlips uint8,
	identityFlips []state.IdentityFlip,
	rewardedExtraFlips map[string]struct{},
	penalized bool,
	missedValidation bool,
	stake *big.Int,
) db.ValidationRewardSummary {
	rewardedFlipsCnt := getRewardedFlips(identityFlips, rewardedExtraFlips)
	var earned *big.Int
	if reward != nil {
		earned = new(big.Int).Set(reward)
	}
	var missed *big.Int
	if share != nil && stake != nil && stake.Sign() > 0 && availableFlips > requiredFlips {
		weight := stakeWeight(stake)
		flipReward := math2.ToInt(decimal.NewFromBigInt(share, 0).Mul(decimal.NewFromFloat32(weight)))
		missed = new(big.Int).Mul(flipReward, new(big.Int).SetUint64(uint64(availableFlips-requiredFlips-rewardedFlipsCnt)))
	}
	var missedReason byte
	if missed != nil && missed.Sign() > 0 {
		if penalized {
			missedReason = missedRewardReasonPenalty
		} else if missedValidation {
			missedReason = missedRewardReasonMissedValidation
		} else {
			missedReason = missedRewardReasonNotAllFlips
		}
	}
	return db.ValidationRewardSummary{
		Earned:       earned,
		Missed:       missed,
		MissedReason: missedRewardReasonOrNil(missedReason),
	}
}

var invitationRewardTypes = [...]stats.RewardType{stats.Invitations, stats.Invitations2, stats.Invitations3, stats.SavedInvite, stats.SavedInviteWin}
var inviteeRewardTypes = [...]stats.RewardType{stats.Invitee1, stats.Invitee2, stats.Invitee3}

func calculateInvitationsRewardSummary(identityRewardsByType map[stats.RewardType]*big.Int, penalized bool) db.ValidationRewardSummary {
	var earned *big.Int
	for _, invitationRewardType := range invitationRewardTypes {
		if reward := identityRewardsByType[invitationRewardType]; reward != nil {
			if earned == nil {
				earned = new(big.Int)
			}
			earned.Add(earned, reward)
		}
	}
	var missedReason byte
	if penalized {
		missedReason = missedRewardReasonPenalty
	}
	return db.ValidationRewardSummary{
		Earned:       earned,
		MissedReason: missedRewardReasonOrNil(missedReason),
	}
}

func calculateInviteeRewardSummary(
	identityRewardsByType map[stats.RewardType]*big.Int,
	penalized bool,
	newState state.IdentityState,
) db.ValidationRewardSummary {

	var earned *big.Int
	for _, inviteeRewardType := range inviteeRewardTypes {
		if reward := identityRewardsByType[inviteeRewardType]; reward != nil {
			if earned == nil {
				earned = new(big.Int)
			}
			earned.Add(earned, reward)
		}
	}
	var missedReason byte
	if penalized {
		missedReason = missedRewardReasonPenalty
	} else if !newState.NewbieOrBetter() {
		missedReason = missedRewardReasonNotValidated
	}
	return db.ValidationRewardSummary{
		Earned:       earned,
		MissedReason: missedRewardReasonOrNil(missedReason),
	}
}

func calculateReportsRewardSummary(
	reward *big.Int,
	share *big.Int,
	penalized bool,
	ignoredGrades bool,
	reportedFlips map[string]struct{},
	rewardedReportedFlips map[string]struct{},
	flipsWithReportConsensus []string,
) db.ValidationRewardSummary {
	var earned *big.Int
	if reward != nil {
		earned = new(big.Int).Set(reward)
	}
	missedReports := 0
	var hasIgnoredSuccessReport bool
	for _, flipWithReportConsensus := range flipsWithReportConsensus {
		if ignoredGrades && !hasIgnoredSuccessReport {
			_, hasIgnoredSuccessReport = reportedFlips[flipWithReportConsensus]
		}
		if _, ok := rewardedReportedFlips[flipWithReportConsensus]; !ok {
			missedReports++
		}
	}
	var missed *big.Int
	var missedReason byte
	if missedReports > 0 {
		missed = new(big.Int).Mul(share, new(big.Int).SetUint64(uint64(missedReports)))
		if missed.Sign() > 0 {
			if penalized {
				missedReason = missedRewardReasonPenalty
			} else if hasIgnoredSuccessReport {
				missedReason = missedRewardReasonIgnoredGrades
			} else {
				missedReason = missedRewardReasonNotAllReports
			}
		}
	}
	return db.ValidationRewardSummary{
		Earned:       earned,
		Missed:       missed,
		MissedReason: missedRewardReasonOrNil(missedReason),
	}
}

func zeroIfNil(value *big.Int) *big.Int {
	if value == nil {
		return new(big.Int)
	}
	return value
}

func missedRewardReasonOrNil(value byte) *byte {
	if value == missedRewardReasonEmpty {
		return nil
	}
	return &value
}
