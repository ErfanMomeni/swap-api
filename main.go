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

type HistoricalRatesResponse struct {
	Base  string             `json:"base"`
	Date  string             `json:"date"`
	Rates map[string]float64 `json:"rates"`
}

type ConvertResponse struct {
	Query  map[string]interface{} `json:"query"`
	Info   map[string]interface{} `json:"info"`
	Date   string                 `json:"date"`
	Result float64                `json:"result"`
}

func main() {
	e := echo.New()
	e.GET("/symbols", GetSymbols)
	e.GET("/:date", GetHistoricalRates)
	e.GET("/convert", Convert)
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
	c.Visit(fmt.Sprintf("https://www.xe.com/currencytables/?from=USD&date=%s", time.Now().Add(-24*time.Hour).Format("2006-01-02")))
	return e.JSON(http.StatusOK, response)
}

func GetHistoricalRates(e echo.Context) error {
	response := new(HistoricalRatesResponse)
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
		response = &HistoricalRatesResponse{
			Base:  e.QueryParam("base"),
			Date:  e.Param("date"),
			Rates: rates,
		}
	})
	c.Visit(fmt.Sprintf("https://www.xe.com/currencytables/?from=%s&date=%s", e.QueryParam("base"), e.Param("date")))
	return e.JSON(http.StatusOK, response)
}

func Convert(e echo.Context) error {
	response := new(ConvertResponse)
	c := colly.NewCollector()
	c.SetRequestTimeout(60 * time.Second)
	c.OnHTML("body", func(r *colly.HTMLElement) {
		s := strings.Split(r.ChildText(".result__BigRate-sc-1bsijpp-1.dPdXSB"), " ")
		result, _ := strconv.ParseFloat(s[0], 64)
		s = strings.Split(r.ChildText(".unit-rates___StyledDiv-sc-1dk593y-0.iGxfWX > p:nth-child(1)"), " ")
		rate, _ := strconv.ParseFloat(s[3], 64)
		s = strings.Split(r.ChildText("div.result__LiveSubText-sc-1bsijpp-2.jcIWiH"), "updated ")
		s = strings.Split(s[1], " ")
		datetime, _ := time.Parse("2006-Jan-02 15:04", fmt.Sprintf("%s-%s-%s %s", string(s[2][:4]), s[0], string(s[1][:2]), s[3]))
		timestamp := datetime.Unix()
		info := map[string]interface{}{
			"rate":      rate,
			"timestamp": timestamp,
		}
		amount, _ := strconv.ParseFloat(e.QueryParam("amount"), 64)
		query := map[string]interface{}{
			"from":   e.QueryParam("from"),
			"to":     e.QueryParam("to"),
			"amount": amount,
		}
		response = &ConvertResponse{
			Query:  query,
			Info:   info,
			Date:   time.Now().Format("2006-01-02"),
			Result: result,
		}
	})
	c.Visit(fmt.Sprintf("https://www.xe.com/currencyconverter/convert/?Amount=%s&From=%s&To=%s", e.QueryParam("amount"), e.QueryParam("from"), e.QueryParam("to")))
	return e.JSON(http.StatusOK, response)
}
