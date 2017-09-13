package main

import (
	"time"
	"sync"
	"fmt"
	"encoding/json"
	"io/ioutil"
)

type Cache struct {
	Data  Data
	timer timer
	dataFileName string
}

type Data struct {
	sync.RWMutex
	Map map[string]interface{} `json:"map"` //stores key-value pair data
	Ttl map[string]time.Time   `json:"ttl"` //stores key and expire time information
}

type timer struct {
	sync.RWMutex
	timer *time.Timer
}

func NewCache(dataFileName ...string) *Cache {
	cache := &Cache{Data: Data{Map: make(map[string]interface{}), Ttl: make(map[string]time.Time)}}
	if len(dataFileName) > 0 {
		cache.dataFileName = dataFileName[0]
	} else {
		cache.dataFileName = cacheDataFileName;
	}
	return cache
}

//persist cache data to disk
func (c *Cache) persist() error {
	c.Data.RLock()
	jsonData, err := json.Marshal(c)
	c.Data.RUnlock()

	if err == nil {
		err = ioutil.WriteFile(c.dataFileName, jsonData, 0644)
	}
	return err
}

//reload data from disk to cache
func (c *Cache) reload() error {
	jsonData, err := ioutil.ReadFile(c.dataFileName)
	if err == nil {
		c.Data.Lock()
		err = json.Unmarshal(jsonData, &c)
		c.Data.Unlock()

		if err == nil {
			go c.checkTtl()
		}
	}
	return err
}

//clear cache data
func (c *Cache) clear() {
	c.Data.RLock()
	c.Data.Map = make(map[string]interface{})
	for i := range c.Data.Map { delete(c.Data.Map, i) }
	for i := range c.Data.Ttl { delete(c.Data.Ttl, i) }
	c.Data.RUnlock()

	c.timer.Lock()
	if c.timer.timer != nil {
		c.timer.timer.Stop()
	}
	c.timer.Unlock()
}

//get cache value by key
func (c *Cache) get(key string) (interface{}, bool) {
	c.Data.RLock()
	defer c.Data.RUnlock()
	value, ok := c.Data.Map[key]
	return value, ok
}

//put key-value pair into the cache
func (c *Cache) put(key string, value interface{}, expire int) {
	c.Data.Lock()
	defer c.Data.Unlock()
	delete(c.Data.Ttl, key) //delete old Ttl if exists
	c.Data.Map[key] = value

	if expire > 0 {
		c.Data.Ttl[key] = time.Now().Add(time.Second * time.Duration(expire))
		go c.checkTtl()
	}
}

//set app timeout on key in seconds
func (c *Cache) expire(key string, expire int) bool {
	result := false
	c.Data.Lock()
	defer c.Data.Unlock()

	delete(c.Data.Ttl, key) //delete old Ttl if exists
	_, ok := c.Data.Map[key]

	if (expire > 0) && (ok == true) {
		c.Data.Ttl[key] = time.Now().Add(time.Second * time.Duration(expire))
		result = true
		go c.checkTtl()
	}
	return result
}

//remove key-value pair from the cache
func (c *Cache) remove(key string) {
	c.Data.Lock()
	defer c.Data.Unlock()
	delete(c.Data.Map, key)
	delete(c.Data.Ttl, key)
}

//get list of cache keys
func (c *Cache) getKeys() []string {
	c.Data.RLock()
	defer c.Data.RUnlock()
	keys := make([]string, 0, len(c.Data.Map))
	for k := range c.Data.Map {
		keys = append(keys, k)
	}
	return keys
}

//returns the remaining time to live of app key that has app timeout
//returns -1 for app key that hasn't app timeout
func (c *Cache) getTtl(key string) int {
	c.Data.RLock()
	value, ok := c.Data.Ttl[key]
	c.Data.RUnlock()

	var ttlValue int = -1
	if ok == true {
		ttlValue = int(time.Until(value).Seconds())
	}
	return ttlValue
}

//check all cache values with ttl, remove expired, find minimum ttl and set timer for it
func (c *Cache) checkTtl() {
	var minTtl time.Duration
	c.Data.Lock()
	for key, expire := range c.Data.Ttl {
		ttlValue := time.Until(expire)
		if ttlValue > 0 {
			if (minTtl == 0) || (ttlValue < minTtl) {
				minTtl = ttlValue
			}
		} else {
			//delete expired key-value pair with current lock
			delete(c.Data.Ttl, key)
			delete(c.Data.Map, key)
			fmt.Printf("`%v` removed\n", key)
		}
	}
	c.Data.Unlock()

	if minTtl > 0 {
		c.timer.Lock()
		//reset timer for the minTtl
		if c.timer.timer == nil {
			c.timer.timer = time.NewTimer(minTtl)
		} else {
			c.timer.timer.Reset(minTtl)
		}
		c.timer.Unlock()
		fmt.Printf("Timer reset with `%v`s duration\n", minTtl.Seconds())

		go func() {
			<-c.timer.timer.C
			c.checkTtl()
		}()
	}
}