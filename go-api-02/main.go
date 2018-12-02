package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

// HTTPBinResponse struct
type HTTPBinResponse struct {
	Args    map[string]string `json:"args"`
	Headers map[string]string `json:"headers"`
	Origin  string            `json:"origin"`
	URL     string            `json:"url"`
}

func main() {
	APIURL := "https://httpbin.org/get?arg1=one&arg2=two"
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

	var r HTTPBinResponse
	json.Unmarshal(body, &r)

	fmt.Printf("%#v\n\n", r)
	fmt.Printf("%v\n\n", r.Headers["User-Agent"])
	fmt.Printf("%v\n", r.Args["arg2"])
}
