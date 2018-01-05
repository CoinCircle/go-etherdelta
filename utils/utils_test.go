package utils

import (
	"testing"
)

func TestRandInt(t *testing.T) {
	n := RandInt(5, 10)

	if !(n >= 5 && n <= 10) {
		t.Errorf("Expected rand in to between 4 < x < 11, got %v", n)
	}
}

func TestKeyExists(t *testing.T) {

}
