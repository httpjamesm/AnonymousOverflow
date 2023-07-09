FROM golang:1.19-alpine as build

WORKDIR /app

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

ENV CGO_ENABLED=0

RUN go build -o anonymousoverflow

FROM scratch

COPY --from=build /app/anonymousoverflow /anonymousoverflow
COPY templates /templates
COPY public /public
COPY --from=build /etc/ssl/certs /etc/ssl/certs

EXPOSE 8080

CMD ["/anonymousoverflow"]