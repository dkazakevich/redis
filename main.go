package main

func main() {
	a := App{}
	a.Initialize()

	//start redis server
	a.Run(":" + serverPort)
}