package indexer

import (
	"github.com/idena-network/idena-go/core/appstate"
	"github.com/idena-network/idena-indexer/db"
)

func detectMinersHistoryItem(prevState *appstate.AppState, newState *appstate.AppState) *db.MinersHistoryItem {
	onlineValidators := newState.ValidatorsCache.OnlineSize()
	onlineMiners := newState.ValidatorsCache.ValidatorsSize()
	if prevState.ValidatorsCache.OnlineSize() == onlineValidators && prevState.ValidatorsCache.ValidatorsSize() == onlineMiners {
		return nil
	}
	return &db.MinersHistoryItem{
		OnlineValidators: uint64(onlineValidators),
		OnlineMiners:     uint64(onlineMiners),
	}
}
