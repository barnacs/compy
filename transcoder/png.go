package transcoder

import (
	"github.com/barnacs/compy/proxy"
	"image/png"
)

type Png struct{}

func (t *Png) Transcode(w *proxy.ResponseWriter, r *proxy.ResponseReader) error {
	img, err := png.Decode(r)
	if err != nil {
		return err
	}
	if err = png.Encode(w, img); err != nil {
		return err
	}
	return nil
}
