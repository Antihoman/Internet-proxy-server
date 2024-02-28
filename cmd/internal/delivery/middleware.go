package delivery

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github/Antihoman/Internet-proxy-server/cmd/internal/domain"
)

type Middleware struct {
	repo Repository
}

func NewMiddleware(repo Repository) Middleware {
	return Middleware{repo: repo}
}

type customRecorder struct {
	http.ResponseWriter

	response []byte
	code     int
}

func (w *customRecorder) WriteHeader(code int) {
	w.code = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *customRecorder) Write(b []byte) (int, error) {
	w.response = append(w.response, b...)
	return w.ResponseWriter.Write(b)
}

func parseReqHeaders(r *http.Request) map[string]string {
	data := make(map[string]string)
	for name, values := range r.Header {
		if name != "Cookie" {
			data[name] = values[0]
		}
	}
	return data
}

func parseReqCookies(r *http.Request) map[string]string {
	data := make(map[string]string)
	for _, cookie := range r.Cookies() {
		data[cookie.Name] = cookie.Value
	}
	return data
}

func parseReqGetParams(r *http.Request) map[string]string {
	data := make(map[string]string)
	query := r.URL.Query()
	fmt.Println("query ==", query)
	for key, values := range query {
		data[key] = values[0]
	}
	return data
}

func parseReqPostParams(r *http.Request) map[string]string {
	err := r.ParseForm()
	if err != nil {
		log.Panic(err)
	}

	data := make(map[string]string)
	for key, values := range r.PostForm {
		data[key] = values[0]
	}
	return data
}

func parseResHeaders(w http.ResponseWriter) map[string]string {
	data := make(map[string]string)
	for name, values := range w.Header() {
		data[name] = values[0]
	}
	return data
}

func (mw *Middleware) Save(upstream http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println(r.Method, r.Host, r.URL.Path)
		r.Header.Set("X-Test", "test")

		recorder := &customRecorder{ResponseWriter: w}

		r.Header.Del("Accept-Encoding")
		recorder.Header().Set("Content-Encoding", "identity")

		getParams := parseReqGetParams(r)
		headers := parseReqHeaders(r)
		cookies := parseReqCookies(r)
		postParams := parseReqPostParams(r)

		upstream.ServeHTTP(recorder, r)

		transaction := domain.HTTPTransaction{
			Time: time.Now(),
			Request: domain.Request{
				Host:       r.Host,
				Method:     r.Method,
				Version:    r.Proto,
				Path:       r.URL.Path,
				Headers:    headers,
				Cookies:    cookies,
				GetParams:  getParams,
				PostParams: postParams,
			},
			Response: domain.Response{
				StatusCode: recorder.code,
				Body:       recorder.response,
				Headers:    parseResHeaders(w),
			},
		}

		err := mw.repo.Add(transaction)
		if err != nil {
			log.Println("error to add request to db", err)
		}

	})
}