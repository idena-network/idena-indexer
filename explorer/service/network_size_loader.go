package service

type NetworkSizeLoader interface {
	Load() (uint64, error)
}
