package main

import (
	"image"
	"image/draw"

	"image/jpeg"
	"image/png"

	"os"
)

type singleImageProcessor struct{
	in image.Image
	out image.Image
	format string
}

func (imgp *singleImageProcessor) Decode(f *os.File) {
	inImg, format, err := image.Decode(f)
	if err != nil {
		perr("Could not decode image:", err)
	}
	imgp.in = inImg
	imgp.format = format
}

func (imgp *singleImageProcessor) Swap(colorInverter inverter) {
	bounds := imgp.in.Bounds()
	rgbImg := image.NewNRGBA(bounds)
	for y := 0; y < bounds.Max.Y; y++ {
		for x := 0; x < bounds.Max.X; x++ {
			c := colorInverter(imgp.in.At(x, y))
			rgbImg.Set(x, y, c)
		}
	}
	imgp.in = rgbImg
}

func (imgp *singleImageProcessor) Colorize(doDither bool, ditherer draw.Drawer) {
	bounds := imgp.in.Bounds()
	outImg := image.NewPaletted(bounds, palette)

	if doDither {
		ditherer.Draw(outImg, bounds, imgp.in, bounds.Min)
	} else {
		draw.Draw(outImg, bounds, imgp.in, bounds.Min, draw.Src)
	}

	imgp.out = outImg
}

func (imgp *singleImageProcessor) Write(f *os.File) {
	switch imgp.format {
		case "jpeg": jpeg.Encode(f, imgp.out, nil)
		case "png": png.Encode(f, imgp.out)
	}
}
