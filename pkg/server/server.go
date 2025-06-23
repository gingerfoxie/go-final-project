package server

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"go1f/pkg/api"

	"github.com/joho/godotenv"
)

func Run() error {

	port := 7540

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	} else {
		envPort := os.Getenv("SERVER_PORT")

		if len(envPort) > 0 {
			if eport, err := strconv.ParseInt(envPort, 10, 32); err == nil {
				port = int(eport)

			}
		}
	}

	log.Printf("Server port: %d\n", port)
	http.Handle("/", http.FileServer(http.Dir("web")))
	api.Init()
	return http.ListenAndServe(fmt.Sprintf(":%d", port), nil)

}
