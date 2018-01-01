package utils

import (
	"math/rand"
	"time"
)

func RandInt(min int, max int) int {
	return min + rand.Intn(max-min)
}

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}
