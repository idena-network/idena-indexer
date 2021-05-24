package indexer

import (
	"github.com/idena-network/idena-go/blockchain/types"
	"github.com/idena-network/idena-go/common"
	"github.com/idena-network/idena-go/core/appstate"
	"github.com/idena-network/idena-indexer/db"
)

func detectDelegationSwitches(
	block *types.Block,
	prevState *appstate.AppState,
	newState *appstate.AppState,
	killedAddrs map[common.Address]struct{},
	switchDelegationTxs []*types.Transaction,
) []*db.DelegationSwitch {
	if block == nil || block.Header == nil || !block.Header.Flags().HasFlag(types.IdentityUpdate) {
		return nil
	}

	var res []*db.DelegationSwitch

	for killedAddr := range killedAddrs {
		prevDelegatee := prevState.State.Delegatee(killedAddr)
		if prevDelegatee != nil {
			res = append(res, &db.DelegationSwitch{
				Delegator: killedAddr,
			})
		}
	}

	delegations := prevState.State.Delegations()
	handledAddrs := make(map[common.Address]struct{})
	for _, delegation := range delegations {
		if _, ok := killedAddrs[delegation.Delegator]; ok {
			continue
		}
		if delegationSwitch := detectDelegationSwitch(delegation.Delegator, prevState, newState); delegationSwitch != nil {
			res = append(res, delegationSwitch)
		}
		handledAddrs[delegation.Delegator] = struct{}{}
	}
	for _, switchDelegationTx := range switchDelegationTxs {
		delegator, _ := types.Sender(switchDelegationTx)
		if _, ok := killedAddrs[delegator]; ok {
			continue
		}
		if _, ok := handledAddrs[delegator]; ok {
			continue
		}
		if delegationSwitch := detectDelegationSwitch(delegator, prevState, newState); delegationSwitch != nil {
			res = append(res, delegationSwitch)
		}
	}

	return res
}

func detectDelegationSwitch(delegator common.Address, prevState *appstate.AppState, newState *appstate.AppState) *db.DelegationSwitch {
	prevDelegatee := prevState.State.Delegatee(delegator)
	newDelegatee := newState.State.Delegatee(delegator)

	if prevDelegatee == nil && newDelegatee == nil {
		return nil
	}

	if prevDelegatee == nil {
		var birthEpoch *uint16
		if v := newState.State.GetIdentity(delegator).Birthday; v > 0 {
			birthEpoch = &v
		}
		return &db.DelegationSwitch{
			Delegator:  delegator,
			Delegatee:  newDelegatee,
			BirthEpoch: birthEpoch,
		}
	}

	if newDelegatee == nil {
		return &db.DelegationSwitch{
			Delegator: delegator,
		}
	}
	if prevDelegatee == newDelegatee {
		return nil
	}

	var birthEpoch *uint16
	if v := newState.State.GetIdentity(delegator).Birthday; v > 0 {
		birthEpoch = &v
	}
	return &db.DelegationSwitch{
		Delegator:  delegator,
		Delegatee:  newDelegatee,
		BirthEpoch: birthEpoch,
	}
}

func detectPoolSizeUpdates(delegationSwitches []*db.DelegationSwitch, addresses []db.Address, epochIdentities []db.EpochIdentity, prevState, newState *appstate.AppState) []db.PoolSize {
	if len(addresses) == 0 && len(delegationSwitches) == 0 && len(epochIdentities) == 0 {
		return nil
	}
	poolSizeByAddress := make(map[common.Address]uint64)

	handleDelegator := func(addr common.Address) {
		if delegatee := prevState.State.Delegatee(addr); delegatee != nil {
			if _, ok := poolSizeByAddress[*delegatee]; !ok {
				poolSizeByAddress[*delegatee] = uint64(newState.ValidatorsCache.PoolSize(*delegatee))
			}
		}
		if delegatee := newState.State.Delegatee(addr); delegatee != nil {
			if _, ok := poolSizeByAddress[*delegatee]; !ok {
				poolSizeByAddress[*delegatee] = uint64(newState.ValidatorsCache.PoolSize(*delegatee))
			}
		}
	}

	handlePool := func(addr common.Address) {
		if prevState.ValidatorsCache.IsPool(addr) || newState.ValidatorsCache.IsPool(addr) {
			if _, ok := poolSizeByAddress[addr]; !ok {
				poolSizeByAddress[addr] = uint64(newState.ValidatorsCache.PoolSize(addr))
			}
		}
	}

	for _, delegationSwitch := range delegationSwitches {
		addr := delegationSwitch.Delegator
		handleDelegator(addr)
	}

	for _, address := range addresses {
		if len(address.StateChanges) == 0 {
			continue
		}
		addr := common.HexToAddress(address.Address)
		handlePool(addr)
		handleDelegator(addr)
	}

	for _, epochIdentity := range epochIdentities {
		addr := common.HexToAddress(epochIdentity.Address)
		handlePool(addr)
		handleDelegator(addr)
	}
	if len(poolSizeByAddress) == 0 {
		return nil
	}
	res := make([]db.PoolSize, 0, len(poolSizeByAddress))
	for poolAddress, size := range poolSizeByAddress {
		res = append(res, db.PoolSize{
			Address: poolAddress,
			Size:    size,
		})
	}
	return res
}
