package drive

import "time"

type Drive interface {
	// Init Initialization processing
	Init() error
	// Acquire key traffic
	Acquire(key string, ttl time.Duration) (bool, error)
	// Clear Expired Data
	Clear() error
}
