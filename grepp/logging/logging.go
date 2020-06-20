/*
package logging - Wrapper around log
*/
package logging

import (
	"io"
	"log"
)

var (
	Trace   *log.Logger
	Debug   *log.Logger
	Info    *log.Logger
	Warning *log.Logger
	Error   *log.Logger
)

func LogInit(th, dh, ih, wh, eh io.Writer) {
	Trace = log.New(th, "TRACE: ", log.Ldate|log.Ltime|log.Lshortfile)
	Debug = log.New(dh, "DEBUG: ", log.Ldate|log.Ltime|log.Lshortfile)
	Info = log.New(ih, "", 0)
	Warning = log.New(wh, "WARNING: ", 0)
	Error = log.New(eh, "ERROR: ", 0)
}
