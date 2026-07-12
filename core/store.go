package core

import (
	"server/config"
	"time"
)

var store map[string]*Obj // #genai: maps must be initialized before writes

func init() {
	store = make(map[string]*Obj)
}

func NewObject(value interface{}, expTimeMs int64, oType uint8, oEncoding uint8) *Obj {
	var expiresAt int64 = -1
	if expTimeMs > 0 {
		expiresAt = time.Now().UnixMilli() + expTimeMs
	}
	return &Obj{Value: value, ExpireAt: expiresAt, TypeEncoding: oType | oEncoding}
}

func Put(key string, obj *Obj) {
	if len(store) >= config.Config.MaxKeys {
		evict()
	}
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
	if obj.ExpireAt > 0 && obj.ExpireAt < time.Now().UnixMilli() {
		delete(store, key)
		return nil
	}
	return obj
}

func Delete(key string) bool {
	_, ok := store[key]
	if !ok {
		return false
	}
	delete(store, key)
	updateKeySpaceStat(0, "keys", len(store))
	return true
}
