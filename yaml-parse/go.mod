module github.com/DavidGamba/dgtools/yaml-parse

require (
	github.com/DavidGamba/dgtools/yamlutils v0.0.0
	github.com/DavidGamba/go-getoptions v0.23.0
)

replace github.com/DavidGamba/dgtools => ../

replace github.com/DavidGamba/dgtools/yamlutils => ../yamlutils

go 1.15
