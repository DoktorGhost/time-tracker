package main

import "time-tracker/internal/server"

func main() {
	err := server.StartServer()
	if err != nil {
		panic(err)
	}
}
