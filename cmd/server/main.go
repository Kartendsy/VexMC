package main

import (
	"vexmc/server"
)

func main() {
	srv := server.NewServer()
	srv.Start()
}
