package restserver

import (
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"encoding/json"
	"strconv"
	"time"
	"github.com/dkazakevich/redis/pkg/cache"
)

type RestServer struct {
	router *mux.Router
	cache  *cache.Cache
}

const (
	baseUrl = "/api/v1/"

	itemNotFoundMsg          = "Cache item not found"
	invalidIndexParamMsg     = "Invalid `listIndex` param. Number required"
	valueNotListMsg          = "Indicated value is not list"
	valueNotDictMsg          = "Indicated value is not dictionary"
	dictItemNotFoundMsg      = "Dictionary item not found"
	invalidPayloadRequestMsg = "Invalid payload request"
	invalidExpireValueMsg    = "Invalid expire value"

	timeoutSetMsg    = "The timeout was set"
	itemDeletedMsg   = "Cache item deleted"
	dataPersistedMsg = "Cache Data persisted"
	dataReloadedMsg  = "Cache Data reloaded"

	keyParam       = "key"
	dictKeyParam   = "dictKey"
	valueParam     = "value"
	expireParam    = "expire"
	listIndexParam = "listIndex"
	errorParam     = "error"
	messageParam   = "message"
)

func (rs *RestServer) Initialize() {
	rs.cache = cache.NewCache(10, time.Second, 10)
	rs.cache.Reload() //load stored cache data

	rs.router = mux.NewRouter()
	rs.initializeRoutes()
}

func (rs *RestServer) Run(port string) {
	log.Fatal(http.ListenAndServe(":"+port, rs.router))
}

func (rs *RestServer) initializeRoutes() {
	rs.router.HandleFunc(baseUrl+"ping", rs.ping).Methods(http.MethodGet)
	rs.router.HandleFunc(baseUrl+"keys", rs.getKeys).Methods(http.MethodGet)
	rs.router.HandleFunc(baseUrl+"values/{key}", rs.getValue).Methods(http.MethodGet)
	rs.router.HandleFunc(baseUrl+"ttl/{key}", rs.getTtl).Methods(http.MethodGet)
	rs.router.HandleFunc(baseUrl+"values/{key}", rs.putValue).Methods(http.MethodPut)
	rs.router.HandleFunc(baseUrl+"expire/{key}", rs.expire).Methods(http.MethodPut)
	rs.router.HandleFunc(baseUrl+"values/{key}", rs.remove).Methods(http.MethodDelete)
	rs.router.HandleFunc(baseUrl+"persist", rs.persist).Methods(http.MethodPost)
	rs.router.HandleFunc(baseUrl+"reload", rs.reload).Methods(http.MethodPost)
}

func (rs *RestServer) ping(w http.ResponseWriter, r *http.Request) {
	respondWithValue(w, http.StatusOK, "ping")
}

//get cache value by key
func (rs *RestServer) getValue(w http.ResponseWriter, r *http.Request) {
	var result interface{}
	vars := mux.Vars(r)
	key := vars[keyParam]

	//get cache value
	value, exists := rs.cache.Get(key)
	if exists == false {
		respondWithError(w, http.StatusNotFound, itemNotFoundMsg)
		return
	} else {
		result = value;

		//get i item from cache list value
		if listIndex := r.FormValue(listIndexParam); listIndex != "" {
			index, err := strconv.Atoi(listIndex)
			if err != nil {
				respondWithError(w, http.StatusBadRequest, invalidIndexParamMsg)
				return
			} else {
				v, exists := value.([]interface{})
				if exists == false {
					respondWithError(w, http.StatusBadRequest, valueNotListMsg)
					return
				} else {
					result = v[index]
				}
			}
		} else {
			//get item by key from dict cache value
			if dictKey := r.FormValue(dictKeyParam); dictKey != "" {
				v, exists := value.(map[string]interface{})
				if exists == false {
					respondWithError(w, http.StatusBadRequest, valueNotDictMsg)
					return
				} else {
					dictValue, exists := v[dictKey]
					if exists == false {
						respondWithError(w, http.StatusNotFound, dictItemNotFoundMsg)
						return
					} else {
						result = dictValue
					}
				}
			}
		}
	}
	respondWithValue(w, http.StatusOK, result)
}

//put key-value pair into the cache
func (rs *RestServer) putValue(w http.ResponseWriter, r *http.Request) {
	//get value from the request body
	var value interface{}
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&value); err != nil {
		respondWithError(w, http.StatusBadRequest, invalidPayloadRequestMsg)
		return
	}
	defer r.Body.Close()

	//indicated expire param
	expire := -1
	if exp := r.FormValue(expireParam); exp != "" {
		var err error
		expire, err = strconv.Atoi(exp)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, invalidExpireValueMsg)
			return
		}
	}

	key := mux.Vars(r)[keyParam]
	rs.cache.Put(key, value, expire)
	result, _ := rs.cache.Get(key)
	respondWithValue(w, http.StatusOK, result)
}

//set a timeout on key in seconds
func (rs *RestServer) expire(w http.ResponseWriter, r *http.Request) {
	var expire int
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&expire); err != nil {
		respondWithError(w, http.StatusBadRequest, invalidPayloadRequestMsg)
		return
	}
	defer r.Body.Close()

	result := rs.cache.Expire(mux.Vars(r)[keyParam], expire)
	if result == false {
		respondWithError(w, http.StatusNotFound, itemNotFoundMsg)
	} else {
		respondWithMessage(w, http.StatusOK, timeoutSetMsg)
	}
}

//returns the remaining time to live of a key that has a timeout
func (rs *RestServer) getTtl(w http.ResponseWriter, r *http.Request) {
	result := rs.cache.GetTtl(mux.Vars(r)[keyParam])
	if result == -2 {
		respondWithError(w, http.StatusNotFound, itemNotFoundMsg)
	} else {
		respondWithValue(w, http.StatusOK, result)
	}
}

//get list of cache keys
func (rs *RestServer) getKeys(w http.ResponseWriter, r *http.Request) {
	respondWithJSON(w, http.StatusOK, map[string][]string{valueParam: rs.cache.GetKeys()})
}

//remove key-value pair from the cache
func (rs *RestServer) remove(w http.ResponseWriter, r *http.Request) {
	key := mux.Vars(r)[keyParam]
	_, exists := rs.cache.Get(key)
	if exists == true {
		rs.cache.Remove(key)
		respondWithMessage(w, http.StatusOK, itemDeletedMsg)
	} else {
		respondWithError(w, http.StatusNotFound, itemNotFoundMsg)
	}
}

//persist cache data
func (rs *RestServer) persist(w http.ResponseWriter, r *http.Request) {
	err := rs.cache.Persist()
	if err == nil {
		respondWithMessage(w, http.StatusOK, dataPersistedMsg)
	} else {
		respondWithError(w, http.StatusInternalServerError, err.Error())
	}
}

//reload persisted data
func (rs *RestServer) reload(w http.ResponseWriter, r *http.Request) {
	err := rs.cache.Reload()
	if err == nil {
		respondWithMessage(w, http.StatusOK, dataReloadedMsg)
	} else {
		respondWithError(w, http.StatusInternalServerError, err.Error())
	}
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{errorParam: message})
}

func respondWithMessage(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{messageParam: message})
}

func respondWithValue(w http.ResponseWriter, code int, value interface{}) {
	respondWithJSON(w, code, map[string]interface{}{valueParam: value})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}
