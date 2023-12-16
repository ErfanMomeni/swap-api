# swap

swap is a free API for exchange rates [published by Xe](https://www.xe.com/currencytables/).

## Installing

To start using swap, install Go and run `go mod tidy`.

## Usage

Get historical rates for any day since Dec 1, 1995.

```http
GET /2006-01-02?base=USD&symbols=JPY,CAD,CNY,CHF
```

Get currency symbols.

```http
GET /symbols
```