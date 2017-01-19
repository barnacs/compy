package transcoder

import (
	"github.com/barnacs/compy/proxy"
	"github.com/chai2010/webp"
	"image/png"
	"net/http"
)

type Png struct{}

func (t *Png) Transcode(w *proxy.ResponseWriter, r *proxy.ResponseReader, headers http.Header) error {
	img, err := png.Decode(r)
	if err != nil {
		return err
	}

	if SupportsWebP(headers) {
		w.Header().Set("Content-Type", "image/webp")
		options := webp.Options{
			Lossless: true,
		}
		if err = webp.Encode(w, img, &options); err != nil {
			return err
		}
	} else {
		if err = png.Encode(w, img); err != nil {
			return err
		}
	}
	return nil
}
