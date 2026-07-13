package core

import "sort"

type PoolItem struct {
	Key            string
	LastAccessedAt uint32
}

type EvictionPool struct {
	pool   []*PoolItem
	keySet map[string]*PoolItem
}

const ePoolSizeMax = 16

var ePool *EvictionPool = newEvictionPool(ePoolSizeMax)

type ByIdleTime []*PoolItem

func (a ByIdleTime) Len() int {
	return len(a)
}

func (a ByIdleTime) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

func (a ByIdleTime) Less(i, j int) bool {
	return getIdleTime(a[i].LastAccessedAt) > getIdleTime(a[j].LastAccessedAt)
}

func (e *EvictionPool) Push(key string, lastAccessedAt uint32) {
	_, ok := e.keySet[key]
	if ok {
		return
	}
	item := &PoolItem{Key: key, LastAccessedAt: lastAccessedAt}
	if len(e.pool) < ePoolSizeMax {
		e.keySet[key] = item
		e.pool = append(e.pool, item)
		sort.Sort(ByIdleTime(e.pool))
	} else if lastAccessedAt > e.pool[len(e.pool)-1].LastAccessedAt {
		e.pool = e.pool[1:]
		e.keySet[key] = item
		e.pool = append(e.pool, item)
	}
}

func (e *EvictionPool) Pop() *PoolItem {
	if len(e.pool) == 0 {
		return nil
	}
	item := e.pool[0]
	e.pool = e.pool[1:]
	return item
}

func newEvictionPool(size int) *EvictionPool {
	return &EvictionPool{
		pool:   make([]*PoolItem, size),
		keySet: make(map[string]*PoolItem),
	}
}
