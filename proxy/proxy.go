package proxy

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"sync/atomic"
)

type Proxy struct {
	transcoders map[string]Transcoder
	ml          *mitmListener
	ReadCount   uint64
	WriteCount  uint64
	user        string
	pass        string
	host        string
	cert        string
}

type Transcoder interface {
	Transcode(*ResponseWriter, *ResponseReader, http.Header) error
}

func New(host string, cert string) *Proxy {
	p := &Proxy{
		transcoders: make(map[string]Transcoder),
		ml:          nil,
		host:        host,
		cert:        cert,
	}
	return p
}

func (p *Proxy) EnableMitm(ca, key string) error {
	cf, err := newCertFaker(ca, key)
	if err != nil {
		return err
	}

	var config *tls.Config
	if p.cert != "" {
		roots, err := x509.SystemCertPool()
		if err != nil {
			return err
		}
		pem, err := ioutil.ReadFile(p.cert)
		if err != nil {
			return err
		}
		ok := roots.AppendCertsFromPEM([]byte(pem))
		if !ok {
			return errors.New("failed to parse root certificate")
		}
		config = &tls.Config{RootCAs: roots}
	}

	p.ml = newMitmListener(cf, config)
	go http.Serve(p.ml, p)
	return nil
}

func (p *Proxy) SetAuthentication(user, pass string) {
	p.user = user
	p.pass = pass
}

func (p *Proxy) AddTranscoder(contentType string, transcoder Transcoder) {
	p.transcoders[contentType] = transcoder
}

func (p *Proxy) Start(host string) error {
	return http.ListenAndServe(host, p)
}

func (p *Proxy) StartTLS(host, cert, key string) error {
	return http.ListenAndServeTLS(host, cert, key, p)
}

func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Printf("serving request: %s", r.URL)
	if err := p.handle(w, r); err != nil {
		log.Printf("%s while serving request: %s", err, r.URL)
	}
}

func (p *Proxy) checkHttpBasicAuth(auth string) bool {
	prefix := "Basic "
	if !strings.HasPrefix(auth, prefix) {
		return false
	}
	decoded, err := base64.StdEncoding.DecodeString(auth[len(prefix):])
	if err != nil {
		return false
	}
	values := strings.SplitN(string(decoded), ":", 2)
	if len(values) != 2 || values[0] != p.user || values[1] != p.pass {
		return false
	}
	return true
}

func (p *Proxy) stripUnsupportedEncodings(in_encs string) string {
	out_encs := ""
	enc := ""

	tokens := strings.Split( strings.Replace( in_encs, " ", "", -1), ",")

	for _,  arg := range tokens {
		enc = strings.Split( arg, ";")[0]
		if enc == "br" || enc == "gzip" {
			if out_encs != "" {
				out_encs += ", "
			}
			out_encs = out_encs + enc
		}
	}
	fmt.Println( tokens)
	return out_encs
}

func (p *Proxy) handle(w http.ResponseWriter, r *http.Request) error {
	// TODO: only HTTPS?
	if p.user != "" {
		if !p.checkHttpBasicAuth(r.Header.Get("Proxy-Authorization")) {
			w.Header().Set("Proxy-Authenticate", "Basic realm=\"Compy\"")
			w.WriteHeader(http.StatusProxyAuthRequired)
			return nil
		}
		r.Header.Del("Proxy-Authorization")
	}

	if r.Method == "CONNECT" {
		return p.handleConnect(w, r)
	}

	if r.Header.Get("Accept-Encoding") != "" {
		supportedComp := p.stripUnsupportedEncodings(r.Header.Get("Accepted-Encoding"))

		if supportedComp != "" {
			r.Header.Set("Accept-Encoding", supportedComp)
		} else {
			r.Header.Del("Accept-Encoding")
		}
	}

	host := r.URL.Host
	if host == "" {
		host = r.Host
	}
	if hostname, err := os.Hostname(); host == p.host || (err == nil && host == hostname+p.host) {
		return p.handleLocalRequest(w, r)
	}

	resp, err := forward(r)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return fmt.Errorf("error forwarding request: %s", err)
	}
	defer resp.Body.Close()
	rw := newResponseWriter(w)
	rr := newResponseReader(resp)
	err = p.proxyResponse(rw, rr, r.Header)
	read := rr.counter.Count()
	written := rw.rw.Count()
	log.Printf("transcoded: %d -> %d (%3.1f%%)", read, written, float64(written)/float64(read)*100)
	atomic.AddUint64(&p.ReadCount, read)
	atomic.AddUint64(&p.WriteCount, written)
	return err
}

func (p *Proxy) handleLocalRequest(w http.ResponseWriter, r *http.Request) error {
	if r.Method == "GET" && (r.URL.Path == "" || r.URL.Path == "/") {
		w.Header().Set("Content-Type", "text/html")
		read := atomic.LoadUint64(&p.ReadCount)
		written := atomic.LoadUint64(&p.WriteCount)
		io.WriteString(w, fmt.Sprintf(`<html>
<head>
<title>compy</title>
</head>
<body>
<h1>compy</h1>
<ul>
<li>total transcoded: %d -> %d (%3.1f%%)</li>
<li><a href="/cacert">CA cert</a></li>
<li><a href="https://github.com/barnacs/compy">GitHub</a></li>
</ul>
</body>
</html>`, read, written, float64(written)/float64(read)*100))
		return nil
	} else if r.Method == "GET" && r.URL.Path == "/cacert" {
		if p.cert == "" {
			http.NotFound(w, r)
			return nil
		}
		w.Header().Set("Content-Type", "application/x-x509-ca-cert")
		http.ServeFile(w, r, p.cert)
		return nil
	} else {
		w.WriteHeader(http.StatusNotImplemented)
		return nil
	}
}

func forward(r *http.Request) (*http.Response, error) {
	if r.URL.Scheme == "" {
		if r.TLS != nil && r.TLS.ServerName == r.Host {
			r.URL.Scheme = "https"
		} else {
			r.URL.Scheme = "http"
		}
	}
	if r.URL.Host == "" {
		r.URL.Host = r.Host
	}
	r.RequestURI = ""
	return http.DefaultTransport.RoundTrip(r)
}

func (p *Proxy) proxyResponse(w *ResponseWriter, r *ResponseReader, headers http.Header) error {
	w.takeHeaders(r)
	transcoder, found := p.transcoders[r.ContentType()]
	if !found {
		return w.ReadFrom(r)
	}
	w.setChunked()
	if err := transcoder.Transcode(w, r, headers); err != nil {
		return fmt.Errorf("transcoding error: %s", err)
	}
	return nil
}

func (p *Proxy) handleConnect(w http.ResponseWriter, r *http.Request) error {
	if p.ml == nil {
		return fmt.Errorf("CONNECT received but mitm is not enabled")
	}
	w.WriteHeader(http.StatusOK)
	var conn net.Conn
	if h, ok := w.(http.Hijacker); ok {
		conn, _, _ = h.Hijack()
	} else {
		fw := w.(FlushWriter)
		fw.Flush()
		mconn := newMitmConn(fw, r.Body, r.RemoteAddr)
		conn = mconn
		defer func() {
			<-mconn.closed
		}()
	}
	sconn, err := p.ml.Serve(conn, r.Host)
	if err != nil {
		conn.Close()
		return err
	}
	sconn.Close() // TODO: reuse this connection for https requests
	return nil
}
