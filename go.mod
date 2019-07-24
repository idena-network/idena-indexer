module github.com/idena-network/idena-indexer

go 1.12

replace github.com/tendermint/iavl => github.com/idena-network/iavl v0.12.3-0.20190724103809-104317193459

require (
	github.com/go-stack/stack v1.8.0
	github.com/idena-network/idena-go v0.4.1-0.20190726045443-36187b8fe733
	github.com/ipsn/go-ipfs v0.0.0-20190407150747-8b9b72514244
	github.com/lib/pq v1.1.1
	github.com/pkg/errors v0.8.1
	github.com/shopspring/decimal v0.0.0-20180709203117-cd690d0c9e24
	gopkg.in/urfave/cli.v1 v1.20.0
)
