package main

import (
	"docfiff/src/api"
)

func main() {
	server := api.New(":5000")
	server.HandleFunc("/compare", server.Compare)
	server.HandleFunc("/findall", server.GetDiffs)
	server.HandleFunc("/findone", server.GetDiffBy)
	server.Start()
}
