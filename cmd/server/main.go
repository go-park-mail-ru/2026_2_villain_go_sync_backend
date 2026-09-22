package main

import (
	"log"
	"net/http"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/server"
)

func main() {
	const addr = ":8080"

	log.Printf("server started on %s", addr)

	if err := http.ListenAndServe(addr, server.New()); err != nil {
		log.Fatal(err)
	}
}
