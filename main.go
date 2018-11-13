package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"golang.org/x/net/html"
)

func validate(reqURL string, c chan string, f chan bool) {

	escapeUrl, _ := url.QueryUnescape(reqURL)
	resp, err := http.Get(escapeUrl)
	defer func() {
		// notify done to finish channel
		f <- true
	}()
	if err != nil {
		fmt.Println("\nHTTP ERROR: Failed to reach \"" + escapeUrl + "\"")
		fmt.Println(err.Error())
		c <- escapeUrl
		return
	}
	if resp.StatusCode == http.StatusNotFound {
		c <- escapeUrl
		return
	}
	return
}

// get href attribute
func getHref(pageURL string, t html.Token) (ok bool, href string) {
	for _, a := range t.Attr {
		if a.Key == "href" {
			u, err := url.Parse(a.Val)
			if err != nil {
				panic(err)
			}

			if strings.Contains(a.Val, "#") {
				// if host contains a #, it's a link on the page, append to pageURL
				href = pageURL + "/" + a.Val
			} else if u.Host == "" {
				// if the host is blank, that means we got a class selector like # and need to build the full url
				pageURL, err := url.Parse(pageURL)
				if err != nil {
					panic(err)
				}

				newURL := &url.URL{
					Scheme: pageURL.Scheme,
					Host:   pageURL.Hostname(),
					Path:   a.Val,
				}
				href = newURL.String()
			} else {
				// this is a weird case for broken links
				if u.Scheme == "" && u.Hostname() == "" {
					continue
				}
				href = a.Val
			}

			ok = true
		}
	}
	// bare return will return all vars defined in function
	return
}

// get all http links from page
func parse(url string, c chan string, f chan bool) {
	resp, err := http.Get(url)
	defer func() {
		// notify done to finish channel
		f <- true
	}()
	if err != nil {
		fmt.Println("ERROR: Failed to crawl \"" + url + "\"")
		return
	}
	defer resp.Body.Close()
	z := html.NewTokenizer(resp.Body)

	for {
		tt := z.Next()

		switch {
		case tt == html.ErrorToken:
			// End of the document, we're done
			return
		case tt == html.StartTagToken:
			t := z.Token()

			isAnchor := t.Data == "a"
			if !isAnchor {
				continue
			}

			ok, url := getHref(url, t)
			if !ok {
				continue
			}

			c <- url
		}
	}
}

func main() {
	start := time.Now()

	foundURLs := make(map[string]bool)
	failedUrls := make([]string, 0)

	/* get urls from command line */
	if len(os.Args) < 2 {
		fmt.Println("please provide url")
		os.Exit(1)
	}

	seedURLs := os.Args[1:]

	c := make(chan string)
	f := make(chan bool)

	// find all links on each provided page
	for _, url := range seedURLs {
		go parse(url, c, f)
		// listen to channel(url, c, f)
	}
	for i := 0; i < len(seedURLs); {
		select {
		case url := <-c:
			foundURLs[url] = true
		case <-f:
			i++
		}
	}

	fmt.Println("\nFound", len(foundURLs), "unique urls:")

	/* print urls and try to access */
	cStatus := make(chan string)
	fStatus := make(chan bool)

	// validate every link on page
	for link := range foundURLs {
		fmt.Println(" - " + link)
		go validate(link, cStatus, fStatus)
	}
	// listen to channel
	for i := 0; i < len(foundURLs); {
		select {
		case url := <-cStatus:
			failedUrls = append(failedUrls, url)
		case <-fStatus:
			i++
		}
	}

	fmt.Println("\nFailed", len(failedUrls), "unique urls:")
	for _, url := range failedUrls {
		fmt.Println(" - " + url)

	}

	elapsed := time.Since(start)
	log.Printf("URL Validation took %s", elapsed)
}
