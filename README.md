# swap

swap is a free API for current and historical foreign exchange rates [published by Xe](https://www.xe.com/currencytables/).

## Installing

To start using swap, install Go and run `go mod tidy`.

## Usage

Get the latest foreign exchange rates.

```http
GET localhost:1323/rates
```

Get historical rates for any day since Dec 1, 1995.

```http
GET localhost:1323/rates?date=2006-01-02
```

Rates are quoted against the USD by default. Quote against a different currency by setting the base parameter in your request.

```http
GET localhost:1323/rates?base=EUR
```

Request specific exchange rates by setting the symbols parameter.

```http
GET localhost:1323/rates?symbols=EUR,GBP
```

Get the currency symbols.

```http
GET localhost:1323/symbols
```