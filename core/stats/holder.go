package stats

type StatsHolder interface {
	GetStats() *Stats
	Disable()
	Enable()
}
