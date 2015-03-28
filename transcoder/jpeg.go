package transcoder

import (
	"github.com/barnacs/compy/proxy"
	"github.com/pixiv/go-libjpeg/jpeg"
)

type Jpeg struct {
	decOptions *jpeg.DecoderOptions
	encOptions *jpeg.EncoderOptions
}

func NewJpeg(quality int) *Jpeg {
	return &Jpeg{
		decOptions: &jpeg.DecoderOptions{},
		encOptions: &jpeg.EncoderOptions{
			Quality:        quality,
			OptimizeCoding: true,
		},
	}
}

func (t *Jpeg) Transcode(w *proxy.ResponseWriter, r *proxy.ResponseReader) error {
	img, err := jpeg.Decode(r, t.decOptions)
	if err != nil {
		return err
	}
	if err = jpeg.Encode(w, img, t.encOptions); err != nil {
		return err
	}
	return nil
}
