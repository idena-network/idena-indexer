package tests

import (
	"github.com/idena-network/idena-go/tests"
	"github.com/idena-network/idena-indexer/db"
	"github.com/idena-network/idena-indexer/tests/common"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func Test_penalty(t *testing.T) {
	dbConnector, dbAccessor := common.InitDefaultPostgres(filepath.Join("..", ".."))
	address1 := tests.GetRandAddr()
	address2 := tests.GetRandAddr()

	// When
	data := createBaseData()
	data.Addresses = []db.Address{{Address: address1.Hex()}, {Address: address2.Hex()}}
	penaltyNotToPay := db.Penalty{Penalty: decimal.New(234, -2), Address: address1.Hex()}
	data.Penalties = []db.Penalty{penaltyNotToPay}
	err := dbAccessor.Save(data)
	// Then
	require.Nil(t, err)

	// When
	data = createBaseData()
	data.Block = createBlock(2)
	penaltyToPay := db.Penalty{Penalty: decimal.New(234, -2), Address: address1.Hex()}
	data.Penalties = []db.Penalty{penaltyToPay}
	err = dbAccessor.Save(data)
	// Then
	require.Nil(t, err)

	// When
	data = createBaseData()
	data.Block = createBlock(3)
	data.BurntPenalties = []db.Penalty{{Address: address2.Hex(), Penalty: decimal.New(101, -2)}}
	err = dbAccessor.Save(data)
	// Then
	require.Nil(t, err)
	//require.True(t, strings.Contains(err.Error(), "there is no penalty to close"))

	// When
	data = createBaseData()
	data.Block = createBlock(4)
	data.BurntPenalties = []db.Penalty{{Address: address1.Hex(), Penalty: decimal.New(101, -2)}}
	err = dbAccessor.Save(data)
	paidPenalties, err2 := common.GetPaidPenalties(dbConnector)
	// Then
	require.Nil(t, err)
	require.Nil(t, err2)
	require.Equal(t, 1, len(paidPenalties))
	require.Equal(t, 4, int(paidPenalties[0].BlockHeight))
	require.Equal(t, 2, int(paidPenalties[0].PenaltyId))
	require.Equal(t, "1.33", paidPenalties[0].Penalty.String())

	// When
	data = createBaseData()
	data.Block = createBlock(5)
	data.BurntPenalties = []db.Penalty{{Address: address1.Hex(), Penalty: decimal.New(101, -2)}}
	err = dbAccessor.Save(data)
	// Then
	require.NotNil(t, err)
	require.True(t, strings.Contains(err.Error(), "latest penalty is already closed"))
}

func Test_PenaltyWithNotPaidPreviousOne(t *testing.T) {
	dbConnector, dbAccessor := common.InitDefaultPostgres(filepath.Join("..", ".."))
	address := tests.GetRandAddr()

	// When
	data := createBaseData()
	data.Addresses = []db.Address{{Address: address.Hex()}}
	data.Penalties = []db.Penalty{{Penalty: decimal.New(234, -2), Address: address.Hex()}}
	err := dbAccessor.Save(data)
	// Then
	require.Nil(t, err)

	// When
	data = createBaseData()
	data.Block = createBlock(2)
	data.Penalties = []db.Penalty{{Penalty: decimal.New(123, -2), Address: address.Hex()}}
	data.BurntPenalties = []db.Penalty{{Address: address.Hex(), Penalty: decimal.New(101, -2)}}
	err = dbAccessor.Save(data)
	paidPenalties, err2 := common.GetPaidPenalties(dbConnector)
	// Then
	require.Nil(t, err)
	require.Nil(t, err2)
	require.Equal(t, 1, len(paidPenalties))
	require.Equal(t, 2, int(paidPenalties[0].BlockHeight))
	require.Equal(t, 1, int(paidPenalties[0].PenaltyId))
	require.Equal(t, "1.33", paidPenalties[0].Penalty.String())
}

func createBaseData() *db.Data {
	return &db.Data{
		Block: createBlock(1),
		Epoch: 10,
	}
}

func createBlock(height uint64) db.Block {
	return db.Block{
		Height: height,
		Hash:   strconv.Itoa(int(height)),
	}
}
