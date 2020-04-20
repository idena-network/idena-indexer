package tests

import (
	"database/sql"
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
