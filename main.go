package main

import (
	"github.com/rs/zerolog/log"
)

func main() {
	store, err := NewPostgresStore()
	if err != nil {
		log.Fatal().AnErr("error", err).Str("message", "database error").Send()
	}

	if err := store.Init(); err != nil {
		log.Fatal().AnErr("error", err).Str("message", "database init error").Send()
	}
	server := NewApiServer(":3000", store)
	server.Run()
}
