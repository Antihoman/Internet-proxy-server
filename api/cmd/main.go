package main

import (
	"fmt"
	"log"
	"net/http"
)

func handler(w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()
	name := params.Get("name")
	response := fmt.Sprintf("Hello, %s!", name)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(response))
}

func main() {
	http.HandleFunc("/", handler)
	log.Fatal(http.ListenAndServe(":80", nil))
}