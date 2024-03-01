package main

import (
	"fmt"
	"log"
	"net/http"
	"regexp"

	"github.com/PuerkitoBio/goquery"

	"github.com/labstack/echo/v4"
)

type VerseResponse struct {
	Verse Verse `json:"verse"`
}
type Verse struct {
	Details Details `json:"details"`
}
type Details struct {
	Text      string `json:"text"`
	Reference string `json:"reference"`
	Version   string `json:"version"`
}

func extractVersion(text string) (string, string) {
	// Define a regular expression to match Bible versions (assuming they are in parentheses)
	re := regexp.MustCompile(`^(.*?)\s*\(([^)]+)\)$`)
	// Find matches in the text
	matches := re.FindStringSubmatch(text)

	// Extract the Bible version (if found)
	if len(matches) >= 2 {
		bibleReference := matches[1]
		bibleVersion := matches[2]
		return bibleReference, bibleVersion
	}
	// Return an empty string if no match is found
	return text, ""
}

func getVOTD(c echo.Context) error {
	url := "https://www.bible.com/verse-of-the-day"
	// Request the HTML page
	response, err := http.Get(url)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, fmt.Sprintf("Error making request: %s", err))
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {
		log.Fatalf("status code error: %d %s", response.StatusCode, response.Status)
	}

	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(response.Body)
	if err != nil {
		log.Fatal("Failed to parse the HTML document", err)
	}

	votdParent := doc.Find("h1").First().Parent()
	innerVOTDHTML := votdParent.Find("div.mbs-3")

	// Find the first a tag
	textHTML := innerVOTDHTML.Find("a").First()
	// Get the Next element
	referenceHTML := textHTML.Next()

	text := textHTML.Text()
	reference, version := extractVersion(referenceHTML.Text())

	resp := &VerseResponse{
		Verse: Verse{
			Details: Details{
				Text:      text,
				Reference: reference,
				Version:   version,
			},
		},
	}

	return c.JSON(http.StatusOK, resp)
}

func main() {

	e := echo.New()
	e.GET("/api/v1/votd", getVOTD)

	e.Logger.Fatal(e.Start(":8300"))

}
