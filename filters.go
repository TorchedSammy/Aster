package main

import (
	"math"
	"image"
	"image/color"

	"golang.org/x/image/draw"
	resiz "github.com/nfnt/resize"
)

func recolorDither(img image.Image, palette color.Palette, algorithm string) (image.Image, error) {
	dither, err := getDitherAlgo(algorithm)
	if err != nil {
		return nil, err
	}

	bounds := img.Bounds()
	out := image.NewPaletted(bounds, palette)
	dither.Draw(out, bounds, img, bounds.Min)

	return out, nil
}

func recolor(img image.Image, palette color.Palette) (image.Image, error) {
	bounds := img.Bounds()
	out := image.NewPaletted(bounds, palette)
	draw.Draw(out, bounds, img, bounds.Min, draw.Src)

	return out, nil
}

func resize(img image.Image, percent int) image.Image {
	size := float64(img.Bounds().Size().X) * (float64(percent) / float64(100))
	return resiz.Resize(uint(math.Floor(size)), 0, img, resiz.NearestNeighbor)
}
