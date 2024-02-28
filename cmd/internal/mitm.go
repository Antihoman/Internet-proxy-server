package mitm

import (
	"crypto/tls"
	"errors"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"sync"
	"time"
)

type Proxy struct {
	Wrap            func(upstream http.Handler) http.Handler
	CA              *tls.Certificate
	TLSServerConfig *tls.Config
	TLSClientConfig *tls.Config
	FlushInterval   time.Duration
}

func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Println("ServeHTTP")
	if r.Method == "CONNECT" {
		log.Println("CONNECT")
		p.serveConnect(w, r)
		return
	}
	log.Println("NO CONNECT")
	rp := &httputil.ReverseProxy{
		Director:      httpDirector,
		FlushInterval: p.FlushInterval,
	}
	p.Wrap(rp).ServeHTTP(w, r)
}

func (p *Proxy) serveConnect(w http.ResponseWriter, r *http.Request) {
	var (
		err           error
		sconn         *tls.Conn
		name          = dnsName(r.Host)
	)
	log.Printf("name : %s", name)
	if name == "" {
		log.Println("cannot determine cert name for " + r.Host)
		http.Error(w, "no upstream", 503)
		return
	}
	provisionalCert, err := p.cert(name)
	if err != nil {
		log.Println("cert", err)
		http.Error(w, "no upstream", 503)
		return
	}
	sConfig := new(tls.Config)
	if p.TLSServerConfig != nil {
		*sConfig = *p.TLSServerConfig
	}
	sConfig.Certificates = []tls.Certificate{*provisionalCert}
	sConfig.GetCertificate = func(hello *tls.ClientHelloInfo) (*tls.Certificate, error) {
		log.Println("GetCertificate")
		cConfig := new(tls.Config)
		if p.TLSClientConfig != nil {
			*cConfig = *p.TLSClientConfig
		}
		cConfig.ServerName = hello.ServerName
		log.Printf("Dial with %s from PROXY", r.Host)
		sconn, err = tls.Dial("tcp", r.Host, cConfig)
		if err != nil {
			log.Println("dial", r.Host, err)
			return nil, err
		}
		return p.cert(hello.ServerName)
	}
	cconn, err := handshake(w, sConfig)
	if err != nil {
		log.Println("handshake", r.Host, err)
		return
	}
	defer cconn.Close()
	if sconn == nil {
		log.Println("could not determine cert name for " + r.Host)
		return
	}
	defer sconn.Close()
	od := &oneShotDialer{c: sconn}
	rp := &httputil.ReverseProxy{
		Director:      httpsDirector,
		Transport:     &http.Transport{DialTLS: od.Dial},
		FlushInterval: p.FlushInterval,
	}
	ch := make(chan int)
	wc := &onCloseConn{cconn, func() { ch <- 0 }}
	http.Serve(&oneShotListener{wc}, p.Wrap(rp))
	<-ch
}

func (p *Proxy) cert(names ...string) (*tls.Certificate, error) {
	return genCert(p.CA, names)
}

var okHeader = []byte("HTTP/1.1 200 OK\r\n\r\n")

func handshake(w http.ResponseWriter, config *tls.Config) (net.Conn, error) {
	raw, _, err := w.(http.Hijacker).Hijack()
	if err != nil {
		http.Error(w, "no upstream", 503)
		return nil, err
	}
	log.Printf("resp OK to client")
	if _, err = raw.Write(okHeader); err != nil {
		raw.Close()
		return nil, err
	}
	conn := tls.Server(raw, config)
	log.Println("Handshake with CLIENT")
	err = conn.Handshake()
	if err != nil {
		conn.Close()
		raw.Close()
		return nil, err
	}
	return conn, nil
}

func httpDirector(r *http.Request) {
	r.URL.Host = r.Host
	r.URL.Scheme = "http"
}

func httpsDirector(r *http.Request) {
	r.URL.Host = r.Host
	r.URL.Scheme = "https"
}

func dnsName(addr string) string {
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return ""
	}
	return host
}

func namesOnCert(conn *tls.Conn) []string {
	c := conn.ConnectionState().PeerCertificates[0]
	if len(c.DNSNames) > 0 {
		return c.DNSNames
	}
	return []string{c.Subject.CommonName}
}

type oneShotDialer struct {
	c  net.Conn
	mu sync.Mutex
}

func (d *oneShotDialer) Dial(network, addr string) (net.Conn, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.c == nil {
		log.Println(">>> Dial nil")
		return nil, errors.New("closed")
	}
	c := d.c
	d.c = nil
	return c, nil
}

type oneShotListener struct {
	c net.Conn
}

func (l *oneShotListener) Accept() (net.Conn, error) {
	if l.c == nil {
		log.Println(">>> Accept nil")
		return nil, errors.New("closed")
	}
	log.Println(">>> Conn Accept BUSY")
	c := l.c
	l.c = nil
	return c, nil
}

func (l *oneShotListener) Close() error {
	return nil
}

func (l *oneShotListener) Addr() net.Addr {
	return l.c.LocalAddr()
}

type onCloseConn struct {
	net.Conn
	f func()
}

func (c *onCloseConn) Close() error {
	if c.f != nil {
		c.f()
		log.Println(">>> Close nil")
		c.f = nil
	}
	return c.Conn.Close()
}
