package main

import (
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"encoding/json"
	"strconv"
)

type App struct {
	Router *mux.Router
}

var c *Cache

func (a *App) Initialize() {

	c = newCache()
	//load stored cache data
	c.reload()

	a.Router = mux.NewRouter()
	a.initializeRoutes()
}

func (a *App) Run(addr string) {
	log.Fatal(http.ListenAndServe(addr, a.Router))
}

func (a *App) initializeRoutes() {

	a.Router.HandleFunc(baseUrl + "ping", a.ping).Methods(http.MethodGet)
	a.Router.HandleFunc(baseUrl + "keys", a.getKeys).Methods(http.MethodGet)
	a.Router.HandleFunc(baseUrl + "values/{key}", a.getValue).Methods(http.MethodGet)
	a.Router.HandleFunc(baseUrl + "ttl/{key}", a.getTtl).Methods(http.MethodGet)
	a.Router.HandleFunc(baseUrl + "values/{key}", a.putValue).Methods(http.MethodPut)
	a.Router.HandleFunc(baseUrl + "expire/{key}", a.expire).Methods(http.MethodPut)
	a.Router.HandleFunc(baseUrl + "values/{key}", a.remove).Methods(http.MethodDelete)
	a.Router.HandleFunc(baseUrl + "persist", a.persist).Methods(http.MethodPost)
	a.Router.HandleFunc(baseUrl + "reload", a.reload).Methods(http.MethodPost)
}

func (a *App) ping(w http.ResponseWriter, r *http.Request) {

	respondWithValue(w, http.StatusOK, "ping")
}

//get cache value by key
func (a *App) getValue(w http.ResponseWriter, r *http.Request) {
	var result interface{}
	vars := mux.Vars(r)
	key := vars[keyParam]

	//get cache value
	value, ok := c.get(key)

	if ok == false {
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
				v, ok := value.([]interface{})
				if ok == false {
					respondWithError(w, http.StatusBadRequest, valueNotListMsg)
					return
				} else {
					result = v[index]
				}
			}
		} else {
			//get item by key from dict cache value
			if dictKey := r.FormValue(dictKeyParam); dictKey != "" {
				v, ok := value.(map[string]interface{})
				if ok == false {
					respondWithError(w, http.StatusBadRequest, valueNotDictMsg)
					return
				} else {
					dictValue, ok := v[dictKey]
					if ok == false {
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
func (a *App) putValue(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars[keyParam]

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

	c.put(key, value, expire)

	result, _ := c.get(key)
	respondWithValue(w, http.StatusOK, result)
}

//set a timeout on key in seconds
func (a *App) expire(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars[keyParam]

	var expire uint
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&expire); err != nil {
		respondWithError(w, http.StatusBadRequest, invalidPayloadRequestMsg)
		return
	}
	defer r.Body.Close()

	result := c.expire(key, expire)

	if result == false {
		respondWithError(w, http.StatusNotFound, itemNotFoundMsg)
	} else {
		respondWithMessage(w, http.StatusOK, timeoutSetMsg)
	}
}

//returns the remaining time to live of a key that has a timeout
func (a *App) getTtl(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars[keyParam]

	result := c.getTtl(key)
	if result == -1 {
		respondWithError(w, http.StatusNotFound, itemNotFoundMsg)
	} else {
		respondWithValue(w, http.StatusOK, result)
	}
}

//get list of cache keys
func (a *App) getKeys(w http.ResponseWriter, r *http.Request) {

	respondWithJSON(w, http.StatusOK, map[string][]string{valueParam: c.getKeys()})
}

//remove key-value pair from the cache
func (a *App) remove(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars[keyParam]

	_, ok := c.get(key)
	if ok == true {
		c.remove(key)
		respondWithMessage(w, http.StatusOK, itemDeletedMsg)
	} else {
		respondWithError(w, http.StatusNotFound, itemNotFoundMsg)
	}
}

//persist cache data
func (a *App) persist(w http.ResponseWriter, r *http.Request) {
	err := c.persist()
	if err == nil {
		respondWithMessage(w, http.StatusOK, dataPersistedMsg)
	} else {
		respondWithError(w, http.StatusInternalServerError, err.Error())
	}
}

//reload persisted data
func (a *App) reload(w http.ResponseWriter, r *http.Request) {
	err := c.reload()
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