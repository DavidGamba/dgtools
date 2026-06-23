@extern(embed)
package myPackage

import "strings"

// NOTE: calls to embed can't be done from a hidden file
_hola: _ @embed(file=hola.txt,type=text)
hola:  strings.TrimSpace(_hola)
