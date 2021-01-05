package postgres

import (
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_calculateEstimatedMaxOracleReward(t *testing.T) {
	votingMinPayment := decimal.New(2, 0)
	res := calculateEstimatedMaxOracleReward(
		decimal.New(100, 0),
		&votingMinPayment,
		50,
		200,
		5,
		51,
		4)

	require.Equal(t, "12.9412", res.StringFixed(4))

	res = calculateEstimatedMaxOracleReward(
		decimal.New(10, 0),
		nil,
		0,
		10,
		1,
		66,
		0)

	require.Equal(t, "15.1515", res.StringFixed(4))
}
