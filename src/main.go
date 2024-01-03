package main

import (
	"docfiff/src/api"
)

func main() {
	server := api.New(":5000")
	server.HandleFunc("/compare", server.Compare)
  server.Start()
}
