package api

import (
	"github.com/idena-network/idena-indexer/core/activity"
	"github.com/idena-network/idena-indexer/core/penalty"
)

type Api struct {
	lastActivities   activity.LastActivitiesHolder
	currentPenalties penalty.CurrentPenaltiesHolder
}

func NewApi(lastActivities activity.LastActivitiesHolder, currentPenalties penalty.CurrentPenaltiesHolder) *Api {
	return &Api{
		lastActivities:   lastActivities,
		currentPenalties: currentPenalties,
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

func (a *Api) GetCurrentPenaltiesCount() uint64 {
	return uint64(len(a.currentPenalties.GetAll()))
}

func (a *Api) GetCurrentPenalties(startIndex, count uint64) []*Penalty {
	var res []*Penalty
	all := a.currentPenalties.GetAll()
	if len(all) > 0 {
		for i := startIndex; i >= 0 && i < startIndex+count && i < uint64(len(all)); i++ {
			la := all[i]
			res = append(res, &Penalty{
				Address: la.Address,
				Penalty: la.Penalty,
			})
		}
	}
	return res
}

func (a *Api) GetCurrentPenalty(address string) *Penalty {
	la := a.currentPenalties.Get(address)
	if la == nil {
		return nil
	}
	return &Penalty{
		Address: la.Address,
		Penalty: la.Penalty,
	}
}
