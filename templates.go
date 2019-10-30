package main

import (
	"os"
	"strings"
	"text/template"
)

type Diff struct {
	Filename string
	Contents string
}

type ColorDiffs struct {
	Color string
	DetectedGeneratedFileChange bool
	Diffs []Diff
}

type DiffResultComment struct {
	JobName    string
	ColorDiffs []ColorDiffs
	GenericMap map[string]string
}

func main() {
	val := DiffResultComment{
		GenericMap: map[string]string{
			"c5.xlarge": "1",
			"c3.xlarge": "2",
		},
		JobName: "deployment-updater",
		ColorDiffs: []ColorDiffs{
			{
				Color: "Blue",
				DetectedGeneratedFileChange: true,
				Diffs: []Diff{
					{ Filename: "a/b/c.yml", Contents: strings.TrimSpace("+hi\n-hello\n\n") },
					{ Filename: "a/b/d.yml", Contents: strings.TrimSpace("+hello\n-hihi\n") },
				},
			},
			{
				Color: "Orange",
				DetectedGeneratedFileChange: false,
				Diffs: []Diff{
					{ Filename: "a/b/c.yml", Contents: strings.TrimSpace("+hi\n-hello\n\n") },
					{ Filename: "a/b/d.yml", Contents: strings.TrimSpace("+hello\n-hihi\n") },
				},
			},
		},
	}

	tmpl := template.Must(template.ParseFiles("./tmpl.tmpl"))
	if err := tmpl.Execute(os.Stdout, val); err != nil {
		panic(err)
	}
}
