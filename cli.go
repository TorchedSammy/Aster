package main

import (
	"fmt"
	"io"
	"image"
	"image/color"
	"os"
	"strconv"
	"strings"

	"github.com/mattn/go-sixel"
	"github.com/chzyer/readline"
	"github.com/TorchedSammy/Aster/script"
)

type cliState struct{
	sourceImage image.Image
	sourceFormat string
	prevImageStates []image.Image
	workingImage image.Image
	palette color.Palette
}

func (s *cliState) pushWorkingImg(img image.Image) {
	s.prevImageStates = append(s.prevImageStates, s.workingImage)
	s.workingImage = img
}

func (s *cliState) undoImg() {
	prevIdx := len(s.prevImageStates) - 1
	prevWorkingImg := s.prevImageStates[prevIdx]
	s.prevImageStates = s.prevImageStates[:prevIdx]

	s.workingImage = prevWorkingImg
}

type userCommand struct{
	name string
	args []string
}

type argDefinition struct{
	typ string
	name string
	defaultVal interface{}
}

type command struct{
	name string
	args []argDefinition
	listener func(...interface{})
}

func runCli() (int, error) {	
	completer := readline.NewPrefixCompleter(
		readline.PcItem("load",
			readline.PcItemDynamic(func(s string) []string {
				fmt.Println(s)

				return []string{}
			}),
		),
	)

	rl, err := readline.NewEx(&readline.Config{
		Prompt: "-> ",
		AutoComplete: completer,
	})
	if err != nil {
		return 1, err
	}

	state := cliState{}
	sixelEncoder := sixel.NewEncoder(os.Stdout)
	interp := script.NewInterp()

	interp.RegisterFunction("hello", script.Fun{
		Caller: func(v []script.Value) []script.Value {
			greet := "world"
			if len(v) > 0 && v[0] != script.EmptyValue && v[0].Kind == script.StringKind {
				greet = v[0].Val
			}

			fmt.Printf("Hello %s!\n", greet)

			return []script.Value{}
		},
	})

	interp.RegisterFunction("prompt", script.Fun{
		Caller: func(v []script.Value) []script.Value {
			if len(v) == 0 {
				fmt.Println("Missing required argument to set prompt")
				return []script.Value{}
			}

			rl.SetPrompt(v[0].Val)
			return []script.Value{}
		},
	})

	interp.RegisterFunction("load", script.Fun{
		Caller: func(v []script.Value) (ret []script.Value) {
			if len(v) == 0 {
				fmt.Println("Missing required path to load image")
				return
			}

			if v[0].Kind != script.StringKind {
				fmt.Println("")
				return
			}

			path := v[0].Val
			inFile, err := os.Open(path)
			if err != nil {
				fmt.Println("Could not open input file")
				return
			}

			inImg, format, err := image.Decode(inFile)
			if err != nil {
				fmt.Println("Could not decode image:", err)
				return
			}

			state.sourceImage = inImg
			state.sourceFormat = format
			state.workingImage = state.sourceImage
			return
		},
	})

	interp.RegisterFunction("palette", script.Fun{
		Caller: func(v []script.Value) (ret []script.Value) {
			if len(v) == 0 {
				// TODO: display palette nicely
				return
			}

			// now we're setting the palette
			var palette color.Palette
			if len(v) == 1 {
				var err error
				palette, err = colorsFromFile(v[0].Val)
				if err != nil {
					fmt.Println(err)
				}
			} else {
				for _, val := range v {
					colorStr := val.Val
					col, err := strToColor(colorStr)
					if err != nil {
						fmt.Fprintln(os.Stderr, "Invalid color", colorStr)
						continue
					}
					palette = append(palette, col)
				}
			}
			state.palette = palette
			return
		},
	})

	interp.RegisterFunction("recolor", script.Fun{
		Caller: func(v []script.Value) (ret []script.Value) {
			res, err := recolor(state.workingImage, state.palette)
			if err != nil {
				fmt.Println(err)
				return
			}

			state.pushWorkingImg(res)
			return
		},
	})

	interp.RegisterFunction("preview", script.Fun{
		Caller: func(v []script.Value) (ret []script.Value) {
			f := 35
			scale := interp.GetGlobal("previewScale")
			fmt.Println(scale)
			if scale.Kind == script.NumberKind {
				f, _ = strconv.Atoi(scale.Val)
			}

			sixelEncoder.Encode(resize(state.workingImage, f))
			return
		},
	})

	interp.RegisterFunction("undo", script.Fun{
		Caller: func(v []script.Value) (ret []script.Value) {
			state.undoImg()
			return
		},
	})

	f, _ := os.Open("test.aster")
	interp.Run(f)

	for {
		line, err := rl.Readline()
		if err != nil {
			if err == io.EOF {
				return 0, nil
			}

			return 2, err
		}

		err = interp.Run(strings.NewReader(line))
		if err != nil {
			fmt.Println(err)
			continue
		}
	}
}
