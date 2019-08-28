package api

import (
	"github.com/idena-network/idena-indexer/core/activity"
)

type Api struct {
	lastActivities activity.LastActivitiesHolder
}

func NewApi(lastActivities activity.LastActivitiesHolder) *Api {
	return &Api{
		lastActivities: lastActivities,
	}
}

func (a *Api) GetLastActivitiesCount() uint64 {
	return uint64(len(a.lastActivities.GetAll()))
}

func (a *Api) GetLastActivities(startIndex, count uint64) []*Activity {
	var res []*Activity
	all := a.lastActivities.GetAll()
	if len(all) > 0 {
		for i := startIndex; i >= 0 && i < startIndex+count && i < uint64(len(all)); i++ {
			la := all[i]
			res = append(res, &Activity{
				Address: la.Address,
				Time:    la.Time,
			})
		}
	}
	return res
}

func (a *Api) GetLastActivity(address string) *Activity {
	la := a.lastActivities.Get(address)
	if la == nil {
		return nil
	}
	return &Activity{
		Address: la.Address,
		Time:    la.Time,
	}
}
