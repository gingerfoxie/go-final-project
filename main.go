package main

import (
	"log"

	"github.com/joho/godotenv"

	"go1f/pkg/db"
	"go1f/pkg/server"
)

func main() {

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	log.Println("Database init (main)")
	err = db.Init()
	if err != nil {
		log.Printf("Init database error: %s", err.Error())
		return
	}
	defer db.Close()

	log.Println("Starting server")
	err = server.Run()

	if err != nil {
		log.Printf("Start server error: %s", err.Error())
		return
	}
}
