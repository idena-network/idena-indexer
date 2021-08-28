package data

import (
	"fmt"
	"github.com/idena-network/idena-go/common/eventbus"
	"github.com/idena-network/idena-indexer/events"
	"github.com/idena-network/idena-indexer/log"
	"github.com/pkg/errors"
	"time"
)

func StartDataService(eventBus eventbus.Bus, dbAccessor DbAccessor, logger log.Logger) {
	service := serviceImpl{
		dbAccessor: dbAccessor,
		logger:     logger,
		state:      &state{},
	}
	eventBus.Subscribe(events.CurrentEpochEventId, func(e eventbus.Event) {
		currentEpochEvent := e.(*events.CurrentEpochEvent)
		epoch := currentEpochEvent.Epoch
		service.state.epoch = &epoch
	})
	eventBus.Subscribe(events.NewEpochEventId, func(e eventbus.Event) {
		newEpochEvent := e.(*events.NewEpochEvent)
		epoch := newEpochEvent.Epoch
		service.state.epoch = &epoch
	})
	go service.observe()
}

type serviceImpl struct {
	dbAccessor DbAccessor
	configFile string
	logger     log.Logger

	state *state
}

type state struct {
	epoch *uint16
}

func (service *serviceImpl) observe() {
	isFirst := true
	for {
		if !isFirst {
			time.Sleep(time.Minute)
		} else {
			isFirst = false
		}
		dataList, err := service.dbAccessor.GetDataList()
		if err != nil {
			service.logger.Warn(errors.Wrap(err, "unable to get data list").Error())
			continue
		}
		service.logger.Debug("Checking data list")
		now := time.Now().UTC()
		for _, dataItem := range dataList {
			if dataItem.RefreshPeriod == nil || dataItem.RefreshProcedure == nil {
				continue
			}
			if *dataItem.RefreshPeriod == refreshPeriodDay {
				currentDay := now.Truncate(time.Hour * 24)
				if dataItem.RefreshTime == nil && dataItem.RefreshDelay != nil {
					refreshTime := currentDay.Add(*dataItem.RefreshDelay)
					if err := service.dbAccessor.UpdateRefreshTime(dataItem.Name, refreshTime); err != nil {
						service.logger.Error(fmt.Sprintf("Unable to update refresh time, name: %v, err: %v", dataItem.Name, err.Error()))
						continue
					}
					service.logger.Info(fmt.Sprintf("Updated delayed refresh time, name: %v", dataItem.Name))
					continue
				}
				if dataItem.RefreshTime == nil || dataItem.RefreshTime.Before(now) {
					nextRefreshTime := currentDay.Add(time.Hour * 24)
					if dataItem.RefreshDelay != nil {
						nextRefreshTime = nextRefreshTime.Add(*dataItem.RefreshDelay)
					}
					if err := service.dbAccessor.Refresh(dataItem.Name, *dataItem.RefreshProcedure, time.Now().UTC(), &nextRefreshTime, nil); err != nil {
						service.logger.Error(fmt.Sprintf("Unable to refresh, name: %v, err: %v", dataItem.Name, err.Error()))
						continue
					}
					service.logger.Info(fmt.Sprintf("Refreshed, name: %v", dataItem.Name))
				}
				continue
			}
			if *dataItem.RefreshPeriod == refreshPeriodEpoch {
				currentEpoch := service.state.epoch
				if currentEpoch == nil {
					service.logger.Info(fmt.Sprintf("Current epoch is not initialized, name: %v", dataItem.Name))
					continue
				}
				if dataItem.RefreshEpoch == nil || *currentEpoch >= *dataItem.RefreshEpoch {
					if dataItem.RefreshTime == nil && dataItem.RefreshDelay != nil {
						refreshTime := now.Add(*dataItem.RefreshDelay)
						if err := service.dbAccessor.UpdateRefreshTime(dataItem.Name, refreshTime); err != nil {
							service.logger.Error(fmt.Sprintf("Unable to update refresh time, name: %v, err: %v", dataItem.Name, err.Error()))
							continue
						}
						service.logger.Info(fmt.Sprintf("Updated delayed refresh time, name: %v", dataItem.Name))
						continue
					}
					if dataItem.RefreshTime == nil || dataItem.RefreshTime.Before(now) {
						nextEpoch := *currentEpoch + 1
						if err := service.dbAccessor.Refresh(dataItem.Name, *dataItem.RefreshProcedure, time.Now().UTC(), nil, &nextEpoch); err != nil {
							service.logger.Error(fmt.Sprintf("Unable to refresh, name: %v, err: %v", dataItem.Name, err.Error()))
							continue
						}
						service.logger.Info(fmt.Sprintf("Refreshed, name: %v", dataItem.Name))
					}
				}
				continue
			}
			service.logger.Error(fmt.Sprintf("Unknown refresh period: %v", *dataItem.RefreshPeriod))
		}
	}
}
