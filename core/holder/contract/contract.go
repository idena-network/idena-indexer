package contract

import (
	"github.com/idena-network/idena-go/blockchain"
	"github.com/idena-network/idena-go/common"
	"github.com/idena-network/idena-go/common/hexutil"
	state2 "github.com/idena-network/idena-go/core/state"
	"github.com/idena-network/idena-go/vm/helpers"
	"github.com/idena-network/idena-indexer/core/holder/state"
	"github.com/idena-network/idena-indexer/core/types"
	"github.com/shopspring/decimal"
	"math/big"
)

type Holder interface {
	GetMultisigState(contractAddress string) (types.Multisig, error)
}

func NewHolder(appStateHolder state.AppStateHolder) Holder {
	return &holderImpl{
		appStateHolder: appStateHolder,
	}
}

type holderImpl struct {
	appStateHolder state.AppStateHolder
}

func (h *holderImpl) GetMultisigState(contractAddress string) (types.Multisig, error) {
	appState, err := h.appStateHolder.GetAppState()
	if err != nil {
		return types.Multisig{}, err
	}
	address := common.HexToAddress(contractAddress)
	addresses, err := readMap(appState.State, address, "addr", "hex", "hex")
	if err != nil {
		return types.Multisig{}, err
	}
	amounts, err := readMap(appState.State, address, "amount", "hex", "dna")
	if err != nil {
		return types.Multisig{}, err
	}
	var res types.Multisig
	for addr, dest := range addresses {
		signer := types.MultisigSigner{
			Address:     addr.(string),
			DescAddress: dest.(string),
		}
		if amount, ok := amounts[addr.(string)]; ok {
			signer.Amount = amount.(decimal.Decimal)
		}
		res.Signers = append(res.Signers, signer)
	}
	return res, nil
}

func readMap(state *state2.StateDB, contract common.Address, mapName, keyFormat, valueFormat string) (map[interface{}]interface{}, error) {
	minKey := []byte(mapName)
	maxKey := []byte(mapName)
	for i := len([]byte(mapName)); i < common.MaxContractStoreKeyLength; i++ {
		maxKey = append(maxKey, 0xFF)
	}
	res := make(map[interface{}]interface{})
	var err error
	prefixLen := len([]byte(mapName))
	state.IterateContractStore(contract, minKey, maxKey, func(key []byte, value []byte) bool {
		k, err := conversion(keyFormat, key[prefixLen:])
		if err != nil {
			return true
		}
		v, err := conversion(valueFormat, value)
		if err != nil {
			return true
		}
		res[k] = v
		return false
	})
	if err != nil {
		return nil, err
	}
	return res, nil
}

func conversion(convertTo string, data []byte) (interface{}, error) {
	switch convertTo {
	case "byte":
		return helpers.ExtractByte(0, data)
	case "uint64":
		return helpers.ExtractUInt64(0, data)
	case "string":
		return string(data), nil
	case "bigint":
		v := new(big.Int)
		v.SetBytes(data)
		return v.String(), nil
	case "hex":
		return hexutil.Encode(data), nil
	case "dna":
		v := new(big.Int)
		v.SetBytes(data)
		return blockchain.ConvertToFloat(v), nil
	default:
		return hexutil.Encode(data), nil
	}
}
