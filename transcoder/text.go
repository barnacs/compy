package transcoder

import (
	"github.com/barnacs/compy/proxy"
	"github.com/tdewolff/minify"
	"github.com/tdewolff/minify/css"
	"github.com/tdewolff/minify/html"
	"github.com/tdewolff/minify/js"
	"github.com/tdewolff/parse"
)

func init() {
	parse.MaxBuf *= 8
}

type Text struct {
	m minify.Minify
}

func NewText() *Text {
	m := minify.New()
	m.AddFunc("text/html", html.Minify)
	m.AddFunc("text/css", css.Minify)
	m.AddFunc("text/javascript", js.Minify)
	m.AddFunc("application/javascript", js.Minify)
	m.AddFunc("application/x-javascript", js.Minify)
	return &Text{
		m: m,
	}
}

func (t *Text) Transcode(w *proxy.ResponseWriter, r *proxy.ResponseReader) error {
	return t.m.Minify(r.ContentType(), w, r)
}
