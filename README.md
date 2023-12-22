# swap

Current and historical foreign exchange rates [published by the Xe](https://www.xe.com).

## Installation

To start using swap, install Go and run `go mod tidy`.

## Usage

Run `go run main.go`.

Get the latest foreign exchange rates.

```http
GET /latest?base=USD&symbols=EUR,CAD
```

Get historical rates for any day since Dec 1, 1995.

```http
GET /2020-01-02?base=USD&symbols=EUR,CAD
```

Get currency symbols.

```http
GET /symbols
```

Convert an amount from one currency to another.

```http
GET /convert?amount=50&from=USD&to=EUR
```

Get historical rates for a time period.

```http
GET /timeseries?start_at=2020-01-02&end_at=2020-01-05&base=USD&symbols=EUR,CAD
```