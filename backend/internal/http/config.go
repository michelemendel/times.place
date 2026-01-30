package http

import (
	"os"
	"strconv"
)

// FreeTierMaxVenues returns the max venues allowed per owner (free tier). Default 2.
// Configure via FREE_TIER_MAX_VENUES environment variable.
func FreeTierMaxVenues() int64 {
	s := os.Getenv("FREE_TIER_MAX_VENUES")
	if s == "" {
		return 2
	}
	n, err := strconv.ParseInt(s, 10, 64)
	if err != nil || n < 1 {
		return 2
	}
	return n
}
