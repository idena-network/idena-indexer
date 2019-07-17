module github.com/idena-network/idena-indexer

go 1.12

replace github.com/tendermint/iavl => github.com/idena-network/iavl v0.0.0-20190701090235-eef65d855b4a

require (
	github.com/go-stack/stack v1.8.0
	github.com/idena-network/idena-go v0.3.2-0.20190717113723-f18f3d93da16
	github.com/ipsn/go-ipfs v0.0.0-20190407150747-8b9b72514244
	github.com/lib/pq v1.1.1
	github.com/pkg/errors v0.8.1
	gopkg.in/urfave/cli.v1 v1.20.0
)
