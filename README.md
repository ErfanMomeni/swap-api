# swap

swap is a free API for current and historical foreign exchange rates [published by Xe](https://www.xe.com/currencytables/).

## Installing

To start using swap, install Go and run `go mod tidy`.

## Usage

Get the latest foreign exchange rates.

```http
GET /latest?base=USD&symbols=CAD,EUR
```

Get historical rates for any day since Dec 1, 1995.

```http
GET /2006-01-02?base=USD&symbols=JPY,CAD,CNY,CHF
```

Get currency symbols.

```http
GET /symbols
```

Convert an amount from one currency to another.

```http
GET /convert?amount=50&from=USD&to=EUR
```