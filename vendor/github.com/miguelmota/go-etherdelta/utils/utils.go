package utils

import (
	"math/rand"
	"time"
)

// RandInt Generate a random integer between a range
func RandInt(min int, max int) int {
	return min + rand.Intn(max-min)
}

// KeyExists Check is a key exists in an interface
func KeyExists(decoded map[string]interface{}, key string) bool {
	val, ok := decoded[key]
	return ok && val != nil
}

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}
