ARG GO_BUILDER_IMAGE=golang:1.22.3-alpine
ARG RUNTIME_IMAGE=alpine:3.20

FROM ${GO_BUILDER_IMAGE} AS builder

ARG APP=api
ARG GOPROXY=https://proxy.golang.org,direct
WORKDIR /src
ENV GOPROXY=${GOPROXY}

COPY go.mod go.sum ./
RUN go mod download

COPY apps ./apps
COPY internal ./internal
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/service ./apps/${APP}

FROM ${RUNTIME_IMAGE}

ARG ALPINE_REPOSITORY=
RUN if [ -n "$ALPINE_REPOSITORY" ]; then \
      sed -i "s|https://dl-cdn.alpinelinux.org/alpine|$ALPINE_REPOSITORY|g" /etc/apk/repositories; \
    fi && \
    apk add --no-cache ca-certificates tzdata font-wqy-zenhei
RUN addgroup -S app && adduser -S app -G app
WORKDIR /app
USER app
COPY --from=builder /out/service /usr/local/bin/service
COPY migrations ./migrations

ENTRYPOINT ["/usr/local/bin/service"]
