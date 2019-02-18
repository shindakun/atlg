package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/microcosm-cc/bluemonday"
	"github.com/russross/blackfriday"
	yaml "gopkg.in/yaml.v2"
)

const delimiter = "---"

type post struct {
	Title       string
	Published   bool
	Description string
	Tags        []string
	CoverImage  string
	Series      string
	PostBody    template.HTML
}

type index struct {
	Pages []Page
}

type Page struct {
	FileName string
	Title    string
}

var indexTempl = `<!DOCTYPE html>
<html lang="en">
  <head>
    <title>shindakun's dev site</title>
    <meta charset="utf-8">
    <meta http-equiv="X-UA-Compatible" content="IE=edge">
    <meta name="HandheldFriendly" content="True">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <meta name="referrer" content="no-referrer-when-downgrade" />
    <meta name="description" content="shindakun's dev site" />
  </head>
	<body>
		<div class="index">
		{{ range $key, $value := .Pages }}
			<a href="/{{ $value.FileName }}">{{ $value.Title }}</a>
    {{ end }}
		</div>
	</body>
	</html>
`

var postTempl = `<!DOCTYPE html>
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

func getContentsOf(r io.Reader) ([]byte, error) {
	return ioutil.ReadAll(r)
}

func parseFrontMatter(b []byte) (map[string]interface{}, error) {
	fm := make(map[string]interface{})
	err := yaml.Unmarshal(b, &fm)
	if err != nil {
		msg := fmt.Sprintf("error: %v\ninput:\n%s", err, b)
		return nil, fmt.Errorf(msg)
	}
	return fm, nil
}

func splitData(fm []byte, delimiter string) ([][]byte, error) {
	b := bytes.Split(fm, []byte(delimiter))
	if len(b) < 3 || len(b[0]) != 0 {
		return nil, fmt.Errorf("Front matter is damaged")
	}
	return b, nil
}

// makePost creates the post struct, returns that and the template HTML
func makePost(fm map[string]interface{}, contents []byte, s [][]byte) (*template.Template, *post, bool) {
	post := &post{}

	post.Published = false
	pubIntf, ok := fm["published"]
	if ok {
		if published, ok := pubIntf.(bool); ok {
			post.Published = published
		}
	}

	if !post.Published {
		return nil, nil, true
	}

	post.Title = ""
	titleIntf, ok := fm["title"]
	if ok {
		if title, ok := titleIntf.(string); ok {
			post.Title = title
		}
	}

	if post.Title == "" {
		return nil, nil, true
	}

	post.Description = ""
	descIntf, ok := fm["description"]
	if ok {
		if description, ok := descIntf.(string); ok {
			post.Description = description
		}
	}

	post.Tags = []string{}
	tagsIntf, ok := fm["tags"]
	if ok {
		if tags, ok := tagsIntf.(string); ok {
			post.Tags = strings.Split(tags, ", ")
		}
	}

	post.CoverImage = ""
	covIntf, ok := fm["cover_image"]
	if ok {
		if coverImage, ok := covIntf.(string); ok {
			post.CoverImage = coverImage
		}
	}

	post.Series = ""
	seriesIntf, ok := fm["series"]
	if ok {
		if series, ok := seriesIntf.(string); ok {
			post.Series = series
		}
	}

	pBody := contents[len(s[1])+(len(delimiter)*2):]

	bf := blackfriday.Run(pBody)

	bm := bluemonday.UGCPolicy()
	bm.AllowAttrs("class").Matching(regexp.MustCompile("^language-[a-zA-Z0-9]+$")).OnElements("code")
	post.PostBody = template.HTML(bm.SanitizeBytes(bf))

	tm := template.Must(template.New("post").Parse(postTempl))
	return tm, post, false
}

func writeIndex(idx index, destination string) error {
	indexFile, err := os.Create(destination + "/" + "index.html")
	if err != nil {
		return err
	}
	defer indexFile.Close()

	buffer := bufio.NewWriter(indexFile)
	tm := template.Must(template.New("index").Parse(indexTempl))
	err = tm.Execute(buffer, idx)
	if err != nil {
		return err
	}
	buffer.Flush()
	return nil
}

func main() {
	var idx index

	destination := flag.String("destination", "", "destination folder")
	source := flag.String("source", "", "source directory")

	flag.Parse()

	if _, err := os.Stat(*destination); os.IsNotExist(err) {
		err := os.Mkdir(*destination, 0777)
		if err != nil {
			panic(err)
		}
	} else {
		log.Panicf("error: destination '%s' already exists", *destination)
	}

	_, err := ioutil.ReadDir(*destination)
	if err != nil {
		panic(err)
	}

	srcDir, err := ioutil.ReadDir(*source)
	if err != nil {
		panic(err)
	}

	for _, file := range srcDir {
		if fileName := file.Name(); strings.HasSuffix(fileName, ".md") {

			openedFile, err := os.Open(fileName)
			if err != nil {
				log.Println(fileName, err)
				continue
			}

			contents, err := getContentsOf(openedFile)
			if err != nil {
				openedFile.Close()
				log.Println(fileName, err)
				continue
			}
			openedFile.Close()

			s, err := splitData(contents, delimiter)
			if err != nil {
				log.Println(fileName, err)
				continue
			}

			fm, err := parseFrontMatter(s[1])
			if err != nil {
				msg := fmt.Sprintf("%v\n", err)
				log.Println(fileName, msg)
				continue
			}

			template, post, skip := makePost(fm, contents, s)
			if !skip {
				trimmedName := strings.TrimSuffix(fileName, ".md")
				outputFile, err := os.Create(*destination + "/" + trimmedName + ".html")
				if err != nil {
					log.Println(err)
					continue
				}

				buffer := bufio.NewWriter(outputFile)
				err = template.Execute(buffer, post)
				if err != nil {
					panic(err)
				}

				buffer.Flush()
				outputFile.Close()

				indexLinks := Page{
					FileName: trimmedName + ".html",
					Title:    post.Title,
				}
				idx.Pages = append(idx.Pages, indexLinks)
			}
		}
	}

	if len(idx.Pages) > 0 {
		err := writeIndex(idx, *destination)
		if err != nil {
			log.Println(err)
		}
	}
}
