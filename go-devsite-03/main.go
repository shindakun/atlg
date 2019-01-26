package main

import (
	"bufio"
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

func getContents(f *string) ([]byte, error) {
	b, err := ioutil.ReadFile(*f)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func parseFM(b *[]byte) (map[string]interface{}, error) {
	m := make(map[string]interface{})
	err := yaml.Unmarshal(*b, &m)
	if err != nil {
		msg := fmt.Sprintf("error: %v\ninput:\n%s", err, b)
		return nil, fmt.Errorf(msg)
	}
	return m, nil
}

func isNil(i interface{}) bool {
	if i != nil {
		return false
	}
	return true
}

func splitData(f *[]byte) ([][]byte, error) {
	b := bytes.Split(*f, []byte(delim))
	if len(b) < 3 || len(b[0]) != 0 {
		return nil, fmt.Errorf("Front matter is damaged")
	}
	return b, nil
}

// makePost creates the post struct, returns that and the template HTML
func makePost(fm map[string]interface{}, contents []byte,
	s [][]byte) (*template.Template, *post) {
	p := &post{}

	if isNil(fm["title"]) {
		panic("isNil tripped at title")
	} else {
		p.Title = fm["title"].(string)
	}
	p.Published = fm["published"].(bool)
	p.Description = fm["description"].(string)

	// TODO: Strip space after comma prior to parse?
	tmp := fm["tags"].(string)
	p.Tags = strings.Split(tmp, ", ")

	p.CoverImage = fm["cover_image"].(string)
	p.Series = fm["series"].(string)

	pBody := contents[len(s[1])+(len(delim)*2):]

	out := blackfriday.Run(pBody)

	bm := bluemonday.UGCPolicy()
	bm.AllowAttrs("class").Matching(regexp.MustCompile("^language-[a-zA-Z0-9]+$")).OnElements("code")
	p.PostBody = template.HTML(bm.SanitizeBytes(out))

	tm := template.Must(template.New("msg").Parse(templ))
	return tm, p
}

func main() {
	d, err := ioutil.ReadDir(".")
	if err != nil {
		panic(err)
	}

	for _, f := range d {
		if t := f.Name(); strings.HasSuffix(t, ".md") {
			contents, err := getContents(&t)
			if err != nil {
				panic(err)
			}
			s, err := splitData(&contents)
			if err != nil {
				panic(err)
			}

			fm, err := parseFM(&s[1])
			if err != nil {
				msg := fmt.Sprintf("error: %v\ninput:\n%s", err, s[1])
				panic(msg)
			}

			tm, p := makePost(fm, contents, s)
			fin := strings.TrimSuffix(t, ".md")
			o, err := os.Create(fin + ".html")
			if err != nil {
				panic(err)
			}
			defer o.Close()

			buf := bufio.NewWriter(o)

			err = tm.Execute(buf, p)
			if err != nil {
				panic(err)
			}
			buf.Flush()
		}
	}
}
