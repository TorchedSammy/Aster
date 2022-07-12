package main

import (
	"encoding/hex"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"os"
	"strings"

	// supported formats
	"image/jpeg"
	"image/png"

	"github.com/spf13/pflag"
)

func main() {
	inFlag := pflag.StringP("input", "i", "", "Path to input image")
	outFlag := pflag.StringP("output", "o", "", "Path to output")
	paletteFlag := pflag.StringP("palette", "p", "", "Palette")
	ditherFlag := pflag.BoolP("dither", "d", true, "Whether to use dithering on the image or not")

	pflag.Parse()
	check(inFlag, "input")
	check(outFlag, "output")
	check(paletteFlag, "palette")

	var palette color.Palette
	for _, colorStr := range strings.Split(*paletteFlag, " ") {
		col, err := strToColor(colorStr)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Invalid color", colorStr)
		}
		palette = append(palette, col)
	}

	if len(palette) < 2 {
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
	outImg := image.NewPaletted(bounds, palette)
	if *ditherFlag {
		dither := draw.FloydSteinberg
		dither.Draw(outImg, bounds, inImg, bounds.Min)
	} else {
		draw.Draw(outImg, bounds, inImg, bounds.Min, draw.Src)
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
