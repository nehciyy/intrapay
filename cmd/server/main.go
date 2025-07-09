package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"

	"github.com/nehciyy/intrapay/internal/api"
	"github.com/nehciyy/intrapay/internal/db"
	"github.com/nehciyy/intrapay/internal/repository"
	"github.com/nehciyy/intrapay/internal/service"
)

func main() {
	// Load .env file
	if _, exists := os.LookupEnv("DATABASE_URL"); !exists {
        err := godotenv.Load()
        if err != nil {
            log.Println("Warning: no .env file found, proceeding without it")
        }
    }

	// Initialize database
	database, err := db.InitDB()
	if err != nil {
		log.Fatal(err)
	}

	// Create repositories
	accountRepo := repository.NewPostgresAccountRepository(database)
	transactionRepo := repository.NewPostgresTransactionRepository(database)

	// Pass both repos to the service
	svc := service.NewService(database, accountRepo, transactionRepo)

	// Initialize API server with DB and service layer
	server := &api.Server{
		Service: svc,
	}

	// Set up routes
	router := mux.NewRouter()
	router.HandleFunc("/accounts", server.CreateAccount).Methods("POST")
	router.HandleFunc("/accounts/{id}", server.GetAccount).Methods("GET")
	router.HandleFunc("/transactions", server.CreateTransaction).Methods("POST")

	// Set port from env or fallback
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Println("intrapay server is running on port", port)
	log.Fatal(http.ListenAndServe(":"+port, router))
}
