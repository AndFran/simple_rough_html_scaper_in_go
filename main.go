package main

import (
	"fmt"
	"golang.org/x/net/html"
	"io"
	"math/rand"
	"net/http"
	"strings"
	"sync"
	"time"
)

type UrlResult struct {
	Url        string
	StatusCode int
	Atags      []string
	Error      error
	Ok         bool
}

func ScrapeUrl(inputUrls <-chan string, outputResults chan<- UrlResult,
	wg *sync.WaitGroup) {
	defer wg.Done() // signal no more work to do when inputUrls closes
	for j := range inputUrls {

		resp, err := http.Get(j)
		if err != nil {
			outputResults <- UrlResult{Url: j, Error: err, Ok: false}
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			outputResults <- UrlResult{Url: j, StatusCode: resp.StatusCode, Ok: false}
			return
		}
		b, _ := io.ReadAll(resp.Body)

		z := html.NewTokenizer(strings.NewReader(string(b)))
		aTags := make([]string, 0)

	loop:
		for {
			token := z.Next()
			switch token {
			case html.StartTagToken:
				t := z.Token()
				if t.Data == "a" {
					for _, v := range t.Attr {
						if v.Key == "href" {
							aTags = append(aTags, v.Val)
						}
					}
				}
			case html.ErrorToken:
				break loop
			}
		}

		// TODO remove random timers for testing
		time.Sleep(time.Duration(rand.Intn(5)) * time.Second)
		outputResults <- UrlResult{Url: j, StatusCode: resp.StatusCode, Ok: true, Atags: aTags}
	}
}

func main() {
	urls := []string{
		"https://www.google.com",
		"https://www.yahoo.com",
		"https://www.facebook.com",
		"https://www.instagram.com",
		"https://www.reddit.com",
		"https://www.twitter.com",
	}

	jobs := make(chan string)
	results := make(chan UrlResult)
	waitForResults := make(chan struct{})

	wg := sync.WaitGroup{}

	// 3 workers
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go ScrapeUrl(jobs, results, &wg)
	}

	go func() {
		// wait for all jobs to complete
		wg.Wait()
		close(results)
	}()

	go func() {
		for r := range results {
			if r.Ok {
				fmt.Println("Results for ", r.Url, ":", len(r.Atags))
			} else {
				fmt.Println(r.Error)
			}
		}
		waitForResults <- struct{}{}
	}()

	for _, url := range urls {
		jobs <- url // send the work
	}
	close(jobs)

	<-waitForResults
}
