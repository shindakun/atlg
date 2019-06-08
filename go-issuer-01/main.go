package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type Payload struct {
	Action string `json:"action"`
	Issue  struct {
		URL           string `json:"url"`
		RepositoryURL string `json:"repository_url"`
		Number        int    `json:"number"`
		Title         string `json:"title"`
		Body          string `json:"body"`
	} `json:"issue"`
	Repository struct {
		Name string `json:"name"`
	} `json:"repository"`
}

func status(res http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(res, "Hello!")
}

func handleWebhook(res http.ResponseWriter, req *http.Request) {
	var Payload Payload
	defer req.Body.Close()

	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&Payload)
	if err != nil {
		http.Error(res, "bad request: "+err.Error(), 400)
		log.Printf("bad request: %v", err.Error())
		return
	}

	log.Printf("%#v\n", Payload.Repository.Name)
	log.Printf("%#v\n", Payload.Issue.URL)
	log.Printf("%#v\n", Payload.Issue.Title)
	log.Printf("%#v\n", Payload.Issue.Body)
	log.Printf("%#v\n", Payload.Issue.Number)
	log.Printf("%#v\n", Payload.Issue.RepositoryURL)
}

func main() {
	log.Println("Issuer")

	http.HandleFunc("/", status)
	http.HandleFunc("/webhook", handleWebhook)
	http.ListenAndServe(":3000", nil)
}
