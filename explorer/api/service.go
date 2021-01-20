package api

import (
	"fmt"
	"github.com/idena-network/idena-go/common/hexutil"
	"github.com/idena-network/idena-indexer/explorer/db"
	"github.com/idena-network/idena-indexer/explorer/db/postgres"
	"github.com/idena-network/idena-indexer/explorer/types"
)

type Service interface {
	db.Accessor
	MemPoolTxs(count uint64) ([]*types.TransactionSummary, error)
}

type MemPool interface {
	GetTransaction(hash string) (*types.TransactionDetail, error)
	GetTransactionRaw(hash string) (hexutil.Bytes, error)
	GetAddressTransactions(address string, count int) ([]*types.TransactionSummary, error)
	GetTransactions(count int) ([]*types.TransactionSummary, error)
}

func NewService(dbAccessor db.Accessor, memPool MemPool) Service {
	return &service{
		Accessor: dbAccessor,
		memPool:  memPool,
	}
}

type service struct {
	db.Accessor
	memPool MemPool
}

func (s *service) Search(value string) ([]types.Entity, error) {
	res, err := s.Accessor.Search(value)
	if err != nil {
		return nil, err
	}
	tx, _ := s.memPool.GetTransaction(value)
	if tx != nil {
		res = append(res, types.Entity{
			Name:     "Transaction",
			Value:    value,
			Ref:      fmt.Sprintf("/api/Transaction/%s", value),
			NameOld:  "Transaction",
			ValueOld: value,
			RefOld:   fmt.Sprintf("/api/Transaction/%s", value),
		})
	}
	return res, err
}

func (s *service) Transaction(hash string) (types.TransactionDetail, error) {
	res, err := s.Accessor.Transaction(hash)
	if err != postgres.NoDataFound {
		return res, err
	}
	var tx *types.TransactionDetail
	tx, err = s.memPool.GetTransaction(hash)
	if tx != nil {
		res = *tx
	}
	return res, err
}

func (s *service) TransactionRaw(hash string) (hexutil.Bytes, error) {
	res, err := s.Accessor.TransactionRaw(hash)
	if err != postgres.NoDataFound {
		return res, err
	}
	res, err = s.memPool.GetTransactionRaw(hash)
	return res, err
}

func (s *service) IdentityTxs(address string, count uint64, continuationToken *string) ([]types.TransactionSummary, *string, error) {
	var res []types.TransactionSummary
	var nextContinuationToken *string
	var err error

	if continuationToken == nil {
		// Mem pool txs
		txs, _ := s.memPool.GetAddressTransactions(address, int(count))
		if len(txs) > 0 {
			res = make([]types.TransactionSummary, 0, len(txs))
			for _, tx := range txs {
				res = append(res, *tx)
				if len(res) == int(count) {
					break
				}
			}
			count = count - uint64(len(res))
		}
	}

	if count > 0 {
		// DB txs
		var txs []types.TransactionSummary
		txs, nextContinuationToken, err = s.Accessor.IdentityTxs(address, count, continuationToken)
		res = append(res, txs...)
	}
	return res, nextContinuationToken, err
}

func (s *service) MemPoolTxs(count uint64) ([]*types.TransactionSummary, error) {
	return s.memPool.GetTransactions(int(count))
}
