package restore

import (
	"github.com/idena-network/idena-go/blockchain"
	"github.com/idena-network/idena-go/common"
	"github.com/idena-network/idena-go/core/appstate"
	"github.com/idena-network/idena-go/core/state"
	"github.com/idena-network/idena-indexer/core/conversion"
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
	if res.Birthdays, err = r.collectBirthdays(); err != nil {
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
	appState, err := r.appState.Readonly(height)
	if err != nil {
		return nil, errors.Errorf("unable to get appState for height %d, err %v", height, err.Error())
	}
	var balances []db.Balance
	appState.State.IterateAccounts(func(key []byte, _ []byte) bool {
		if key == nil {
			return true
		}
		address := conversion.BytesToAddr(key)
		convertedAddress := conversion.ConvertAddress(address)
		balances = append(balances, db.Balance{
			Address: convertedAddress,
			Balance: blockchain.ConvertToFloat(appState.State.GetBalance(address)),
			Stake:   blockchain.ConvertToFloat(appState.State.GetStakeBalance(address)),
		})
		return false
	})
	return balances, nil
}

func (r *Restorer) collectBirthdays() ([]db.Birthday, error) {
	head := r.chain.Head
	if head == nil {
		return nil, errors.New("blockchain header is nil")
	}
	height := head.Height() - 1
	appState, err := r.appState.Readonly(height)
	if err != nil {
		return nil, errors.Errorf("unable to get appState for height %d, err %v", height, err.Error())
	}
	var res []db.Birthday
	appState.State.IterateOverIdentities(func(addr common.Address, identity state.Identity) {
		res = append(res, db.Birthday{
			Address:    conversion.ConvertAddress(addr),
			BirthEpoch: uint64(identity.Birthday),
		})
	})
	return res, nil
}
