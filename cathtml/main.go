package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"text/template"

	"github.com/DavidGamba/go-getoptions"
)

var logger = log.New(io.Discard, "", log.LstdFlags)

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
	opt.String("toc-prefix", "", opt.Description("If defined, adds table of contents with the given words as the Prefix + index."))
	opt.Bool("toc-from-file", false, opt.Description("If true, adds table of contents with the filename as the link"))
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
	tocPrefix := opt.Value("toc-prefix").(string)
	tocFromFile := opt.Value("toc-from-file").(bool)
	tocSkip := opt.Value("toc-skip").(int)
	tocTitle := opt.Value("toc-title").(string)

	toc := false
	if tocFromFile || tocPrefix != "" {
		toc = true
	}

	bodyEntries := []FileData{}
	for _, file := range files {
		b, err := os.ReadFile(file)
		if err != nil {
			return err
		}
		fd := FileData{
			Content: b,
			Name:    file,
		}
		bodyEntries = append(bodyEntries, fd)
	}

	data := HTMLData{
		Title:            title,
		Stylesheets:      stylesheets,
		BodyEntries:      bodyEntries,
		BodyEntriesClass: classSeparator,
		TOC:              toc,
		TOCPrefix:        tocPrefix,
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

type FileData struct {
	Content []byte
	Name    string
}

// HTMLData - Struct that holds HTML content
type HTMLData struct {
	Title            string
	Stylesheets      []string
	BodyEntries      []FileData
	BodyEntriesClass string
	TOC              bool
	TOCPrefix        string
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
				{{- if ne $.TOCPrefix "" }}
				<li><a href="#{{ $.TOCPrefix }}_{{ toc_index $i $.TOCSkip }}">{{ $.TOCPrefix }} {{ toc_index $i $.TOCSkip }}</a></li>
				{{- else }}
				<li><a href="#{{ $e.Name }}">{{ $e.Name }}</a></li>
				{{- end -}}
				{{- end -}}
			{{ end }}
			</ul>
		</div>
		<div class="pagebreak"></div>
		{{ end }}
		{{- range $i, $a := .BodyEntries }}
		<div{{with $.BodyEntriesClass }} class="{{ . }}" {{ end }}>
			{{- if ge $i $.TOCSkip }}
			{{- if ne $.TOCPrefix "" }}
			<h2 id="{{ $.TOCPrefix }}_{{ toc_index $i $.TOCSkip }}"><a href="#toc">{{ $.TOCPrefix }} {{ toc_index $i $.TOCSkip }}</a></h2>
			{{- else }}
			<h2 id="{{ .Name }}"><a href="#toc">{{ .Name }}</a></h2>
			{{- end }}
			{{ end }}
			{{ printf "%s" .Content }}
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
