package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"github.com/dkazakevich/redis/pkg/restserver"
)

type configuration struct {
	ServerPort string `json:"serverPort"`
}

const (
	configFile        = "conf.json"
	defaultServerPort = "8080"
)

func main() {
	var rest restserver.RestServer
	rest.Initialize()

	configuration := configuration{}
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
	rest.Run(port) //start redis server
}
