module github.com/idena-network/idena-indexer

go 1.16

replace github.com/cosmos/iavl => github.com/idena-network/iavl v0.12.3-0.20210604085842-854e73deab29

require (
	github.com/cosmos/iavl v0.15.3
	github.com/deckarep/golang-set v1.7.1
	github.com/go-stack/stack v1.8.1
	github.com/golang/protobuf v1.5.2
	github.com/gorilla/handlers v1.4.2
	github.com/gorilla/mux v1.7.3
	github.com/idena-network/idena-go v0.28.7
	github.com/ipfs/go-cid v0.1.0
	github.com/ipfs/interface-go-ipfs-core v0.5.2
	github.com/lib/pq v1.1.1
	github.com/mholt/archiver/v3 v3.5.1-0.20210112195346-074da64920d3
	github.com/patrickmn/go-cache v2.1.0+incompatible
	github.com/pkg/errors v0.9.1
	github.com/shopspring/decimal v0.0.0-20200227202807-02e2044944cc
	github.com/stretchr/testify v1.7.0
	github.com/tendermint/tm-db v0.6.4
	golang.org/x/image v0.0.0-20190802002840-cff245a6509b
	gopkg.in/urfave/cli.v1 v1.20.0
)
