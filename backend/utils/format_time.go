package utils

import (
	"time"
)

func FormatTime(t time.Time) float32 {
	return float32(t.Unix())
}
