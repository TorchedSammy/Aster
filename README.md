# Aster
> 🌼 Command line image colorizer utility.

Aster is a simple command line tool to recolor images into a specific palette.  
It is based on [felix-u's imgclr tool](https://github.com/felix-u/imgclr);
This is an alternate version written in Go.

| Original                           | Recolored                     |
| ---------------------------------- | ----------------------------- |
| ![](samples/ghbanner/original.jpg) | ![](samples/ghbanner/res.jpg) |

> Made with this palette: *#0e1112 #181d1f #212629 #35383b #4e5256 #666b70 #181d1f #7f9aa3 #1c2124*

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
