package utils

import (
	"math/rand"
	"time"
)

func RandInt(min int, max int) int {
	return min + rand.Intn(max-min)
}

func KeyExists(decoded map[string]interface{}, key string) bool {
	val, ok := decoded[key]
	return ok && val != nil
}

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}
