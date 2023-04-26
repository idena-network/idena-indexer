package verification

import (
	"database/sql"
	"github.com/idena-network/idena-go/common"
	"github.com/idena-network/idena-indexer/core/conversion"
	"github.com/pkg/errors"
	"time"
)

type State = uint8

const (
	StatePending  = State(0)
	StateVerified = State(1)
	StateFailed   = State(2)
)

type VerifierDb interface {
	SavePendingVerification(contractAddress common.Address, data []byte, fileName string) (usrErr, err error)
	GetPendingVerification() (*PendingVerification, error)
	UpdateVerificationState(contractAddress common.Address, state State, data []byte, errorMessage *string) error
}

type PendingVerification struct {
	Address common.Address
	Code    []byte
	Data    []byte
}

type VerifierPostgres struct {
	db *sql.DB
}

func NewVerifierPostgres(connStr string) *VerifierPostgres {
	dbAccessor, err := sql.Open("postgres", connStr)
	if err != nil {
		panic(err)
	}
	dbAccessor.SetMaxOpenConns(5)
	dbAccessor.SetMaxIdleConns(5)
	dbAccessor.SetConnMaxLifetime(5 * time.Minute)
	return &VerifierPostgres{
		db: dbAccessor,
	}
}

func (vdb *VerifierPostgres) SavePendingVerification(contractAddress common.Address, data []byte, fileName string) (usrErr, err error) {
	const query = "SELECT coalesce(save_contract_pending_verification($1, $2, $3, $4), '');"
	timestamp := time.Now().UTC().Unix()
	var userErrorMsg string
	if err = vdb.db.QueryRow(query,
		conversion.ConvertAddress(contractAddress),
		timestamp,
		data,
		fileName,
	).Scan(&userErrorMsg); err != nil {
		return nil, err
	}
	if len(userErrorMsg) > 0 {
		return errors.New(userErrorMsg), nil
	}
	return nil, nil
}

func (vdb *VerifierPostgres) GetPendingVerification() (*PendingVerification, error) {
	const query = `SELECT a.address, coalesce(c.code, ''::bytea), cv.data
FROM contract_verifications cv
         LEFT JOIN addresses a ON a.id = cv.contract_address_id
         LEFT JOIN contracts c ON c.contract_address_id = cv.contract_address_id
         WHERE cv.state=$1 LIMIT 1`
	rows, err := vdb.db.Query(query, StatePending)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	if !rows.Next() {
		return nil, nil
	}
	var res PendingVerification
	var address string
	err = rows.Scan(&address, &res.Code, &res.Data)
	if err != nil {
		return nil, err
	} else {
		res.Address = common.HexToAddress(address)
		return &res, nil
	}
}

func (vdb *VerifierPostgres) UpdateVerificationState(contractAddress common.Address, state State, data []byte, errorMessage *string) error {
	const query = "call update_contract_verification_state($1, $2, $3, $4, $5);"
	timestamp := time.Now().UTC().Unix()
	_, err := vdb.db.Exec(query,
		conversion.ConvertAddress(contractAddress),
		state,
		timestamp,
		data,
		errorMessage,
	)
	return err
}
