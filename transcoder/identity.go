package transcoder

import (
	"github.com/barnacs/compy/proxy"
)

type Identity struct{}

func (i *Identity) Transcode(w *proxy.ResponseWriter, r *proxy.ResponseReader) error {
	return w.ReadFrom(r)
}
