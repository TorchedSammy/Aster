package main

import (
	"image"
	"image/draw"

	"image/gif"

	"os"
)

type gifProcessor struct{
	gif *gif.GIF
	imgs []image.Image
	processed []*image.Paletted
}

func (g *gifProcessor) Decode(f *os.File) {
	maybeGif, _ := gif.DecodeAll(f) // TODO: handle error
	g.gif = maybeGif
	g.imgs = make([]image.Image, len(g.gif.Image))

	for i, im := range g.gif.Image {
		g.imgs[i] = im
	}
}

func (g *gifProcessor) Swap(colorInverter inverter) {
	for i, im := range g.gif.Image {
		bounds := im.Bounds()
		rgbImg := image.NewNRGBA(bounds)
		for y := 0; y < bounds.Max.Y; y++ {
			for x := 0; x < bounds.Max.X; x++ {
				c := colorInverter(im.At(x, y))
				rgbImg.Set(x, y, c)
			}
		}
		g.imgs[i] = rgbImg
	}
}

func (g *gifProcessor) Colorize(doDither bool, ditherer draw.Drawer) {
	for _, im := range g.imgs {
		bounds := im.Bounds()
		outImg := image.NewPaletted(bounds, palette)

		if doDither {
			ditherer.Draw(outImg, bounds, im, bounds.Min)
		} else {
			draw.Draw(outImg, bounds, im, bounds.Min, draw.Src)
		}

		g.processed = append(g.processed, outImg)
	}
}

func (g *gifProcessor) Write(f *os.File) {
	newGif := &gif.GIF{
		Image: g.processed,
		Delay: g.gif.Delay,
		LoopCount: g.gif.LoopCount,
		Disposal: g.gif.Disposal,
		Config: g.gif.Config,
		BackgroundIndex: g.gif.BackgroundIndex,
	}

	gif.EncodeAll(f, newGif)
}
