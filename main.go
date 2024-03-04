package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/PuerkitoBio/goquery"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type votd struct {
	Verse Verse `json:"verse"`
}

type Verse struct {
	Date      string `json:"date"`
	Text      string `json:"text"`
	Reference string `json:"reference"`
	Combined  string `json:"combined"`
}

type votw struct {
	Verses []votd `json:"verses"`
}

func getPageContent(contentType string) (*goquery.Selection, error) {
	url := "https://www.bible.com/verse-of-the-day"
	// Request the HTML page
	response, err := http.Get(url)
	if err != nil {
		return nil, err
		// return c.JSON(http.StatusInternalServerError, fmt.Sprintf("Error making request: %s", err))
		// log.Fatalf("Error making request: %s", err)
	}

	defer response.Body.Close()

	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(response.Body)
	if err != nil {
		return nil, err
		// log.Fatal("Failed to parse the HTML document", err)
	}

	// Find the parent element
	parentSelector := "main>div.w-full>div.w-full>div"
	parent := doc.Find(parentSelector)

	votdHTML := parent.Children().Eq(0) //!For VOTD
	votwHTML := parent.Children().Eq(2) //!For VOTW

	if contentType == "day" {
		return votdHTML, nil
	} else if contentType == "week" {
		return votwHTML, nil
	} else {
		return nil, fmt.Errorf("unable to get content")
	}
}

func getVOTD(c echo.Context) error {
	// Used to parse and get the  content for verse of the day
	contentType := "day"
	votdHTML, err := getPageContent(contentType)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}

	dateHTML := votdHTML.Find("p").First()
	textHTML := votdHTML.Find("div.mbs-3>a").First()
	referenceHTML := textHTML.Next()

	date := dateHTML.Text()
	text := textHTML.Text()
	reference := referenceHTML.Text()

	votdResponse := votd{
		Verse: Verse{
			Date:      date,
			Text:      text,
			Reference: reference,
			Combined:  fmt.Sprintf("%s - %s", text, reference),
		},
	}

	return c.JSON(http.StatusOK, votdResponse)
}

func getVOTW(c echo.Context) error {
	// Used to parse the content and get the bible verse of the week
	votwResponse := votw{}
	contentType := "week"

	votwHTML, err := getPageContent(contentType)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}
	days := votwHTML.Find("div.mlb-2")

	days.Each(func(i int, element *goquery.Selection) {
		dateHTML := element.Find("p").First()
		textHTML := element.Find("a").First()
		referenceHTML := textHTML.Next()

		date := dateHTML.Text()
		text := textHTML.Text()
		reference := referenceHTML.Text()

		votdResponse := votd{
			Verse: Verse{
				Date:      date,
				Text:      text,
				Reference: reference,
				Combined:  fmt.Sprintf("%s - %s", text, reference),
			},
		}
		votwResponse.Verses = append(votwResponse.Verses, votdResponse)
	})
	return c.JSON(http.StatusOK, votwResponse)
}

func main() {

	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.GET("/api/v1/votd", getVOTD)
	e.GET("/api/v1/votw", getVOTW)

	httpPort := os.Getenv("PORT")
	if httpPort == "" {
		httpPort = "8330"
	}

	e.Logger.Fatal(e.Start(":" + httpPort))
}
