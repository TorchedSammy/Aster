package main

import (
	"image"
	"image/color"
	"image/draw"

	"image/jpeg"
	"image/png"

	//"math"
	"os"

	//"github.com/lucasb-eyer/go-colorful"
	"golang.org/x/exp/rand"
	"gonum.org/v1/gonum/stat/distuv"
)

type clutProcessor struct{
	in image.Image
	out image.Image
	format string
}

func (imgp *clutProcessor) Decode(f *os.File) {
	inImg, format, err := image.Decode(f)
	if err != nil {
		perr("Could not decode image:", err)
	}
	imgp.in = inImg
	imgp.format = format
}

func (imgp *clutProcessor) Swap(colorInverter inverter) {
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

func (imgp *clutProcessor) Colorize(doDither bool, ditherer draw.Drawer) {
	// step 1: generate half clut
	level := 8 // ??
	clutSize := level * level
	clutImgSize := clutSize * level // 8 * 8 * 8 (cube, makes sense right?) 
	clutBounds := image.Rectangle{Min: image.Point{0, 0}, Max: image.Point{clutImgSize, clutImgSize}}
	clut := image.NewNRGBA(clutBounds)

    p := 0;
    //t := float64(0.5)
    dist := distuv.Normal{
		Mu: 0,
		Sigma: 20,
		Src: rand.NewSource(42080085),
    }
    iterations := 512
    for blue := 0; blue < clutSize; blue++ {
        for green := 0; green < clutSize; green++ {
            for red := 0; red < clutSize; red++ {
                r := (red * 255) / (clutSize - 1);
                g := (green * 255) / (clutSize - 1);
                b := (blue * 255) / (clutSize - 1);

                x := p % clutImgSize;
                y := (p - x) / clutImgSize;

				clutColor := color.RGBA{uint8(r), uint8(g), uint8(b), 0xff}
				/*
				// 3rd party blend thing attempt that didnt work well
				cA, _ := colorful.MakeColor(clutColor)
				cB, _ := colorful.MakeColor(palette[palette.Index(clutColor)])
				correctedColor := cA.BlendRgb(cB, t)
				*/
				/*
				// simple method
				correctedColor := palette[palette.Index(clutColor)]
				*/

				// gaussian sampling (that also doesnt work lol)
				// im stupid
				mean := []float64{0, 0, 0}
				for n := 0; n < iterations; n++ {
					_ = n // dont ask
					color := color.RGBA{
						uint8(float64(clutColor.R) + dist.Rand()),
						uint8(float64(clutColor.G) + dist.Rand()),
						uint8(float64(clutColor.B) + dist.Rand()),
						0xff,
					}

					corrected := palette.Convert(color)
					r, g, b, _ := corrected.RGBA()
					mean[0] += float64(r / 257) / float64(iterations)
					mean[1] += float64(g / 257) / float64(iterations)
					mean[2] += float64(b / 257) / float64(iterations)
				}
				correctedColor := color.RGBA{
					uint8((mean[0])),
					uint8((mean[1])),
					uint8((mean[2])),
					0xff,
				}

				clut.Set(x, y, correctedColor)
                p++;
            }
        }
    }

	imgp.out = clut

/*
	bounds := imgp.in.Bounds()
	outImg := image.NewNRGBA(bounds)
	for y := 0; y < bounds.Max.Y; y++ {
		for x := 0; x < bounds.Max.X; x++ {
			i := imgp.in.At(x, y).(color.RGBA)
			r := uint32(i.R) * uint32(clutSize - 1) / 255;
			g := uint32(i.G) * uint32(clutSize - 1) / 255;
			b := uint32(i.B) * uint32(clutSize - 1) / 255;

			xx := (int(r) % clutSize) + (int(g) % (level)) * clutSize;
			yy := (int(b) * level) + (int(g) / level);

			outImg.Set(x, y, clut.At(xx, yy))
		}
	}

	imgp.out = outImg
*/
	/*
	bounds := imgp.in.Bounds()
	outImg := image.NewPaletted(bounds, palette)

	if doDither {
		ditherer.Draw(outImg, bounds, imgp.in, bounds.Min)
	} else {
		draw.Draw(outImg, bounds, imgp.in, bounds.Min, draw.Src)
	}

	*/
}

func (imgp *clutProcessor) Write(f *os.File) {
	switch imgp.format {
		case "jpeg": jpeg.Encode(f, imgp.out, nil)
		case "png": png.Encode(f, imgp.out)
	}
}
