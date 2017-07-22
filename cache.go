package main

import (
	"time"
	"sync"
	"fmt"
	"encoding/json"
	"io/ioutil"
)

const CacheDataFileName = "cacheData.json"

type Cache struct {
	sync.RWMutex
	Data  map[string]interface{} `json:"data"`
	Ttl   map[string]time.Time  `json:"ttl"`
	timer *time.Timer
}

func newCache() *Cache {
	return &Cache{Data: make(map[string]interface{}), Ttl: make(map[string]time.Time)}
}

func (c *Cache) persist() error {
	c.RLock()
	jsonData, err := json.Marshal(c)
	if err == nil {
		err = ioutil.WriteFile(CacheDataFileName, jsonData, 0644)
	}
	c.RUnlock()

	return err
}

func (c *Cache) reload() error {
	c.Lock()
	jsonData, err := ioutil.ReadFile(CacheDataFileName)
	if err == nil {
		err = json.Unmarshal(jsonData, &c)
		if err == nil {
			c.checkTtl()
		}
	}
	c.Unlock()

	return err
}

func (c *Cache) get(key string) (interface{}, bool) {
	c.RLock()
	value, ok := c.Data[key]
	c.RUnlock()

	return value, ok
}

func (c *Cache) put(key string, value interface{}, expire int) {
	c.Lock()
	//delete old Ttl if exists
	delete(c.Ttl, key)

	c.Data[key] = value

	if expire > 0 {
		c.Ttl[key] = time.Now().Add(time.Second*time.Duration(expire))
		c.checkTtl()
	}
	c.Unlock()

	//fmt.Printf("Put `%v` key with '%v' value\n", key, value)
}

func (c *Cache) expire(key string, expire int) bool {
	c.Lock()
	//delete old Ttl if exists
	delete(c.Ttl, key)

	value := c.Data[key]

	result := false
	if (expire > 0) && (value != nil) {
		c.Ttl[key] = time.Now().Add(time.Second*time.Duration(expire))
		c.checkTtl()
		result = true
	}
	c.Unlock()

	return result
}

func (c *Cache) remove(key string) {
	c.Lock()
	delete(c.Data, key)
	delete(c.Ttl, key)
	c.Unlock()
}

func (c *Cache) getKeys() []string {
	c.RLock()
	keys := make([]string, 0, len(c.Data))
	for k := range c.Data {
		keys = append(keys, k)
	}
	c.RUnlock()

	return keys
}

func (c *Cache) getTtl(key string) int {
	c.RLock()
	value, ok := c.Ttl[key]
	c.RUnlock()

	var ttlValue int = -1
	if ok == true {
		ttlValue = int(time.Until(value).Seconds())
	}

	return ttlValue
}

//check all cache values with Ttl, remove expired, find minimum Ttl and set timer for it
func (c *Cache) checkTtl() {
	if c.timer != nil {
		//stop the current timer
		c.timer.Stop()
	}

	var minTtl time.Duration = -1
	for key, expire := range c.Ttl {
		ttlValue := time.Until(expire)
		if ttlValue > 0 {
			if minTtl < 0 {
				minTtl = ttlValue
			}
			if ttlValue < minTtl {
				minTtl = ttlValue
			}
		} else {
			//delete from Ttl with current lock
			delete(c.Ttl, key)
			delete(c.Data, key)
			fmt.Printf("`%v` removed\n", key)
		}
	}

	if minTtl > 0 {
		if c.timer == nil {
			c.timer = time.NewTimer(minTtl)
		} else {
			c.timer.Reset(minTtl)
		}
		fmt.Printf("Timer reset with `%v` duration\n", minTtl.Seconds())

		go func() {
			<-c.timer.C
			c.checkTtl()
		}()
	}
}