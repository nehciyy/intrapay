package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"

	"github.com/nehciyy/intrapay.git/internal/api"
	"github.com/nehciyy/intrapay.git/internal/db"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	database, err := db.InitDB()
	if err != nil {
		log.Fatal(err)
	}

	router := mux.NewRouter()
	server := &api.Server{DB: database}

	router.HandleFunc("/accounts", server.CreateAccount).Methods("POST")
	router.HandleFunc("/accounts/{id}", server.GetAccount).Methods("GET")
	router.HandleFunc("/transactions", server.CreateTransaction).Methods("POST")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Println("intrapay server is running on port", port)
	log.Fatal(http.ListenAndServe(":"+port, router))
}
