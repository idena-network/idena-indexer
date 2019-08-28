package api

import (
	"time"
)

type Activity struct {
	Address string    `json:"address"`
	Time    time.Time `json:"timestamp"`
}
