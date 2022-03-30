package state

import (
	"github.com/idena-network/idena-go/blockchain"
	"github.com/idena-network/idena-go/core/appstate"
)

type AppStateHolder interface {
	GetAppState() (*appstate.AppState, error)
}

func NewAppStateHolder(appState *appstate.AppState, chain *blockchain.Blockchain) AppStateHolder {
	return &appStateHolderImpl{
		appState: appState,
		chain:    chain,
	}
}

type appStateHolderImpl struct {
	appState *appstate.AppState
	chain    *blockchain.Blockchain
}

func (a *appStateHolderImpl) GetAppState() (*appstate.AppState, error) {
	return a.appState.Readonly(a.chain.Head.Height())
}
