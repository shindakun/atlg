package main

import (
	"fmt"
	"io/ioutil"

	"github.com/shindakun/atlg/go-api-03/client"
	"github.com/shindakun/envy"
)

const apiurl = "https://api.mailgun.net/v3/youremaildomain.com"

func main() {
	mgKey, err := envy.Get("MGKEY")
	if err != nil {
		panic(err)
	}

	mgc := client.NewMgClient(apiurl, mgKey)

	req, err := mgc.FormatEmailRequest("<Name> some@email.domain",
		"other@email.domain", "Test email", "This is a test email!")
	if err != nil {
		panic(err)
	}

	res, err := mgc.Client.Do(req)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}

	fmt.Println(string(body))
}
