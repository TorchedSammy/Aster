package main

import (
	"image"
	"image/color"
	"image/draw"

	"image/jpeg"
	"image/png"

	"math"
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

func uint8Add(a uint8, b float64) uint8 {
	res := int(math.Round(float64(a) + b))
	return uint8Conv(res)
}

func uint8Conv(num int) uint8 {
	if num > math.MaxUint8 {
		return math.MaxUint8
	}

	return uint8(num)
}

func (imgp *clutProcessor) Colorize(doDither bool, ditherer draw.Drawer) {
	// step 1: generate half clut
	level := 6 // ??
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
                r := (red * 0xff) / (clutSize - 1);
                g := (green * 0xff) / (clutSize - 1);
                b := (blue * 0xff) / (clutSize - 1);

                x := p % clutImgSize;
                y := (p - x) / clutImgSize;

				clutColor := color.NRGBA{uint8(r), uint8(g), uint8(b), 0xff}
				//correctedColor := clutColor
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
					_ = n
					c := color.NRGBA{
						uint8Add(clutColor.R, math.Round(dist.Rand())),
						uint8Add(clutColor.G, math.Round(dist.Rand())),
						uint8Add(clutColor.B, math.Round(dist.Rand())),
						0xff,
					}

					corrected := palette.Convert(c).(color.NRGBA)
					mean[0] += float64(corrected.R) / float64(iterations)
					mean[1] += float64(corrected.G) / float64(iterations)
					mean[2] += float64(corrected.B) / float64(iterations)
				}
				println("x, y: ", x, y, " color: ", uint8Conv(int(math.Round(mean[0]))), uint8Conv(int(math.Round(mean[1]))), uint8Conv(int(math.Round(mean[2]))))
				correctedColor := color.NRGBA{
					uint8Conv(int(math.Round(mean[0]))),
					uint8Conv(int(math.Round(mean[1]))),
					uint8Conv(int(math.Round(mean[2]))),
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
			iR, iG, iB, _ := imgp.in.At(x, y).RGBA()
			r := (uint32(iR / 255) * uint32(clutSize - 1)) / 255;
			g := (uint32(iG / 255) * uint32(clutSize - 1)) / 255;
			b := (uint32(iB / 255) * uint32(clutSize - 1)) / 255;

			xx := (int(r) % clutSize) + (int(g) % (level)) * clutSize;
			yy := (int(b) * level) + (int(g) / level);

			outImg.Set(x, y, clut.At(xx, yy))
		}
	}

	imgp.out = outImg
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
