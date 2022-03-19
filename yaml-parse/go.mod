module github.com/DavidGamba/dgtools/yaml-parse

require (
	github.com/DavidGamba/dgtools/yamlutils v0.0.0
	github.com/DavidGamba/go-getoptions v0.25.3
)

require (
	github.com/DavidGamba/dgtools/trees v0.0.0 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
)

replace github.com/DavidGamba/dgtools/yamlutils => ../yamlutils

replace github.com/DavidGamba/dgtools/trees => ../trees

go 1.18
