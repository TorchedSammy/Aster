package main

import (
	"fmt"
	"io"
	"image"
	"image/color"
	"os"
	"strings"

	"github.com/mattn/go-sixel"
	"github.com/chzyer/readline"
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

	for {
		line, err := rl.Readline()
		if err != nil {
			if err == io.EOF {
				return 0, nil
			}

			return 2, err
		}

		cmd := parseLine(line)
		switch cmd.name {
			case "hello":
				if len(cmd.args) == 0 {
					fmt.Println("Hello world!")
				} else {
					fmt.Printf("Hello %s!", cmd.args[1])
				}
			case "prompt":
				if len(cmd.args) == 0 {
					fmt.Println("Missing required argument to set prompt")
					continue
				}

				rl.SetPrompt(cmd.args[0])
			case "load":
				if len(cmd.args) == 0 {
					fmt.Println("Missing required path to load image")
				}

				path := cmd.args[0]
				inFile, err := os.Open(path)
				if err != nil {
					fmt.Println("Could not open input file")
					continue
				}

				inImg, format, err := image.Decode(inFile)
				if err != nil {
					fmt.Println("Could not decode image:", err)
					continue
				}

				state.sourceImage = inImg
				state.sourceFormat = format
				state.workingImage = state.sourceImage
			case "palette":
				if len(cmd.args) == 0 {
					// TODO: display palette nicely
					continue
				}

				// now we're setting the palette
				var palette color.Palette
				if len(cmd.args) == 1 {
					var err error
					palette, err = colorsFromFile(cmd.args[0])
					if err != nil {
						fmt.Println(err)
					}
				} else {
					for _, colorStr := range cmd.args {
						col, err := strToColor(colorStr)
						if err != nil {
							fmt.Fprintln(os.Stderr, "Invalid color", colorStr)
							continue
						}
						palette = append(palette, col)
					}
				}
				state.palette = palette
			case "recolor":
				var dither bool
				if len(cmd.args) != 0 {
					if cmd.args[0] == "@dither" {
						dither = true
					}
				}

				var res image.Image
				var err error
				if dither {
					res, err = recolorDither(state.workingImage, state.palette, "floydsteinberg")
				} else {
					res, err = recolor(state.workingImage, state.palette)
				}

				if err != nil {
					fmt.Println(err)
					continue
				}
				state.pushWorkingImg(res)
			case "preview":
				sixelEncoder.Encode(resize(state.workingImage, 40))
			case "undo":
				state.undoImg()
		}
	}
}

func parseLine(line string) userCommand {
	line = strings.TrimSpace(line)

	fields := split(line)
	cmd := userCommand{}

	args := []string{}
	for i, field := range fields {
		if i == 0 {
			cmd.name = strings.ToLower(field)
			continue
		}

		args = append(args, field)
	}
	cmd.args = args

	return cmd
}

// stolen from hilbish
func split(str string) []string {
	strParts := []string{}
	sb := &strings.Builder{}
	quoted := false

	for i, r := range str {
		if r == '"' {
			quoted = !quoted
		} else if r == ' ' && str[i - 1] == '\\' {
			sb.WriteRune(r)
		} else if !quoted && r == ' ' {
			strParts = append(strParts, sb.String())
			sb.Reset()
		} else {
			sb.WriteRune(r)
		}
	}
	if strings.HasSuffix(str, " ") {
		strParts = append(strParts, "")
	}

	if sb.Len() > 0 {
		strParts = append(strParts, sb.String())
	}

	return strParts
}
