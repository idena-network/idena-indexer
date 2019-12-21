package monitoring

import (
	"fmt"
	"github.com/idena-network/idena-indexer/log"
	"sort"
	"sync"
	"time"
)

type PerformanceMonitor interface {
	Start(name string, details string) uint32
	Complete(id uint32)
}

type emptyPerformanceMonitor struct {
}

func NewEmptyPerformanceMonitor() PerformanceMonitor {
	return &emptyPerformanceMonitor{}
}

func (pm *emptyPerformanceMonitor) Start(name string, details string) uint32 {
	// do nothing
	return 0
}

func (pm *emptyPerformanceMonitor) Complete(id uint32) {
	// do nothing
}

type performanceMonitorImpl struct {
	recordsById map[uint32]*record
	mutex       sync.Mutex
	counter     uint32
	log         log.Logger
}

type record struct {
	id      uint32
	name    string
	details string
	start   time.Time
	finish  *time.Time
}

type report struct {
	maxReqCount int
	nameReports []nameReport
}

type nameReport struct {
	name            string
	count           int
	averageDuration time.Duration
	topMax          []*record
	topMin          []*record
}

func NewPerformanceMonitor(interval time.Duration, log log.Logger) PerformanceMonitor {
	pm := &performanceMonitorImpl{
		log:         log,
		recordsById: make(map[uint32]*record),
	}
	go pm.loop(interval)
	return pm
}

func (pm *performanceMonitorImpl) Start(name string, details string) uint32 {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()
	pm.counter++
	id := pm.counter
	pm.recordsById[id] = &record{
		id:      id,
		name:    name,
		details: details,
		start:   time.Now(),
	}
	return id
}

func (pm *performanceMonitorImpl) Complete(id uint32) {
	finish := time.Now()
	pm.mutex.Lock()
	defer pm.mutex.Unlock()
	rec, ok := pm.recordsById[id]
	if !ok {
		pm.log.Error(fmt.Sprintf("Record %v not found", id))
		return
	}
	if rec.finish != nil {
		pm.log.Error(fmt.Sprintf("Record %v already completed", id))
		return
	}
	rec.finish = &finish
}

func (pm *performanceMonitorImpl) loop(interval time.Duration) {
	for {
		time.Sleep(interval)
		pm.report()
	}
}

func (pm *performanceMonitorImpl) report() {
	recordsById := pm.copyRecords()
	report := getReport(recordsById)
	pm.logReport(report)
}

func (pm *performanceMonitorImpl) copyRecords() map[uint32]*record {
	res := make(map[uint32]*record)
	pm.mutex.Lock()
	defer pm.mutex.Unlock()
	for id, rec := range pm.recordsById {
		if rec.finish != nil {
			delete(pm.recordsById, id)
		}
		res[id] = rec
	}
	return res
}

func getReport(recordsById map[uint32]*record) report {
	completedRecordsByName := getCompletedRecordsByName(recordsById)
	report := report{}
	for name, nameRecords := range completedRecordsByName {
		report.nameReports = append(report.nameReports, getNameReport(name, nameRecords))
	}
	sort.Slice(report.nameReports, func(i, j int) bool {
		return report.nameReports[i].averageDuration > report.nameReports[j].averageDuration
	})
	report.maxReqCount = determineMaxReqCount(recordsById)
	return report
}

func getCompletedRecordsByName(recordsById map[uint32]*record) map[string][]*record {
	res := make(map[string][]*record)
	for _, rec := range recordsById {
		if rec.finish == nil {
			continue
		}
		res[rec.name] = append(res[rec.name], rec)
	}
	return res
}

func getNameReport(name string, nameRecords []*record) nameReport {
	type recordWrapper struct {
		record   *record
		duration time.Duration
	}
	var recordWrappers []recordWrapper

	var totalDuration time.Duration
	for _, record := range nameRecords {
		duration := record.finish.Sub(record.start)
		totalDuration += duration
		recordWrappers = append(recordWrappers, recordWrapper{
			record:   record,
			duration: duration,
		})
	}
	sort.Slice(recordWrappers, func(i, j int) bool {
		return recordWrappers[i].duration < recordWrappers[j].duration
	})
	var averageDuration time.Duration
	count := len(recordWrappers)
	if count > 0 {
		averageDuration = time.Duration(float64(totalDuration) / float64(count))
	}
	var topMin, topMax []*record
	i := 0
	for len(topMin) < 3 && i < count {
		topMin = append(topMin, recordWrappers[i].record)
		i++
	}
	i = count - 1
	for len(topMax) < 3 && i >= 0 {
		topMax = append(topMax, recordWrappers[i].record)
		i--
	}
	return nameReport{
		name:            name,
		count:           count,
		averageDuration: averageDuration,
		topMax:          topMax,
		topMin:          topMin,
	}
}

func determineMaxReqCount(recordsById map[uint32]*record) int {
	type event struct {
		time    time.Time
		isStart bool
	}
	var events []event
	for _, record := range recordsById {
		events = append(events, event{
			time:    record.start,
			isStart: true,
		})
		if record.finish != nil {
			events = append(events, event{
				time:    *record.finish,
				isStart: false,
			})
		}
	}
	sort.Slice(events, func(i, j int) bool {
		if events[i].time.Before(events[j].time) {
			return true
		}
		if events[i].time.After(events[j].time) {
			return false
		}
		return events[i].isStart
	})
	cnt := 0
	res := 0
	for _, event := range events {
		if event.isStart {
			cnt++
			if cnt > res {
				res = cnt
			}
		} else {
			cnt--
		}
	}
	return res
}

func (pm *performanceMonitorImpl) logReport(report report) {
	pm.log.Info("============== Monitoring report ==============")
	pm.log.Info(fmt.Sprintf("max req count: %v", report.maxReqCount))
	for _, nameReport := range report.nameReports {
		pm.log.Info(fmt.Sprintf("name: %v", nameReport.name))
		pm.log.Info(fmt.Sprintf("    cnt: %v, avg: %v", nameReport.count, nameReport.averageDuration))
		pm.log.Info("    top max:")
		for _, record := range nameReport.topMax {
			pm.log.Info(fmt.Sprintf("        d: %v, details: %v", record.finish.Sub(record.start), record.details))
		}
		pm.log.Info("    top min:")
		for _, record := range nameReport.topMin {
			pm.log.Info(fmt.Sprintf("        d: %v, details: %v", record.finish.Sub(record.start), record.details))
		}
	}
}
