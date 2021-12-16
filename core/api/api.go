package api

import (
	"github.com/idena-network/idena-go/common"
	"github.com/idena-network/idena-go/common/hexutil"
	"github.com/idena-network/idena-go/crypto"
	"github.com/idena-network/idena-indexer/core/holder/online"
	"github.com/idena-network/idena-indexer/core/holder/state"
	"github.com/idena-network/idena-indexer/core/holder/transaction"
	"github.com/idena-network/idena-indexer/core/holder/upgrade"
	"github.com/idena-network/idena-indexer/core/mempool"
	"github.com/idena-network/idena-indexer/core/types"
	"github.com/idena-network/idena-indexer/db"
	"github.com/pkg/errors"
	"strconv"
)

type Api struct {
	onlineIdentities online.CurrentOnlineIdentitiesHolder
	upgradesVoting   upgrade.UpgradesVotingHolder
	memPool          transaction.MemPool
	contractsMemPool mempool.Contracts
	stateHolder      state.Holder
}

func NewApi(
	onlineIdentities online.CurrentOnlineIdentitiesHolder,
	upgradesVoting upgrade.UpgradesVotingHolder,
	memPool transaction.MemPool,
	contractsMemPool mempool.Contracts,
	stateHolder state.Holder,
) *Api {
	return &Api{
		onlineIdentities: onlineIdentities,
		upgradesVoting:   upgradesVoting,
		memPool:          memPool,
		contractsMemPool: contractsMemPool,
		stateHolder:      stateHolder,
	}
}

func (a *Api) GetOnlineIdentitiesCount() uint64 {
	return uint64(len(a.onlineIdentities.GetAll()))
}

func (a *Api) GetOnlineIdentities(count uint64, continuationToken *string) ([]*types.OnlineIdentity, *string, error) {
	var startIndex uint64
	if continuationToken != nil {
		var err error
		if startIndex, err = strconv.ParseUint(*continuationToken, 10, 64); err != nil {
			return nil, nil, errors.New("invalid continuation token")
		}
	}
	var res []*types.OnlineIdentity
	all := a.onlineIdentities.GetAll()
	var nextContinuationToken *string
	if len(all) > 0 {
		for i := startIndex; i >= 0 && i < startIndex+count && i < uint64(len(all)); i++ {
			res = append(res, convertOnlineIdentity(all[i]))
		}
		if uint64(len(all)) > startIndex+count {
			t := strconv.FormatUint(startIndex+count, 10)
			nextContinuationToken = &t
		}
	}
	return res, nextContinuationToken, nil
}

// Deprecated
func (a *Api) GetOnlineIdentitiesOld(startIndex, count uint64) []*types.OnlineIdentity {
	var res []*types.OnlineIdentity
	all := a.onlineIdentities.GetAll()
	if len(all) > 0 {
		for i := startIndex; i >= 0 && i < startIndex+count && i < uint64(len(all)); i++ {
			res = append(res, convertOnlineIdentity(all[i]))
		}
	}
	return res
}

func (a *Api) GetOnlineIdentity(address string) *types.OnlineIdentity {
	oi := a.onlineIdentities.Get(address)
	if oi == nil {
		return nil
	}
	return convertOnlineIdentity(oi)
}

func (a *Api) GetOnlineCount() uint64 {
	return uint64(a.onlineIdentities.GetOnlineCount())
}

func convertOnlineIdentity(oi *online.Identity) *types.OnlineIdentity {
	if oi == nil {
		return nil
	}
	res := &types.OnlineIdentity{
		Address:      oi.Address,
		LastActivity: oi.LastActivity,
		Penalty:      oi.Penalty,
		Online:       oi.Online,
	}
	if oi.Delegatee != nil {
		res.Delegetee = &types.OnlineIdentity{
			Address:      oi.Delegatee.Address,
			LastActivity: oi.Delegatee.LastActivity,
			Penalty:      oi.Delegatee.Penalty,
			Online:       oi.Delegatee.Online,
		}
	}
	return res
}

func (a *Api) SignatureAddress(value, signature string) (string, error) {
	hash := crypto.Hash([]byte(value))
	hash = crypto.Hash(hash[:])
	signatureBytes, err := hexutil.Decode(signature)
	if err != nil {
		return "", err
	}
	pubKey, err := crypto.Ecrecover(hash[:], signatureBytes)
	if err != nil {
		return "", err
	}
	addr, err := crypto.PubKeyBytesToAddress(pubKey)
	if err != nil {
		return "", err
	}
	return addr.Hex(), nil
}

func (a *Api) UpgradeVoting() []*types.UpgradeVotes {
	votes := a.upgradesVoting.Get()
	var res []*types.UpgradeVotes
	if len(votes) > 0 {
		res = make([]*types.UpgradeVotes, len(votes))
		for i, v := range votes {
			res[i] = &types.UpgradeVotes{
				Upgrade: v.Upgrade,
				Votes:   v.Votes,
			}
		}
	}
	return res
}

func (a *Api) MemPoolTransaction(hash string) (*types.TransactionDetail, error) {
	return a.memPool.GetTransaction(hash)
}

func (a *Api) MemPoolTransactionRaw(hash string) (hexutil.Bytes, error) {
	return a.memPool.GetTransactionRaw(hash)
}

func (a *Api) MemPoolAddressTransactions(address string, count uint64) ([]*types.TransactionSummary, error) {
	return a.memPool.GetAddressTransactions(address, int(count))
}

func (a *Api) MemPoolTransactions(count uint64) ([]*types.TransactionSummary, error) {
	return a.memPool.GetTransactions(int(count))
}

func (a *Api) MemPoolTransactionsCount() (int, error) {
	return a.memPool.GetTransactionsCount()
}

func (a *Api) MemPoolOracleVotingContractDeploys(author string) ([]db.OracleVotingContract, error) {
	return a.contractsMemPool.GetOracleVotingContractDeploys(common.HexToAddress(author))
}

func (a *Api) MemPoolAddressContractTxs(address, contractAddress string) ([]db.Transaction, error) {
	return a.contractsMemPool.GetAddressContractTxs(address, contractAddress)
}

func (a *Api) ValidatorsCount() uint64 {
	return uint64(a.onlineIdentities.ValidatorsCount())
}

func (a *Api) Validators(count uint64, continuationToken *string) ([]*types.Validator, *string, error) {
	var startIndex uint64
	if continuationToken != nil {
		var err error
		if startIndex, err = strconv.ParseUint(*continuationToken, 10, 64); err != nil {
			return nil, nil, errors.New("invalid continuation token")
		}
	}
	var res []*types.Validator
	all := a.onlineIdentities.Validators()
	var nextContinuationToken *string
	if len(all) > 0 {
		for i := startIndex; i >= 0 && i < startIndex+count && i < uint64(len(all)); i++ {
			res = append(res, all[i])
		}
		if uint64(len(all)) > startIndex+count {
			t := strconv.FormatUint(startIndex+count, 10)
			nextContinuationToken = &t
		}
	}
	return res, nextContinuationToken, nil
}

func (a *Api) OnlineValidatorsCount() uint64 {
	return uint64(a.onlineIdentities.OnlineValidatorsCount())
}

func (a *Api) OnlineValidators(count uint64, continuationToken *string) ([]*types.Validator, *string, error) {
	var startIndex uint64
	if continuationToken != nil {
		var err error
		if startIndex, err = strconv.ParseUint(*continuationToken, 10, 64); err != nil {
			return nil, nil, errors.New("invalid continuation token")
		}
	}
	var res []*types.Validator
	all := a.onlineIdentities.OnlineValidators()
	var nextContinuationToken *string
	if len(all) > 0 {
		for i := startIndex; i >= 0 && i < startIndex+count && i < uint64(len(all)); i++ {
			res = append(res, all[i])
		}
		if uint64(len(all)) > startIndex+count {
			t := strconv.FormatUint(startIndex+count, 10)
			nextContinuationToken = &t
		}
	}
	return res, nextContinuationToken, nil
}

func (a *Api) IdentityWithProof(epoch uint64, address string) (*hexutil.Bytes, error) {
	return a.stateHolder.IdentityWithProof(epoch, common.HexToAddress(address))
}
