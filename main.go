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
	jsonData, err := ioutil.ReadFile(configFile)
	if err == nil {
		err = json.Unmarshal(jsonData, &configuration)
		fmt.Println("Configuration data loaded from file.")
	} else {
		fmt.Println("Can't load app configuration file: ", err)
	}

	port := defaultServerPort
	if configuration.ServerPort != "" {
		port = configuration.ServerPort
	}
	fmt.Printf("Running server on the %v port", port)
	a.Run(":" + port)	//start redis server
}