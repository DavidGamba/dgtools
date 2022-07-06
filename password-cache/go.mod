module github.com/DavidGamba/dgtools/password-cache

go 1.15

require (
	github.com/DavidGamba/dgtools/run v0.6.0 // indirect
	github.com/DavidGamba/go-getoptions v0.21.0
	github.com/jsipprell/keyctl v1.0.4-0.20211208153515-36ca02672b6c
	golang.org/x/crypto v0.0.0-20201002170205-7f63de1d35b0

	// workaround for error: //go:linkname must refer to declared function or variable
	golang.org/x/sys v0.0.0-20220422013727-9388b58f7150 // indirect
)
