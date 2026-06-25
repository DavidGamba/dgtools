@extern(embed)
package myPackage

import "strings"

// NOTE: calls to embed from a hidden file require CUE v0.18.0+
_hola: _ @embed(file=hola.txt,type=text)
hola:  strings.TrimSpace(_hola)
