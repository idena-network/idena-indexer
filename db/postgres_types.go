package db

import (
	"database/sql/driver"
	"fmt"
)

func (r *Reward) Value() (driver.Value, error) {
	return fmt.Sprintf("(%v,%v,%v,%v)", r.Address, r.Balance, r.Stake, r.Type), nil
}
