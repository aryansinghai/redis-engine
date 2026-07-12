package core

import "server/config"

func evict() {
	switch config.Config.EVICTION_POLICY {
	case "allkeys-random":
		evictAllKeysRandom()
	default:
		evictFirst()
	}
}

func evictFirst() {
	for k := range store {
		delete(store, k)
		return
	}
}

func evictAllKeysRandom() {
	evictCount := int(float64(len(store)) * config.Config.EVICTION_RATIO)

	for k := range store {
		Delete(k)
		evictCount--
		if evictCount <= 0 {
			break
		}
	}
}
