package api

import (
	"github.com/idena-network/idena-go/common/hexutil"
	"github.com/idena-network/idena-go/crypto"
	"github.com/idena-network/idena-go/rlp"
	"github.com/idena-network/idena-indexer/core/holder/online"
)

type Api struct {
	onlineIdentities online.CurrentOnlineIdentitiesHolder
}

func NewApi(onlineIdentities online.CurrentOnlineIdentitiesHolder) *Api {
	return &Api{
		onlineIdentities: onlineIdentities,
	}
}

func (a *Api) GetOnlineIdentitiesCount() uint64 {
	return uint64(len(a.onlineIdentities.GetAll()))
}

func (a *Api) GetOnlineIdentities(startIndex, count uint64) []*OnlineIdentity {
	var res []*OnlineIdentity
	all := a.onlineIdentities.GetAll()
	if len(all) > 0 {
		for i := startIndex; i >= 0 && i < startIndex+count && i < uint64(len(all)); i++ {
			res = append(res, convertOnlineIdentity(all[i]))
		}
	}
	return res
}

func (a *Api) GetOnlineIdentity(address string) *OnlineIdentity {
	oi := a.onlineIdentities.Get(address)
	if oi == nil {
		return nil
	}
	return convertOnlineIdentity(oi)
}

func (a *Api) GetOnlineCount() uint64 {
	return uint64(a.onlineIdentities.GetOnlineCount())
}

func convertOnlineIdentity(oi *online.Identity) *OnlineIdentity {
	if oi == nil {
		return nil
	}
	return &OnlineIdentity{
		Address:      oi.Address,
		LastActivity: oi.LastActivity,
		Penalty:      oi.Penalty,
		Online:       oi.Online,
	}
}

func (a *Api) SignatureAddress(value, signature string) (string, error) {
	hash := rlp.Hash(value)
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
