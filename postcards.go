package postcards2diane

import (
	"bytes"
	"errors"
	"image"
	"image/color"
	"image/draw"
	"image/png"

	"golang.org/x/image/math/fixed"

	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
)

var postcardSizes = map[string]postcardSize{
	"6x11": postcardSize{
		dimensions: image.Rect(0, 0, 4500, 2500),
		safe:       image.Rect(152, 152, 4348, 2348),
	},
}

type postcardSize struct {
	dimensions image.Rectangle
	safe       image.Rectangle
}

func New(size string, lines []string, message string) *Postcard {
	return &Postcard{
		size:       size,
		font:       fonts[defaultFont],
		palette:    append(color.Palette{}, color.Black, color.White),
		background: color.White,
		text:       color.Black,
		lines:      lines,
		message:    message,
	}
}

type Postcard struct {
	size       string
	font       *truetype.Font
	palette    color.Palette
	background color.Color
	text       color.Color
	lines      []string
	message    string
}

func (p *Postcard) SetFontPreferences(fontNames ...string) bool {
	for _, name := range fontNames {
		f, ok := fonts[name]
		if ok {
			p.font = f
			return true
		}
	}
	return false
}

// Render draws the postcard image with the provided text and
// encodes it as a png.
func (p *Postcard) Render() ([]byte, error) {
	const startingFontSize = 500.0

	fontSize := startingFontSize
	img, err := p.render(fontSize)
	for err == errTooBig {
		fontSize = 3.0 * fontSize / 4.0
		img, err = p.render(fontSize)
	}
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	err = png.Encode(&buf, img)
	return buf.Bytes(), err
}

var errTooBig = errors.New("message too big")

func (p *Postcard) render(fontSize float64) (*image.Paletted, error) {
	const safeBufferPixels = 40

	sz := postcardSizes[p.size]
	img := image.NewPaletted(sz.dimensions, p.palette)
	draw.Draw(img, sz.dimensions, image.White, image.ZP, draw.Src)

	c := freetype.NewContext()
	c.SetFont(p.font)
	c.SetDPI(400)
	c.SetFontSize(fontSize)
	c.SetClip(sz.dimensions)
	c.SetDst(img)
	c.SetSrc(image.Black)

	var err error
	fixedSafeRect := fixed.R(sz.safe.Min.X, sz.safe.Min.Y, sz.safe.Max.X, sz.safe.Max.Y)
	pt := freetype.Pt(
		sz.safe.Min.X+safeBufferPixels,
		sz.safe.Min.Y+safeBufferPixels+int(c.PointToFixed(fontSize)>>6),
	)
	startingX := pt.X
	for _, line := range p.lines {
		pt, err = c.DrawString(line, pt)
		if err != nil {
			return nil, err
		}

		// Make sure that drawing the string didn't land us outside of
		// the safe printing area.
		if !pt.In(fixedSafeRect) {
			return nil, errTooBig
		}

		pt.X = startingX
		pt.Y += c.PointToFixed(fontSize * 1.1)
	}
	return img, nil
}
