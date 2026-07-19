package core

import (
	"log"
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
	const sampleSize = 20
	checked := 0
	expired := 0

	for key, obj := range store {
		if checked >= sampleSize {
			break
		}
		checked++
		if isExpired(obj) {
			Delete(key) // #genai: only remove keys that actually expired
			expired++
		}
	}

	if checked == 0 {
		return 0
	}
	return float32(expired) / float32(checked)
}
