package delivery

import (
	"crypto/tls"
	"net/http"
	"net/http/httputil"
	"time"
)

type Proxy struct {
	CA *tls.Certificate

	TLSServerConfig *tls.Config

	TLSClientConfig *tls.Config

	FlushInterval time.Duration
}

func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == "CONNECT" {
		p.serveConnect(w, r)
		return
	}
	reverseProxy := &httputil.ReverseProxy{
		Director:      httpDirector,
		FlushInterval: p.FlushInterval,
	}
	reverseProxy.ServeHTTP(w, r)
}

func httpDirector(r *http.Request) {
	r.URL.Host = r.Host
	r.URL.Scheme = "http"
}

func httpsDirector(r *http.Request) {
	r.URL.Host = r.Host
	r.URL.Scheme = "https"
}
