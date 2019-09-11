package common

import "github.com/idena-network/idena-go/common"

func ConvertAddress(address common.Address) string {
	return address.Hex()
}

func BytesToAddr(bytes []byte) common.Address {
	addr := common.Address{}
	addr.SetBytes(bytes[1:])
	return addr
}
