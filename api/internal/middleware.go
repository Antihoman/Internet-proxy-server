package delivery

import (
	"log"
	"net/http"
)

func Log(upstream http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println(r.Method, r.Host, r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		upstream.ServeHTTP(w, r)
	})
}