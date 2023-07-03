package stats

import (
	"github.com/golang/protobuf/proto"
	"github.com/idena-network/idena-go/api"
	"github.com/idena-network/idena-go/blockchain/attachments"
	"github.com/idena-network/idena-go/blockchain/types"
	"github.com/idena-network/idena-go/common"
	"github.com/idena-network/idena-go/common/hexutil"
	config2 "github.com/idena-network/idena-go/config"
	"github.com/idena-network/idena-go/core/appstate"
	"github.com/idena-network/idena-go/node"
	"github.com/idena-network/idena-go/vm"
	"github.com/idena-network/idena-go/vm/helpers"
	"github.com/idena-network/idena-indexer/db"
	"github.com/idena-network/idena-indexer/log"
	models "github.com/idena-network/idena-wasm-binding/lib/protobuf"
	"github.com/pkg/errors"
	"math/big"
	"strings"
)

type TokenBalance = db.TokenBalance

const (
	eventNameTransfer  = "transfer"
	methodNameBalance  = "getBalance"
	methodNameName     = "name"
	methodNameSymbol   = "symbol"
	methodNameDecimals = "decimals"
	eventNameAirdrop   = "airdrop"
	methodNameBurn     = "burn"
)

type Token struct {
	Name, Symbol string
	Decimals     byte
}

type TokenContractHolder interface {
	Info(appState *appstate.AppState, contractAddress common.Address) (Token, error)
	Balance(appState *appstate.AppState, contractAddress common.Address, address []byte) (*big.Int, error)
}

type TokenContractHolderImpl struct {
	cfg     *config2.Config
	nodeCtx *node.NodeCtx
}

func (t *TokenContractHolderImpl) ProvideNodeCtx(nodeCtx *node.NodeCtx, cfg *config2.Config) {
	t.nodeCtx = nodeCtx
	t.cfg = cfg
}

func (t *TokenContractHolderImpl) Info(appState *appstate.AppState, contractAddress common.Address) (Token, error) {
	var name, symbol string
	var decimals byte

	if outputData, err := t.call(contractAddress, methodNameName, nil, appState); err != nil {
		return Token{}, err
	} else {
		name = string(outputData)
	}

	if outputData, err := t.call(contractAddress, methodNameSymbol, nil, appState); err != nil {
		return Token{}, err
	} else {
		symbol = string(outputData)
	}

	if outputData, err := t.call(contractAddress, methodNameDecimals, nil, appState); err != nil {
		return Token{}, err
	} else if v, err := helpers.ExtractByte(0, outputData); err != nil {
		return Token{}, err
	} else {
		decimals = v
	}
	return Token{
		Name:     name,
		Symbol:   symbol,
		Decimals: decimals,
	}, nil
}

func (t *TokenContractHolderImpl) Balance(appState *appstate.AppState, contractAddress common.Address, address []byte) (*big.Int, error) {
	args, _ := api.DynamicArgs{&api.DynamicArg{
		Index:  0,
		Format: "hex",
		Value:  hexutil.Encode(address),
	}}.ToSlice()
	outputData, err := t.call(contractAddress, methodNameBalance, args, appState)
	if err != nil {
		return nil, err
	}
	res := new(big.Int)
	res.SetBytes(outputData)
	return res, nil
}

func (t *TokenContractHolderImpl) call(contractAddress common.Address, method string, args [][]byte, appState *appstate.AppState) ([]byte, error) {
	attachment := attachments.CreateCallContractAttachment(method, args...)
	payload, err := attachment.ToBytes()
	if err != nil {
		return nil, errors.Wrap(err, "failed to build payload")
	}
	tx := &types.Transaction{
		Type:    types.CallContractTx,
		To:      &contractAddress,
		Payload: payload,
	}
	virtualMachine := vm.NewVmImpl(appState, t.nodeCtx.Blockchain, t.nodeCtx.Blockchain.Head, nil, t.cfg)
	txReceipt := virtualMachine.Run(tx, nil, -1, true)
	if !txReceipt.Success {
		return nil, errors.Wrapf(txReceipt.Error, "failed to call %v", methodNameBalance)
	}

	protoModel := &models.ActionResult{}
	if err := proto.Unmarshal(txReceipt.ActionResult, protoModel); err != nil {
		return nil, errors.Wrapf(txReceipt.Error, "invalid ActionResult")
	}
	return protoModel.OutputData, nil
}

type tokenBalanceUpdate struct {
	contractAddress common.Address
	address         common.Address
	addressBytes    []byte
	balance         *big.Int
}

type tokenBalanceUpdateCollector struct {
	updates                     []*tokenBalanceUpdate
	updatesByContractAndAddress map[common.Address]map[common.Address]*tokenBalanceUpdate
	holder                      TokenContractHolder
}

func newTokenBalanceUpdateCollector(holder TokenContractHolder) *tokenBalanceUpdateCollector {
	return &tokenBalanceUpdateCollector{
		holder: holder,
	}
}

func (c *tokenBalanceUpdateCollector) applyTxReceipt(txReceipt *types.TxReceipt, appState *appstate.AppState) []TokenBalance {
	add := func(update tokenBalanceUpdate) {
		var contractUpdates map[common.Address]*tokenBalanceUpdate
		var contractUpdatesOk bool

		if c.updatesByContractAndAddress == nil {
			c.updatesByContractAndAddress = make(map[common.Address]map[common.Address]*tokenBalanceUpdate)
		} else {
			contractUpdates, contractUpdatesOk = c.updatesByContractAndAddress[update.contractAddress]
		}

		if !contractUpdatesOk {
			contractUpdates = make(map[common.Address]*tokenBalanceUpdate)
			c.updatesByContractAndAddress[update.contractAddress] = contractUpdates
		}

		if v, ok := contractUpdates[update.address]; !ok {
			c.updates = append(c.updates, &update)
			contractUpdates[update.address] = &update
		} else {
			v.balance = nil
		}
	}

	if strings.ToLower(txReceipt.Method) == methodNameBurn && len(txReceipt.Events) == 0 {
		if update, ok := detectTokenBalanceUpdate(txReceipt.From.Bytes(), txReceipt.ContractAddress); ok {
			add(update)
		}
	}

	for _, event := range txReceipt.Events {
		for _, update := range detectTokenBalanceUpdates(event) {
			add(update)
		}
	}

	balanceUpdates := make([]TokenBalance, 0, len(c.updates))
	for _, update := range c.updates {
		balance := update.balance
		if balance == nil {
			var err error
			balance, err = c.holder.Balance(appState, update.contractAddress, update.addressBytes)
			if err != nil {
				log.Error("failed to get token balance, contract: %v, address: %v, err: %v", update.contractAddress.Hex(), update.address.Hex(), err)
				continue
			}
			update.balance = balance
		}
		balanceUpdates = append(balanceUpdates, TokenBalance{
			Address:         update.address,
			ContractAddress: update.contractAddress,
			Balance:         balance,
		})
	}

	return balanceUpdates
}

func detectTokenBalanceUpdates(event *types.TxEvent) []tokenBalanceUpdate {
	switch strings.ToLower(event.EventName) {
	case eventNameTransfer:
		return detectTransferTokenBalanceUpdates(event)
	case eventNameAirdrop:
		var res []tokenBalanceUpdate
		if bu, ok := detectAirdropTokenBalanceUpdate(event); ok {
			res = append(res, bu)
		}
		return res
	default:
		return nil
	}
}

func detectTransferTokenBalanceUpdates(event *types.TxEvent) []tokenBalanceUpdate {
	if len(event.Data) != 3 {
		return nil
	}
	var res []tokenBalanceUpdate
	if senderUpdate, ok := detectTokenBalanceUpdate(event.Data[0], event.Contract); ok {
		res = append(res, senderUpdate)
	}
	if receiverUpdate, ok := detectTokenBalanceUpdate(event.Data[1], event.Contract); ok {
		res = append(res, receiverUpdate)
	}
	return res
}

func detectAirdropTokenBalanceUpdate(event *types.TxEvent) (tokenBalanceUpdate, bool) {
	if len(event.Data) != 2 {
		return tokenBalanceUpdate{}, false
	}
	return detectTokenBalanceUpdate(event.Data[0], event.Contract)
}

func detectTokenBalanceUpdate(holder []byte, contract common.Address) (tokenBalanceUpdate, bool) {
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

	address := bytesToAddress(holder)
	if address == common.EmptyAddress {
		return tokenBalanceUpdate{}, false
	}
	return tokenBalanceUpdate{
		address:         address,
		addressBytes:    holder,
		contractAddress: contract,
	}, true
}

type tokenDetector struct {
	holder TokenContractHolder
}

func newTokenDetector(holder TokenContractHolder) *tokenDetector {
	return &tokenDetector{
		holder: holder,
	}
}

func (c *tokenDetector) detectTokens(contracts []common.Address, appState *appstate.AppState) []db.Token {
	var res []db.Token
	for _, contract := range contracts {
		info, err := c.holder.Info(appState, contract)
		if err != nil {
			continue
		}
		res = append(res, db.Token{
			ContractAddress: contract,
			Name:            info.Name,
			Symbol:          info.Symbol,
			Decimals:        info.Decimals,
		})
	}
	return res
}
