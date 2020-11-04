package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"text/template"

	"github.com/DavidGamba/go-getoptions"
)

var logger = log.New(ioutil.Discard, "", log.LstdFlags)

func main() {
	os.Exit(program(os.Args))
}

func program(args []string) int {
	opt := getoptions.New()
	opt.Self("", "Allows concatenating multiple HTML file portions into a single one")
	opt.Bool("help", false, opt.Alias("?"))
	opt.Bool("debug", false, opt.GetEnv("DEBUG"))
	opt.String("title", "")
	opt.String("class-separator", "")
	opt.StringSlice("stylesheets", 1, 99)
	opt.StringSlice("files", 1, 99)
	remaining, err := opt.Parse(args[1:])
	if opt.Called("help") {
		fmt.Println(opt.Help())
		return 1
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
		return 1
	}
	if opt.Called("debug") {
		logger.SetOutput(os.Stderr)
	}
	logger.Println(remaining)
	err = realMain(opt)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
		return 1
	}
	return 0
}

func realMain(opt *getoptions.GetOpt) error {
	title := opt.Value("title").(string)
	stylesheets := opt.Value("stylesheets").([]string)
	files := opt.Value("files").([]string)
	classSeparator := opt.Value("class-separator").(string)

	bodyEntries := [][]byte{}
	for _, file := range files {
		b, err := ioutil.ReadFile(file)
		if err != nil {
			return err
		}
		bodyEntries = append(bodyEntries, b)
	}

	data := HTMLData{
		Title:            title,
		Stylesheets:      stylesheets,
		BodyEntries:      bodyEntries,
		BodyEntriesClass: classSeparator,
	}

	out, err := htmlOutput(data)
	if err != nil {
		return err
	}
	fmt.Println(out)

	return nil
}

// HTMLData - Struct that holds HTML content
type HTMLData struct {
	Title            string
	Stylesheets      []string
	BodyEntries      [][]byte
	BodyEntriesClass string
}

func htmlOutput(data HTMLData) (string, error) {
	const tpl = `
<!DOCTYPE html>
<html>
	<head>
		<meta charset="UTF-8">
		{{with .Title}}<title>{{ .Title }}</title>{{end}}
		{{ range .Stylesheets }}<link rel="stylesheet" type="text/css" href="{{ . }}"/>
		{{ end }}
	</head>
	<body>
		{{ range .BodyEntries }}<div{{with $.BodyEntriesClass }} class="{{ . }}" {{ end }}>
			{{ printf "%s" . }}
		</div>
		{{ end }}
	</body>
</html>
`
	t, err := template.New("tpl").Parse(tpl)
	if err != nil {
		return "", err
	}
	out := ""
	buf := bytes.NewBufferString(out)
	err = t.Execute(buf, data)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}
