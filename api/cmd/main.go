package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github/Antihoman/Internet-proxy-server/pkg/mongoclient"
	"github/Antihoman/Internet-proxy-server/pkg/repository"
	internal "github/Antihoman/Internet-proxy-server/api/internal"
)

const URI = "mongodb://root:root@mongo:27017"

func main() {
	log.SetPrefix("[WEB-API] ")
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	client, closeConn, err := mongoclient.NewMongoClient(URI)
	if err != nil {
		log.Fatal(err)
	}

	defer closeConn()
	repo, err := repository.NewRepository(client)
	if err != nil {
		log.Fatal(err)
	}

	handler := internal.NewHandler(&repo)

	r := mux.NewRouter()

	r.Use(internal.Log)

	r.HandleFunc("/requests", handler.Requests)
	r.HandleFunc("/requests/{id}", handler.RequestByID)
	r.HandleFunc("/scan/{id}", handler.ScanByID)
	r.HandleFunc("/repeat/{id}", handler.RepeatByID)

	log.Println("Web-api :8000")
	log.Fatal(http.ListenAndServe(":8000", r))
}