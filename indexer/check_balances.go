package indexer

import (
	"fmt"
	"github.com/idena-network/idena-go/blockchain"
	"github.com/idena-network/idena-go/common"
	"github.com/idena-network/idena-go/core/appstate"
	"github.com/idena-network/idena-go/core/state"
	"github.com/idena-network/idena-indexer/db"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
	"strings"
)

type balanceData struct {
	address string
	balance decimal.Decimal
}

func checkBalances(appState *appstate.AppState, dbAccessor db.Accessor) {
	cnt, err := dbAccessor.BalanceUpdateGapCnt()
	if err != nil {
		panic("failed to get balance update gap cnt")
	}
	if cnt > 0 {
		panic(fmt.Sprintf("balance update gap cnt: %v", cnt))
	}

	cnt, err = dbAccessor.BurntCoinsInconsistencyCnt()
	if err != nil {
		panic("failed to get burnt coins inconsistency cnt")
	}
	if cnt > 0 {
		panic(fmt.Sprintf("burnt coins inconsistency cnt: %v", cnt))
	}

	stateBalancesByAddr := make(map[string]balanceData)
	appState.State.IterateOverAccounts(func(addr common.Address, account state.Account) {
		addressStr := strings.ToLower(addr.Hex())
		stateBalancesByAddr[addressStr] = balanceData{
			address: addressStr,
			balance: blockchain.ConvertToFloat(account.Balance),
		}
	})

	dbBalances, err := dbAccessor.Balances()
	if err != nil {
		panic("failed to get balances from db to check")
	}

	dbLatestBalanceUpdates, err := dbAccessor.LatestBalanceUpdates()
	if err != nil {
		panic("failed to get latest balance updates from db to check")
	}

	toMap := func(list []db.Balance) map[string]balanceData {
		res := make(map[string]balanceData)
		for _, v := range list {
			address := strings.ToLower(v.Address)
			res[address] = balanceData{
				address: address,
				balance: v.Balance,
			}
		}
		return res
	}

	dbBalancesByAddr := toMap(dbBalances)
	dbLatestBalanceUpdatesByAddr := toMap(dbLatestBalanceUpdates)

	copyBalanceMap := func(src map[string]balanceData) map[string]balanceData {
		res := make(map[string]balanceData, len(src))
		for k, v := range src {
			res[k] = v
		}
		return res
	}

	check := func(balances1, balances2 map[string]balanceData) error {
		b1 := copyBalanceMap(balances1)
		b2 := copyBalanceMap(balances2)

		for address1, info1 := range b1 {
			info2, ok := b2[address1]
			delete(b1, address1)
			delete(b2, address1)
			if !ok {
				if info1.balance.IsZero() {
					continue
				}
				return errors.Errorf("address: %v, balance 1: %v, balance 2: NO", address1, info1.balance)
			}
			if info1.balance.Cmp(info2.balance) == 0 {
				continue
			}
			return errors.Errorf("address: %v, balance 1: %v, balance 2: %v", address1, info1.balance, info2.balance)
		}
		for address2, info2 := range b2 {
			if !info2.balance.IsZero() {
				return errors.Errorf("address: %v, balance 1: NO, balance 2: %v", address2, info2.balance)
			}
		}
		return nil
	}

	if err := check(stateBalancesByAddr, dbBalancesByAddr); err != nil {
		panic(errors.Wrap(err, "state and db balances check failed"))
	}
	if err := check(stateBalancesByAddr, dbLatestBalanceUpdatesByAddr); err != nil {
		panic(errors.Wrap(err, "state and db latest balance updates check failed"))
	}
}
