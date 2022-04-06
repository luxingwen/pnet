package utils

import "time"

func StopTimer(timer *time.Timer) {
	if !timer.Stop() {
		select {
		case <-timer.C:
		default:
		}
	}
}

func ResetTimer(timer *time.Timer, duration time.Duration) {
	StopTimer(timer)
	timer.Reset(duration)
}
