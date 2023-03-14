package tokens

import (
	"database/sql"
	"fmt"
	"github.com/idena-network/idena-go/common"
	"github.com/idena-network/idena-go/core/appstate"
	"github.com/idena-network/idena-indexer/core/stats"
	"github.com/idena-network/idena-indexer/db"
	"github.com/idena-network/idena-indexer/log"
	"math/big"
)

func CollectData(connStr string, holder *stats.TokenContractHolderImpl, appState *appstate.AppState) error {
	contracts, err := loadContracts(connStr)
	if err != nil {
		return err
	}
	if len(contracts) == 0 {
		return nil
	}

	log.Info("start collecting token balances")
	defer log.Info("collecting token balances completed")

	var allTokens []db.Token
	var allBalances []db.TokenBalance

	for _, c := range contracts {
		ok, err := isToken(c.address, holder, appState)
		if err != nil {
			return err
		}
		if !ok {
			continue
		}

		token, err := token(c.address, holder, appState)
		if err != nil {
			return err
		}
		allTokens = append(allTokens, token)

		balances, err := balances(c.address, appState)
		if err != nil {
			return err
		}
		if len(balances) == 0 {
			log.Warn("empty balances", "token", c.address.Hex())
		}
		allBalances = append(allBalances, balances...)
	}

	if err := save(connStr, allTokens, allBalances); err != nil {
		return err
	}
	return nil
}

type contract struct {
	address common.Address
}

func loadContracts(connStr string) ([]contract, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	const query = `SELECT a.address
FROM contracts c
         LEFT JOIN addresses a ON a.id = c.contract_address_id
WHERE c.type = 6 AND not exists(SELECT 1 FROM tokens)
ORDER BY c.tx_id`

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []contract
	for rows.Next() {
		item := contract{}
		var address string
		err := rows.Scan(&address)
		item.address = common.HexToAddress(address)
		if err != nil {
			return nil, err
		}
		res = append(res, item)
	}
	return res, nil
}

func isToken(address common.Address, holder *stats.TokenContractHolderImpl, appState *appstate.AppState) (bool, error) {
	_, err := holder.Balance(appState, address, common.EmptyAddress[:])
	if err != nil {
		log.Debug(fmt.Sprintf("%v is not a token: %v", address.Hex(), err))
	}
	return err == nil, nil
}

func token(address common.Address, holder *stats.TokenContractHolderImpl, appState *appstate.AppState) (db.Token, error) {
	var res db.Token
	res.ContractAddress = address
	info, err := holder.Info(appState, address)
	if err == nil {
		res.Symbol = info.Symbol
		res.Name = info.Name
		res.Decimals = info.Decimals
	} else {
		log.Warn("failed to extract token info", "err", err)
	}
	return res, nil
}

func balances(address common.Address, appState *appstate.AppState) ([]db.TokenBalance, error) {
	bytesToAddress := func(bytes []byte) common.Address {
		var res common.Address
		if len(bytes) < len(res) {
			var extendedBytes [20]byte
			copy(extendedBytes[:], bytes)
			bytes = extendedBytes[:]
		}
		res.SetBytes(bytes)
		return res
	}
	var balances []db.TokenBalance
	appState.State.IterateContractStore(address, append([]byte("b:"), common.MinAddr[:]...), append([]byte("b:"), common.MaxAddr[:]...), func(key []byte, value []byte) bool {
		tokenHolder := bytesToAddress(key[len([]byte("b:")):])
		if tokenHolder != common.EmptyAddress {
			balance := new(big.Int).SetBytes(value)
			balances = append(balances, db.TokenBalance{
				Address:         tokenHolder,
				ContractAddress: address,
				Balance:         balance,
			})
		}
		return false
	})
	return balances, nil
}

func save(connStr string, tokens []db.Token, balances []db.TokenBalance) error {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return err
	}
	defer db.Close()

	for _, token := range tokens {
		if _, err := db.Exec(`
INSERT INTO tokens (contract_address_id, "name", symbol, decimals)
VALUES ((SELECT id FROM addresses WHERE lower(address)=lower($1)), limited_text($2, 50), limited_text($3, 10), null_if_zero($4))
`, token.ContractAddress.Hex(), token.Name, token.Symbol, token.Decimals); err != nil {
			return err
		}
	}

	for _, balance := range balances {
		if balance.Balance.Sign() <= 0 {
			continue
		}
		if _, err := db.Exec(`
INSERT INTO token_balances (contract_address_id, address, balance)
VALUES ((SELECT id FROM addresses WHERE lower(address)=lower($1)), lower($2), $3::numeric)
`, balance.ContractAddress.Hex(), balance.Address.Hex(), balance.Balance.String()); err != nil {
			return err
		}
	}

	return nil
}
