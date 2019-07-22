Payments
--------

## Project purpose

Payment system, provides ability to transfer money between accounts. 

System provide reports: 
 - all registered accounts; 
 - all registered payments (transfers).

## Usage

### Command-line flags

 - Web server:
   - `-http_address` _string_ -- Http address for web server running (default "0.0.0.0:8080")
 - Database:
   - `-db_address` _string_ -- Address to connect to PostgreSQL server (default "localhost:5432")
   - `-database` _string_ -- PostgreSQL database name (default "payments")
   - `-db_user` _string_ -- PostgreSQL connection user (default "postgres")
   - `-db_password` _string_ -- PostgreSQL connection password
   - `-pool_size` _int_ -- PostgreSQL connection pool size (default 10)
   - `-app_name` _string_ -- PostgreSQL application name (for logging) (default "payments")
   - `-db_log` -- Switch for statements logging

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

### Step 1. Build docker image
```bash
 docker build -t payments-app .
```

### Step 2. Run it

```bash
docker run --rm -p 8099:8080 payments-app --db_address=192.168.0.1:5432 --db_password=${DB_PASSWORD}
```

## How to run tests

```bash
go test -race ./...

```

## How to run code linting

```bash
golangci-lint run --presets=bugs,complexity,format
```

## How to start contributing

TBA