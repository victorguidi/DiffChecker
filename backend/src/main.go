package main

import (
	"docfiff/src/api"
)

// TODO: Add integration with the Database

func main() {
	server := api.New(":5000")
	server.HandleFunc("/compare", server.Compare)
	server.Start()
}
