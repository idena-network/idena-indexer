package monitoring

import (
	"fmt"
	"github.com/idena-network/idena-indexer/log"
	"strings"
	"time"
)

type PerformanceMonitor interface {
	Start(name string)
	Complete(name string)
}

type emptyPerformanceMonitor struct {
}

func (pm *emptyPerformanceMonitor) Start(name string) {
	// do nothing
}
func (pm *emptyPerformanceMonitor) Complete(name string) {
	// do nothing
}

type performanceMonitorImpl struct {
	log         log.Logger
	records     []*record
	startTimes  []*eventType
	blocksToLog int
}

type record struct {
	name       string
	duration   time.Duration
	subRecords []*record
}

type eventType struct {
	name       string
	time       time.Time
	subRecords []*record
}

func NewEmptyPerformanceMonitor() PerformanceMonitor {
	return &emptyPerformanceMonitor{}
}

func NewPerformanceMonitor(blocksToLog int, log log.Logger) PerformanceMonitor {
	return &performanceMonitorImpl{
		log:         log,
		blocksToLog: blocksToLog,
	}
}

func (pm *performanceMonitorImpl) Start(name string) {
	pm.startTimes = append(pm.startTimes, &eventType{
		name: name,
		time: time.Now(),
	})
}

func (pm *performanceMonitorImpl) Complete(name string) {
	t := time.Now()
	eventType := pm.startTimes[len(pm.startTimes)-1]
	if eventType.name != name {
		panic(fmt.Sprintf("unexpected name to complete: %v, expected: %v", name, eventType.name))
	}
	pm.startTimes = pm.startTimes[:len(pm.startTimes)-1]

	a := &record{
		name:       name,
		duration:   t.Sub(eventType.time),
		subRecords: eventType.subRecords,
	}

	if len(pm.startTimes) > 0 {
		parentStartTime := pm.startTimes[len(pm.startTimes)-1]
		parentStartTime.subRecords = append(parentStartTime.subRecords, a)
	} else {
		pm.records = append(pm.records, a)

		if len(pm.records) == pm.blocksToLog {
			records := pm.records
			pm.records = nil
			go pm.logReport(records)
		}
	}
}

func (pm *performanceMonitorImpl) logReport(records []*record) {
	totals := make(map[string]time.Duration)
	collectTotals(totals, records)
	pm.logRecords(totals, records[:1], 0, len(records))

}

func collectTotals(totals map[string]time.Duration, records []*record) {
	for _, r := range records {
		if _, ok := totals[r.name]; !ok {
			totals[r.name] = r.duration
		} else {
			totals[r.name] = totals[r.name] + r.duration
		}
		collectTotals(totals, r.subRecords)
	}
}

func (pm *performanceMonitorImpl) logRecords(totals map[string]time.Duration, records []*record, level int, count int) {
	gap := strings.Repeat(" ", level*5)
	for _, r := range records {
		pm.log.Info(fmt.Sprintf("%v%v: %v/%v=%v", gap, r.name, totals[r.name], count, time.Duration(int64(totals[r.name])/int64(count))))
		pm.logRecords(totals, r.subRecords, level+1, count)
	}
}
