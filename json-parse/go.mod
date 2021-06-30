module github.com/DavidGamba/dgtools/json-parse

require (
	github.com/DavidGamba/dgtools v0.0.0
	github.com/DavidGamba/dgtools/jsonutils v0.0.0
	github.com/DavidGamba/go-getoptions v0.21.0
)

replace github.com/DavidGamba/dgtools => ../
replace github.com/DavidGamba/dgtools/jsonutils => ../jsonutils

go 1.16
