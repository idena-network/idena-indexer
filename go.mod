module github.com/idena-network/idena-indexer

go 1.13

replace github.com/tendermint/iavl => github.com/idena-network/iavl v0.12.3-0.20200414113415-041d1524315e

require (
	github.com/alecthomas/template v0.0.0-20190718012654-fb15b899a751
	github.com/deckarep/golang-set v1.7.1
	github.com/go-stack/stack v1.8.0
	github.com/gorilla/handlers v1.4.2
	github.com/gorilla/mux v1.7.3
	github.com/idena-network/idena-go v0.21.0
	github.com/ipfs/go-cid v0.0.6
	github.com/lib/pq v1.1.1
	github.com/patrickmn/go-cache v2.1.0+incompatible
	github.com/pierrec/lz4 v2.5.2+incompatible // indirect
	github.com/pkg/errors v0.9.1
	github.com/shopspring/decimal v0.0.0-20200227202807-02e2044944cc
	github.com/stretchr/testify v1.6.0
	github.com/swaggo/http-swagger v0.0.0-20200308142732-58ac5e232fba
	github.com/swaggo/swag v1.6.7
	github.com/tendermint/tm-db v0.5.1
	golang.org/x/image v0.0.0-20190802002840-cff245a6509b
	gopkg.in/urfave/cli.v1 v1.20.0
)
