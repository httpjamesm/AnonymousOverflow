FROM golang:1.22.1-alpine3.19 AS build

WORKDIR /app

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

# Architecture and OS are set dynamically (by BuildKit)
ARG TARGETOS
ARG TARGETARCH
ENV CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH

RUN go build -o anonymousoverflow && go build -o healthcheck ./src/healthcheck

FROM scratch

COPY --from=build /app/anonymousoverflow /anonymousoverflow
COPY --from=build /app/healthcheck /healthcheck
COPY --from=build /etc/ssl/certs /etc/ssl/certs

HEALTHCHECK --interval=60s --timeout=5s --start-period=2s --retries=3 CMD [ "/healthcheck","http://localhost:8080/healthz" ]

EXPOSE 8080

CMD ["/anonymousoverflow"]
