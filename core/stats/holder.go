package stats

type BlockStatsHolder interface {
	GetStats() *Stats
}
