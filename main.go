package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

type Configuration struct {
	ServerPort string `json:"serverPort"`
}

func main() {
	a := App{}
	a.Initialize()

	configuration := Configuration{}
	jsonData, err := ioutil.ReadFile("conf.json")
	if err == nil {
		err = json.Unmarshal(jsonData, &configuration)
	}

	if err != nil {
		fmt.Println("Can't load a configuration file: ", err)
	}

	port := serverPort
	if configuration.ServerPort != "" {
		port = configuration.ServerPort
	}

	fmt.Printf("Running server on the %v port", port)

	//start redis server
	a.Run(":" + port)
}