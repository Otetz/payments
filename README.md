Payments
--------

## Project purpose

## Dependencies

- [go-kit](http://github.com/go-kit/kit) -- toolkit for building microservices, recommended by design;
- [govalidator](http://github.com/asaskevich/govalidator) -- package of validators and sanitizers for strings, 
numerical, slices and structures;
- [decimal](http://github.com/shopspring/decimal) -- arbitrary-precision fixed-point decimal numbers in go; 
- [uuid](http://github.com/google/uuid) -- go package for UUIDs based on RFC 4122 and DCE 1.1;
- [gorilla/mux](http://github.com/gorilla/mux) -- a powerful HTTP router and URL matcher for building Go web servers;
- [prometheus client](http://github.com/prometheus/client_golang) -- prometheus instrumentation library for Go
applications;
- [go-cmp](https://github.com/google/go-cmp) -- package for comparing Go values in tests;
- [go-pg](https://github.com/go-pg/pg) -- golang ORM with focus on PostgreSQL features and performance.

## How to set up

## How to run tests

```bash
go test -race ./...

```

## How to run code linting

```bash
golangci-lint run --presets=bugs,complexity,format
```

## How to start contributing

