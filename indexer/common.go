package indexer

import "github.com/idena-network/idena-go/common"

func ConvertAddress(address common.Address) string {
	return address.Hex()
}
