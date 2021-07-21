package data

import "time"

type Data struct {
	Name             string
	RefreshProcedure *string
	RefreshPeriod    *refreshPeriod
	RefreshDelay     *time.Duration
	RefreshTime      *time.Time
	RefreshEpoch     *uint16
}

type refreshPeriod = string

const (
	refreshPeriodEpoch = "e"
	refreshPeriodDay   = "d"
)
