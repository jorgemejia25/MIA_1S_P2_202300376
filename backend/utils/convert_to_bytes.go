package utils

import (
	"errors"
)

func ConvertToBytes(size int, unit string) (int64, error) {
	switch unit {
	case "B":
		return int64(size), nil
	case "K":
		return int64(size) * 1024, nil
	case "M":
		return int64(size) * 1024 * 1024, nil
	default:
		return 0, errors.New("invalid unit")
	}
}
