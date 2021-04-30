module github.com/idena-network/idena-indexer

go 1.13

replace github.com/cosmos/iavl => github.com/idena-network/iavl v0.12.3-0.20210112075003-70ccb13c86c9

require (
	github.com/deckarep/golang-set v1.7.1
	github.com/go-stack/stack v1.8.0
	github.com/gorilla/handlers v1.4.2
	github.com/gorilla/mux v1.7.3
	github.com/idena-network/idena-go v0.25.1-0.20210323080324-2e4c5dd63563
	github.com/ipfs/go-cid v0.0.7
	github.com/ipfs/go-ipfs-pinner v0.1.1 // indirect
	github.com/lib/pq v1.1.1
	github.com/pkg/errors v0.9.1
	github.com/shopspring/decimal v0.0.0-20200227202807-02e2044944cc
	github.com/stretchr/testify v1.7.0
	github.com/tendermint/tm-db v0.6.4
	golang.org/x/image v0.0.0-20190802002840-cff245a6509b
	gopkg.in/urfave/cli.v1 v1.20.0
)
