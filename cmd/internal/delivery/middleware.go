package delivery

import (
	"bytes"
	"log"
	"net/http"
	"time"

	"github/Antihoman/Internet-proxy-server/pkg/domain"
	"go.mongodb.org/mongo-driver/bson/primitive"
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
	for key, values := range query {
		data[key] = values[0]
	}
	return data
}

func parseReqPostParams(r *http.Request) map[string]string {
	err := r.ParseForm()
	if err != nil {
		log.Println(err)
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

type wrappedRequest struct {
	*http.Request
	body []byte
}

func (wr *wrappedRequest) Read(p []byte) (n int, err error) {
	return bytes.NewReader(wr.body).Read(p)
}

func (wr *wrappedRequest) Close() error {
	return nil
}

func (mw *Middleware) Save(upstream http.Handler, isSecure bool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println(r.Method, r.Host, r.URL.Path)
		r.Header.Set("X-Proxy", "yes")

		recorder := &customRecorder{ResponseWriter: w}

		r.Header.Del("Accept-Encoding")
		recorder.Header().Set("Content-Encoding", "identity")
		objectID := primitive.NewObjectID()
		recorder.Header().Set("X-Transaction-Id", objectID.Hex())

		reqGetParams := parseReqGetParams(r)
		reqHeaders := parseReqHeaders(r)
		reqCookies := parseReqCookies(r)
		reqPostParams := parseReqPostParams(r)

		var err error
		var reqBody []byte

		var protocol string
		if isSecure {
			protocol = "https"
		} else {
			protocol = "http"
		}

		upstream.ServeHTTP(recorder, r)

		transaction := domain.HTTPTransaction{
			ID:   objectID,
			Time: time.Now(),
			Request: domain.Request{
				Host:       r.Host,
				Method:     r.Method,
				Version:    r.Proto,
				Path:       r.URL.Path,
				Headers:    reqHeaders,
				Cookies:    reqCookies,
				Protocol:   protocol,
				GetParams:  reqGetParams,
				PostParams: reqPostParams,
				RawBody:    reqBody,
			},
			Response: domain.Response{
				StatusCode:    recorder.code,
				RawBody:       recorder.response,
				Headers:       parseResHeaders(w),
				ContentLenght: len(recorder.response),
			},
		}

		err = mw.repo.Add(transaction)
		if err != nil {
			http.Error(w, "Error to add request to db", http.StatusInternalServerError)
			log.Println("error to add request to db", err)
			return
		}

	})
}