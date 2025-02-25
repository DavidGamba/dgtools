@extern(embed)

package file5

import "strings"

en: "hello"

_es: _ @embed(file=hola.txt,type=text)
es:  strings.TrimSpace(_es)
