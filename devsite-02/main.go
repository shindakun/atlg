package main

import (
	"bytes"
	"fmt"
	"html/template"
	"io/ioutil"
	"os"
	"regexp"
	"strings"

	"github.com/microcosm-cc/bluemonday"
	"github.com/russross/blackfriday"
	yaml "gopkg.in/yaml.v2"
)

const delim = "---"

type post struct {
	Title       string
	Published   bool
	Description string
	Tags        []string
	CoverImage  string
	Series      string
	PostBody    template.HTML
}

var templ = `<!DOCTYPE html>
<html lang="en">
  <head>
    <title>{{.Title}}</title>
    <meta charset="utf-8">
    <meta http-equiv="X-UA-Compatible" content="IE=edge">
    <meta name="HandheldFriendly" content="True">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <meta name="referrer" content="no-referrer-when-downgrade" />
    <meta name="description" content="{{.Description}}" />
  </head>
	<body>
		<div class="post">
			<h1>{{.Title}}</h1>
			{{.PostBody}}
		</div>
	</body>
	</html>
	`

func loadFile(s string) (b []byte, err error) {
	f, err := ioutil.ReadFile(s)
	if err != nil {
		return nil, err
	}
	return f, nil
}

func isNil(i interface{}) bool {
	if i != nil {
		return false
	}
	return true
}

func main() {
	f, err := loadFile("test.md")
	if err != nil {
		panic(err)
	}

	// Not the best test
	b := bytes.Split(f, []byte(delim))
	if len(b) < 3 || len(b[0]) != 0 {
		panic(fmt.Errorf("Front matter is damaged"))
	}

	m := make(map[string]interface{})
	err = yaml.Unmarshal([]byte(b[1]), &m)
	if err != nil {
		msg := fmt.Sprintf("error: %v\ninput:\n%s", err, b[1])
		panic(msg)
	}

	p := &post{}

	if isNil(m["title"]) {
		panic(err)
	} else {
		p.Title = m["title"].(string)
	}
	p.Published = m["published"].(bool)
	p.Description = m["description"].(string)

	// TODO: Strip space after comma prior to parse?
	tmp := m["tags"].(string)
	p.Tags = strings.Split(tmp, ", ")

	p.CoverImage = m["cover_image"].(string)
	p.Series = m["series"].(string)

	pBody := f[len(b[1])+(len(delim)*2):]

	out := blackfriday.Run(pBody)

	bm := bluemonday.UGCPolicy()
	bm.AllowAttrs("class").Matching(regexp.MustCompile("^language-[a-zA-Z0-9]+$")).OnElements("code")
	p.PostBody = template.HTML(bm.SanitizeBytes(out))
	// p.PostBody = template.HTML(bluemonday.UGCPolicy().SanitizeBytes(out))

	t := template.Must(template.New("msg").Parse(templ))

	err = t.Execute(os.Stdout, p)
	if err != nil {
		panic(err)
	}
}
