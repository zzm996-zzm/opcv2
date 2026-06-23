FROM golang:1.22.3-alpine AS builder

ARG APP=api
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY apps ./apps
COPY internal ./internal
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/service ./apps/${APP}

FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata
RUN addgroup -S app && adduser -S app -G app
WORKDIR /app
USER app
COPY --from=builder /out/service /usr/local/bin/service
COPY migrations ./migrations

ENTRYPOINT ["/usr/local/bin/service"]
