package main

import (
	"bufio"
	"image/color"
	"io"
	"os"
	"strings"
)

func colorsFromFile(path string) (color.Palette, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	var palette color.Palette
	reader := bufio.NewReader(f)

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}

		line = strings.Trim(line, "\n")
		col, err := strToColor(line)
		if err != nil {
			pwarn("Invalid color", line)
			continue
		}
		palette = append(palette, col)
	}

	return palette, nil
}
