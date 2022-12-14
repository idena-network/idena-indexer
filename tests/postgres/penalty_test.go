package tests

import (
	"github.com/idena-network/idena-go/tests"
	"github.com/idena-network/idena-indexer/db"
	"github.com/idena-network/idena-indexer/tests/common"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
	"path/filepath"
	"strconv"
	"testing"
)

func Test_penalty(t *testing.T) {
	_, dbAccessor := common.InitDefaultPostgres(filepath.Join("..", ".."))
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
	err = dbAccessor.Save(data)
	// Then
	require.Nil(t, err)
	//require.True(t, strings.Contains(err.Error(), "there is no penalty to close"))

	// When
	data = createBaseData()
	data.Block = createBlock(4)
	err = dbAccessor.Save(data)
	// Then
	require.Nil(t, err)
}

func Test_penaltySeconds(t *testing.T) {
	_, dbAccessor := common.InitDefaultPostgres(filepath.Join("..", ".."))
	address1 := tests.GetRandAddr()
	address2 := tests.GetRandAddr()
	address3 := tests.GetRandAddr()

	// When
	data := createBaseData()
	data.Addresses = []db.Address{{Address: address1.Hex()}, {Address: address3.Hex()}}
	data.Penalties = []db.Penalty{
		{Seconds: 1000, Address: address1.Hex()},
	}
	err := dbAccessor.Save(data)
	// Then
	require.Nil(t, err)

	// When
	data = createBaseData()
	data.Block = createBlock(2)
	data.Penalties = []db.Penalty{
		{Seconds: 2000, Address: address1.Hex(), InheritedFrom: address2.Hex()},
	}
	err = dbAccessor.Save(data)
	// Then
	require.Nil(t, err)

	// When
	data = createBaseData()
	data.Block = createBlock(3)
	data.Penalties = []db.Penalty{
		{Seconds: 3000, Address: address1.Hex(), InheritedFrom: address2.Hex()},
		{Seconds: 4000, Address: address3.Hex()},
	}
	err = dbAccessor.Save(data)
	// Then
	require.Nil(t, err)

	// When
	data = createBaseData()
	data.Block = createBlock(4)
	err = dbAccessor.Save(data)
	// Then
	require.Nil(t, err)
}

func Test_PenaltyWithNotPaidPreviousOne(t *testing.T) {
	_, dbAccessor := common.InitDefaultPostgres(filepath.Join("..", ".."))
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
	err = dbAccessor.Save(data)
	// Then
	require.Nil(t, err)
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
