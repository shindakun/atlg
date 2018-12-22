package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"text/template"

	c "github.com/shindakun/atlg/go-api-03/client"
	"github.com/shindakun/envy"
)

const apiurl = "https://api.mailgun.net/v3/youremaildomain.com"

type Users []struct {
	Name     string `json:"name"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

func checkStatusCode(s int) error {
	if s != 200 {
		err := fmt.Sprintf("unexpected status code: %d", s)
		return errors.New(err)
	}
	return nil
}

func sendEmail(buf bytes.Buffer, email string) {
	mgKey, err := envy.Get("MGKEY")
	if err != nil {
		panic(err)
	}

	mgc := c.NewMgClient(apiurl, mgKey)

	req, err := mgc.FormatEmailRequest(email, "youremailaddress@",
		"Test email", buf.String())
	if err != nil {
		panic(err)
	}

	res, err := mgc.Client.Do(req)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()

	if err = checkStatusCode(res.StatusCode); err != nil {
		panic(err)
	}
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}

	fmt.Println(string(body))
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

	if err = checkStatusCode(resp.StatusCode); err != nil {
		panic(err)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	var u Users
	json.Unmarshal(body, &u)

	msgText := "Hi {{.Name}}! Or should I call you, {{.Username}}? There is a new post!\n\n\n"
	t := template.Must(template.New("msg").Parse(msgText))

	for _, v := range u {
		var buf bytes.Buffer

		err := t.Execute(&buf, v)
		if err != nil {
			panic(err)
		}
		sendEmail(buf, v.Email)
	}
}
