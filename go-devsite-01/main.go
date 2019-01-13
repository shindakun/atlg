package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"strings"

	yaml "gopkg.in/yaml.v2"
)

const delim = "---"

type post struct {
	title       string
	published   bool
	description string
	tags        []string
	coverImage  string
	series      string
}

func loadFile(s string) (b []byte, err error) {
	f, err := ioutil.ReadFile(s)
	if err != nil {
		return nil, err
	}
	return f, nil
}

func main() {
	f, err := loadFile("test.md")
	if err != nil {
		panic(err)
	}

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
	p.title = m["title"].(string)
	p.published = m["published"].(bool)
	p.description = m["description"].(string)

	// TODO: Strip space after comma prior to parse?
	tmp := m["tags"].(string)
	p.tags = strings.Split(tmp, ", ")

	p.coverImage = m["cover_image"].(string)
	p.series = m["series"].(string)

	fmt.Printf("%#v\n", p)
}
