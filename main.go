package main

import (
	"fmt"
	"go1f/pkg/db"
	"go1f/pkg/server"
)

func main() {

	err := db.Init()
	fmt.Println("Database init  main")
	if err != nil {
		fmt.Printf("Init database error: %s", err.Error())
		return
	}

	err = server.Run()

	if err != nil {
		fmt.Printf("Start server error: %s", err.Error())
		return
	}
}
