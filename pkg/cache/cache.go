package cache

import (
	"time"
	"sync"
	"encoding/json"
	"io/ioutil"
	"math/rand"
	"hash/fnv"
)

const defaultCacheDataFileName = "cacheData.json"
const defultGarbageCollectionCheckItems = 20

type Cache struct {
	sync.RWMutex
	Data                        []*Data
	GarbageCollectionInterval   time.Duration
	GarbageCollectionCheckItems int
	DataFileName                string
}

type Data struct {
	sync.RWMutex
	Map        map[string]*Item `json:"map"`        //key-item (with value) pair data
	ExpireKeys map[int]string   `json:"expireKeys"` //expire keys with indexes for garbage collection to check random key expiration
}

type Item struct {
	Value          interface{} //key's value
	Expiration     int64       //expiration time
	ExpireKeyIndex int         //expire index in the ExpireKeys data structure for garbage collection for synchronization
}

//partitions - number of cache parts separated for parallel data access
//garbageCollectionInterval - period of time for periodical GC call
//garbageCollectionCheckItems - number of random keys that expiration will be checked for every partition in process of GC
//dataFileName - file name for persist cache data
func NewCache(partitions int, garbageCollectionInterval time.Duration, garbageCollectionCheckItems int,
	dataFileName ...string) *Cache {
	c := &Cache{Data: make([]*Data, partitions), GarbageCollectionInterval: garbageCollectionInterval}
	for i := range c.Data {
		c.Data[i] = &Data{Map: make(map[string]*Item), ExpireKeys: make(map[int]string)}
	}

	if garbageCollectionCheckItems > 0 {
		c.GarbageCollectionCheckItems = garbageCollectionCheckItems
	} else {
		c.GarbageCollectionCheckItems = defultGarbageCollectionCheckItems
	}

	if len(dataFileName) > 0 {
		c.DataFileName = dataFileName[0]
	} else {
		c.DataFileName = defaultCacheDataFileName;
	}

	if garbageCollectionInterval > 0 {
		go c.garbageCollection(garbageCollectionInterval)
	}

	return c
}

//persist cache data to disk
func (c *Cache) Persist() error {
	c.RLock()
	jsonData, err := json.Marshal(c)
	c.RUnlock()

	if err == nil {
		err = ioutil.WriteFile(c.DataFileName, jsonData, 0644)
	}
	return err
}

//reload data from disk to cache
func (c *Cache) Reload() error {
	jsonData, err := ioutil.ReadFile(c.DataFileName)
	if err == nil {
		c.Lock()
		defer c.Unlock()
		err = json.Unmarshal(jsonData, &c)
	}
	return err
}

//clear cache data
func (c *Cache) Clear() {
	for i := range c.Data {
		data := c.Data[i]
		data.Lock()
		data.Map = make(map[string]*Item)
		data.ExpireKeys = make(map[int]string)
		data.Unlock()
	}
}

//get cache partition based on key hash
func (c *Cache) getPartition(key string) *Data {
	hash := fnv.New32a()
	hash.Write([]byte(key))
	c.RLock()
	defer c.RUnlock()
	partition := hash.Sum32() % uint32(len(c.Data))
	return c.Data[partition]
}

//get cache value by key
func (c *Cache) Get(key string) (interface{}, bool) {
	var value interface{}
	data := c.getPartition(key)
	data.RLock()
	defer data.RUnlock()

	item, exists := c.get(data, key)
	if exists {
		value = item.Value
	}
	return value, exists
}

//check if the indicated key exists and delete if expired
func (c *Cache) get(data *Data, key string) (*Item, bool) {
	item, exists := data.Map[key]
	if exists && (item.Expiration != -1) && (item.Expiration-time.Now().UnixNano() < 0) {
		go c.Remove(key)
		item = nil
		exists = false
	}
	return item, exists
}

//put/replace key-value pair into the cache
func (c *Cache) Put(key string, value interface{}, expire int) {
	data := c.getPartition(key)
	data.Lock()
	defer data.Unlock()
	_, exists := data.Map[key]
	if exists {
		data.Map[key].Value = value
	} else {
		data.Map[key] = &Item{value, -1, -1}
	}
	c.updateExpiration(data, key, expire)
}

//set app timeout on key in seconds
func (c *Cache) Expire(key string, expire int) bool {
	data := c.getPartition(key)
	data.Lock()
	defer data.Unlock()
	_, exists := c.get(data, key)
	if exists {
		c.updateExpiration(data, key, expire)
	}
	return exists
}

func (c *Cache) updateExpiration(data *Data, key string, expire int) {
	item, exists := data.Map[key]
	if exists {
		if item.ExpireKeyIndex == -1 {
			if expire > 0 {
				newIndex := len(data.ExpireKeys)
				item.Expiration = time.Now().Add(time.Second * time.Duration(expire)).UnixNano()
				item.ExpireKeyIndex = newIndex
				data.ExpireKeys[newIndex] = key
			}
		} else {
			if expire > 0 {
				item.Expiration = time.Now().Add(time.Second * time.Duration(expire)).UnixNano()
			} else {
				c.removeExpiration(data, item.ExpireKeyIndex)
			}
		}
	}
}

//remove key-value pair from the cache
func (c *Cache) Remove(key string) {
	data := c.getPartition(key)
	data.Lock()
	defer data.Unlock()
	c.remove(data, key)
}

func (c *Cache) remove(data *Data, key string) {
	item, exists := data.Map[key]
	if exists {
		if item.ExpireKeyIndex != -1 {
			c.removeExpiration(data, item.ExpireKeyIndex)
		}
		delete(data.Map, key)
	}
}

func (c *Cache) removeExpiration(data *Data, expireKeyIndex int) {
	key := data.ExpireKeys[expireKeyIndex]
	data.Map[key].ExpireKeyIndex = -1

	//put lastExpireIndex into the removing index
	lastExpireIndex := len(data.ExpireKeys) - 1
	if expireKeyIndex != lastExpireIndex {
		lastKey := data.ExpireKeys[lastExpireIndex]
		data.Map[lastKey].ExpireKeyIndex = expireKeyIndex
		data.ExpireKeys[expireKeyIndex] = lastKey
	}
	delete(data.ExpireKeys, lastExpireIndex)
}

//get list of cache keys
func (c *Cache) GetKeys() []string {
	keys := make([]string, 0)
	c.RLock()
	defer c.RUnlock()
	for i := range c.Data {
		data := c.Data[i]
		data.RLock()
		for key, item := range data.Map {
			if (item.Expiration == -1) || (item.Expiration-time.Now().UnixNano() > 0) {
				keys = append(keys, key)
			}
		}
		data.RUnlock()
	}
	return keys
}

//returns the remaining time to live in seconds of a key that has a timeout
//returns -1 if the key exists but has no associated expire
//returns -2 if the key does not exist
func (c *Cache) GetTtl(key string) int {
	var result = -2
	data := c.getPartition(key)
	data.RLock()
	defer data.RUnlock()

	item, exists := c.get(data, key)
	if exists {
		if item.Expiration == -1 {
			result = -1
		} else {
			result = int((item.Expiration - time.Now().UnixNano()) / int64(time.Second))
		}
	}
	return result
}

//GC do the following steps every `garbageCollectionInterval` interval:
//1. Test `garbageCollectionCheckItems` number random keys from the every partition set of keys with an associated expire
//2. Delete all the keys found expired
//3. If more than 25% of keys were expired, start again from step 1
func (c *Cache) garbageCollection(garbageCollectionInterval time.Duration) {
	ticker := time.NewTicker(garbageCollectionInterval)
	for {
		select {
		case <-ticker.C:
			c.RLock()
			for i := range c.Data {
				data := c.Data[i]
				data.RLock()
				size := len(data.ExpireKeys)
				data.RUnlock()

				if size > 0 {
					runGCAgain := true
					for runGCAgain {
						data.Lock()
						maxIndex := len(data.ExpireKeys) - 1
						deletedCount := 0
						for i := 0; (i < c.GarbageCollectionCheckItems) && (maxIndex >= 0); i++ {
							//get random key
							randIndex := 0
							if maxIndex > 0 {
								randIndex = rand.Intn(maxIndex)
							}
							key := data.ExpireKeys[randIndex]
							item := data.Map[key]
							if (item.Expiration != -1) && (item.Expiration-time.Now().UnixNano() < 0) {
								deletedCount++
								c.remove(data, key)
								maxIndex--
							}
						}
						data.Unlock()
						//repeat while deleted > 25%
						if (deletedCount * 100 / c.GarbageCollectionCheckItems) <= 25 {
							runGCAgain = false
						}
					}
				}
			}
			c.RUnlock()
		}
	}
}
