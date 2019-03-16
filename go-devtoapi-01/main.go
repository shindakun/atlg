package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
	"time"
)

// Article JSON struct
type Article struct {
	TypeOf                 string        `json:"type_of"`
	ID                     int           `json:"id"`
	Title                  string        `json:"title"`
	Description            string        `json:"description"`
	CoverImage             string        `json:"cover_image"`
	PublishedAt            time.Time     `json:"published_at"`
	ReadablePublishDate    string        `json:"readable_publish_date"`
	SocialImage            string        `json:"social_image"`
	TagList                string        `json:"tag_list"`
	Slug                   string        `json:"slug"`
	Path                   string        `json:"path"`
	URL                    string        `json:"url"`
	CanonicalURL           string        `json:"canonical_url"`
	CommentsCount          int           `json:"comments_count"`
	PositiveReactionsCount int           `json:"positive_reactions_count"`
	BodyHTML               string        `json:"body_html"`
	LtagStyle              []interface{} `json:"ltag_style"`
	LtagScript             []interface{} `json:"ltag_script"`
	User                   struct {
		Name            string `json:"name"`
		Username        string `json:"username"`
		TwitterUsername string `json:"twitter_username"`
		GithubUsername  string `json:"github_username"`
		WebsiteURL      string `json:"website_url"`
		ProfileImage    string `json:"profile_image"`
		ProfileImage90  string `json:"profile_image_90"`
	} `json:"user"`
}

// Articles array JSON struct
type Articles []struct {
	TypeOf                 string    `json:"type_of"`
	ID                     int32     `json:"id"`
	Title                  string    `json:"title"`
	Description            string    `json:"description"`
	CoverImage             string    `json:"cover_image"`
	PublishedAt            time.Time `json:"published_at"`
	TagList                []string  `json:"tag_list"`
	Slug                   string    `json:"slug"`
	Path                   string    `json:"path"`
	URL                    string    `json:"url"`
	CanonicalURL           string    `json:"canonical_url"`
	CommentsCount          int       `json:"comments_count"`
	PositiveReactionsCount int       `json:"positive_reactions_count"`
	User                   struct {
		Name            string      `json:"name"`
		Username        string      `json:"username"`
		TwitterUsername string      `json:"twitter_username"`
		GithubUsername  interface{} `json:"github_username"`
		WebsiteURL      string      `json:"website_url"`
		ProfileImage    string      `json:"profile_image"`
		ProfileImage90  string      `json:"profile_image_90"`
	} `json:"user"`
}

// DevtoClient struct
type DevtoClient struct {
	DevtoAPIURL string
	Client      *http.Client
}

// New returns our DevtoClient
func New(apiurl string, client *http.Client) *DevtoClient {
	if client == nil {
		client = http.DefaultClient
	}
	return &DevtoClient{
		apiurl,
		client,
	}
}

// FormatPagedRequest reutnrs *http.Request ready to do() to get one page
func (dtc DevtoClient) FormatPagedRequest(param, paramValue string) (*http.Request, error) {
	URL := dtc.DevtoAPIURL
	if param == "page" && paramValue != "" {
		URL = dtc.DevtoAPIURL + "articles/?" + param + "=" + paramValue
		fmt.Printf("%v\n", URL)
	}
	return http.NewRequest(http.MethodGet, URL, nil)
}

// FormatArticleRequest returns http.Request ready to do() and get an article
func (dtc DevtoClient) FormatArticleRequest(i int32) (*http.Request, error) {
	URL := fmt.Sprintf(dtc.DevtoAPIURL+"articles/%d", i)
	return http.NewRequest(http.MethodGet, URL, nil)
}

func getArticle(dtc *DevtoClient, i int32, wg *sync.WaitGroup) {
	defer wg.Done()
	r, err := dtc.FormatArticleRequest(i)
	if err != nil {
		panic(err)
	}

	resp, err := dtc.Client.Do(r)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	//var article Article
	//json.Unmarshal(body, &article)
	//fmt.Printf("%v", article.BodyHTML)

	fileName := fmt.Sprintf("%d.json", i)
	ioutil.WriteFile("./out/"+fileName, body, 0666)
}

func main() {
	dtc := New("https://dev.to/api/", nil)
	doit := true
	c := 1

	for doit {
		req, err := dtc.FormatPagedRequest("page", fmt.Sprintf("%d", c))
		if err != nil {
			panic(err)
		}
		resp, err := dtc.Client.Do(req)
		if err != nil {
			panic(err)
		}
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			panic(err)
		}

		var wg sync.WaitGroup
		var articles Articles
		err = json.Unmarshal(body, &articles)
		if err != nil {
			panic(err)
		}
		wg.Add(len(articles))

		for i := range articles {
			go getArticle(dtc, articles[i].ID, &wg)
		}
		wg.Wait()

		if string(body) != "[]" {
			c++
			continue
		}
		doit = false
	}
}
