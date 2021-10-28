package events

import (
	"github.com/idena-network/idena-go/common/eventbus"
	"github.com/idena-network/idena-go/core/state"
)

const (
	NewEpochEventId     = eventbus.EventID("new-epoch")
	NewBlockEventId     = eventbus.EventID("new-block")
	CurrentEpochEventId = eventbus.EventID("current-epoch")
)

type NewEpochEvent struct {
	Epoch       uint16
	EpochHeight uint64
}

func (e *NewEpochEvent) EventID() eventbus.EventID {
	return NewEpochEventId
}

type NewBlockEvent struct {
	Height      uint64
	EpochPeriod state.ValidationPeriod
}

func (e *NewBlockEvent) EventID() eventbus.EventID {
	return NewBlockEventId
}

type CurrentEpochEvent struct {
	Epoch       uint16
	Height      uint64
	EpochHeight uint64
	EpochPeriod state.ValidationPeriod
}

func (e *CurrentEpochEvent) EventID() eventbus.EventID {
	return CurrentEpochEventId
}
