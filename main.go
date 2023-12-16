package main

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/labstack/echo/v4"
)

type SymbolsResponse struct {
	Symbols map[string]string `json:"symbols"`
}

type RatesResponse struct {
	Base  string             `json:"base"`
	Date  string             `json:"date"`
	Rates map[string]float64 `json:"rates"`
}

const BASE_URL = "https://www.xe.com/currencytables"

func main() {
	e := echo.New()
	e.GET("/symbols", GetSymbols)
	e.GET("/:date", GetRates)
	e.Logger.Fatal(e.Start(":1323"))
}

func GetSymbols(e echo.Context) error {
	response := new(SymbolsResponse)
	symbols := make(map[string]string)
	c := colly.NewCollector()
	c.SetRequestTimeout(60 * time.Second)
	c.OnHTML("table.currencytables__Table-sc-xlq26m-3 > tbody", func(r *colly.HTMLElement) {
		r.ForEach("tr", func(_ int, r *colly.HTMLElement) {
			symbols[r.ChildText("*[scope]")] = r.ChildText("td:nth-child(2)")
		})
		response = &SymbolsResponse{
			Symbols: symbols,
		}
	})
	c.Visit(fmt.Sprintf("%s/?from=USD&date=%s", BASE_URL, time.Now().Add(-24*time.Hour).Format("2006-01-02")))
	return e.JSON(http.StatusOK, response)
}

func GetRates(e echo.Context) error {
	response := new(RatesResponse)
	rates := make(map[string]float64)
	c := colly.NewCollector()
	c.SetRequestTimeout(60 * time.Second)
	c.OnHTML("table.currencytables__Table-sc-xlq26m-3 > tbody", func(r *colly.HTMLElement) {
		r.ForEach("tr", func(_ int, r *colly.HTMLElement) {
			for _, s := range strings.Split(e.QueryParam("symbols"), ",") {
				if s == r.ChildText("*[scope]") {
					currency, _ := strconv.ParseFloat(r.ChildText("td:nth-child(3)"), 64)
					rates[r.ChildText("*[scope]")] = currency
				}
			}
		})
		response = &RatesResponse{
			Base:  e.QueryParam("base"),
			Date:  e.Param("date"),
			Rates: rates,
		}
	})
	c.Visit(fmt.Sprintf("%s/?from=%s&date=%s", BASE_URL, e.QueryParam("base"), e.Param("date")))
	return e.JSON(http.StatusOK, response)
}
