package main

import (
	"testing"
	"reflect"
	"time"
)

var cache *Cache = NewCache("testCacheData.json")

func TestCachePutAndGetString(t *testing.T) {
	cache.put(stringKey, stringValue, 20)
	result, ok := cache.get(stringKey)
	assertEquals(t, true, ok)
	assertEquals(t, stringValue, result)
	assertEquals(t, true, cache.getTtl(stringKey) > 0)
}

func TestCachePutAndGetDict(t *testing.T) {
	dictValue["planetMix"] = []interface{}{"Jupiter", "Saturn", "Mars"}
	cache.put(dictKey, dictValue, -1)
	result, ok := cache.get(dictKey)
	assertEquals(t, true, ok)
	assertEqualsDeep(t, dictValue["planetMix"], result.(map[string]interface{})["planetMix"])
}

func TestCachePutAndGetList(t *testing.T) {
	cache.put(listKey, listValue, -1)
	result, ok := cache.get(listKey)
	assertEquals(t, true, ok)
	assertEquals(t, listValue[1], result.([]interface{})[1])
}

func TestCacheKeys(t *testing.T) {
	keys := cache.getKeys()
	assertEquals(t, 3, len(keys))
}

func TestCacheDelete(t *testing.T) {
	cache.remove(listKey)
	_, ok := cache.get(listKey)
	assertEquals(t, false, ok)
}

func TestCacheExpireAndCheckTtl(t *testing.T) {
	cache.put(tempStringKey, tempStringValue, -1)
	cache.expire(tempStringKey, 2)
	assertEquals(t, true, cache.getTtl(tempStringKey) > 0)

	time.Sleep(3 * time.Second)
	assertEquals(t, -1, cache.getTtl(tempStringKey))
	_, ok := cache.get(tempStringKey)
	assertEquals(t, false, ok)
}

func TestCachePersist(t *testing.T) {
	keys := cache.getKeys()
	err := cache.persist()
	assertEquals(t, nil, err)

	cache.clear()
	assertEquals(t, 0, len(cache.getKeys()))

	err = cache.reload()
	assertEquals(t, nil, err)
	assertEquals(t, len(keys), len(cache.getKeys()))
}

func assertEqualsDeep(t *testing.T, expected interface{}, actual interface{}) {
	if !reflect.DeepEqual(expected, actual) {
		t.Fatalf("Expected: '%v'. Actual: '%v'", expected, actual)
	}
}