package postgres

import (
	"encoding/json"
	"github.com/idena-network/idena-indexer/explorer/types"
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_bigTimestamp(t *testing.T) {
	type Wrapper struct {
		Timestamp types.JSONTime `json:"timestamp" example:"2020-01-01T00:00:00Z"`
	}
	timestamp := int64(1614033480000)
	wrapper := new(Wrapper)
	wrapper.Timestamp = types.JSONTime(timestampToTimeUTC(timestamp))
	jsonBytes, err := json.Marshal(wrapper)
	require.Nil(t, err)
	require.NotEmpty(t, t, jsonBytes)
}
