package entity

import "time"

type Container struct {
	IP          string
	IsReachable bool
	LastPing    time.Time
}
