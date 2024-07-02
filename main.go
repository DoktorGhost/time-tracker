package main

import (
	"time-tracker/internal/server"
)

//	@title			Тайм-Трекер API
//	@version		1.0

// @host		localhost:8080

func main() {
	err := server.StartServer()
	if err != nil {
		panic(err)
	}
}
