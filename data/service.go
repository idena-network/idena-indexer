package data

import (
	"fmt"
	"github.com/idena-network/idena-go/common/eventbus"
	state2 "github.com/idena-network/idena-go/core/state"
	"github.com/idena-network/idena-indexer/events"
	"github.com/idena-network/idena-indexer/log"
	"github.com/pkg/errors"
	"time"
)

func StartDataService(eventBus eventbus.Bus, dbAccessor DbAccessor, logger log.Logger) {
	service := serviceImpl{
		dbAccessor: dbAccessor,
		logger:     logger,
	}
	eventBus.Subscribe(events.CurrentEpochEventId, func(e eventbus.Event) {
		currentEpochEvent := e.(*events.CurrentEpochEvent)
		service.state = &state{
			epoch:       currentEpochEvent.Epoch,
			epochHeight: currentEpochEvent.EpochHeight,
			height:      currentEpochEvent.Height,
			epochPeriod: currentEpochEvent.EpochPeriod,
		}
	})
	eventBus.Subscribe(events.NewEpochEventId, func(e eventbus.Event) {
		newEpochEvent := e.(*events.NewEpochEvent)
		service.state.epoch = newEpochEvent.Epoch
		service.state.epochHeight = newEpochEvent.EpochHeight
	})
	eventBus.Subscribe(events.NewBlockEventId, func(e eventbus.Event) {
		newBlockEvent := e.(*events.NewBlockEvent)
		service.state.height = newBlockEvent.Height
		service.state.epochPeriod = newBlockEvent.EpochPeriod
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
	epoch       uint16
	epochHeight uint64
	height      uint64
	epochPeriod state2.ValidationPeriod
}

func (service *serviceImpl) observe() {
	isFirst := true
	for {
		if !isFirst {
			time.Sleep(time.Minute)
		} else {
			isFirst = false
		}
		if service.state == nil {
			service.logger.Info("State not initialized yet to check data list")
			continue
		}
		dataList, err := service.dbAccessor.GetDataList()
		if err != nil {
			service.logger.Warn(errors.Wrap(err, "unable to get data list").Error())
			continue
		}
		service.logger.Debug("Checking data list")
		now := time.Now().UTC()
		for _, dataItem := range dataList {
			start := time.Now()
			if dataItem.RefreshPeriod == nil || dataItem.RefreshProcedure == nil {
				continue
			}
			if *dataItem.RefreshPeriod == refreshPeriodDay || *dataItem.RefreshPeriod == refreshPeriodHour {
				const minutesAfterValidation = 30
				const blocksAfterValidation = 3 * minutesAfterValidation
				isValidation := service.state.epochPeriod >= state2.FlipLotteryPeriod || service.state.height-service.state.epochHeight < blocksAfterValidation
				var period time.Duration
				switch *dataItem.RefreshPeriod {
				case refreshPeriodDay:
					period = time.Hour * 24
				case refreshPeriodHour:
					period = time.Hour
				}
				truncatedNow := now.Truncate(period)
				if dataItem.RefreshTime == nil && dataItem.RefreshDelay != nil {
					refreshTime := truncatedNow.Add(*dataItem.RefreshDelay)
					if err := service.dbAccessor.UpdateRefreshTime(dataItem.Name, refreshTime); err != nil {
						service.logger.Error(fmt.Sprintf("Unable to update refresh time, name: %v, err: %v", dataItem.Name, err.Error()), "d", time.Since(start))
						continue
					}
					service.logger.Info(fmt.Sprintf("Updated delayed refresh time, name: %v", dataItem.Name), "d", time.Since(start))
					continue
				}
				if dataItem.RefreshTime == nil || dataItem.RefreshTime.Before(now) {
					if isValidation {
						refreshTime := now.Add(time.Minute * minutesAfterValidation)
						if err := service.dbAccessor.UpdateRefreshTime(dataItem.Name, refreshTime); err != nil {
							service.logger.Error(fmt.Sprintf("Unable to update refresh time due to validation, name: %v, err: %v", dataItem.Name, err.Error()), "d", time.Since(start))
							continue
						}
						service.logger.Info(fmt.Sprintf("Updated refresh time due to validation, name: %v", dataItem.Name), "d", time.Since(start))
						continue
					}
					nextRefreshTime := truncatedNow.Add(period)
					if dataItem.RefreshDelay != nil {
						nextRefreshTime = nextRefreshTime.Add(*dataItem.RefreshDelay)
					}
					if err := service.dbAccessor.Refresh(dataItem.Name, *dataItem.RefreshProcedure, time.Now().UTC(), &nextRefreshTime, nil); err != nil {
						service.logger.Error(fmt.Sprintf("Unable to refresh, name: %v, err: %v", dataItem.Name, err.Error()), "d", time.Since(start))
						continue
					}
					service.logger.Info(fmt.Sprintf("Refreshed, name: %v", dataItem.Name), "d", time.Since(start))
				}
				continue
			}
			if *dataItem.RefreshPeriod == refreshPeriodEpoch {
				currentEpoch := service.state.epoch
				if dataItem.RefreshEpoch == nil || currentEpoch >= *dataItem.RefreshEpoch {
					if dataItem.RefreshTime == nil && dataItem.RefreshDelay != nil {
						refreshTime := now.Add(*dataItem.RefreshDelay)
						if err := service.dbAccessor.UpdateRefreshTime(dataItem.Name, refreshTime); err != nil {
							service.logger.Error(fmt.Sprintf("Unable to update refresh time, name: %v, err: %v", dataItem.Name, err.Error()), "d", time.Since(start))
							continue
						}
						service.logger.Info(fmt.Sprintf("Updated delayed refresh time, name: %v", dataItem.Name), "d", time.Since(start))
						continue
					}
					if dataItem.RefreshTime == nil || dataItem.RefreshTime.Before(now) {
						nextEpoch := currentEpoch + 1
						if err := service.dbAccessor.Refresh(dataItem.Name, *dataItem.RefreshProcedure, time.Now().UTC(), nil, &nextEpoch); err != nil {
							service.logger.Error(fmt.Sprintf("Unable to refresh, name: %v, err: %v", dataItem.Name, err.Error()), "d", time.Since(start))
							continue
						}
						service.logger.Info(fmt.Sprintf("Refreshed, name: %v", dataItem.Name), "d", time.Since(start))
					}
				}
				continue
			}
			service.logger.Error(fmt.Sprintf("Unknown refresh period: %v", *dataItem.RefreshPeriod))
		}
	}
}
