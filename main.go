package main

import "log"

func main() {
	store, err := NewPostgresStorage()
	if err != nil {
		log.Fatal(err)
	}

	if err := store.Start(); err != nil {
		log.Fatal(err)
	}
	server := NewAPIServer(":8080", store)
	server.Start()
}