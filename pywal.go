package main

import (
	"encoding/json"
	"image/color"
	"os"
	"os/user"
)

type pywalStore struct{
	Special map[string]string `json:"special"` // dont really think this is needed
	Colors map[string]string `json:"colors"`
}

func pywalColors() (color.Palette, error) {
	curuser, _ := user.Current()
	colorsFile := curuser.HomeDir + "/.cache/wal/colors.json"

	f, err := os.ReadFile(colorsFile)
	if err != nil {
		return nil, err
	}

	var palette color.Palette
	wal := pywalStore{}

	err = json.Unmarshal(f, &wal)
	if err != nil {
		return nil, err
	}

	appendC := func(val string) {
		col, err := strToColor(val)
		if err != nil {
			pwarn("Invalid color", val)
			return
		}
		palette = append(palette, col)
	}

	for _, val := range wal.Colors {
		appendC(val)
	}

	return palette, nil
}
