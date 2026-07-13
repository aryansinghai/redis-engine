package core

import (
	"server/config"
	"time"
)

func getCurrentClock() uint32 {
	return uint32(time.Now().Unix()) & 0x00FFFFFF
}

func getIdleTime(lastAccessedAt uint32) uint32 {
	c := getCurrentClock()
	if c < lastAccessedAt {
		return c - lastAccessedAt
	}
	return c + 0x1000000 - lastAccessedAt
}

func populateEvictionPool() {
	sampleSize := 5
	for k := range store {
		ePool.Push(k, store[k].LastAccessedAt)
		sampleSize--
		if sampleSize <= 0 {
			break
		}
	}
}

func evict() {
	switch config.Config.EVICTION_POLICY {
	case "allkeys-random":
		evictAllKeysRandom()
	case "allkeys-lru":
		evictAllKeysLRU()
	default:
		evictFirst()
	}
}

func evictAllKeysLRU() {
	populateEvictionPool()
	evictCount := int16(config.Config.EVICTION_RATIO * float64(config.Config.MaxKeys))
	for i := 0; i < int(evictCount) && len(ePool.pool) > 0; i++ {
		item := ePool.Pop()
		if item == nil {
			return
		}
		Delete(item.Key)
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
