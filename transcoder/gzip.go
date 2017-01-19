package transcoder

import (
	"compress/gzip"
	"github.com/barnacs/compy/proxy"
	"net/http"
	"strings"
)

type Gzip struct {
	proxy.Transcoder
	CompressionLevel int
	SkipGzipped      bool
}

func (t *Gzip) Transcode(w *proxy.ResponseWriter, r *proxy.ResponseReader, headers http.Header) error {
	if t.decompress(r) {
		gzr, err := gzip.NewReader(r.Reader)
		if err != nil {
			return err
		}
		defer gzr.Close()
		r.Reader = gzr
		r.Header().Del("Content-Encoding")
		w.Header().Del("Content-Encoding")
	}

	shouldGzip := false
	for _, v := range strings.Split(headers.Get("Accept-Encoding"), ", ") {
		if strings.SplitN(v, ";", 2)[0] == "gzip" {
			shouldGzip = true
			break
		}
	}

	if shouldGzip && compress(r) {
		gzw, err := gzip.NewWriterLevel(w.Writer, t.CompressionLevel)
		if err != nil {
			return err
		}
		defer gzw.Close()
		w.Writer = gzw
		w.Header().Set("Content-Encoding", "gzip")
	}
	return t.Transcoder.Transcode(w, r, headers)
}

func (t *Gzip) decompress(r *proxy.ResponseReader) bool {
	return !t.SkipGzipped && r.Header().Get("Content-Encoding") == "gzip"
}

func compress(r *proxy.ResponseReader) bool {
	return r.Header().Get("Content-Encoding") == ""
}
