package main

import (
	"log"

	"github.com/AlekseyZapadovnikov/DelayedNotifier/config"
	"github.com/AlekseyZapadovnikov/DelayedNotifier/internal/web"
	"github.com/joho/godotenv"
)

func main() {

	if err := godotenv.Load(); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	conf, err := config.LoadConfig()
	if err != nil {
		panic(err)
	}

	server := web.NewServer(conf, struct{}{})
	server.Start()
}