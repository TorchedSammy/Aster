package main

import (
	"encoding/hex"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"math"
	"os"
	"strings"
	"sync"

	// supported formats
	"image/jpeg"
	"image/png"

	"github.com/spf13/pflag"
	"github.com/lucasb-eyer/go-colorful"
	"github.com/brouxco/dithering"
)

func main() {
	inFlag := pflag.StringP("input", "i", "", "Path to input image")
	outFlag := pflag.StringP("output", "o", "", "Path to output")
	paletteFlag := pflag.StringP("palette", "p", "", "Palette of the image")
	ditherFlag := pflag.BoolP("dither", "d", true, "Whether to use dithering on the image or not")
	ditherAlgoFlag := pflag.StringP("ditherAlgorithm", "D", "floydsteinberg", "The dithering algorithm to use.")
	swapFlag := pflag.BoolP("swap", "s", false, "Swap luminance of image before colorizing")
	swapOnlyFlag := pflag.BoolP("swapOnly", "S", false, "Only swap luminance and dont colorize. This implies -s (luminance swap)")
	grayscaleSwapFlag := pflag.BoolP("grayscaleSwap", "g", false, "Only invert parts of the image that are calculated to be grayscale (blacks/whites)")
	chunkedFlag := pflag.BoolP("chunk", "c", true, "Process colorization by rectangular chunks")
	chunksFlag := pflag.IntP("chunkAmount", "C", 4, "Amount of chunks to create out of the image")

	pflag.Parse()
	check(inFlag, "input")
	check(outFlag, "output")
//	check(paletteFlag, "palette")

	if *swapOnlyFlag {
		f := true
		swapFlag = &f
	}

	var op draw.Drawer = draw.Src
	if *ditherFlag {
		var err error
		op, err = getDitherAlgo(*ditherAlgoFlag)
		if err != nil {
			perr("Invalid dither algorithm", *ditherAlgoFlag)
		}
	}
	var palette color.Palette
	if len(*paletteFlag) != 0 {
		for _, colorStr := range strings.Split(*paletteFlag, " ") {
			col, err := strToColor(colorStr)
			if err != nil {
				fmt.Fprintln(os.Stderr, "Invalid color", colorStr)
				continue
			}
			palette = append(palette, col)
		}
	}

	if len(palette) < 2 && !*swapOnlyFlag {
		perr("Provided palette has less than 2 colors")
	}

	inFile, err := os.Open(*inFlag)
	if err != nil {
		perr("Could not open input file")
	}
	outFile, err := os.Create(*outFlag)
	if err != nil {
		perr("Could not create output file")
	}

	inImg, format, err := image.Decode(inFile)
	if err != nil {
		perr("Could not decode image:", err)
	}
	bounds := inImg.Bounds()

	var outImg draw.Image

	var colorInverter func(color.Color) color.Color = normalInverter
	if *grayscaleSwapFlag {
		colorInverter = grayscaleInverter
	}

	if *swapFlag {
		rgbImg := image.NewNRGBA(bounds)
		for y := 0; y < bounds.Max.Y; y++ {
			for x := 0; x < bounds.Max.X; x++ {
				c := colorInverter(inImg.At(x, y))
				rgbImg.Set(x, y, c)
			}
		}
		inImg = rgbImg
		outImg = rgbImg
	}

	if !*swapOnlyFlag {
		outImg = image.NewPaletted(bounds, palette)
		if !*chunkedFlag {
			op.Draw(outImg, bounds, inImg, bounds.Min)
			return
		}

		n := *chunksFlag
		cols := math.Ceil(math.Sqrt(float64(n)))
		rows := float64(n) / cols
		w := float64(bounds.Max.X) / cols
		h := float64(bounds.Max.Y) / rows

		wg := sync.WaitGroup{}
		wg.Add(n)

		for y := 0; y < int(rows); y++ {
			for x := 0; x < int(cols); x++ {
				go func(x, y int) {
					defer wg.Done()
					chunk := image.Rectangle{
						Min: image.Point{x * int(w), y * int(h)},
						Max: image.Point{(x + 1) * int(w), (y + 1) * int(h)},
					}
					p := image.Point{x * int(w), y * int(h)}
					op.Draw(outImg, chunk, inImg, p)
				}(x, y)
			}
		}

		wg.Wait()
	}

	switch format {
		case "jpeg": jpeg.Encode(outFile, outImg, nil)
		case "png": png.Encode(outFile, outImg)
	}
}

func check(flagVal *string, name string) {
	if *flagVal == "" {
		fmt.Fprintln(os.Stderr, "Missing flag", name)
		pflag.Usage()
		os.Exit(1)
	}
}

func perr(str ...interface{}) {
	fmt.Fprintln(os.Stderr, str...)
	os.Exit(1)
}

func normalInverter(cl color.Color) color.Color {
	clr, _ := colorful.MakeColor(cl)
	h, s, l := clr.Hsl()

	swap := colorful.Hsl(h, s, 1 - l)
	return swap
}

func grayscaleInverter(cl color.Color) color.Color {
	clr, _ := colorful.MakeColor(cl)
	h, s, l := clr.Hsl()

	r := math.Round(clr.R * 100) / 100
	g := math.Round(clr.B * 100) / 100
	b := math.Round(clr.G * 100) / 100
	diff := math.Abs((r - b) / g)
	deriv := 0.1
	if /* s > 0.5 && */ diff >= deriv {
		return cl
	}

	swap := colorful.Hsl(h, s, 1 - l)
	return swap
}

func strToColor(str string) (color.Color, error) {
	if str[0] == '#' {
		// hexadecimal
		normalStr, err := normalizeHex(str)
		if err != nil {
			return nil, err
		}

		b, err := hex.DecodeString(normalStr)
		if err != nil {
			return nil, err
		}
		
		//                  r     g     b    a
		return color.RGBA{b[0], b[1], b[2], 0xff}, nil
	}

	return nil, errors.New("Invalid format for color")
}

func normalizeHex(hx string) (string, error) {
	switch len(hx) {
		case 4: // #fff
			longHex := ""
	
			for i, c := range hx {
				if i == 0 { continue }
				longHex = longHex + string(c) + string(c)
			}

			return longHex, nil
		case 7: // #ffffff
			return hx[1:], nil
		default:
			return "", errors.New("invalid string for hex")
	}
}

func getDitherAlgo(algo string) (alg draw.Drawer, err error) {
	switch algo {
		case "floydsteinberg": alg = draw.FloydSteinberg // stdlib floyd is more optimized, from reading the code
		case "atkinson": alg = dithering.NewDither(dithering.Atkinson)
		case "jjn", "jarvisjudiceninke": alg = dithering.NewDither(dithering.JarvisJudiceNinke)
		default: err = errors.New("invalid dither algorithm")
	}

	return
}
