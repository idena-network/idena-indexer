module github.com/idena-network/idena-indexer

go 1.12

replace github.com/tendermint/iavl => github.com/idena-network/iavl v0.12.3-0.20190724103809-104317193459

require (
	github.com/go-stack/stack v1.8.0
	github.com/idena-network/idena-go v0.5.1-0.20190729054315-6b8be903f16a
	github.com/ipfs/go-cid v0.0.2
	github.com/lib/pq v1.1.1
	github.com/pkg/errors v0.8.1
	github.com/shopspring/decimal v0.0.0-20180709203117-cd690d0c9e24
	gopkg.in/urfave/cli.v1 v1.20.0
)
