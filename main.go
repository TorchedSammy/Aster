package main

import (
	"encoding/hex"
	"errors"
	"fmt"
	"image"
	"image/color"
	"golang.org/x/image/draw"
	"math"
	"os"
	"strings"

	"github.com/spf13/pflag"
	"github.com/lucasb-eyer/go-colorful"
	"github.com/brouxco/dithering"
)

type imageProcessor interface{
	Decode(*os.File)
	Swap(inverter)
	Colorize(bool, draw.Drawer)
	Write(*os.File)
}

type processImage struct{
	imgs []image.Image
	processed []image.Image
	format string
}
type inverter func(color.Color) color.Color
var queue []imageProcessor
var palette color.Palette

func main() {
	inFlag := pflag.StringP("input", "i", "", "Path to input image")
	outFlag := pflag.StringP("output", "o", "", "Path to output")
	paletteFlag := pflag.StringP("palette", "p", "", "Palette of the image")
	paletteFileFlag := pflag.StringP("paletteFile", "P", "", "Path to a file which contains newline delimited colors")
	ditherFlag := pflag.BoolP("dither", "d", true, "Whether to use dithering on the image or not")
	ditherAlgoFlag := pflag.StringP("ditherAlgorithm", "D", "floydsteinberg", "The dithering algorithm to use.")
	swapFlag := pflag.BoolP("swap", "s", false, "Swap luminance of image before colorizing")
	swapOnlyFlag := pflag.BoolP("swapOnly", "S", false, "Only swap luminance and dont colorize. This implies -s (luminance swap)")
	grayscaleSwapFlag := pflag.BoolP("grayscaleSwap", "g", false, "Only invert parts of the image that are calculated to be grayscale (blacks/whites)")
	pywalFlag := pflag.BoolP("pywal", "w", false, "Use pywal colors as the palette")
	xresourcesFlag := pflag.BoolP("xresources", "x", false, "Use colors from Xresources as the palette")
	cliFlag := pflag.BoolP("cli", "c", false, "Use the Aster command line shell")

	pflag.Parse()

	if *cliFlag {
		exit, err := runCli()
		if err != nil {
			fmt.Println(err)
		}
		os.Exit(exit)
	}
	check(inFlag, "input")
	check(outFlag, "output")
//	check(paletteFlag, "palette")

	if *swapOnlyFlag {
		f := true
		swapFlag = &f
	}

	var dither draw.Drawer
	if *ditherFlag {
		var err error
		dither, err = getDitherAlgo(*ditherAlgoFlag)
		if err != nil {
			perr("Invalid dither algorithm", *ditherAlgoFlag)
		}
	}

	if *paletteFileFlag != "" {
		var err error
		palette, err = colorsFromFile(*paletteFileFlag)
		if err != nil {
			perr(err.Error())
		}
	}

	if *xresourcesFlag {
		var err error
		palette, err = xresourcesColors()
		if err != nil {
			perr(err.Error())
		}
	}

	if *pywalFlag {
		var err error
		palette, err = pywalColors()
		if err != nil {
			perr(err.Error())
		}
	}

	if len(*paletteFlag) != 0 && len(palette) == 0 {
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

	var colorInverter inverter = normalInverter
	if *grayscaleSwapFlag {
		colorInverter = grayscaleInverter
	}

	var processor imageProcessor
	splits := strings.Split(*inFlag, ".")
	ext := splits[len(splits) - 1]
	switch ext {
		case "gif": processor = &gifProcessor{}
		case "png", "jpg", "jpeg": processor = &singleImageProcessor{}
	}
	queue := append(queue, processor)

	for _, im := range queue {
		im.Decode(inFile)

		if *swapFlag {
			im.Swap(colorInverter)
		}

		if !*swapOnlyFlag {
			im.Colorize(*ditherFlag, dither)
		}

		im.Write(outFile)
	}
}

func check(flagVal *string, name string) {
	if *flagVal == "" {
		fmt.Fprintln(os.Stderr, "Missing flag", name)
		pflag.Usage()
		os.Exit(1)
	}
}

func pwarn(str ...interface{}) {
	fmt.Fprintln(os.Stderr, str...)
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
	switch strings.ToLower(algo) {
		case "floydsteinberg": alg = draw.FloydSteinberg // stdlib floyd is more optimized, from reading the code
		case "atkinson": alg = dithering.NewDither(dithering.Atkinson)
		case "jjn", "jarvisjudiceninke": alg = dithering.NewDither(dithering.JarvisJudiceNinke)
		default: err = fmt.Errorf("invalid dither algorithm %s", algo)
	}

	return
}
