package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

func createDestination(s string) error {
	if _, err := os.Stat(s); os.IsNotExist(err) {
		err := os.Mkdir(s, 0777)
		if err != nil {
			return err
		}
	} else {
		// already exists, probably, maybe, hopefully
		return nil
	}
	return nil
}

func doMove(file, dir string) error {
	err := os.Rename(file, dir+"/"+file)
	if err != nil {
		return err
	}
	return nil
}

func prepAndMove(files map[string][]string) {
	for i, list := range files {
		switch i {
		case "text", "txt", "md":
			dir := "documents"
			err := createDestination(dir)
			if err != nil {
				msg := fmt.Sprintf("An error occured creating destiation.\n%s", err)
				fmt.Println(msg)
				os.Exit(1)
			}
			for j := range list {
				doMove(list[j], dir)
			}
		case "png", "jpg", "gif", "webp":
			dir := "images"
			err := createDestination(dir)
			if err != nil {
				msg := fmt.Sprintf("An error occured creating destiation.\n%s", err)
				fmt.Println(msg)
				os.Exit(1)
			}
			for j := range list {
				doMove(list[j], dir)
			}
		}
	}
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
		}
	}

	prepAndMove(m)
}
