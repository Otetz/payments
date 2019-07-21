##################################
# STEP 1 build executable binary #
##################################
FROM golang:1.12-alpine AS builder

WORKDIR /go/src/github.com/otetz/payments
COPY . .

ENV GO111MODULE=on

RUN apk --update add git

RUN go get -d -v ./...
RUN go install -v ./...
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -ldflags="-w -s" -o /go/bin/payments

##############################
# STEP 2 build a small image #
##############################
FROM alpine:3.10

RUN addgroup -S payments \
    && adduser -S payments -G payments -s /bin/sh \
    && apk --update add curl

COPY --from=builder /go/bin/payments /go/bin/payments

WORKDIR /go/bin/

EXPOSE 8080
USER payments

ENTRYPOINT ["/go/bin/payments"]
CMD []