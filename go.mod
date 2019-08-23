module github.com/idena-network/idena-indexer

go 1.12

replace github.com/tendermint/iavl => github.com/idena-network/iavl v0.12.3-0.20190724103809-104317193459

require (
	github.com/deckarep/golang-set v1.7.1
	github.com/go-stack/stack v1.8.0
	github.com/gorilla/handlers v1.4.2
	github.com/gorilla/mux v1.7.3
	github.com/idena-network/idena-go v0.6.5-0.20190822201722-7bff5bd30dd1
	github.com/ipfs/go-cid v0.0.2
	github.com/lib/pq v1.1.1
	github.com/pkg/errors v0.8.1
	github.com/shopspring/decimal v0.0.0-20180709203117-cd690d0c9e24
	golang.org/x/image v0.0.0-20190802002840-cff245a6509b
	gopkg.in/urfave/cli.v1 v1.20.0
)
