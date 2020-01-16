package db

import (
	"database/sql"
	"fmt"
	"github.com/pkg/errors"
)

type context struct {
	epoch                   uint64
	blockHeight             uint64
	flipIdsPerCid           map[string]int64
	txIdsPerHash            map[string]int64
	epochIdentityIdsPerAddr map[string]int64
	a                       *postgresAccessor
	tx                      *sql.Tx
}

func newContext(a *postgresAccessor, tx *sql.Tx, epoch uint64, blockHeight uint64) *context {
	return &context{
		a:           a,
		tx:          tx,
		epoch:       epoch,
		blockHeight: blockHeight,
	}
}

func (c *context) flipId(cid string) (int64, error) {
	if c.flipIdsPerCid == nil {
		c.flipIdsPerCid = make(map[string]int64)
	}
	if id, present := c.flipIdsPerCid[cid]; present {
		return id, nil
	}
	id, err := c.a.getFlipId(c.tx, cid)
	if err != nil {
		return 0, err
	}
	c.flipIdsPerCid[cid] = id
	return id, nil
}

func (c *context) txId(hash string) (int64, error) {
	if txId, present := c.txIdsPerHash[hash]; !present {
		return 0, errors.New(fmt.Sprintf("Id for tx %s not found", hash))
	} else {
		return txId, nil
	}
}
