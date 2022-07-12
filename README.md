# Aster
> 🌼 Command line image colorizer utility.

Aster is a simple command line tool to recolor images into a specific palette.  
It is based on [felix-u's imgclr tool](https://github.com/felix-u/imgclr);
this is somewhat a Go rewrite.

Work in progress!

# Features
- [x] Recolor images
## Formats
- [x] JPEG
- [ ] PNG
- [ ] GIF

# Install
`go install github.com/TorchedSammy/Aster`

# Usage
`aster -i in.jpg -o out.jpg -p "#fff #000"` will recolor `in.jpg` to black
and white, and write the result to `out.jpg`.

# License
MIT
