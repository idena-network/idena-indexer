package db

import "github.com/idena-network/idena-go/common/hexutil"

type FlipContent struct {
	LeftOrder  []uint16
	RightOrder []uint16
	Pics       []hexutil.Bytes
}
