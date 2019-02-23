package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"reflect"
	"sort"
	"strings"
)

func plainList(m map[string][]string, v []string) {
	for _, value := range v {
		for _, file := range m[value] {
			fmt.Println(file)
		}
	}
}

func nestedList(m map[string][]string, v []string) {
	for i, value := range v {
		fmt.Println(v[i])
		for _, file := range m[value] {
			fmt.Println(" - ", file)
		}
	}
}

func jsonList(m map[string][]string) {
	j, err := json.Marshal(m)
	if err != nil {
		log.Panicf("Error marshalling JSON. %s", err)
	}
	fmt.Printf("%s", j)
}

func main() {
	wd, err := os.Getwd()
	if err != nil {
		msg := fmt.Sprintf("An error occured getting the current working directory.\n%s", err)
		fmt.Println(msg)
		os.Exit(1)
	}

	dir, err := ioutil.ReadDir(wd)
	if err != nil {
		msg := fmt.Sprintf("An error occured reading the current working directory.\n%s", err)
		fmt.Println(msg)
		os.Exit(1)
	}

	var m = make(map[string][]string)
	for _, file := range dir {
		if !file.IsDir() {
			fileName := file.Name()
			ext := strings.Split(fileName, ".")
			if len(ext) > 1 {
				m[ext[len(ext)-1]] = append(m[ext[len(ext)-1]], fileName)
			}
			if len(ext) == 1 {
				m["no-ext"] = append(m["no-ext"], fileName)
			}
			sort.Strings(m[ext[len(ext)-1]])
		}
	}
	values := reflect.ValueOf(m).MapKeys()

	var extensions []string
	for _, value := range values {
		extensions = append(extensions, value.String())
	}
	sort.Strings(extensions)

	if len(os.Args) > 1 {
		switch arg := os.Args[1]; arg {
		case "plain":
			plainList(m, extensions)
		case "nested":
			nestedList(m, extensions)
		case "json":
			jsonList(m)
		default:
			fmt.Println("Usage: gls [plain|nested|json]")
		}
	} else {
		nestedList(m, extensions)
	}
}
