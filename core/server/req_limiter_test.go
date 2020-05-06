package server

import (
	"github.com/patrickmn/go-cache"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func Test_checkReqLimit(t *testing.T) {
	limiter := &reqLimiter{
		reqCountsByClientId: cache.New(time.Second*5, time.Minute),
		reqLimit:            10,
	}

	for i := 0; i < 10; i++ {
		require.Nil(t, limiter.checkReqLimit("client1"))
		require.Nil(t, limiter.checkReqLimit("client2"))
	}

	require.Equal(t, errReqLimitExceed, limiter.checkReqLimit("client1"))
	require.Equal(t, errReqLimitExceed, limiter.checkReqLimit("client2"))
	require.Nil(t, limiter.checkReqLimit("client3"))

	time.Sleep(time.Second * 6)

	require.Nil(t, limiter.checkReqLimit("client1"))
	require.Nil(t, limiter.checkReqLimit("client2"))
	require.Nil(t, limiter.checkReqLimit("client3"))

	limiter.reqLimit = 0
	limiter.reqCountsByClientId = nil
	for i := 0; i < 1000; i++ {
		require.Nil(t, limiter.checkReqLimit("client1"))
	}
}
