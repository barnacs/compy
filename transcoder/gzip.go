package transcoder

import (
	"compress/gzip"
	"github.com/barnacs/compy/proxy"
)

type Gzip struct {
	proxy.Transcoder
	SkipGzipped bool
}

func (t *Gzip) Transcode(w *proxy.ResponseWriter, r *proxy.ResponseReader) error {
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
	if compress(r) {
		gzw := gzip.NewWriter(w.Writer)
		defer gzw.Flush()
		w.Writer = gzw
		w.Header().Set("Content-Encoding", "gzip")
	}
	return t.Transcoder.Transcode(w, r)
}

func (t *Gzip) decompress(r *proxy.ResponseReader) bool {
	return !t.SkipGzipped && r.Header().Get("Content-Encoding") == "gzip"
}

func compress(r *proxy.ResponseReader) bool {
	return r.Header().Get("Content-Encoding") == ""
}
