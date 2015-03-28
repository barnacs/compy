package proxy

import (
	"fmt"
	"log"
	"net/http"
)

type Proxy struct {
	transcoders map[string]Transcoder
	ml          *mitmListener
}

type Transcoder interface {
	Transcode(*ResponseWriter, *ResponseReader) error
}

func New() *Proxy {
	p := &Proxy{
		transcoders: make(map[string]Transcoder),
		ml:          nil,
	}
	return p
}

func (p *Proxy) EnableMitm(ca, key string) error {
	cf, err := newCertFaker(ca, key)
	if err != nil {
		return err
	}
	p.ml = newMitmListener(cf)
	go http.Serve(p.ml, p)
	return nil
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
	if err := p.handle(w, r); err != nil {
		log.Printf("%s while serving request: %s", err, r.URL)
	}
}

func (p *Proxy) handle(w http.ResponseWriter, r *http.Request) error {
	if r.Method == "CONNECT" {
		return p.handleConnect(w, r.Host)
	}
	resp, err := forward(r)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return fmt.Errorf("error forwarding request: %s", err)
	}
	defer resp.Body.Close()
	return p.proxyResponse(newResponseWriter(w), newResponseReader(resp))
}

func forward(r *http.Request) (*http.Response, error) {
	if r.URL.Scheme == "" {
		r.URL.Scheme = "https"
	}
	if r.URL.Host == "" {
		r.URL.Host = r.Host
	}
	r.RequestURI = ""
	return http.DefaultTransport.RoundTrip(r)
}

func (p *Proxy) proxyResponse(w *ResponseWriter, r *ResponseReader) error {
	w.takeHeaders(r)
	transcoder, found := p.transcoders[r.ContentType()]
	if !found {
		return w.writeFrom(r)
	}
	w.setChunked()
	if err := transcoder.Transcode(w, r); err != nil {
		return fmt.Errorf("transcoding error: %s", err)
	}
	return nil
}

func (p *Proxy) handleConnect(w http.ResponseWriter, host string) error {
	if p.ml == nil {
		return fmt.Errorf("CONNECT received but mitm is not enabled")
	}
	h, ok := w.(http.Hijacker)
	if !ok {
		return fmt.Errorf("connection cannot be hijacked")
	}
	w.WriteHeader(http.StatusOK)
	conn, _, err := h.Hijack()
	if err != nil {
		return err
	}
	sconn, err := p.ml.Serve(conn, host)
	if err != nil {
		conn.Close()
		return err
	}
	sconn.Close() // TODO: reuse this connection for https requests
	return nil
}
