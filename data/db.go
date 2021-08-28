package data

import "time"

type DbAccessor interface {
	GetDataList() ([]Data, error)
	UpdateRefreshTime(name string, refreshTime time.Time) error
	Refresh(name, refreshProcedure string, time time.Time, nextRefreshTime *time.Time, refreshEpoch *uint16) error
}
