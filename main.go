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
	Success bool              `json:"success"`
	Symbols map[string]string `json:"symbols"`
}

type RatesResponse struct {
	Success bool               `json:"success"`
	Base    string             `json:"base"`
	Date    string             `json:"date"`
	Rates   map[string]float64 `json:"rates"`
}

type ConvertResponse struct {
	Success bool                   `json:"success"`
	Query   map[string]interface{} `json:"query"`
	Info    map[string]interface{} `json:"info"`
	Date    string                 `json:"date"`
	Result  float64                `json:"result"`
}

func main() {
	e := echo.New()
	e.GET("/symbols", GetSymbols)
	e.GET("/:date", GetHistoricalRates)
	e.GET("/convert", Convert)
	e.GET("/latest", GetLatestRates)
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
			Success: true,
			Symbols: symbols,
		}
	})
	c.Visit(fmt.Sprintf("https://www.xe.com/currencytables/?from=USD&date=%s", time.Now().Add(-24*time.Hour).Format("2006-01-02")))
	return e.JSON(http.StatusOK, response)
}

func GetHistoricalRates(e echo.Context) error {
	response := new(RatesResponse)
	rates := make(map[string]float64)
	c := colly.NewCollector()
	c.SetRequestTimeout(60 * time.Second)
	c.OnHTML("table.currencytables__Table-sc-xlq26m-3 > tbody", func(r *colly.HTMLElement) {
		r.ForEach("tr", func(_ int, r *colly.HTMLElement) {
			for _, symbol := range strings.Split(e.QueryParam("symbols"), ",") {
				if symbol == r.ChildText("*[scope]") {
					currency, _ := strconv.ParseFloat(r.ChildText("td:nth-child(3)"), 64)
					rates[r.ChildText("*[scope]")] = currency
				}
			}
		})
		response = &RatesResponse{
			Success: true,
			Base:    e.QueryParam("base"),
			Date:    e.Param("date"),
			Rates:   rates,
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
		data := strings.Split(r.ChildText(".result__BigRate-sc-1bsijpp-1.dPdXSB"), " ")
		result, _ := strconv.ParseFloat(strings.Replace(data[0], ",", "", -1), 64)
		amount, _ := strconv.ParseFloat(e.QueryParam("amount"), 64)
		rate := result / amount
		data = strings.Split(r.ChildText("div.result__LiveSubText-sc-1bsijpp-2.jcIWiH"), "updated ")
		data = strings.Split(data[1], " ")
		datetime, _ := time.Parse("2006-Jan-02 15:04", fmt.Sprintf("%s-%s-%s %s", string(data[2][:4]), data[0], string(data[1][:2]), data[3]))
		timestamp := datetime.Unix()
		info := map[string]interface{}{
			"rate":      rate,
			"timestamp": timestamp,
		}
		query := map[string]interface{}{
			"from":   e.QueryParam("from"),
			"to":     e.QueryParam("to"),
			"amount": amount,
		}
		response = &ConvertResponse{
			Success: true,
			Query:   query,
			Info:    info,
			Date:    time.Now().Format("2006-01-02"),
			Result:  result,
		}
	})
	c.Visit(fmt.Sprintf("https://www.xe.com/currencyconverter/convert/?Amount=%s&From=%s&To=%s", e.QueryParam("amount"), e.QueryParam("from"), e.QueryParam("to")))
	return e.JSON(http.StatusOK, response)
}

func GetLatestRates(e echo.Context) error {
	response := new(RatesResponse)
	rates := make(map[string]float64)
	c := colly.NewCollector()
	c.SetRequestTimeout(60 * time.Second)
	for _, symbol := range strings.Split(e.QueryParam("symbols"), ",") {
		c.OnHTML("body", func(r *colly.HTMLElement) {
			data := strings.Split(r.ChildText(".result__BigRate-sc-1bsijpp-1.dPdXSB"), " ")
			rate, _ := strconv.ParseFloat(strings.Replace(data[0], ",", "", -1), 64)
			rates[symbol] = rate
		})
		c.Visit(fmt.Sprintf("https://www.xe.com/currencyconverter/convert/?Amount=10&From=%s&To=%s", e.QueryParam("base"), symbol))
	}
	response = &RatesResponse{
		Success: true,
		Base:    e.QueryParam("base"),
		Date:    time.Now().Format("2006-01-02"),
		Rates:   rates,
	}
	return e.JSON(http.StatusOK, response)
}
