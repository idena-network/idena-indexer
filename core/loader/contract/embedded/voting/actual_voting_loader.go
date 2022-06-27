package voting

import "github.com/idena-network/idena-go/common"

type ActualOracleVotingsLoader interface {
	ActualOracleVotings() ([]common.Address, error)
}
