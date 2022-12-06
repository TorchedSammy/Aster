package main

import (
	"bufio"
	"image/color"
	"io"
	"os"
	"os/user"
	"strings"
)

func xresourcesColors() (color.Palette, error) {
	curuser, _ := user.Current()
	colorsFile := curuser.HomeDir + "/.Xresources"

	f, err := os.Open(colorsFile)
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

		if strings.HasPrefix(line, "*.color") {
			parts := strings.Split(line, ":")
			col, err := strToColor(strings.TrimSpace(parts[1]))
			if err != nil {
				pwarn("Invalid color", parts[1])
				continue
			}
			palette = append(palette, col)
		}
	}

	return palette, nil
}
