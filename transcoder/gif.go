package transcoder

import (
	"github.com/barnacs/compy/proxy"
	"image/gif"
	"net/http"
)

type Gif struct{}

func (t *Gif) Transcode(w *proxy.ResponseWriter, r *proxy.ResponseReader, headers http.Header) error {
	img, err := gif.Decode(r)
	if err != nil {
		return err
	}
	if err = gif.Encode(w, img, nil); err != nil {
		return err
	}
	return nil
}
