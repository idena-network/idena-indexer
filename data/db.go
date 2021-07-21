package data

import "time"

type DbAccessor interface {
	GetDataList() ([]Data, error)
	UpdateRefreshTime(name string, refreshTime time.Time) error
	Refresh(name, refreshProcedure string, refreshTime *time.Time, refreshEpoch *uint16) error
}
