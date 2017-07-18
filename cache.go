package main

import (
	"time"
	"sync"
	"fmt"
)

type Cache struct {
	sync.RWMutex
	data map[string]interface{}
	ttl map[string]time.Time
	timer *time.Timer
}

func newCache() *Cache {
	return &Cache{data: make(map[string]interface{}), ttl: make(map[string]time.Time)}
}

func (c *Cache) get(key string) (interface{}, bool) {
	c.RLock()
	value, ok := c.data[key]
	c.RUnlock()

	return value, ok
}

func (c *Cache) put(key string, value interface{}, expire int) {
	c.Lock()
	//delete old ttl if exists
	delete(c.ttl, key)

	c.data[key] = value

	if expire > 0 {
		c.ttl[key] = time.Now().Add(time.Second*time.Duration(expire))
		c.checkTtl()
	}
	c.Unlock()

	fmt.Printf("Put `%v` key with '%v' value\n", key, value)
}

func (c *Cache) remove(key string) {
	c.Lock()
	delete(c.data, key)
	delete(c.ttl, key)
	c.Unlock()
}

func (c *Cache) getKeys() []string {
	c.RLock()
	keys := make([]string, 0, len(c.data))
	for k := range c.data {
		keys = append(keys, k)
	}
	c.RUnlock()

	return keys
}

func (c *Cache) getTtl(key string) (int, bool) {
	c.RLock()
	value, ok := c.ttl[key]
	c.RUnlock()

	var ttlValue int = -1
	if ok == true {
		ttlValue = int(time.Until(value).Seconds())
	}

	return ttlValue, ok
}

//check all cache values with ttl, remove expired, find minimum ttl and set timer for it
func (c *Cache) checkTtl() {
	fmt.Println("checkTtl()")
	if c.timer != nil {
		//stop the current timer
		c.timer.Stop()
	}

	var minTtl time.Duration = -1
	for key, expire := range c.ttl {
		ttlValue := time.Until(expire)
		if ttlValue > 0 {
			if minTtl < 0 {
				minTtl = ttlValue
			}
			if ttlValue < minTtl {
				minTtl = ttlValue
			}
		} else {
			//delete from ttl with current lock
			delete(c.ttl, key)
			delete(c.data, key)
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