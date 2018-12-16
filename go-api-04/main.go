package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"text/template"
)

type Users []struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

func main() {
	APIURL := "https://jsonplaceholder.typicode.com/users"
	req, err := http.NewRequest(http.MethodGet, APIURL, nil)
	if err != nil {
		panic(err)
	}
	client := http.DefaultClient
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	var u Users
	json.Unmarshal(body, &u)

	msgText := "To: {{.Email}}\nHi {{.Name}}! There is a new post!\n\n\n"
	t := template.Must(template.New("msg").Parse(msgText))

	for _, r := range u {
		err := t.Execute(os.Stdout, r)
		if err != nil {
			panic(err)
		}
	}
}
