package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/PuerkitoBio/goquery"
	//https://www.devdungeon.com/content/web-scraping-go
)

var logo string

// GetTitlePage ...
func GetTitlePage(domain string) string {

	var res string

	// Make HTTP GET request
	response, err := http.Get("https://www." + domain)
	if err != nil {
		log.Fatal(err)
	}
	defer response.Body.Close()
	GetLogoPage(response.Body)

	// Get the response body as a string
	dataInBytes, err := ioutil.ReadAll(response.Body)
	pageContent := string(dataInBytes)

	// Find a substr
	titleStartIndex := strings.Index(pageContent, "<title>")
	if titleStartIndex == -1 {
		fmt.Println("No title element found")

		res = "No title element found"
		return res
	}
	// The start index of the title is the index of the first
	// character, the < symbol. We don't want to include
	// <title> as part of the final value, so let's offset
	// the index by the number of characers in <title>
	titleStartIndex += 7

	// Find the index of the closing tag
	titleEndIndex := strings.Index(pageContent, "</title>")
	if titleEndIndex == -1 {
		fmt.Println("No closing tag for title found.")
		os.Exit(0)
		res = "No closing tag for title found."
		return res
	}

	// (Optional)
	// Copy the substring in to a separate variable so the
	// variables with the full document data can be garbage collected
	pageTitle := []byte(pageContent[titleStartIndex:titleEndIndex])

	res = string(pageTitle)
	return strings.TrimSpace(res)
}

func GetLogoPage(body io.ReadCloser) {

	// Create a goquery document from the HTTP response
	document, err := goquery.NewDocumentFromReader(body)
	if err != nil {
		log.Fatal("Error loading HTTP response body. ", err)
	}

	// Find all links and process them with the function
	// defined earlier selector
	document.Find("link").Each(processElement)
}

// This will get called for each HTML element found
func processElement(index int, element *goquery.Selection) {

	rel, exists2 := element.Attr("rel")
	if exists2 {

		if rel == "shortcut icon" {
			fmt.Println("Good job!")
			href, _ := element.Attr("href")
			fmt.Println(href)
			logo = href
		}
	}

}
