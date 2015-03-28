package transcoder

import (
	"github.com/barnacs/compy/proxy"
	"io"
)

type Identity struct{}

func (i *Identity) Transcode(w *proxy.ResponseWriter, r *proxy.ResponseReader) error {
	_, err := io.Copy(w, r)
	return err
}
