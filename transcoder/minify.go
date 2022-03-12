package transcoder

import (
	"github.com/barnacs/compy/proxy"
	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/css"
	"github.com/tdewolff/minify/v2/html"
	"github.com/tdewolff/minify/v2/js"
	"net/http"
)

type Minifier struct {
	m *minify.M
}

func NewMinifier() *Minifier {
	m := minify.New()
	m.AddFunc("text/html", html.Minify)
	m.AddFunc("text/css", css.Minify)
	m.AddFunc("text/javascript", js.Minify)
	m.AddFunc("application/javascript", js.Minify)
	m.AddFunc("application/x-javascript", js.Minify)
	return &Minifier{
		m: m,
	}
}

func (t *Minifier) Transcode(w *proxy.ResponseWriter, r *proxy.ResponseReader, headers http.Header) error {
	return t.m.Minify(r.ContentType(), w, r)
}
