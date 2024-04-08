package main

import (
	"net/http"

	"log"
	"test/controllers"
	"test/models"

	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()

	handler := controllers.New()

	server := &http.Server{
		Addr:    "0.0.0.0:8008",
		Handler: handler,
	}

	// fmt.Print("Running on port ", server.Addr)
	log.Println("Server started on :8008")

	models.ConnectDatabase()

	server.ListenAndServe()
}
