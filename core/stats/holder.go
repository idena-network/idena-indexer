package stats

type StatsHolder interface {
	GetStats() *Stats
}
