package main

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/labstack/echo/v4"
)

type Symbols struct {
	Success bool              `json:"success"`
	Symbols map[string]string `json:"symbols"`
}

type Rates struct {
	Success bool               `json:"success"`
	Base    string             `json:"base"`
	Date    string             `json:"date"`
	Rates   map[string]float64 `json:"rates"`
}

type ConvertedCurrency struct {
	Success bool                   `json:"success"`
	Query   map[string]interface{} `json:"query"`
	Info    map[string]interface{} `json:"info"`
	Date    string                 `json:"date"`
	Result  float64                `json:"result"`
}

type TimeSriesRates struct {
	Success bool                          `json:"success"`
	Base    string                        `json:"base"`
	StartAt string                        `json:"start_at"`
	EndAt   string                        `json:"end_at"`
	Rates   map[string]map[string]float64 `json:"rates"`
}

func main() {
	e := echo.New()
	e.GET("/symbols", GetSymbols)
	e.GET("/:date", GetHistoricalRates)
	e.GET("/convert", GetConvertedCurrency)
	e.GET("/latest", GetLatestRates)
	e.GET("/timeseries", GetTimeSriesRates)
	e.Logger.Fatal(e.Start(":1323"))
}

func GetSymbols(e echo.Context) error {
	response := new(Symbols)
	symbols := make(map[string]string)
	c := colly.NewCollector()
	c.SetRequestTimeout(60 * time.Second)
	c.OnHTML("table.currencytables__Table-sc-xlq26m-3 > tbody", func(c *colly.HTMLElement) {
		c.ForEach("tr", func(_ int, c *colly.HTMLElement) {
			symbols[c.ChildText("*[scope]")] = c.ChildText("td:nth-child(2)")
		})
	})
	c.Visit(fmt.Sprintf("https://www.xe.com/currencytables/?from=USD&date=%s", time.Now().Add(-48*time.Hour).Format("2006-01-02")))
	if len(symbols) != 0 {
		response = &Symbols{
			Success: true,
			Symbols: symbols,
		}
	}
	return e.JSON(http.StatusOK, response)
}

func GetHistoricalRates(e echo.Context) error {
	response := new(Rates)
	rates := make(map[string]float64)
	c := colly.NewCollector()
	c.SetRequestTimeout(60 * time.Second)
	c.OnHTML("table.currencytables__Table-sc-xlq26m-3 > tbody", func(c *colly.HTMLElement) {
		c.ForEach("tr", func(_ int, c *colly.HTMLElement) {
			for _, symbol := range strings.Split(e.QueryParam("symbols"), ",") {
				if symbol == c.ChildText("*[scope]") {
					r, _ := strconv.ParseFloat(c.ChildText("td:nth-child(3)"), 64)
					rates[c.ChildText("*[scope]")] = r
				}
			}
		})
	})
	c.Visit(fmt.Sprintf("https://www.xe.com/currencytables/?from=%s&date=%s", e.QueryParam("base"), e.Param("date")))
	if len(rates) != 0 {
		response = &Rates{
			Success: true,
			Base:    e.QueryParam("base"),
			Date:    e.Param("date"),
			Rates:   rates,
		}
	}
	return e.JSON(http.StatusOK, response)
}

func GetConvertedCurrency(e echo.Context) error {
	response := new(ConvertedCurrency)
	info := make(map[string]interface{})
	query := make(map[string]interface{})
	var result float64
	c := colly.NewCollector()
	c.SetRequestTimeout(60 * time.Second)
	c.OnHTML("body", func(c *colly.HTMLElement) {
		data := strings.Split(c.ChildText(".result__BigRate-sc-1bsijpp-1.dPdXSB"), " ")
		result, _ = strconv.ParseFloat(strings.Replace(data[0], ",", "", -1), 64)
		amount, _ := strconv.ParseFloat(e.QueryParam("amount"), 64)
		rate := result / amount
		data = strings.Split(c.ChildText("div.result__LiveSubText-sc-1bsijpp-2.jcIWiH"), "updated ")
		data = strings.Split(data[1], " ")
		datetime, _ := time.Parse("2006-Jan-02 15:04", fmt.Sprintf("%s-%s-%s %s", string(data[2][:4]), data[0], string(data[1][:2]), data[3]))
		timestamp := datetime.Unix()
		info["rate"] = rate
		info["timestamp"] = timestamp
		query["from"] = e.QueryParam("from")
		query["to"] = e.QueryParam("to")
		query["amount"] = amount
	})
	c.Visit(fmt.Sprintf("https://www.xe.com/currencyconverter/convert/?Amount=%s&From=%s&To=%s", e.QueryParam("amount"), e.QueryParam("from"), e.QueryParam("to")))
	if len(info) != 0 && len(query) != 0 {
		response = &ConvertedCurrency{
			Success: true,
			Query:   query,
			Info:    info,
			Date:    time.Now().Format("2006-01-02"),
			Result:  result,
		}
	}
	return e.JSON(http.StatusOK, response)
}

func GetLatestRates(e echo.Context) error {
	response := new(Rates)
	rates := make(map[string]float64)
	var wg sync.WaitGroup
	for _, symbol := range strings.Split(e.QueryParam("symbols"), ",") {
		wg.Add(1)
		go func(symbol string) {
			defer wg.Done()
			c := colly.NewCollector()
			c.SetRequestTimeout(60 * time.Second)
			c.OnHTML("body", func(c *colly.HTMLElement) {
				data := strings.Split(c.ChildText(".result__BigRate-sc-1bsijpp-1.dPdXSB"), " ")
				rate, _ := strconv.ParseFloat(strings.Replace(data[0], ",", "", -1), 64)
				rates[symbol] = rate
			})
			c.Visit(fmt.Sprintf("https://www.xe.com/currencyconverter/convert/?Amount=1&From=%s&To=%s", e.QueryParam("base"), symbol))
		}(symbol)
	}
	wg.Wait()
	if len(rates) != 0 {
		response = &Rates{
			Success: true,
			Base:    e.QueryParam("base"),
			Date:    time.Now().Format("2006-01-02"),
			Rates:   rates,
		}
	}
	return e.JSON(http.StatusOK, response)
}

func GetTimeSriesRates(e echo.Context) error {
	response := new(TimeSriesRates)
	allRates := make(map[string]map[string]float64)
	rates := make(map[string]float64)
	var wg sync.WaitGroup
	startAt, _ := time.Parse("2006-01-02", e.QueryParam("start_at"))
	endAt, _ := time.Parse("2006-01-02", e.QueryParam("end_at"))
	for {
		if startAt.Before(endAt) || startAt.Equal(endAt) {
			wg.Add(1)
			go func(startAt time.Time) {
				defer wg.Done()
				c := colly.NewCollector()
				c.SetRequestTimeout(60 * time.Second)
				c.OnHTML("table.currencytables__Table-sc-xlq26m-3 > tbody", func(c *colly.HTMLElement) {
					c.ForEach("tr", func(_ int, c *colly.HTMLElement) {
						for _, symbol := range strings.Split(e.QueryParam("symbols"), ",") {
							if symbol == c.ChildText("*[scope]") {
								currency, _ := strconv.ParseFloat(c.ChildText("td:nth-child(3)"), 64)
								rates[c.ChildText("*[scope]")] = currency
							}
						}
					})
				})
				c.Visit(fmt.Sprintf("https://www.xe.com/currencytables/?from=%s&date=%s", e.QueryParam("base"), startAt.Format("2006-01-02")))
				if len(rates) != 0 {
					allRates[startAt.Format("2006-01-02")] = rates
				}
			}(startAt)
			startAt = startAt.Add(24 * time.Hour)
		} else {
			break
		}
	}
	wg.Wait()
	if len(rates) != 0 {
		response = &TimeSriesRates{
			Success: true,
			StartAt: e.QueryParam("start_at"),
			EndAt:   e.QueryParam("end_at"),
			Base:    e.QueryParam("base"),
			Rates:   allRates,
		}
	}
	return e.JSON(http.StatusOK, response)
}
