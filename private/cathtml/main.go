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
	opt.String("title", "", opt.Alias("t"))
	opt.String("class-separator", "")
	opt.String("toc", "", opt.Description("If defined, adds table of contents with the given words as the Title + index."))
	opt.String("toc-title", "Table Of Contents", opt.Description(""))
	opt.Int("toc-skip", 0)
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
	toc := opt.Value("toc").(string)
	tocSkip := opt.Value("toc-skip").(int)
	tocTitle := opt.Value("toc-title").(string)

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
		TOC:              toc,
		TOCSkip:          tocSkip,
		TOCTitle:         tocTitle,
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
	TOC              string
	TOCSkip          int
	TOCTitle         string
}

func htmlOutput(data HTMLData) (string, error) {
	const tpl = `
<!DOCTYPE html>
<html>
	<head>
		<meta charset="UTF-8">
			{{- with .Title }}<title>{{ . }}</title>{{ end }}
		{{ range .Stylesheets }}<link rel="stylesheet" type="text/css" href="{{ . }}"/>
		{{- end }}
	</head>
	<body>
		{{- with .TOC -}}
		<div id="toc" class="toc_container" >
			<h2 class="toc_title">{{ $.TOCTitle }}</h2>
			<ul class="toc_list">
			  {{- range $i, $e := $.BodyEntries }}
				{{- if ge $i $.TOCSkip }}
				<li><a href="#{{ $.TOC }}_{{ toc_index $i $.TOCSkip }}">{{ $.TOC }} {{ toc_index $i $.TOCSkip }}</a></li>
				{{- end -}}
			{{ end }}
			</ul>
		</div>
		<div class="pagebreak"></div>
		{{ end }}
		{{- range $i, $a := .BodyEntries }}
		<div{{with $.BodyEntriesClass }} class="{{ . }}" {{ end }}>
			{{- if ge $i $.TOCSkip }}
			<h2 id="{{ $.TOC }}_{{ toc_index $i $.TOCSkip }}"><a href="#toc">{{ $.TOC }} {{ toc_index $i $.TOCSkip }}</a></h2>
			{{ end }}
			{{ printf "%s" . }}
		</div>
		<div class="pagebreak"></div>
		{{ end }}
	</body>
</html>
`
	t, err := template.New("tpl").
		Funcs(template.FuncMap{
			"toc_index": func(a, b int) int { return a - b + 1 },
		}).Parse(tpl)
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
