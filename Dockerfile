FROM golang:latest as build

WORKDIR /go/src/app

COPY . .
COPY go.mod go.sum ./
RUN go mod tidy
RUN go build -o docker-scraper

FROM chromedp/headless-shell:latest

RUN apt-get update && apt install dumb-init -y
ENTRYPOINT ["dumb-init", "--"]

COPY --from=build /go/src/app/docker-scraper /tmp
COPY --from=build /go/src/app/credentials /credentials
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

CMD ["/tmp/docker-scraper"]