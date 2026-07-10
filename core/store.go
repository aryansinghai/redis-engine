package core

import (
	"server/config"
	"time"
)

var store = make(map[string]*Object) // #genai: maps must be initialized before writes

type Object struct {
	Value     interface{}
	ExpiresAt int64
}

func NewObject(value interface{}, expTimeMs int64) *Object {
	var expiresAt int64 = -1
	if expTimeMs != -1 {
		expiresAt = time.Now().UnixMilli() + expTimeMs
	}
	return &Object{Value: value, ExpiresAt: expiresAt}
}

func Put(key string, obj *Object) {
	if len(store) >= config.Config.MaxKeys {
		evict()
	}
	store[key] = obj
}

func Get(key string) *Object {
	obj, ok := store[key]
	if !ok {
		return nil
	}
	if obj.ExpiresAt != -1 && obj.ExpiresAt < time.Now().UnixMilli() {
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
	return true
}
