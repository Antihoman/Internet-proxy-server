package delivery

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/url"

	"github.com/gorilla/mux"
	"github/Antihoman/Internet-proxy-server/pkg/domain"
)

const (
	attackVector           = `vulnerable'"><img src onerror=alert()>`
	resHeaderTransactionID = "X-Transaction-Id"
	proxyURL               = "http://proxy:8080"
)

type Handler struct {
	repo Repository
}

func NewHandler(repo Repository) Handler {
	return Handler{repo: repo}
}

func (h *Handler) Requests(w http.ResponseWriter, r *http.Request) {
	transactions, err := h.repo.GetAll()
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Error to get all requests"))
	}

	reqs := []TransactionDTO{}
	for _, tr := range transactions {
		req := TransactionDTO{
			ID:            tr.ID.(string),
			Host:          tr.Request.Host,
			Method:        tr.Request.Method,
			Path:          tr.Request.Path,
			StatusCode:    tr.Response.StatusCode,
			ContentLenght: tr.Response.ContentLenght,
		}
		reqs = append(reqs, req)
	}
	response, err := json.Marshal(reqs)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
	}
	w.WriteHeader(http.StatusOK)
	w.Write(response)

}

func (h *Handler) RequestByID(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	id := mux.Vars(r)["id"]
	transaction, err := h.repo.GetByID(id)

	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Error to get transaction by id"))
	}

	response, err := json.Marshal(transaction)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
	w.WriteHeader(http.StatusOK)
	w.Write(response)

}

func (h *Handler) ScanByID(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	id := mux.Vars(r)["id"]
	transaction, err := h.repo.GetByID(id)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Error to get transaction by id"))
		return
	}

	vulnGetParams := []string{}
	vulnPostParams := []string{}

	transactionsIDs := []string{}

	for key, value := range transaction.Request.GetParams {
		transaction.Request.GetParams[key] = `vulnerable'"><img src onerror=alert()>`
		resRepeat, err := RepeatRequest(transaction)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Error to repeat request"))
			return
		}
		body, err := io.ReadAll(resRepeat.Body)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Error to read body from repeated request"))
			return
		}
		transactionsIDs = append(transactionsIDs, resRepeat.Header[resHeaderTransactionID][0])
		if bytes.Contains(body, []byte(attackVector)) {
			vulnGetParams = append(vulnGetParams, key)
		}
		transaction.Request.GetParams[key] = value
	}

	for key, value := range transaction.Request.PostParams {
		transaction.Request.PostParams[key] = `vulnerable'"><img src onerror=alert()>`
		resRepeat, err := RepeatRequest(transaction)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Error to repeat request"))
			return
		}
		body, err := io.ReadAll(resRepeat.Body)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Error to read body from repeated request"))
			return
		}
		transactionsIDs = append(transactionsIDs, resRepeat.Header[resHeaderTransactionID][0])
		if bytes.Contains(body, []byte(attackVector)) {
			vulnPostParams = append(vulnGetParams, key)
		}

		transaction.Request.PostParams[key] = value
	}

	isVuln := false
	if len(vulnGetParams)+len(vulnPostParams) != 0 {
		isVuln = true
	}

	res, err := json.Marshal(map[string]interface{}{
		"body": map[string]interface{}{
			"request_id":             id,
			"scan_requests":          transactionsIDs,
			"is_vulnerable":          isVuln,
			"vulnerable_post_params": vulnPostParams,
			"vulnerable_get_params":  vulnGetParams,
		},
	})

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusOK)
	w.Write(res)
}

func (h *Handler) RepeatByID(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	id := mux.Vars(r)["id"]
	transaction, err := h.repo.GetByID(id)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Error to get transaction by id"))
		return
	}

	resRepeat, err := RepeatRequest(transaction)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Error to repeat request"))
		return

	}
	res, err := json.Marshal(map[string]interface{}{
		"body": map[string]interface{}{
			"current_request_id":  resRepeat.Header[resHeaderTransactionID][0],
			"repeated_request_id": id,
			"status_code":         resRepeat.Status,
			"content_length":      resRepeat.ContentLength,
		},
	})

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}

	w.Write(res)
	w.WriteHeader(http.StatusOK)
}

func RepeatRequest(transaction domain.HTTPTransaction) (*http.Response, error) {
	proxyURL, err := url.Parse(proxyURL)
	if err != nil {
		return nil, err
	}
	transport := &http.Transport{
		Proxy: http.ProxyURL(proxyURL),
	}
	client := &http.Client{
		Transport: transport,
	}

	u, err := url.Parse(transaction.Request.Host + transaction.Request.Path)
	if err != nil {
		return nil, err
	}

	query := u.Query()
	for key, value := range transaction.Request.GetParams {
		query.Add(key, value)
	}
	u.RawQuery = query.Encode()

	req, err := http.NewRequest(transaction.Request.Method,
		transaction.Request.Protocol+"://"+u.String(),
		bytes.NewBuffer(transaction.Response.RawBody))
	if err != nil {
		return nil, err
	}

	for key, value := range transaction.Request.Headers {
		req.Header.Set(key, value)
	}

	for key, value := range transaction.Request.Cookies {
		req.AddCookie(&http.Cookie{Name: key, Value: value})
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}