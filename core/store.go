package core

import (
	"server/config"
	"time"
)

var store map[string]*Obj // #genai: maps must be initialized before writes
var expiresAt map[*Obj]uint64

func init() {
	store = make(map[string]*Obj)
	expiresAt = make(map[*Obj]uint64)
}

func setExpiry(obj *Obj, expTimeMs int64) {
	expiresAt[obj] = uint64(time.Now().UnixMilli() + expTimeMs)
}

func isExpired(obj *Obj) bool {
	return expiresAt[obj] > 0 && expiresAt[obj] <= uint64(time.Now().UnixMilli())
}

func NewObject(value interface{}, expTimeMs int64, oType uint8, oEncoding uint8) *Obj {
	obj := &Obj{
		Value:          value,
		TypeEncoding:   oType | oEncoding,
		LastAccessedAt: getCurrentClock(),
	}
	if expTimeMs > 0 {
		setExpiry(obj, expTimeMs)
	}
	return obj
}

func Put(key string, obj *Obj) {
	if len(store) >= config.Config.MaxKeys {
		evict()
	}
	obj.LastAccessedAt = getCurrentClock()
	store[key] = obj
	if KeySpaceStat[0] == nil {
		KeySpaceStat[0] = make(map[string]int)
	}
	updateKeySpaceStat(0, "keys", len(store))
}

func Get(key string) *Obj {
	obj, ok := store[key]
	if !ok {
		return nil
	}
	if isExpired(obj) {
		delete(store, key)
		return nil
	}
	return obj
}

func Delete(key string) bool {
	obj, ok := store[key]
	if !ok {
		return false
	}
	delete(store, key)
	delete(expiresAt, obj)
	updateKeySpaceStat(0, "keys", len(store))
	return true
}
