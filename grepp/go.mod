module github.com/DavidGamba/dgtools/grepp

go 1.21

require (
	github.com/DavidGamba/ffind v0.6.1
	github.com/DavidGamba/go-getoptions v0.29.1-0.20240105065605-b7ada72ecac6
	github.com/mgutz/ansi v0.0.0-20200706080929-d51e80ef957d
)

require (
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect

	// workaround for error: //go:linkname must refer to declared function or variable
	golang.org/x/sys v0.16.0 // indirect
)
