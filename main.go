package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	config, err := NewConfig("config.json")
	if err != nil {
		log.Fatal(err)
	}

	db, err := sqlx.Open("sqlite3", config.dbSourceFile)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	_, err = os.Stat(config.DbFilePath)
	if config.InMemory || os.IsNotExist(err) {
		initSchema(db)
	}

	initData(db, config.CbsApiUrl, config.Sports, config.ImportBatchSize)

	r := chi.NewRouter()
	r.Use(middleware.Logger, middleware.Recoverer, middleware.RequestID)

	store := NewPlayerStore(db)
	service := &Service{players: store}

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		var query GetPlayersQuery
		if err := json.NewDecoder(r.Body).Decode(&query); err != nil {
			log.Println("failed to decode request")
			if err := json.NewEncoder(w).Encode(ServiceResponse[[]PlayerResponse]{Data: nil, statusCode: http.StatusBadRequest}); err != nil {
				log.Printf("failed to encode get response: %v\n", err)
			}

			return
		}
		resp := service.GetPlayers(query)
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			log.Printf("failed to encode get response: %v\n", err)
		}
	})

	fmt.Println("starting server at http://localhost:8080")
	http.ListenAndServe(":"+config.Port, r)
}
