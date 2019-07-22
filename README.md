Payments
--------

## Table of Contents

<!-- TOC depthFrom:2 depthTo:6 updateOnSave:true withLinks:true -->

- [Table of Contents](#table-of-contents)
- [Project purpose](#project-purpose)
- [Usage](#usage)
    - [Command-line flags](#command-line-flags)
- [Dependencies](#dependencies)
- [How to set up](#how-to-set-up)
    - [Step 1. Build docker image](#step-1-build-docker-image)
    - [Step 2. Run it](#step-2-run-it)
- [How to run tests](#how-to-run-tests)
- [How to run code linting](#how-to-run-code-linting)
- [How to Contribute](#how-to-contribute)
    - [Support](#support)
    - [What to contribute](#what-to-contribute)
    - [Advice](#advice)

<!-- /TOC -->

## Project purpose

Payment system, provides ability to transfer money between accounts. 

System also provide reports: 
 - all registered accounts; 
 - all registered payments (transfers).

API Documentation see [here](./docs/api.md)

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

## How to Contribute

We definitely welcome patches and contribution to this project!

### Support

If you do have a contribution to the package, feel free to create a Pull Request or an Issue.

### What to contribute

If you don't know what to do, there are some features and functions that need to be done

   - Refactor code
   - Edit docs and README: spellcheck, grammar and typo check
   - Create actual list of contributors and projects that currently using this package
   - Resolve issues and bugs
   - Implement benchmarking
   - Implement batch of examples
   - Look at forks for new features and fixes

### Advice

Feel free to create what you want, but keep in mind when you implement new features:

  - Code must be clear and readable, names of variables/constants clearly describes what they are doing
  - Public functions must be documented and described in source file
  - There are must be unit-tests for any new functions and improvements
