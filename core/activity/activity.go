package activity

import (
	"github.com/idena-network/idena-go/blockchain"
	"sort"
	"strings"
	"time"
)

type Activity struct {
	Address string
	Time    time.Time
}

type LastActivitiesHolder interface {
	GetAll() []*Activity
	Get(address string) *Activity
}

func NewLastActivitiesCache(offlineDetector *blockchain.OfflineDetector) LastActivitiesHolder {
	cache := &lastActivitiesCache{}
	cache.initialize(offlineDetector)
	return cache
}

type lastActivitiesCache struct {
	activities           []*Activity
	activitiesPerAddress map[string]*Activity
}

type lastActivitiesCacheUpdater struct {
	cache           *lastActivitiesCache
	offlineDetector *blockchain.OfflineDetector
}

func (cache *lastActivitiesCache) GetAll() []*Activity {
	return cache.activities
}

func (cache *lastActivitiesCache) Get(address string) *Activity {
	return cache.activitiesPerAddress[strings.ToLower(address)]
}

func (cache *lastActivitiesCache) set(activities []*Activity, activitiesPerAddress map[string]*Activity) {
	cache.activities = activities
	cache.activitiesPerAddress = activitiesPerAddress
}

func (cache *lastActivitiesCache) initialize(offlineDetector *blockchain.OfflineDetector) {
	updater := lastActivitiesCacheUpdater{
		cache:           cache,
		offlineDetector: offlineDetector,
	}
	go updater.loop()
}

func (updater *lastActivitiesCacheUpdater) loop() {
	for {
		time.Sleep(time.Second * 5)

		activityMap := updater.offlineDetector.GetActivityMap()
		var activities []*Activity
		activitiesPerAddress := make(map[string]*Activity, len(activityMap))
		for address, activityTime := range activityMap {
			addressStr := address.Hex()
			activity := &Activity{
				Address: addressStr,
				Time:    activityTime,
			}
			activities = append(activities, activity)
			activitiesPerAddress[strings.ToLower(addressStr)] = activity
		}

		sort.Slice(activities, func(i, j int) bool {
			return activities[i].Time.After(activities[j].Time)
		})

		updater.cache.set(activities, activitiesPerAddress)
	}
}
