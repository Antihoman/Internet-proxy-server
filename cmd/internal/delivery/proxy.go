package delivery

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"net/http/httputil"
	"time"
)

type Proxy struct {
	Wrap func(upstream http.Handler) http.Handler
	CA *tls.Certificate

	TLSServerConfig *tls.Config

	TLSClientConfig *tls.Config

	FlushInterval time.Duration
}

func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == "CONNECT" {
		fmt.Println("START CONNECT from ", r.URL)
		p.serveConnect(w, r)
		fmt.Println("END CONNECT from ", r.URL)
		return
	}
	reverseProxy := &httputil.ReverseProxy{
		Director:      httpDirector,
		FlushInterval: p.FlushInterval,
	}
	p.Wrap(reverseProxy).ServeHTTP(w, r)
}

func httpDirector(r *http.Request) {
	r.URL.Host = r.Host
	r.URL.Scheme = "http"
}

func httpsDirector(r *http.Request) {
	r.URL.Host = r.Host
	r.URL.Scheme = "https"
}