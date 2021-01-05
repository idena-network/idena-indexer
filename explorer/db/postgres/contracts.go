package postgres

import (
	"database/sql"
	math2 "github.com/idena-network/idena-go/common/math"
	"github.com/idena-network/idena-indexer/explorer/types"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
	"math"
	"strings"
	"time"
)

const (
	oracleVotingContractsAllQuery      = "oracleVotingContractsAll.sql"
	oracleVotingContractsAllOpenQuery  = "oracleVotingContractsAllOpen.sql"
	oracleVotingContractsByOracleQuery = "oracleVotingContractsByOracle.sql"
	ovcAllSortedByDtQuery              = "ovcAllSortedByDt.sql"
	ovcAllOpenSortedByDtQuery          = "ovcAllOpenSortedByDt.sql"
	ovcByOracleSortedByDtQuery         = "ovcByOracleSortedByDt.sql"
	oracleVotingContractQuery          = "oracleVotingContract.sql"
	lastBlockFeeRateQuery              = "lastBlockFeeRate.sql"

	oracleVotingStateOpen       = "open"
	oracleVotingStateVoted      = "voted"
	oracleVotingStateCounting   = "counting"
	oracleVotingStatePending    = "pending"
	oracleVotingStateArchive    = "archive"
	oracleVotingStateTerminated = "terminated"
)

type contractsFilter struct {
	authorAddress   *string
	stateOpen       bool
	stateVoted      bool
	stateCounting   bool
	statePending    bool
	stateArchive    bool
	stateTerminated bool
	sortByReward    bool
	all             bool
}

func createContractsFilter(authorAddress string, states []string, all bool, continuationToken *string) (*contractsFilter, error) {
	res := &contractsFilter{}
	for _, state := range states {
		switch strings.ToLower(state) {
		case oracleVotingStateOpen:
			res.stateOpen = true
		case oracleVotingStateVoted:
			res.stateVoted = true
		case oracleVotingStateCounting:
			res.stateCounting = true
		case oracleVotingStatePending:
			res.statePending = true
		case oracleVotingStateArchive:
			res.stateArchive = true
		case oracleVotingStateTerminated:
			res.stateTerminated = true
		default:
			return nil, errors.Errorf("unknown state %v", state)
		}
	}
	res.sortByReward = (res.stateOpen || res.statePending) && !(res.stateVoted || res.stateCounting || res.stateArchive || res.stateTerminated)
	if len(authorAddress) > 0 {
		res.authorAddress = &authorAddress
	}
	res.all = all
	if !res.sortByReward {
		if _, err := parseUintContinuationToken(continuationToken); err != nil {
			return nil, err
		}
	}
	return res, nil
}

func (a *postgresAccessor) OracleVotingContracts(authorAddress, oracleAddress string, states []string, all bool, count uint64, continuationToken *string) ([]types.OracleVotingContract, *string, error) {
	filter, err := createContractsFilter(authorAddress, states, all, continuationToken)
	if err != nil {
		return nil, nil, err
	}
	var rows *sql.Rows
	if filter.all {
		if !filter.stateOpen && !filter.stateVoted {
			var queryName string
			if filter.sortByReward {
				queryName = oracleVotingContractsAllQuery
			} else {
				queryName = ovcAllSortedByDtQuery
			}
			rows, err = a.db.Query(a.getQuery(queryName), filter.authorAddress, oracleAddress,
				filter.statePending, filter.stateCounting, filter.stateArchive, filter.stateTerminated, count+1,
				continuationToken)
		} else {
			var queryName string
			if filter.sortByReward {
				queryName = oracleVotingContractsAllOpenQuery
			} else {
				queryName = ovcAllOpenSortedByDtQuery
			}
			rows, err = a.db.Query(a.getQuery(queryName), filter.authorAddress, oracleAddress,
				filter.statePending, filter.stateOpen, filter.stateVoted, filter.stateCounting, filter.stateArchive,
				filter.stateTerminated, count+1, continuationToken)
		}
	} else {
		var queryName string
		if filter.sortByReward {
			queryName = oracleVotingContractsByOracleQuery
		} else {
			queryName = ovcByOracleSortedByDtQuery
		}
		rows, err = a.db.Query(a.getQuery(queryName), filter.authorAddress, oracleAddress,
			filter.statePending, filter.stateOpen, filter.stateVoted, filter.stateCounting, filter.stateArchive,
			filter.stateTerminated, count+1, continuationToken)
	}

	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	res, lastContinuationToken, err := a.readOracleVotingContracts(rows)
	if err != nil {
		return nil, nil, err
	}

	var nextContinuationToken *string
	if len(res) > 0 && len(res) == int(count)+1 {
		nextContinuationToken = lastContinuationToken
		res = res[:len(res)-1]
	}
	return res, nextContinuationToken, nil
}

func (a *postgresAccessor) OracleVotingContract(address, oracle string) (types.OracleVotingContract, error) {
	rows, err := a.db.Query(a.getQuery(oracleVotingContractQuery), address, oracle)
	if err != nil {
		return types.OracleVotingContract{}, err
	}
	defer rows.Close()
	contracts, _, err := a.readOracleVotingContracts(rows)
	if err != nil {
		return types.OracleVotingContract{}, err
	}
	if len(contracts) == 0 {
		return types.OracleVotingContract{}, NoDataFound
	}
	return contracts[0], nil
}

func (a *postgresAccessor) readOracleVotingContracts(rows *sql.Rows) ([]types.OracleVotingContract, *string, error) {
	var res []types.OracleVotingContract
	var lastContinuationToken string
	var curItem *types.OracleVotingContract
	var isFirst bool
	var networkSize *uint64
	for rows.Next() {
		item := types.OracleVotingContract{}
		var option, optionVotes, countingBlock, committeeEpoch sql.NullInt64
		var createTime, startTime, headBlockTimestamp int64
		var votingFinishTime, publicVotingFinishTime, finishTime, terminationTime sql.NullInt64
		var minPayment, totalReward NullDecimal
		var headBlockHeight uint64
		if err := rows.Scan(
			&lastContinuationToken,
			&item.ContractAddress,
			&item.Author,
			&item.Balance,
			&item.Fact,
			&item.VoteProofsCount,
			&item.VotesCount,
			&item.State,
			&option,
			&optionVotes,
			&createTime,
			&startTime,
			&headBlockHeight,
			&headBlockTimestamp,
			&votingFinishTime,
			&publicVotingFinishTime,
			&countingBlock,
			&minPayment,
			&item.Quorum,
			&item.CommitteeSize,
			&item.VotingDuration,
			&item.PublicVotingDuration,
			&item.WinnerThreshold,
			&item.OwnerFee,
			&item.IsOracle,
			&committeeEpoch,
			&finishTime,
			&terminationTime,
			&totalReward,
			&item.Stake,
		); err != nil {
			return nil, nil, err
		}
		if curItem == nil || curItem.ContractAddress != item.ContractAddress {
			if curItem != nil {
				res = append(res, *curItem)
			}
			curItem = &item
			isFirst = true
		}
		if option.Valid {
			curItem.Votes = append(curItem.Votes, types.OracleVotingContractOptionVotes{
				Option: byte(option.Int64),
				Count:  uint64(optionVotes.Int64),
			})
		}
		if !isFirst {
			continue
		}
		item.CreateTime = timestampToTimeUTC(createTime)
		item.StartTime = timestampToTimeUTC(startTime)
		itemState := strings.ToLower(item.State)
		if countingBlock.Valid {
			if itemState == oracleVotingStateOpen || itemState == oracleVotingStateVoted || itemState == oracleVotingStateCounting || itemState == oracleVotingStateArchive {
				headBlockTime := timestampToTimeUTC(headBlockTimestamp)

				if networkSize == nil {
					size, err := a.networkSizeLoader.Load()
					if err != nil {
						return nil, nil, errors.Wrap(err, "Unable to load network size")
					}
					networkSize = &size
				}
				d, _ := item.Stake.Mul(decimal.NewFromInt(int64(*networkSize))).Div(decimal.NewFromInt(100)).Float64()
				terminationDays := uint64(math.Round(math.Pow(d, 1.0/3)))
				const blocksInDay = 4320

				estimatedTerminationTime := headBlockTime.Add(time.Second * 20 * time.Duration(uint64(countingBlock.Int64)-headBlockHeight+curItem.PublicVotingDuration+terminationDays*blocksInDay))
				item.EstimatedTerminationTime = &estimatedTerminationTime
				if itemState == oracleVotingStateOpen || itemState == oracleVotingStateVoted || itemState == oracleVotingStateCounting {
					estimatedPublicVotingFinishTime := headBlockTime.Add(time.Second * 20 * time.Duration(uint64(countingBlock.Int64)-headBlockHeight+curItem.PublicVotingDuration))
					item.EstimatedPublicVotingFinishTime = &estimatedPublicVotingFinishTime
					if itemState == oracleVotingStateOpen || itemState == oracleVotingStateVoted {
						estimatedVotingFinishTime := headBlockTime.Add(time.Second * 20 * time.Duration(uint64(countingBlock.Int64)-headBlockHeight))
						item.EstimatedVotingFinishTime = &estimatedVotingFinishTime
					}
				}
			}
		}
		if committeeEpoch.Valid {
			v := uint64(committeeEpoch.Int64)
			item.CommitteeEpoch = &v
		}
		if minPayment.Valid {
			item.MinPayment = &minPayment.Decimal
		}
		if votingFinishTime.Valid {
			v := timestampToTimeUTC(votingFinishTime.Int64)
			item.VotingFinishTime = &v
		}
		if publicVotingFinishTime.Valid {
			v := timestampToTimeUTC(publicVotingFinishTime.Int64)
			item.PublishVotingFinishTime = &v
		}
		if finishTime.Valid {
			v := timestampToTimeUTC(finishTime.Int64)
			item.FinishTime = &v
		}
		if terminationTime.Valid {
			v := timestampToTimeUTC(terminationTime.Int64)
			item.TerminationTime = &v
		}
		if totalReward.Valid {
			item.TotalReward = &totalReward.Decimal
		}

		if itemState == oracleVotingStatePending || itemState == oracleVotingStateOpen || itemState == oracleVotingStateVoted || itemState == oracleVotingStateCounting {
			item.EstimatedOracleReward = calculateEstimatedOracleReward(item.Balance, item.MinPayment, item.OwnerFee, item.CommitteeSize, item.VoteProofsCount)
			item.EstimatedMaxOracleReward = calculateEstimatedMaxOracleReward(item.Balance, item.MinPayment, item.OwnerFee, item.CommitteeSize, item.Quorum, item.WinnerThreshold, item.VoteProofsCount)
			item.EstimatedTotalReward = calculateEstimatedTotalReward(item.Balance, item.MinPayment, item.OwnerFee, item.VoteProofsCount)
		}

		isFirst = false
	}
	if curItem != nil {
		res = append(res, *curItem)
	}
	return res, &lastContinuationToken, nil
}

func calculateEstimatedOracleReward(balance decimal.Decimal, votingMinPaymentP *decimal.Decimal, ownerFee uint8, committeeSize, votesCnt uint64) *decimal.Decimal {
	var ownerReward, votingMinPayment decimal.Decimal
	if votingMinPaymentP != nil {
		votingMinPayment = *votingMinPaymentP
	}
	potentialBalance := balance
	if committeeSize == 0 {
		committeeSize = 1
	}
	if votesCnt > committeeSize {
		committeeSize = votesCnt
	}
	if committeeSize > votesCnt && votingMinPayment.Sign() == 1 {
		potentialBalance = potentialBalance.Add(votingMinPayment.Mul(decimal.NewFromInt(int64(committeeSize - votesCnt))))
	}
	committeeSizeD := decimal.NewFromInt(int64(committeeSize))
	if ownerFee > 0 {
		ownerReward = potentialBalance.Sub(votingMinPayment.Mul(committeeSizeD)).Mul(decimal.NewFromFloat(float64(ownerFee) / 100.0))
	}
	oracleReward := potentialBalance.Sub(ownerReward).Div(committeeSizeD)
	return &oracleReward
}

func calculateEstimatedMaxOracleReward(balance decimal.Decimal, votingMinPaymentP *decimal.Decimal, ownerFee uint8, committeeSize uint64, quorum, winnerThreshold byte, votesCnt uint64) *decimal.Decimal {
	quorumSizeF := float64(committeeSize) * float64(quorum) / 100.0
	quorumSize := uint64(quorumSizeF)
	if quorumSizeF > float64(quorumSize) || quorumSize == 0 {
		quorumSize += 1
	}
	minVotesCnt := math2.Max(quorumSize, votesCnt)

	var votingMinPayment decimal.Decimal
	if votingMinPaymentP != nil {
		votingMinPayment = *votingMinPaymentP
	}

	potentialBalance := balance
	if minVotesCnt > votesCnt && votingMinPayment.Sign() == 1 {
		potentialBalance = potentialBalance.Add(votingMinPayment.Mul(decimal.NewFromInt(int64(minVotesCnt - votesCnt))))
	}

	var ownerReward decimal.Decimal
	minVotesCntD := decimal.NewFromInt(int64(minVotesCnt))
	if ownerFee > 0 {
		ownerReward = potentialBalance.Sub(votingMinPayment.Mul(minVotesCntD)).Mul(decimal.NewFromFloat(float64(ownerFee) / 100.0))
	}

	oracleReward := potentialBalance.Sub(ownerReward).Div(minVotesCntD.Mul(decimal.New(int64(winnerThreshold), -2)))
	return &oracleReward
}

func calculateEstimatedTotalReward(balance decimal.Decimal, votingMinPaymentP *decimal.Decimal, ownerFee uint8, votesCnt uint64) *decimal.Decimal {
	var votingMinPayment decimal.Decimal
	if votingMinPaymentP != nil {
		votingMinPayment = *votingMinPaymentP
	}
	votesCntD := decimal.NewFromInt(int64(votesCnt))
	ownerReward := balance.Sub(votingMinPayment.Mul(votesCntD)).Mul(decimal.NewFromFloat(float64(ownerFee) / 100.0))
	totalReward := balance.Sub(ownerReward)
	return &totalReward
}

func (a *postgresAccessor) EstimatedOracleRewards(committeeSize uint64) ([]types.EstimatedOracleReward, error) {
	return a.estimatedOracleRewardsCache.get(committeeSize)
}

func (a *postgresAccessor) lastBlockFeeRate() (decimal.Decimal, error) {
	rows, err := a.db.Query(a.getQuery(lastBlockFeeRateQuery))
	var res decimal.Decimal
	if err != nil {
		return res, err
	}
	defer rows.Close()
	if rows.Next() {
		if err = rows.Scan(&res); err != nil {
			return res, err
		}
	}
	return res, nil
}
