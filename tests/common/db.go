package common

import (
	"database/sql"
	"github.com/idena-network/idena-go/common/hexutil"
	"github.com/idena-network/idena-indexer/db"
	"github.com/idena-network/idena-indexer/explorer/db/postgres"
	"github.com/shopspring/decimal"
)

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
		var committeeRewardShare postgres.NullDecimal
		var penaltyOld, penaltyNew postgres.NullDecimal
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
		var penaltyOld, penaltyNew postgres.NullDecimal
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

type PaidPenalty struct {
	PenaltyId   uint64
	Penalty     decimal.Decimal
	BlockHeight uint64
}

func GetPaidPenalties(db *sql.DB) ([]PaidPenalty, error) {
	rows, err := db.Query(`select penalty_id, penalty, block_height from paid_penalties`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []PaidPenalty
	for rows.Next() {
		item := PaidPenalty{}
		err := rows.Scan(
			&item.PenaltyId,
			&item.Penalty,
			&item.BlockHeight,
		)
		if err != nil {
			return nil, err
		}
		res = append(res, item)
	}
	return res, nil
}

type EpochIdentity struct {
	Id           int
	ShortAnswers int
	LongAnswers  int
}

func GetEpochIdentities(db *sql.DB) ([]EpochIdentity, error) {
	rows, err := db.Query(`select address_state_id, short_answers, long_answers from epoch_identities`)
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
