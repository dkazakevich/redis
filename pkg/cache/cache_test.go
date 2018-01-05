package cache

import (
	"testing"
	"time"
	"sync"
	"strconv"
	"log"
	"math/rand"
	"github.com/dkazakevich/redis/internal/testutil"
	"fmt"
)

const (
	performanceIterations = 640000
	stringKey             = "sixthMonth"
	stringValue           = "June"
	tempStringKey         = "tempString"
	tempStringValue       = "temp string value"
	dictKey               = "planets"
	listKey               = "cars"
)

var (
	c = NewCache(10, time.Second, 10, "testCacheData.json")

	dictValue = map[string]interface{}{"planet1": "Mercury", "planet2": "Venus", "planet3": "Earth"}
	listValue = []interface{}{"Toyota", "Opel", "Ford"}
)

func TestCachePutAndGetString(t *testing.T) {
	c.Put(stringKey, stringValue, 20)
	result, exists := c.Get(stringKey)
	testutil.AssertEquals(t, true, exists)
	testutil.AssertEquals(t, stringValue, result)
	ttl := c.GetTtl(stringKey)
	fmt.Println(ttl)
	testutil.AssertEquals(t, true, (ttl > 0) && (ttl <= 20))
}

func TestCachePutAndGetDict(t *testing.T) {
	dictValue["planetMix"] = []interface{}{"Jupiter", "Saturn", "Mars"}
	c.Put(dictKey, dictValue, -1)
	result, exists := c.Get(dictKey)
	testutil.AssertEquals(t, true, exists)
	testutil.AssertEqualsDeep(t, dictValue["planetMix"], result.(map[string]interface{})["planetMix"])
	testutil.AssertEquals(t, -1, c.GetTtl(dictKey))
}

func TestCachePutAndGetList(t *testing.T) {
	c.Put(listKey, listValue, -1)
	result, exists := c.Get(listKey)
	testutil.AssertEquals(t, true, exists)
	testutil.AssertEquals(t, listValue[1], result.([]interface{})[1])
	testutil.AssertEquals(t, -1, c.GetTtl(listKey))
}

func TestCacheKeys(t *testing.T) {
	keys := c.GetKeys()
	testutil.AssertEquals(t, 3, len(keys))
}

func TestCacheDelete(t *testing.T) {
	c.Remove(listKey)
	_, exists := c.Get(listKey)
	testutil.AssertEquals(t, false, exists)
	testutil.AssertEquals(t, -2, c.GetTtl(listKey))
}

func TestCacheExpireAndCheckTtl(t *testing.T) {
	c.Put(tempStringKey, tempStringValue, -1)
	testutil.AssertEquals(t, -1, c.GetTtl(tempStringKey))
	c.Expire(tempStringKey, 2)
	ttl := c.GetTtl(tempStringKey)
	testutil.AssertEquals(t, true, (ttl > 0) && (ttl <= 2))

	time.Sleep(3 * time.Second)
	testutil.AssertEquals(t, -2, c.GetTtl(tempStringKey))
	_, exists := c.Get(tempStringKey)
	testutil.AssertEquals(t, false, exists)
}

func TestCachePersist(t *testing.T) {
	keys := c.GetKeys()
	err := c.Persist()
	testutil.AssertEquals(t, nil, err)

	c.Clear()
	testutil.AssertEquals(t, 0, len(c.GetKeys()))

	err = c.Reload()
	testutil.AssertEquals(t, nil, err)
	testutil.AssertEquals(t, len(keys), len(c.GetKeys()))
}

func TestPutThenGetPerformance(t *testing.T) {
	data := make([]string, performanceIterations)
	for i := range data {
		data[i] = stringKey + /*strconv.Itoa(i)//*/ strconv.Itoa(rand.Intn(performanceIterations))
	}

	start := time.Now()
	var wg sync.WaitGroup
	wg.Add(performanceIterations)
	for i := range data {
		value := data[i]
		go func() {
			defer wg.Done()
			c.Put(value, value, rand.Intn(30))
		}()
	}
	wg.Wait()
	log.Printf("%v times put took %v", performanceIterations, time.Since(start))

	start = time.Now()
	wg.Add(performanceIterations)
	for i := range data {
		value := data[i]
		go func() {
			defer wg.Done()
			c.Get(value)
		}()
	}
	wg.Wait()
	log.Printf("%v times get took %v", performanceIterations, time.Since(start))

	keys := c.GetKeys()
	log.Printf("Keys size: %v", len(keys))
	time.Sleep(10 * time.Second)
	keys = c.GetKeys()
	log.Printf("Keys size: %v", len(keys))

	for i := range keys {
		c.Remove(keys[i])
	}
}

func TestPutAndGetPerformance(t *testing.T) {
	data := make([]string, performanceIterations)
	for i := range data {
		data[i] = stringKey + /*strconv.Itoa(i)//*/ strconv.Itoa(rand.Intn(performanceIterations))
	}

	start := time.Now()
	var wg sync.WaitGroup
	wg.Add(2 * performanceIterations)
	for i := range data {
		value := data[i]
		go func() {
			defer wg.Done()
			c.Put(value, value, rand.Intn(30))
		}()
		go func() {
			defer wg.Done()
			c.Get(stringKey + strconv.Itoa(rand.Intn(performanceIterations)))
		}()
	}
	wg.Wait()
	log.Printf("%v times put and get took %v", performanceIterations, time.Since(start))

	keys := c.GetKeys()
	log.Printf("Keys size: %v", len(keys))
	time.Sleep(10 * time.Second)
	keys = c.GetKeys()
	log.Printf("Keys size: %v", len(keys))

	for i := range keys {
		c.Remove(keys[i])
	}
}
