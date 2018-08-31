package transcoder

import (
	"compress/gzip"
	"net/http"
	"strings"

	"github.com/barnacs/compy/proxy"
	brotlidec "gopkg.in/kothar/brotli-go.v0/dec"
	brotlienc "gopkg.in/kothar/brotli-go.v0/enc"
)

type Zip struct {
	proxy.Transcoder
	BrotliCompressionLevel int
	GzipCompressionLevel   int
	SkipCompressed         bool
}

func (t *Zip) Transcode(w *proxy.ResponseWriter, r *proxy.ResponseReader, headers http.Header) error {
	shouldBrotli := false
	shouldGzip := false
	for _, v := range strings.Split(headers.Get("Accept-Encoding"), ", ") {
		switch strings.SplitN(v, ";", 2)[0] {
		case "br":
			shouldBrotli = true
		case "gzip":
			shouldGzip = true
		}
	}

	// always gunzip if the client supports Brotli
	if r.Header().Get("Content-Encoding") == "gzip" && (shouldBrotli || !t.SkipCompressed) {
		gzr, err := gzip.NewReader(r.Reader)
		if err != nil {
			return err
		}
		defer gzr.Close()
		r.Reader = gzr
		r.Header().Del("Content-Encoding")
		w.Header().Del("Content-Encoding")
	}

	if r.Header().Get("Content-Encoding") == "br" && !t.SkipCompressed {
		brr := brotlidec.NewBrotliReader(r.Reader)
		defer brr.Close()
		r.Reader = brr
		r.Header().Del("Content-Encoding")
		w.Header().Del("Content-Encoding")
	}

	if shouldBrotli && compress(r) {
		params := brotlienc.NewBrotliParams()
		params.SetQuality(t.BrotliCompressionLevel)
		brw := brotlienc.NewBrotliWriter(params, w.Writer)
		defer brw.Close()
		w.Writer = brw
		w.Header().Set("Content-Encoding", "br")
	} else if shouldGzip && compress(r) {
		gzw, err := gzip.NewWriterLevel(w.Writer, t.GzipCompressionLevel)
		if err != nil {
			return err
		}
		defer gzw.Close()
		w.Writer = gzw
		w.Header().Set("Content-Encoding", "gzip")
	}
	return t.Transcoder.Transcode(w, r, headers)
}

func compress(r *proxy.ResponseReader) bool {
	return r.Header().Get("Content-Encoding") == ""
}
