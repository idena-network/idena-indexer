package common

import (
	"database/sql"
	"github.com/idena-network/idena-go/common/hexutil"
	"github.com/idena-network/idena-indexer/db"
	"github.com/shopspring/decimal"
)

type NullDecimal struct {
	Decimal decimal.Decimal
	Valid   bool
}

func (n *NullDecimal) Scan(value interface{}) error {
	n.Valid = value != nil
	n.Decimal = decimal.Decimal{}
	if n.Valid {
		return n.Decimal.Scan(value)
	}
	return nil
}

type BalanceUpdate struct {
	Address              string
	BalanceOld           decimal.Decimal
	StakeOld             decimal.Decimal
	PenaltyOld           *decimal.Decimal
	BalanceNew           decimal.Decimal
	StakeNew             decimal.Decimal
	PenaltyNew           *decimal.Decimal
	Reason               db.BalanceUpdateReason
	BlockHeight          int
	TxId                 *int
	LastBlockHeight      *int
	CommitteeRewardShare *decimal.Decimal
	BlocksCount          *int
}

func GetBalanceUpdates(db *sql.DB) ([]BalanceUpdate, error) {
	rows, err := db.Query(`select a.address, 
       bu.balance_old, 
       bu.stake_old,
       bu.penalty_old,
       bu.balance_new, 
       bu.stake_new,
       bu.penalty_new,
       bu.reason,
       bu.block_height,
       bu.tx_id,
       bu.last_block_height,
       bu.committee_reward_share,
       bu.blocks_count
       from balance_updates bu 
           join addresses a on a.id=bu.address_id order by bu.id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []BalanceUpdate
	for rows.Next() {
		item := BalanceUpdate{}
		var txId, lastBlockHeight, blocksCount sql.NullInt32
		var committeeRewardShare NullDecimal
		var penaltyOld, penaltyNew NullDecimal
		err := rows.Scan(
			&item.Address,
			&item.BalanceOld,
			&item.StakeOld,
			&penaltyOld,
			&item.BalanceNew,
			&item.StakeNew,
			&penaltyNew,
			&item.Reason,
			&item.BlockHeight,
			&txId,
			&lastBlockHeight,
			&committeeRewardShare,
			&blocksCount,
		)
		if txId.Valid {
			v := int(txId.Int32)
			item.TxId = &v
		}
		if lastBlockHeight.Valid {
			v := int(lastBlockHeight.Int32)
			item.LastBlockHeight = &v
		}
		if committeeRewardShare.Valid {
			item.CommitteeRewardShare = &committeeRewardShare.Decimal
		}
		if blocksCount.Valid {
			v := int(blocksCount.Int32)
			item.BlocksCount = &v
		}
		if penaltyOld.Valid {
			item.PenaltyOld = &penaltyOld.Decimal
		}
		if penaltyNew.Valid {
			item.PenaltyNew = &penaltyNew.Decimal
		}
		if err != nil {
			return nil, err
		}
		res = append(res, item)
	}
	return res, nil
}

type CommitteeRewardBalanceUpdate struct {
	BlockHeight int
	Address     string
	BalanceOld  decimal.Decimal
	StakeOld    decimal.Decimal
	PenaltyOld  *decimal.Decimal
	BalanceNew  decimal.Decimal
	StakeNew    decimal.Decimal
	PenaltyNew  *decimal.Decimal
}

func GetCommitteeRewardBalanceUpdates(db *sql.DB) ([]CommitteeRewardBalanceUpdate, error) {
	rows, err := db.Query(`select a.address, 
       bu.balance_old, 
       bu.stake_old,
       bu.penalty_old,
       bu.balance_new, 
       bu.stake_new,
       bu.penalty_new,
       bu.block_height
       from latest_committee_reward_balance_updates bu 
           join addresses a on a.id=bu.address_id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []CommitteeRewardBalanceUpdate
	for rows.Next() {
		item := CommitteeRewardBalanceUpdate{}
		var penaltyOld, penaltyNew NullDecimal
		err := rows.Scan(
			&item.Address,
			&item.BalanceOld,
			&item.StakeOld,
			&penaltyOld,
			&item.BalanceNew,
			&item.StakeNew,
			&penaltyNew,
			&item.BlockHeight,
		)
		if err != nil {
			return nil, err
		}
		if penaltyOld.Valid {
			item.PenaltyOld = &penaltyOld.Decimal
		}
		if penaltyNew.Valid {
			item.PenaltyNew = &penaltyNew.Decimal
		}
		res = append(res, item)
	}
	return res, nil
}

type EpochIdentity struct {
	Id                    int
	ShortAnswers          int
	LongAnswers           int
	TotalValidationReward decimal.Decimal
}

func GetEpochIdentities(db *sql.DB) ([]EpochIdentity, error) {
	rows, err := db.Query(`select address_state_id, short_answers, long_answers, total_validation_reward from epoch_identities`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []EpochIdentity
	for rows.Next() {
		item := EpochIdentity{}
		err := rows.Scan(
			&item.Id,
			&item.ShortAnswers,
			&item.LongAnswers,
			&item.TotalValidationReward,
		)
		if err != nil {
			return nil, err
		}
		res = append(res, item)
	}
	return res, nil
}

type AddressSummary struct {
	AddressId       int
	Flips           int
	WrongWordsFlips int
}

func GetAddressSummaries(db *sql.DB) ([]AddressSummary, error) {
	rows, err := db.Query(`select address_id, flips, wrong_words_flips from address_summaries`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []AddressSummary
	for rows.Next() {
		item := AddressSummary{}
		err := rows.Scan(
			&item.AddressId,
			&item.Flips,
			&item.WrongWordsFlips,
		)
		if err != nil {
			return nil, err
		}
		res = append(res, item)
	}
	return res, nil
}

type TransactionRaw struct {
	TxId int
	Raw  hexutil.Bytes
}

func GetTransactionRaws(db *sql.DB) ([]TransactionRaw, error) {
	rows, err := db.Query(`select tx_id, raw from transaction_raws`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []TransactionRaw
	for rows.Next() {
		item := TransactionRaw{}
		err := rows.Scan(
			&item.TxId,
			&item.Raw,
		)
		if err != nil {
			return nil, err
		}
		res = append(res, item)
	}
	return res, nil
}

type ValidationReward struct {
	Address string
	Epoch   int
	Balance decimal.Decimal
	Stake   decimal.Decimal
	Type    int
}

func GetValidationRewards(db *sql.DB) ([]ValidationReward, error) {
	rows, err := db.Query(`select a.address, ei.epoch, vr.balance, vr.stake, vr.type from validation_rewards vr
join epoch_identities ei on ei.address_state_id=vr.ei_address_state_id
join address_states s on s.id=ei.address_state_id
join addresses a on a.id=s.address_id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []ValidationReward
	for rows.Next() {
		item := ValidationReward{}
		err := rows.Scan(
			&item.Address,
			&item.Epoch,
			&item.Balance,
			&item.Stake,
			&item.Type,
		)
		if err != nil {
			return nil, err
		}
		res = append(res, item)
	}
	return res, nil
}

type ReportedFlipRewards struct {
	Address  string
	Epoch    int
	FlipTxId int
	Balance  decimal.Decimal
	Stake    decimal.Decimal
}

func GetReportedFlipRewards(db *sql.DB) ([]ReportedFlipRewards, error) {
	rows, err := db.Query(`select a.address, rfr.epoch, rfr.flip_tx_id, rfr.balance, rfr.stake from reported_flip_rewards rfr
join addresses a on a.id=rfr.address_id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []ReportedFlipRewards
	for rows.Next() {
		item := ReportedFlipRewards{}
		err := rows.Scan(
			&item.Address,
			&item.Epoch,
			&item.FlipTxId,
			&item.Balance,
			&item.Stake,
		)
		if err != nil {
			return nil, err
		}
		res = append(res, item)
	}
	return res, nil
}

type Contract struct {
	TxId    int
	Address string
	Type    int
	Stake   decimal.Decimal
}

func GetContracts(db *sql.DB) ([]Contract, error) {
	rows, err := db.Query(`select c.tx_id, a.address, c.type, c.stake from contracts c join addresses a on a.id=c.contract_address_id order by c.tx_id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []Contract
	for rows.Next() {
		item := Contract{}
		err := rows.Scan(
			&item.TxId,
			&item.Address,
			&item.Type,
			&item.Stake,
		)
		if err != nil {
			return nil, err
		}
		res = append(res, item)
	}
	return res, nil
}

type OracleVotingContract struct {
	TxId                 int
	StartTime            int64
	VotingDuration       int
	VotingMinPayment     *decimal.Decimal
	Fact                 []byte
	PublicVotingDuration int
	WinnerThreshold      int
	Quorum               int
	CommitteeSize        int
	OwnerFee             int
	Epoch                int
}

func GetOracleVotingContracts(db *sql.DB) ([]OracleVotingContract, error) {
	rows, err := db.Query(`select * from oracle_voting_contracts order by contract_tx_id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []OracleVotingContract
	for rows.Next() {
		item := OracleVotingContract{}
		var votingMinPayment NullDecimal
		err := rows.Scan(
			&item.TxId,
			&item.StartTime,
			&item.VotingDuration,
			&votingMinPayment,
			&item.Fact,
			&item.PublicVotingDuration,
			&item.WinnerThreshold,
			&item.Quorum,
			&item.CommitteeSize,
			&item.OwnerFee,
			&item.Epoch,
		)
		if err != nil {
			return nil, err
		}
		if votingMinPayment.Valid {
			item.VotingMinPayment = &votingMinPayment.Decimal
		}
		res = append(res, item)
	}
	return res, nil
}

type SortedOracleVotingContract struct {
	TxId          int
	Author        string
	SortKey       *string
	State         int
	CountingBlock *int
	Epoch         *int
}

func GetSortedOracleVotingContracts(db *sql.DB) ([]SortedOracleVotingContract, error) {
	rows, err := db.Query(`select t.contract_tx_id, a.address, t.sort_key, t.state, t.counting_block, t.epoch from sorted_oracle_voting_contracts t join addresses a on a.id=t.author_address_id order by contract_tx_id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []SortedOracleVotingContract
	for rows.Next() {
		item := SortedOracleVotingContract{}
		var countingBlock, epoch sql.NullInt32
		var sortKey sql.NullString
		err := rows.Scan(
			&item.TxId,
			&item.Author,
			&sortKey,
			&item.State,
			&countingBlock,
			&epoch,
		)
		if err != nil {
			return nil, err
		}
		if sortKey.Valid {
			item.SortKey = &sortKey.String
		}
		if countingBlock.Valid {
			v := int(countingBlock.Int32)
			item.CountingBlock = &v
		}
		if epoch.Valid {
			v := int(epoch.Int32)
			item.Epoch = &v
		}
		res = append(res, item)
	}
	return res, nil
}

type SortedOracleVotingContractCommittee struct {
	TxId    int
	Author  string
	SortKey *string
	State   int
	Address string
	Voted   bool
}

func GetSortedOracleVotingContractCommittees(db *sql.DB) ([]SortedOracleVotingContractCommittee, error) {
	rows, err := db.Query(`select c.contract_tx_id, aa.address, c.sort_key, c.state, a.address, c.voted from sorted_oracle_voting_contract_committees c 
    join addresses a on a.id=c.address_id
join addresses aa on aa.id=c.author_address_id
order by c.contract_tx_id, a.address`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []SortedOracleVotingContractCommittee
	for rows.Next() {
		item := SortedOracleVotingContractCommittee{}
		var sortKey sql.NullString
		err := rows.Scan(
			&item.TxId,
			&item.Author,
			&sortKey,
			&item.State,
			&item.Address,
			&item.Voted,
		)
		if err != nil {
			return nil, err
		}
		if sortKey.Valid {
			item.SortKey = &sortKey.String
		}
		res = append(res, item)
	}
	return res, nil
}

type OracleVotingSummary struct {
	ContractTxId         int
	VoteProofs           int
	Votes                int
	FinishTimestamp      *int64
	TerminationTimestamp *int64
	TotalReward          *decimal.Decimal
	Stake                decimal.Decimal
	SecretVotesCount     *int
	EpochWithoutGrowth   *int
}

func GetOracleVotingSummaries(db *sql.DB) ([]OracleVotingSummary, error) {
	rows, err := db.Query(`select * from oracle_voting_contract_summaries order by contract_tx_id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []OracleVotingSummary
	for rows.Next() {
		item := OracleVotingSummary{}
		var finishTimestamp, terminationTimestamp, secretVotesCount, epochWithoutGrowth sql.NullInt64
		var totalReward NullDecimal
		err := rows.Scan(
			&item.ContractTxId,
			&item.VoteProofs,
			&item.Votes,
			&finishTimestamp,
			&terminationTimestamp,
			&totalReward,
			&item.Stake,
			&secretVotesCount,
			&epochWithoutGrowth,
		)
		if err != nil {
			return nil, err
		}
		if finishTimestamp.Valid {
			item.FinishTimestamp = &finishTimestamp.Int64
		}
		if terminationTimestamp.Valid {
			item.TerminationTimestamp = &terminationTimestamp.Int64
		}
		if totalReward.Valid {
			item.TotalReward = &totalReward.Decimal
		}
		if secretVotesCount.Valid {
			v := int(secretVotesCount.Int64)
			item.SecretVotesCount = &v
		}
		if epochWithoutGrowth.Valid {
			v := int(epochWithoutGrowth.Int64)
			item.EpochWithoutGrowth = &v
		}
		res = append(res, item)
	}
	return res, nil
}

type OracleVotingResult struct {
	ContractTxId int
	Option       int
	Count        int
	AllCount     *int
}

func GetOracleVotingResults(db *sql.DB) ([]OracleVotingResult, error) {
	rows, err := db.Query(`select * from oracle_voting_contract_results order by contract_tx_id, option`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []OracleVotingResult
	for rows.Next() {
		item := OracleVotingResult{}
		var allCount sql.NullInt32
		err := rows.Scan(
			&item.ContractTxId,
			&item.Option,
			&item.Count,
			&allCount,
		)
		if err != nil {
			return nil, err
		}
		if allCount.Valid {
			v := int(allCount.Int32)
			item.AllCount = &v
		}
		res = append(res, item)
	}
	return res, nil
}

//type OracleVotingContractState struct {
//	TxId         int
//	PrevTxId     *int
//	NextTxId     *int
//	ContractTxId int
//	State        int
//	Reason       int
//}
//
//func GetOracleVotingContractStates(db *sql.DB) ([]OracleVotingContractState, error) {
//	rows, err := db.Query(`select * from oracle_voting_contract_states order by state_tx_id`)
//	if err != nil {
//		return nil, err
//	}
//	defer rows.Close()
//	var res []OracleVotingContractState
//	for rows.Next() {
//		item := OracleVotingContractState{}
//		var prevTxId, nextTxId sql.NullInt32
//		err := rows.Scan(
//			&item.TxId,
//			&prevTxId,
//			&nextTxId,
//			&item.ContractTxId,
//			&item.State,
//			&item.Reason,
//		)
//		if err != nil {
//			return nil, err
//		}
//		if prevTxId.Valid {
//			v := int(prevTxId.Int32)
//			item.PrevTxId = &v
//		}
//		if nextTxId.Valid {
//			v := int(nextTxId.Int32)
//			item.NextTxId = &v
//		}
//		res = append(res, item)
//	}
//	return res, nil
//}

type OracleVotingContractCallStart struct {
	TxId             int
	ContractTxId     int
	StartBlockHeight int
	Epoch            int
	VotingMinPayment *decimal.Decimal
	VrfSeed          []byte
	State            int
}

func GetOracleVotingContractCallStarts(db *sql.DB) ([]OracleVotingContractCallStart, error) {
	rows, err := db.Query(`select * from oracle_voting_contract_call_starts order by call_tx_id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []OracleVotingContractCallStart
	for rows.Next() {
		item := OracleVotingContractCallStart{}
		var votingMinPayment NullDecimal
		err := rows.Scan(
			&item.TxId,
			&item.ContractTxId,
			&item.StartBlockHeight,
			&item.Epoch,
			&votingMinPayment,
			&item.VrfSeed,
			&item.State,
		)
		if err != nil {
			return nil, err
		}
		if votingMinPayment.Valid {
			item.VotingMinPayment = &votingMinPayment.Decimal
		}
		res = append(res, item)
	}
	return res, nil
}

type OracleVotingContractCallVoteProof struct {
	TxId         int
	ContractTxId int
	Address      string
	VoteHash     []byte
}

func GetOracleVotingContractCallVoteProofs(db *sql.DB) ([]OracleVotingContractCallVoteProof, error) {
	rows, err := db.Query(`select t.call_tx_id, t.ov_contract_tx_id, a.address, t.vote_hash from oracle_voting_contract_call_vote_proofs t join addresses a on a.id=t.address_id order by t.call_tx_id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []OracleVotingContractCallVoteProof
	for rows.Next() {
		item := OracleVotingContractCallVoteProof{}
		err := rows.Scan(
			&item.TxId,
			&item.ContractTxId,
			&item.Address,
			&item.VoteHash,
		)
		if err != nil {
			return nil, err
		}
		res = append(res, item)
	}
	return res, nil
}

type OracleVotingContractCallVote struct {
	TxId             int
	ContractTxId     int
	Vote             byte
	Salt             []byte
	OptionVotes      *int
	OptionAllVotes   *int
	SecretVotesCount *int
	Delegatee        *string
	PrevPoolVote     *int
	PrevOptionVotes  *int
}

func GetOracleVotingContractCallVotes(db *sql.DB) ([]OracleVotingContractCallVote, error) {
	rows, err := db.Query(`select t.call_tx_id, t.ov_contract_tx_id, t.vote, t.salt, t.option_votes, t.option_all_votes, 
       t.secret_votes_count, a.address, t.prev_pool_vote, t.prev_option_votes from oracle_voting_contract_call_votes  t
           left join addresses a on a.id=t.delegatee_address_id
order by call_tx_id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []OracleVotingContractCallVote
	for rows.Next() {
		var optionVotes, optionAllVotes, secretVotesCount, prevPoolVote, prevOptionVotes sql.NullInt32
		var delegatee sql.NullString
		item := OracleVotingContractCallVote{}
		err := rows.Scan(
			&item.TxId,
			&item.ContractTxId,
			&item.Vote,
			&item.Salt,
			&optionVotes,
			&optionAllVotes,
			&secretVotesCount,
			&delegatee,
			&prevPoolVote,
			&prevOptionVotes,
		)
		if err != nil {
			return nil, err
		}
		if optionVotes.Valid {
			v := int(optionVotes.Int32)
			item.OptionVotes = &v
		}
		if optionAllVotes.Valid {
			v := int(optionAllVotes.Int32)
			item.OptionAllVotes = &v
		}
		if secretVotesCount.Valid {
			v := int(secretVotesCount.Int32)
			item.SecretVotesCount = &v
		}
		if delegatee.Valid {
			item.Delegatee = &delegatee.String
		}
		if prevPoolVote.Valid {
			v := int(prevPoolVote.Int32)
			item.PrevPoolVote = &v
		}
		if prevOptionVotes.Valid {
			v := int(prevOptionVotes.Int32)
			item.PrevOptionVotes = &v
		}
		res = append(res, item)
	}
	return res, nil
}

type OracleVotingContractCallFinish struct {
	TxId         int
	ContractTxId int
	Result       *byte
	Fund         decimal.Decimal
	OracleReward decimal.Decimal
	OwnerReward  decimal.Decimal
	State        int
}

func GetOracleVotingContractCallFinishes(db *sql.DB) ([]OracleVotingContractCallFinish, error) {
	rows, err := db.Query(`select * from oracle_voting_contract_call_finishes order by call_tx_id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []OracleVotingContractCallFinish
	for rows.Next() {
		item := OracleVotingContractCallFinish{}
		var result sql.NullInt32
		err := rows.Scan(
			&item.TxId,
			&item.ContractTxId,
			&result,
			&item.Fund,
			&item.OracleReward,
			&item.OwnerReward,
			&item.State,
		)
		if err != nil {
			return nil, err
		}
		if result.Valid {
			v := byte(result.Int32)
			item.Result = &v
		}
		res = append(res, item)
	}
	return res, nil
}

type OracleVotingContractCallProlongation struct {
	TxId               int
	ContractTxId       int
	Epoch              int
	StartBlock         *int
	VrfSeed            []byte
	EpochWithoutGrowth *int
	ProlongVoteCount   *int
}

func GetOracleVotingContractCallProlongations(db *sql.DB) ([]OracleVotingContractCallProlongation, error) {
	rows, err := db.Query(`select * from oracle_voting_contract_call_prolongations order by call_tx_id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []OracleVotingContractCallProlongation
	for rows.Next() {
		item := OracleVotingContractCallProlongation{}
		var startBlock, epochWithoutGrowth, prolongVoteCount sql.NullInt32
		err := rows.Scan(
			&item.TxId,
			&item.ContractTxId,
			&item.Epoch,
			&startBlock,
			&item.VrfSeed,
			&epochWithoutGrowth,
			&prolongVoteCount,
		)
		if err != nil {
			return nil, err
		}
		if startBlock.Valid {
			v := int(startBlock.Int32)
			item.StartBlock = &v
		}
		if epochWithoutGrowth.Valid {
			v := int(epochWithoutGrowth.Int32)
			item.EpochWithoutGrowth = &v
		}
		if prolongVoteCount.Valid {
			v := int(prolongVoteCount.Int32)
			item.ProlongVoteCount = &v
		}
		res = append(res, item)
	}
	return res, nil
}

type OracleVotingContractCallAddStake struct {
	TxId         int
	ContractTxId int
}

func GetOracleVotingContractCallAddStakes(db *sql.DB) ([]OracleVotingContractCallAddStake, error) {
	rows, err := db.Query(`select * from oracle_voting_contract_call_add_stakes order by call_tx_id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []OracleVotingContractCallAddStake
	for rows.Next() {
		item := OracleVotingContractCallAddStake{}
		err := rows.Scan(
			&item.TxId,
			&item.ContractTxId,
		)
		if err != nil {
			return nil, err
		}
		res = append(res, item)
	}
	return res, nil
}

type OracleVotingContractTermination struct {
	TxId         int
	ContractTxId int
	Fund         *decimal.Decimal
	OracleReward *decimal.Decimal
	OwnerReward  *decimal.Decimal
}

func GetOracleVotingContractTerminations(db *sql.DB) ([]OracleVotingContractTermination, error) {
	rows, err := db.Query(`select * from oracle_voting_contract_terminations order by termination_tx_id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []OracleVotingContractTermination
	for rows.Next() {
		var fund, oracleReward, ownerReward NullDecimal
		item := OracleVotingContractTermination{}
		err := rows.Scan(
			&item.TxId,
			&item.ContractTxId,
			&fund,
			&oracleReward,
			&ownerReward,
		)
		if err != nil {
			return nil, err
		}
		if fund.Valid {
			item.Fund = &fund.Decimal
		}
		if oracleReward.Valid {
			item.OracleReward = &oracleReward.Decimal
		}
		if ownerReward.Valid {
			item.OwnerReward = &ownerReward.Decimal
		}
		res = append(res, item)
	}
	return res, nil
}

//type OracleVotingContractCommittee struct {
//	TxId         int
//	ContractTxId int
//	Address      string
//}
//
//func GetOracleVotingContractCommittees(db *sql.DB) ([]OracleVotingContractCommittee, error) {
//	rows, err := db.Query(`select t.tx_id, t.ov_contract_tx_id, a.address from oracle_voting_contract_committees t join addresses a on a.id=t.address_id order by tx_id`)
//	if err != nil {
//		return nil, err
//	}
//	defer rows.Close()
//	var res []OracleVotingContractCommittee
//	for rows.Next() {
//		item := OracleVotingContractCommittee{}
//		err := rows.Scan(
//			&item.TxId,
//			&item.ContractTxId,
//			&item.Address,
//		)
//		if err != nil {
//			return nil, err
//		}
//		res = append(res, item)
//	}
//	return res, nil
//}

type TxReceipt struct {
	TxId    int
	Success bool
	GasUsed int
	GasCost string
	Method  string
	Error   string
}

func GetTxReceipts(db *sql.DB) ([]TxReceipt, error) {
	rows, err := db.Query(`select * from tx_receipts order by tx_id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []TxReceipt
	for rows.Next() {
		item := TxReceipt{}
		var method, errorMsg sql.NullString
		err := rows.Scan(
			&item.TxId,
			&item.Success,
			&item.GasUsed,
			&item.GasCost,
			&method,
			&errorMsg,
		)
		if err != nil {
			return nil, err
		}
		if method.Valid {
			item.Method = method.String
		}
		if errorMsg.Valid {
			item.Error = errorMsg.String
		}
		res = append(res, item)
	}
	return res, nil
}

type TimeLockContract struct {
	TxId      int
	Timestamp int64
}

func GetTimeLockContracts(db *sql.DB) ([]TimeLockContract, error) {
	rows, err := db.Query(`select * from time_lock_contracts order by contract_tx_id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []TimeLockContract
	for rows.Next() {
		item := TimeLockContract{}
		err := rows.Scan(
			&item.TxId,
			&item.Timestamp,
		)
		if err != nil {
			return nil, err
		}
		res = append(res, item)
	}
	return res, nil
}

type TimeLockContractCallTransfer struct {
	TxId         int
	ContractTxId int
	Dest         string
	Amount       decimal.Decimal
}

func GetTimeLockContractCallTransfers(db *sql.DB) ([]TimeLockContractCallTransfer, error) {
	rows, err := db.Query(`select t.call_tx_id, t.tl_contract_tx_id, a.address, t.amount from time_lock_contract_call_transfers t join addresses a on a.id=t.dest_address_id order by t.call_tx_id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []TimeLockContractCallTransfer
	for rows.Next() {
		item := TimeLockContractCallTransfer{}
		err := rows.Scan(
			&item.TxId,
			&item.ContractTxId,
			&item.Dest,
			&item.Amount,
		)
		if err != nil {
			return nil, err
		}
		res = append(res, item)
	}
	return res, nil
}

type TimeLockContractTermination struct {
	TxId         int
	ContractTxId int
	Dest         string
}

func GetTimeLockContractTerminations(db *sql.DB) ([]TimeLockContractTermination, error) {
	rows, err := db.Query(`select t.termination_tx_id, t.tl_contract_tx_id, a.address from time_lock_contract_terminations t join addresses a on a.id=t.dest_address_id order by t.termination_tx_id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []TimeLockContractTermination
	for rows.Next() {
		item := TimeLockContractTermination{}
		err := rows.Scan(
			&item.TxId,
			&item.ContractTxId,
			&item.Dest,
		)
		if err != nil {
			return nil, err
		}
		res = append(res, item)
	}
	return res, nil
}

type MultisigContract struct {
	TxId     int
	MinVotes int
	MaxVotes int
	State    int
}

func GetMultisigContracts(db *sql.DB) ([]MultisigContract, error) {
	rows, err := db.Query(`select * from multisig_contracts order by contract_tx_id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []MultisigContract
	for rows.Next() {
		item := MultisigContract{}
		err := rows.Scan(
			&item.TxId,
			&item.MinVotes,
			&item.MaxVotes,
			&item.State,
		)
		if err != nil {
			return nil, err
		}
		res = append(res, item)
	}
	return res, nil
}

type MultisigContractCallAdd struct {
	TxId         int
	ContractTxId int
	Address      string
	NewState     *int
}

func GetMultisigContractCallAdds(db *sql.DB) ([]MultisigContractCallAdd, error) {
	rows, err := db.Query(`select t.call_tx_id, t.ms_contract_tx_id, a.address, t.new_state from multisig_contract_call_adds t join addresses a on a.id=t.address_id order by t.call_tx_id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []MultisigContractCallAdd
	for rows.Next() {
		item := MultisigContractCallAdd{}
		var state sql.NullInt32
		err := rows.Scan(
			&item.TxId,
			&item.ContractTxId,
			&item.Address,
			&state,
		)
		if err != nil {
			return nil, err
		}
		if state.Valid {
			v := int(state.Int32)
			item.NewState = &v
		}
		res = append(res, item)
	}
	return res, nil
}

type MultisigContractCallSend struct {
	TxId         int
	ContractTxId int
	Address      string
	Amount       decimal.Decimal
}

func GetMultisigContractCallSends(db *sql.DB) ([]MultisigContractCallSend, error) {
	rows, err := db.Query(`select t.call_tx_id, t.ms_contract_tx_id, a.address, t.amount from multisig_contract_call_sends t join addresses a on a.id=t.dest_address_id order by t.call_tx_id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []MultisigContractCallSend
	for rows.Next() {
		item := MultisigContractCallSend{}
		err := rows.Scan(
			&item.TxId,
			&item.ContractTxId,
			&item.Address,
			&item.Amount,
		)
		if err != nil {
			return nil, err
		}
		res = append(res, item)
	}
	return res, nil
}

type MultisigContractCallPush struct {
	TxId           int
	ContractTxId   int
	Address        string
	Amount         decimal.Decimal
	VoteAddressCnt int
	VoteAmountCnt  int
}

func GetMultisigContractCallPushes(db *sql.DB) ([]MultisigContractCallPush, error) {
	rows, err := db.Query(`select t.call_tx_id, t.ms_contract_tx_id, a.address, t.amount, t.vote_address_cnt, t.vote_amount_cnt from multisig_contract_call_pushes t join addresses a on a.id=t.dest_address_id order by t.call_tx_id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []MultisigContractCallPush
	for rows.Next() {
		item := MultisigContractCallPush{}
		err := rows.Scan(
			&item.TxId,
			&item.ContractTxId,
			&item.Address,
			&item.Amount,
			&item.VoteAddressCnt,
			&item.VoteAmountCnt,
		)
		if err != nil {
			return nil, err
		}
		res = append(res, item)
	}
	return res, nil
}

type MultisigContractTermination struct {
	TxId         int
	ContractTxId int
	Dest         string
}

func GetMultisigContractTerminations(db *sql.DB) ([]MultisigContractTermination, error) {
	rows, err := db.Query(`select t.termination_tx_id, t.ms_contract_tx_id, a.address from multisig_contract_terminations t join addresses a on a.id=t.dest_address_id order by t.termination_tx_id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []MultisigContractTermination
	for rows.Next() {
		item := MultisigContractTermination{}
		err := rows.Scan(
			&item.TxId,
			&item.ContractTxId,
			&item.Dest,
		)
		if err != nil {
			return nil, err
		}
		res = append(res, item)
	}
	return res, nil
}

type OracleLockContract struct {
	TxId                int
	OracleVotingAddress string
	Value               int
	SuccessAddress      string
	FailAddress         string
}

func GetOracleLockContracts(db *sql.DB) ([]OracleLockContract, error) {
	rows, err := db.Query(`select t.contract_tx_id, a1.address, t.value, a2.address, a3.address from oracle_lock_contracts t
    join addresses a1 on a1.id=t.oracle_voting_address_id
    join addresses a2 on a2.id=t.success_address_id
    join addresses a3 on a3.id=t.fail_address_id
order by contract_tx_id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []OracleLockContract
	for rows.Next() {
		item := OracleLockContract{}
		err := rows.Scan(
			&item.TxId,
			&item.OracleVotingAddress,
			&item.Value,
			&item.SuccessAddress,
			&item.FailAddress,
		)
		if err != nil {
			return nil, err
		}
		res = append(res, item)
	}
	return res, nil
}

type OracleLockContractCallCheckOracleVoting struct {
	TxId               int
	ContractTxId       int
	OracleVotingResult *int
}

func GetOracleLockContractCallCheckOracleVotings(db *sql.DB) ([]OracleLockContractCallCheckOracleVoting, error) {
	rows, err := db.Query(`select * from oracle_lock_contract_call_check_oracle_votings order by call_tx_id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []OracleLockContractCallCheckOracleVoting
	for rows.Next() {
		item := OracleLockContractCallCheckOracleVoting{}
		var ovResult sql.NullInt32
		err := rows.Scan(
			&item.TxId,
			&item.ContractTxId,
			&ovResult,
		)
		if err != nil {
			return nil, err
		}
		if ovResult.Valid {
			v := int(ovResult.Int32)
			item.OracleVotingResult = &v
		}
		res = append(res, item)
	}
	return res, nil
}

type OracleLockContractCallPush struct {
	TxId               int
	ContractTxId       int
	Success            bool
	OracleVotingResult int
	Transfer           decimal.Decimal
}

func GetOracleLockContractCallPushes(db *sql.DB) ([]OracleLockContractCallPush, error) {
	rows, err := db.Query(`select * from oracle_lock_contract_call_pushes order by call_tx_id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []OracleLockContractCallPush
	for rows.Next() {
		item := OracleLockContractCallPush{}
		//var ovResult sql.NullInt32
		err := rows.Scan(
			&item.TxId,
			&item.ContractTxId,
			&item.Success,
			&item.OracleVotingResult,
			&item.Transfer,
		)
		if err != nil {
			return nil, err
		}
		//if ovResult.Valid {
		//	v := int(ovResult.Int32)
		//	item.OracleVotingResult = &v
		//}
		res = append(res, item)
	}
	return res, nil
}

type OracleLockContractTermination struct {
	TxId         int
	ContractTxId int
	Dest         string
}

func GetOracleLockContractTerminations(db *sql.DB) ([]OracleLockContractTermination, error) {
	rows, err := db.Query(`select t.termination_tx_id, t.ol_contract_tx_id, a.address from oracle_lock_contract_terminations t join addresses a on a.id=t.dest_address_id order by t.termination_tx_id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []OracleLockContractTermination
	for rows.Next() {
		item := OracleLockContractTermination{}
		err := rows.Scan(
			&item.TxId,
			&item.ContractTxId,
			&item.Dest,
		)
		if err != nil {
			return nil, err
		}
		res = append(res, item)
	}
	return res, nil
}

type RefundableOracleLockContract struct {
	TxId                int
	OracleVotingAddress string
	Value               int
	SuccessAddress      *string
	FailAddress         *string
	RefundDelay         int
	DepositDeadline     int
	OracleVotingFee     int
}

func GetRefundableOracleLockContracts(db *sql.DB) ([]RefundableOracleLockContract, error) {
	rows, err := db.Query(`select t.contract_tx_id, a1.address, t.value, a2.address, a3.address, t.refund_delay, t.deposit_deadline, t.oracle_voting_fee from refundable_oracle_lock_contracts t
    join addresses a1 on a1.id=t.oracle_voting_address_id
    left join addresses a2 on a2.id=t.success_address_id
    left join addresses a3 on a3.id=t.fail_address_id
order by contract_tx_id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []RefundableOracleLockContract
	for rows.Next() {
		item := RefundableOracleLockContract{}
		err := rows.Scan(
			&item.TxId,
			&item.OracleVotingAddress,
			&item.Value,
			&item.SuccessAddress,
			&item.FailAddress,
			&item.RefundDelay,
			&item.DepositDeadline,
			&item.OracleVotingFee,
		)
		if err != nil {
			return nil, err
		}
		res = append(res, item)
	}
	return res, nil
}

type RefundableOracleLockContractTermination struct {
	TxId         int
	ContractTxId int
	Dest         string
}

func GetRefundableOracleLockContractTerminations(db *sql.DB) ([]RefundableOracleLockContractTermination, error) {
	rows, err := db.Query(`select t.termination_tx_id, t.ol_contract_tx_id, a.address from refundable_oracle_lock_contract_terminations t join addresses a on a.id=t.dest_address_id order by t.termination_tx_id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []RefundableOracleLockContractTermination
	for rows.Next() {
		item := RefundableOracleLockContractTermination{}
		err := rows.Scan(
			&item.TxId,
			&item.ContractTxId,
			&item.Dest,
		)
		if err != nil {
			return nil, err
		}
		res = append(res, item)
	}
	return res, nil
}

type RefundableOracleLockContractCallDeposit struct {
	TxId         int
	ContractTxId int
	OwnSum       decimal.Decimal
	Sum          decimal.Decimal
	Fee          decimal.Decimal
}

func GetRefundableOracleLockContractCallDeposits(db *sql.DB) ([]RefundableOracleLockContractCallDeposit, error) {
	rows, err := db.Query(`select * from refundable_oracle_lock_contract_call_deposits order by call_tx_id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []RefundableOracleLockContractCallDeposit
	for rows.Next() {
		item := RefundableOracleLockContractCallDeposit{}
		err := rows.Scan(
			&item.TxId,
			&item.ContractTxId,
			&item.OwnSum,
			&item.Sum,
			&item.Fee,
		)
		if err != nil {
			return nil, err
		}
		res = append(res, item)
	}
	return res, nil
}

type RefundableOracleLockContractCallPush struct {
	TxId               int
	ContractTxId       int
	OracleVotingExists bool
	OracleVotingResult *int
	Transfer           decimal.Decimal
	RefundBlock        *int
}

func GetRefundableOracleLockContractCallPushes(db *sql.DB) ([]RefundableOracleLockContractCallPush, error) {
	rows, err := db.Query(`select * from refundable_oracle_lock_contract_call_pushes order by call_tx_id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []RefundableOracleLockContractCallPush
	for rows.Next() {
		item := RefundableOracleLockContractCallPush{}
		var ovResult sql.NullInt32
		var refundBlock sql.NullInt32
		err := rows.Scan(
			&item.TxId,
			&item.ContractTxId,
			&item.OracleVotingExists,
			&ovResult,
			&item.Transfer,
			&refundBlock,
		)
		if err != nil {
			return nil, err
		}
		if ovResult.Valid {
			v := int(ovResult.Int32)
			item.OracleVotingResult = &v
		}
		if refundBlock.Valid {
			v := int(refundBlock.Int32)
			item.RefundBlock = &v
		}
		res = append(res, item)
	}
	return res, nil
}

type RefundableOracleLockContractCallRefund struct {
	TxId         int
	ContractTxId int
	Balance      decimal.Decimal
	Coef         float64
}

func GetRefundableOracleLockContractCallRefunds(db *sql.DB) ([]RefundableOracleLockContractCallRefund, error) {
	rows, err := db.Query(`select * from refundable_oracle_lock_contract_call_refunds order by call_tx_id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []RefundableOracleLockContractCallRefund
	for rows.Next() {
		item := RefundableOracleLockContractCallRefund{}
		err := rows.Scan(
			&item.TxId,
			&item.ContractTxId,
			&item.Balance,
			&item.Coef,
		)
		if err != nil {
			return nil, err
		}
		res = append(res, item)
	}
	return res, nil
}

type Address struct {
	Id      int
	Address string
}

func GetAddresses(db *sql.DB) ([]Address, error) {
	rows, err := db.Query(`select id, address from addresses`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []Address
	for rows.Next() {
		item := Address{}
		err := rows.Scan(
			&item.Id,
			&item.Address,
		)
		if err != nil {
			return nil, err
		}
		res = append(res, item)
	}
	return res, nil
}

//type Change struct {
//	Id          int
//	BlockHeight int
//}

//func GetChanges(db *sql.DB) ([]Change, error) {
//	rows, err := db.Query(`select id, block_height from changes order by id`)
//	if err != nil {
//		return nil, err
//	}
//	defer rows.Close()
//	var res []Change
//	for rows.Next() {
//		item := Change{}
//		err := rows.Scan(
//			&item.Id,
//			&item.BlockHeight,
//		)
//		if err != nil {
//			return nil, err
//		}
//		res = append(res, item)
//	}
//	return res, nil
//}

type OracleVotingContractResultChange struct {
	ChangeId int
}

func GetOracleVotingContractResultChanges(db *sql.DB) ([]OracleVotingContractResultChange, error) {
	rows, err := db.Query(`select change_id from oracle_voting_contract_results_changes order by change_id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []OracleVotingContractResultChange
	for rows.Next() {
		item := OracleVotingContractResultChange{}
		err := rows.Scan(
			&item.ChangeId,
		)
		if err != nil {
			return nil, err
		}
		res = append(res, item)
	}
	return res, nil
}

type OracleVotingContractSummaryChange struct {
	ChangeId int
}

func GetOracleVotingContractSummaryChanges(db *sql.DB) ([]OracleVotingContractSummaryChange, error) {
	rows, err := db.Query(`select change_id from oracle_voting_contract_summaries_changes order by change_id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []OracleVotingContractSummaryChange
	for rows.Next() {
		item := OracleVotingContractSummaryChange{}
		err := rows.Scan(
			&item.ChangeId,
		)
		if err != nil {
			return nil, err
		}
		res = append(res, item)
	}
	return res, nil
}

type SortedOracleVotingContractChange struct {
	ChangeId int
}

func GetSortedOracleVotingContractChanges(db *sql.DB) ([]SortedOracleVotingContractChange, error) {
	rows, err := db.Query(`select change_id from sorted_oracle_voting_contracts_changes order by change_id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []SortedOracleVotingContractChange
	for rows.Next() {
		item := SortedOracleVotingContractChange{}
		err := rows.Scan(
			&item.ChangeId,
		)
		if err != nil {
			return nil, err
		}
		res = append(res, item)
	}
	return res, nil
}

type SortedOracleVotingContractCommitteeChange struct {
	ChangeId int
	Deleted  *bool
}

func GetSortedOracleVotingContractCommitteeChanges(db *sql.DB) ([]SortedOracleVotingContractCommitteeChange, error) {
	rows, err := db.Query(`select change_id, deleted from sorted_oracle_voting_contract_committees_changes order by change_id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []SortedOracleVotingContractCommitteeChange
	for rows.Next() {
		item := SortedOracleVotingContractCommitteeChange{}
		var deleted sql.NullBool
		err := rows.Scan(
			&item.ChangeId,
			&deleted,
		)
		if err != nil {
			return nil, err
		}
		if deleted.Valid {
			v := deleted.Bool
			item.Deleted = &v
		}
		res = append(res, item)
	}
	return res, nil
}

type ContractTxBalanceUpdate struct {
	Id           int
	ContractTxId int
	Address      string
	ContractType int
	TxId         int
	CallMethod   *int
	BalanceOld   *decimal.Decimal
	BalanceNew   *decimal.Decimal
}

func GetContractTxBalanceUpdates(db *sql.DB) ([]ContractTxBalanceUpdate, error) {
	rows, err := db.Query(`select t.id, t.contract_tx_id, a.address, t.contract_type, t.tx_id, t.call_method, t.balance_old, t.balance_new from contract_tx_balance_updates t join addresses a on a.id=t.address_id order by t.tx_id, t.address_id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []ContractTxBalanceUpdate
	for rows.Next() {
		item := ContractTxBalanceUpdate{}
		var balanceOld, balanceNew NullDecimal
		var callMethod sql.NullInt32
		err := rows.Scan(
			&item.Id,
			&item.ContractTxId,
			&item.Address,
			&item.ContractType,
			&item.TxId,
			&callMethod,
			&balanceOld,
			&balanceNew,
		)
		if err != nil {
			return nil, err
		}
		if balanceOld.Valid {
			item.BalanceOld = &balanceOld.Decimal
		}
		if balanceNew.Valid {
			item.BalanceNew = &balanceNew.Decimal
		}
		if callMethod.Valid {
			v := int(callMethod.Int32)
			item.CallMethod = &v
		}
		res = append(res, item)
	}
	return res, nil
}

type RewardBounds struct {
	Epoch      uint64
	Type       byte
	MinAmount  decimal.Decimal
	MinAddress string
	MaxAmount  decimal.Decimal
	MaxAddress string
}

func GetRewardBounds(db *sql.DB) ([]RewardBounds, error) {
	rows, err := db.Query(`select t.epoch, t.bound_type, t.min_amount, mina.address, t.max_amount, maxa.address from epoch_reward_bounds t
  join addresses mina on mina.id=t.min_address_id
  join addresses maxa on maxa.id=t.max_address_id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []RewardBounds
	for rows.Next() {
		item := RewardBounds{}
		err := rows.Scan(
			&item.Epoch,
			&item.Type,
			&item.MinAmount,
			&item.MinAddress,
			&item.MaxAmount,
			&item.MaxAddress,
		)
		if err != nil {
			return nil, err
		}
		res = append(res, item)
	}
	return res, nil
}

type PoolsSummary struct {
	Count int
}

func GetPoolsSummaries(db *sql.DB) ([]PoolsSummary, error) {
	rows, err := db.Query(`select count from pools_summary`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []PoolsSummary
	for rows.Next() {
		item := PoolsSummary{}
		err := rows.Scan(
			&item.Count,
		)
		if err != nil {
			return nil, err
		}
		res = append(res, item)
	}
	return res, nil
}

type PoolSize struct {
	Address        string
	TotalDelegated int
	Size           int
}

func GetPoolSizes(db *sql.DB) ([]PoolSize, error) {
	rows, err := db.Query(`select a.address, t.total_delegated, t.size from pool_sizes t join addresses a on a.id=t.address_id order by t.address_id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []PoolSize
	for rows.Next() {
		item := PoolSize{}
		err := rows.Scan(
			&item.Address,
			&item.TotalDelegated,
			&item.Size,
		)
		if err != nil {
			return nil, err
		}
		res = append(res, item)
	}
	return res, nil
}

type Delegation struct {
	Delegator  string
	Delegatee  string
	BirthEpoch *int
}

func GetDelegations(db *sql.DB) ([]Delegation, error) {
	rows, err := db.Query(`select a1.address, a2.address, t.birth_epoch from delegations t join addresses a1 on a1.id=t.delegator_address_id join addresses a2 on a2.id=t.delegatee_address_id  order by t.delegator_address_id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []Delegation
	for rows.Next() {
		item := Delegation{}
		var birthEpoch sql.NullInt32
		err := rows.Scan(
			&item.Delegator,
			&item.Delegatee,
			&birthEpoch,
		)
		if birthEpoch.Valid {
			v := int(birthEpoch.Int32)
			item.BirthEpoch = &v
		}
		if err != nil {
			return nil, err
		}
		res = append(res, item)
	}
	return res, nil
}

type RewardedInvitation struct {
	InviteTxId  int
	BlockHeight int
	RewardType  int
	EpochHeight *int
}

func GetRewardedInvitations(db *sql.DB) ([]RewardedInvitation, error) {
	rows, err := db.Query(`select * from rewarded_invitations order by invite_tx_id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []RewardedInvitation
	for rows.Next() {
		item := RewardedInvitation{}
		var epochHeight sql.NullInt32
		err := rows.Scan(
			&item.InviteTxId,
			&item.BlockHeight,
			&item.RewardType,
			&epochHeight,
		)
		if err != nil {
			return nil, err
		}
		if epochHeight.Valid {
			v := int(epochHeight.Int32)
			item.EpochHeight = &v
		}
		res = append(res, item)
	}
	return res, nil
}

type UpgradeVotingHistoryItem struct {
	BlockHeight int
	Upgrade     int
	Votes       int
}

func GetUpgradeVotingHistory(db *sql.DB) ([]UpgradeVotingHistoryItem, error) {
	rows, err := db.Query(`select block_height, upgrade, votes from upgrade_voting_history order by block_height, upgrade`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []UpgradeVotingHistoryItem
	for rows.Next() {
		item := UpgradeVotingHistoryItem{}
		err := rows.Scan(
			&item.BlockHeight,
			&item.Upgrade,
			&item.Votes,
		)
		if err != nil {
			return nil, err
		}
		res = append(res, item)
	}
	return res, nil
}

func GetUpgradeVotingShortHistory(db *sql.DB) ([]UpgradeVotingHistoryItem, error) {
	rows, err := db.Query(`select block_height, upgrade, votes from upgrade_voting_short_history order by block_height, upgrade`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []UpgradeVotingHistoryItem
	for rows.Next() {
		item := UpgradeVotingHistoryItem{}
		err := rows.Scan(
			&item.BlockHeight,
			&item.Upgrade,
			&item.Votes,
		)
		if err != nil {
			return nil, err
		}
		res = append(res, item)
	}
	return res, nil
}

type UpgradeVotingHistorySummary struct {
	Upgrade int
	Items   int
}

func GetUpgradeVotingHistorySummaries(db *sql.DB) ([]UpgradeVotingHistorySummary, error) {
	rows, err := db.Query(`select upgrade, items from upgrade_voting_history_summary order by upgrade`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []UpgradeVotingHistorySummary
	for rows.Next() {
		item := UpgradeVotingHistorySummary{}
		err := rows.Scan(
			&item.Upgrade,
			&item.Items,
		)
		if err != nil {
			return nil, err
		}
		res = append(res, item)
	}
	return res, nil
}
