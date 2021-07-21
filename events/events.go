package events

import "github.com/idena-network/idena-go/common/eventbus"

const (
	NewEpochEventId     = eventbus.EventID("new-epoch")
	CurrentEpochEventId = eventbus.EventID("current-epoch")
)

type NewEpochEvent struct {
	Epoch uint16
}

func (e *NewEpochEvent) EventID() eventbus.EventID {
	return NewEpochEventId
}

type CurrentEpochEvent struct {
	Epoch uint16
}

func (e *CurrentEpochEvent) EventID() eventbus.EventID {
	return CurrentEpochEventId
}
