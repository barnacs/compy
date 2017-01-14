package transcoder

import (
	"github.com/barnacs/compy/proxy"
	"net/http"
)

type Identity struct{}

func (i *Identity) Transcode(w *proxy.ResponseWriter, r *proxy.ResponseReader, headers http.Header) error {
	return w.ReadFrom(r)
}
