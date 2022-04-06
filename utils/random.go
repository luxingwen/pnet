package utils

import (
	mrand "math/rand"
	"time"
)

func RandDuration(average time.Duration, delta float64) time.Duration {
	return time.Duration((1 - delta + 2*mrand.Float64()*delta) * float64(average))
}
