package core

import (
	"log"
	"time"
)

func CleanupExpiredKeys() {
	for {
		frac := expireSample()

		if frac < 0.25 {
			break
		}
	}
	log.Printf("Cleanup expired keys. Total keys: %d", len(store))
}

func expireSample() float32 {
	var limit int = 20
	var count int = 0

	for key, obj := range store {
		if obj.ExpireAt != -1 {
			limit--
			if obj.ExpireAt < time.Now().UnixMilli() {
				delete(store, key)
				count++
			}
		}
		if limit <= 0 {
			break
		}
	}

	return float32(count) / float32(20.0)
}
