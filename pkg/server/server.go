package server

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"go1f/pkg/api"
)

func Run() error {

	port := 7540

	envPort := os.Getenv("TODO_PORT")
	if len(envPort) > 0 {
		if eport, err := strconv.ParseInt(envPort, 10, 32); err == nil {
			port = int(eport)
		}
	}

	log.Printf("Server port: %d\n", port)
	http.Handle("/", http.FileServer(http.Dir("web")))
	api.Init()
	return http.ListenAndServe(fmt.Sprintf(":%d", port), nil)

}
