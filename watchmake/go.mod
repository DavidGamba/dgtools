module github.com/DavidGamba/dgtools/watchmake

go 1.16

require (
	github.com/DavidGamba/dgtools/private/hclutils v0.0.0
	github.com/DavidGamba/dgtools/run v0.3.0
	github.com/DavidGamba/go-getoptions v0.23.0
	github.com/davecgh/go-spew v1.1.1
	github.com/fsnotify/fsnotify v1.4.9
	github.com/google/go-cmp v0.4.0 // indirect
	github.com/hashicorp/hcl/v2 v2.9.0
	github.com/kylelemons/godebug v1.1.0 // indirect
	github.com/zclconf/go-cty v1.8.0
)

replace github.com/DavidGamba/dgtools/private/hclutils => ../private/hclutils
