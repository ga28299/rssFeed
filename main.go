package main

import (
	"fmt"
	"log"
	"os"
)

func main() {
	fmt.Println("hello world!")

	portString := os.Getenv("PORT")
	if portString == "" {
		log.Fatal("No port set. Check env file")
	}
	fmt.Println("port is set to", portString)
}
