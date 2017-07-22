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
	c.reload()

	a.Router = mux.NewRouter()
	a.initializeRoutes()
}

func (a *App) Run(addr string) {
	log.Fatal(http.ListenAndServe(addr, a.Router))
}

func (a *App) initializeRoutes() {

	a.Router.HandleFunc("/api/v1/keys", a.getKeys).Methods("GET")
	a.Router.HandleFunc("/api/v1/values/{key}", a.getValue).Methods("GET")
	a.Router.HandleFunc("/api/v1/ttl/{key}", a.getTtl).Methods("GET")
	a.Router.HandleFunc("/api/v1/values/{key}", a.putValue).Methods("PUT")
	a.Router.HandleFunc("/api/v1/expire/{key}", a.expire).Methods("PUT")
	a.Router.HandleFunc("/api/v1/values/{key}", a.deleteValue).Methods("DELETE")
	a.Router.HandleFunc("/api/v1/persist", a.persist).Methods("POST")
	a.Router.HandleFunc("/api/v1/reload", a.reload).Methods("POST")
}

func (a *App) getValue(w http.ResponseWriter, r *http.Request) {
	var result interface{}
	vars := mux.Vars(r)
	key := vars["key"]

	//get cache value
	value, ok := c.get(key)

	if ok == false {
		respondWithError(w, http.StatusNotFound, "Cache item not found")
		return
	} else {
		result = value;
		//get i item from list cache value
		if posStr := r.FormValue("pos"); posStr != "" {
			pos, err := strconv.Atoi(posStr)
			if err != nil {
				respondWithError(w, http.StatusBadRequest, "Invalid `pos` param. Number required")
				return
			} else {
				v, ok := value.([]interface{})
				if ok == false {
					respondWithError(w, http.StatusBadRequest, "Indicated value is not list")
					return
				} else {
					result = v[pos]
				}
			}
		} else {
			//get item by key from dict cache value
			if dictKey := r.FormValue("key"); dictKey != "" {
				v, ok := value.(map[string]interface{})
				if ok == false {
					respondWithError(w, http.StatusBadRequest, "Indicated value is not dictionary")
					return
				} else {
					dictValue, ok := v[dictKey]
					if ok == false {
						respondWithError(w, http.StatusNotFound, "Dictionary item not found")
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

func (a *App) putValue(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["key"]

	var value interface{}
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&value); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	//indicated expire param
	expire := -1
	if exp := r.FormValue("expire"); exp != "" {
		var err error
		expire, err = strconv.Atoi(exp)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid value expire")
			return
		}
	}

	c.put(key, value, expire)

	result, _ := c.get(key)
	respondWithValue(w, http.StatusOK, result)
}

func (a *App) expire(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["key"]

	var expire int
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&expire); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	result := c.expire(key, expire)

	respondWithValue(w, http.StatusOK, result)
}

//get list of cache keys
func (a *App) getKeys(w http.ResponseWriter, r *http.Request) {

	respondWithJSON(w, http.StatusOK, map[string][]string{"keys": c.getKeys()})
}

func (a *App) deleteValue(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["key"]

	_, ok := c.get(key)
	if ok == true {
		c.remove(key)
		respondWithValue(w, http.StatusOK, "Cache item deleted")
	} else {
		respondWithError(w, http.StatusNotFound, "Cache item not found")
	}
}

func (a *App) persist(w http.ResponseWriter, r *http.Request) {
	err := c.persist()
	if err == nil {
		respondWithJSON(w, http.StatusOK, "Cache Data persisted")
	} else {
		respondWithError(w, http.StatusInternalServerError, err.Error())
	}
}

func (a *App) reload(w http.ResponseWriter, r *http.Request) {
	err := c.reload()
	if err == nil {
		respondWithJSON(w, http.StatusOK, "Cache Data reloaded")
	} else {
		respondWithError(w, http.StatusInternalServerError, err.Error())
	}
}

//get Ttl of cache value
func (a *App) getTtl(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["key"]

	respondWithValue(w, http.StatusOK, c.getTtl(key))
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

func respondWithValue(w http.ResponseWriter, code int, value interface{}) {
	respondWithJSON(w, code, map[string]interface{}{"value": value})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}