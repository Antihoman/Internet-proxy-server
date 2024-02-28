package delivery

type Request struct {
	ID         string `json:"id"`
	Host       string `json:"host"`
	Method     string `json:"method"`
	Path       string `json:"path"`
	StatusCode int    `json:"status_code"`
}