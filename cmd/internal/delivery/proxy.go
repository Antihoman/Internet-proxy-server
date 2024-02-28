package delivery

import (
	"crypto/tls"
	"net/http"
	"net/http/httputil"
	"time"
)

type Proxy struct {
	Wrap func(upstream http.Handler, isSecure bool) http.Handler

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
	p.Wrap(reverseProxy, false).ServeHTTP(w, r)
}

func httpDirector(r *http.Request) {
	r.URL.Host = r.Host
	r.URL.Scheme = "http"
}

func httpsDirector(r *http.Request) {
	r.URL.Host = r.Host
	r.URL.Scheme = "https"
}