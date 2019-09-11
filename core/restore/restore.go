package restore

import (
	"github.com/idena-network/idena-go/blockchain"
	"github.com/idena-network/idena-go/core/appstate"
	"github.com/idena-network/idena-indexer/core/common"
	"github.com/idena-network/idena-indexer/db"
	"github.com/idena-network/idena-indexer/log"
	"github.com/pkg/errors"
	"time"
)

type Restorer struct {
	db       db.Accessor
	appState *appstate.AppState
	chain    *blockchain.Blockchain
}

func NewRestorer(db db.Accessor, appState *appstate.AppState, chain *blockchain.Blockchain) *Restorer {
	return &Restorer{
		db:       db,
		appState: appState,
		chain:    chain,
	}
}

func (r *Restorer) Restore() {
	for {
		if err := r.tryToRestore(); err != nil {
			log.Error(err.Error())
			time.Sleep(time.Second * 5)
			continue
		}
		return
	}
}

func (r *Restorer) tryToRestore() error {
	data, err := r.collectData()
	if err != nil {
		return errors.Wrapf(err, "unable to collect data to restore")
	}
	err = r.db.SaveRestoredData(data)
	return errors.Wrapf(err, "unable to save restored data")
}

func (r *Restorer) collectData() (*db.RestoredData, error) {
	res := &db.RestoredData{}
	var err error
	if res.Balances, err = r.collectBalances(); err != nil {
		return nil, err
	}
	return res, nil
}

func (r *Restorer) collectBalances() ([]db.Balance, error) {
	head := r.chain.Head
	if head == nil {
		return nil, errors.New("blockchain header is nil")
	}
	height := head.Height() - 1
	appState := r.appState.Readonly(height)
	if appState == nil {
		return nil, errors.Errorf("appState for height=%d is absent", height)
	}
	var balances []db.Balance
	appState.State.IterateAccounts(func(key []byte, _ []byte) bool {
		if key == nil {
			return true
		}
		address := common.BytesToAddr(key)
		convertedAddress := common.ConvertAddress(address)
		balances = append(balances, db.Balance{
			Address: convertedAddress,
			Balance: blockchain.ConvertToFloat(appState.State.GetBalance(address)),
			Stake:   blockchain.ConvertToFloat(appState.State.GetStakeBalance(address)),
		})
		return false
	})
	return balances, nil
}
