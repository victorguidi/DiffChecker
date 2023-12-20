package main

import (
	"docfiff/src/api"
)

func main() {
	api := api.New(":5000")
	api.HandleFunc("/compare", api.Compare)
	api.Start()
}
