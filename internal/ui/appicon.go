package ui

import (
	"bytes"
	_ "embed"
	"image"
	"image/png"

	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/styles/properties"
)

// iconPNG is the app icon, drawn once at design time (256×256).
//
//go:embed assets/icon.png
var iconPNG []byte

// iconImage is iconPNG decoded once, for the title bar.
var iconImage = func() image.Image {
	img, err := png.Decode(bytes.NewReader(iconPNG))
	if err != nil {
		return image.NewNRGBA(image.Rect(0, 0, 1, 1))
	}
	return img
}()

// logo shows the app icon at size dp.
func logo(size int) giode.Element {
	return giode.Image(iconImage, giode.Styles{Width: properties.Width(size), Height: properties.Height(size), Fit: giode.FitContain})
}
