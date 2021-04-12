package indexer

import (
	"github.com/idena-network/idena-go/blockchain/types"
	"github.com/idena-network/idena-go/common"
	"github.com/idena-network/idena-go/common/eventbus"
	"github.com/idena-network/idena-go/core/appstate"
	"github.com/idena-network/idena-go/crypto"
	"github.com/idena-network/idena-go/tests"
	"github.com/idena-network/idena-indexer/db"
	"github.com/stretchr/testify/require"
	db2 "github.com/tendermint/tm-db"
	"testing"
)

func Test_detectDelegationSwitches(t *testing.T) {

	prevState, newState := initAppState(), initAppState()

	// Case 1
	{
		block := &types.Block{
			Header: &types.Header{
				ProposedHeader: &types.ProposedHeader{
					Flags: types.IdentityUpdate,
				},
			},
		}
		// When
		switches := detectDelegationSwitches(block, prevState, newState, nil, nil)
		// Then
		require.Nil(t, switches)
	}

	// Case 2
	{
		addr := tests.GetRandAddr()
		prevState.State.ToggleDelegationAddress(addr, common.Address{})
		block := &types.Block{
			Header: &types.Header{
				ProposedHeader: &types.ProposedHeader{
					Flags: types.IdentityUpdate,
				},
			},
		}
		// When
		switches := detectDelegationSwitches(block, prevState, newState, nil, nil)
		// Then
		require.Nil(t, switches)
	}

	// Case 3
	{
		addr1, addr2, addr3, addr4, addr5 := tests.GetRandAddr(), tests.GetRandAddr(), tests.GetRandAddr(),
			tests.GetRandAddr(), tests.GetRandAddr()

		prevState.State.ToggleDelegationAddress(addr1, common.Address{})
		prevState.State.SetDelegatee(addr1, tests.GetRandAddr())
		newState.State.SetBirthday(addr1, 2)

		prevState.State.ToggleDelegationAddress(addr2, common.Address{})
		newState.State.SetDelegatee(addr2, addr3)
		newState.State.SetBirthday(addr2, 3)

		prevState.State.SetDelegatee(addr4, tests.GetRandAddr())
		newState.State.SetBirthday(addr4, 4)

		prevState.State.SetDelegatee(addr5, tests.GetRandAddr())
		prevState.State.ToggleDelegationAddress(addr5, common.Address{})
		newState.State.SetBirthday(addr5, 5)
		newState.State.SetDelegatee(addr5, tests.GetRandAddr())

		killedAddrs := make(map[common.Address]struct{})
		killedAddrs[addr4] = struct{}{}
		killedAddrs[addr5] = struct{}{}
		killedAddrs[tests.GetRandAddr()] = struct{}{}

		block := &types.Block{
			Header: &types.Header{
				ProposedHeader: &types.ProposedHeader{
					Flags: types.IdentityUpdate,
				},
			},
		}

		// When
		switches := detectDelegationSwitches(block, prevState, newState, killedAddrs, nil)

		// Then
		require.Len(t, switches, 4)

		switchesByDelegator := make(map[common.Address]*db.DelegationSwitch, len(switches))
		for _, delegation := range switches {
			switchesByDelegator[delegation.Delegator] = delegation
		}

		delegation := switchesByDelegator[addr1]
		require.Nil(t, delegation.Delegatee)
		require.Equal(t, addr1, delegation.Delegator)
		require.Nil(t, delegation.BirthEpoch)

		delegation = switchesByDelegator[addr2]
		require.Equal(t, addr3, *delegation.Delegatee)
		require.Equal(t, addr2, delegation.Delegator)
		require.Equal(t, uint16(3), *delegation.BirthEpoch)

		delegation = switchesByDelegator[addr4]
		require.Nil(t, delegation.Delegatee)
		require.Equal(t, addr4, delegation.Delegator)
		require.Nil(t, delegation.BirthEpoch)

		delegation = switchesByDelegator[addr5]
		require.Nil(t, delegation.Delegatee)
		require.Equal(t, addr5, delegation.Delegator)
		require.Nil(t, delegation.BirthEpoch)
	}

	// Case 4
	{
		prevState, newState = initAppState(), initAppState()
		key1, _ := crypto.GenerateKey()
		addr1 := crypto.PubkeyToAddress(key1.PublicKey)

		key2, _ := crypto.GenerateKey()
		addr2 := crypto.PubkeyToAddress(key2.PublicKey)

		key5, _ := crypto.GenerateKey()
		addr5 := crypto.PubkeyToAddress(key5.PublicKey)

		addr3, addr4 := tests.GetRandAddr(), tests.GetRandAddr()

		var delegationSwitchTxs []*types.Transaction

		prevState.State.ToggleDelegationAddress(addr1, common.Address{})
		tx1, _ := types.SignTx(&types.Transaction{
			Type: types.DelegateTx,
		}, key1)
		delegationSwitchTxs = append(delegationSwitchTxs, tx1)
		prevState.State.SetDelegatee(addr1, tests.GetRandAddr())
		newState.State.SetBirthday(addr1, 2)

		tx2, _ := types.SignTx(&types.Transaction{
			Type: types.DelegateTx,
		}, key2)
		delegationSwitchTxs = append(delegationSwitchTxs, tx2)
		newState.State.SetDelegatee(addr2, addr3)
		newState.State.SetBirthday(addr2, 3)

		prevState.State.SetDelegatee(addr4, tests.GetRandAddr())
		newState.State.SetBirthday(addr4, 4)

		prevState.State.SetDelegatee(addr5, tests.GetRandAddr())
		tx5, _ := types.SignTx(&types.Transaction{
			Type: types.DelegateTx,
		}, key5)
		delegationSwitchTxs = append(delegationSwitchTxs, tx5)
		newState.State.SetBirthday(addr5, 5)
		newState.State.SetDelegatee(addr5, tests.GetRandAddr())

		killedAddrs := make(map[common.Address]struct{})
		killedAddrs[addr4] = struct{}{}
		killedAddrs[addr5] = struct{}{}
		killedAddrs[tests.GetRandAddr()] = struct{}{}

		block := &types.Block{
			Header: &types.Header{
				ProposedHeader: &types.ProposedHeader{
					Flags: types.IdentityUpdate,
				},
			},
		}

		// When
		switches := detectDelegationSwitches(block, prevState, newState, killedAddrs, delegationSwitchTxs)

		// Then
		require.Len(t, switches, 4)

		switchesByDelegator := make(map[common.Address]*db.DelegationSwitch, len(switches))
		for _, delegation := range switches {
			switchesByDelegator[delegation.Delegator] = delegation
		}

		delegation := switchesByDelegator[addr1]
		require.Nil(t, delegation.Delegatee)
		require.Equal(t, addr1, delegation.Delegator)
		require.Nil(t, delegation.BirthEpoch)

		delegation = switchesByDelegator[addr2]
		require.Equal(t, addr3, *delegation.Delegatee)
		require.Equal(t, addr2, delegation.Delegator)
		require.Equal(t, uint16(3), *delegation.BirthEpoch)

		delegation = switchesByDelegator[addr4]
		require.Nil(t, delegation.Delegatee)
		require.Equal(t, addr4, delegation.Delegator)
		require.Nil(t, delegation.BirthEpoch)

		delegation = switchesByDelegator[addr5]
		require.Nil(t, delegation.Delegatee)
		require.Equal(t, addr5, delegation.Delegator)
		require.Nil(t, delegation.BirthEpoch)
	}

}

func initAppState() *appstate.AppState {
	memDb := db2.NewMemDB()
	appState, _ := appstate.NewAppState(memDb, eventbus.New())
	return appState
}
