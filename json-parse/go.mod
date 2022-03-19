module github.com/DavidGamba/dgtools/json-parse

go 1.18

require (
	github.com/DavidGamba/dgtools/jsonutils v0.0.0
	github.com/DavidGamba/go-getoptions v0.25.3
)

require github.com/DavidGamba/dgtools/trees v0.0.0 // indirect

replace github.com/DavidGamba/dgtools/jsonutils => ../jsonutils

replace github.com/DavidGamba/dgtools/trees => ../trees
