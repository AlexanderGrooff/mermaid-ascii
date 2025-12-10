FROM golang:1.25-alpine AS builder
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o mermaid-ascii .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
RUN addgroup -g 1000 mermaid && \
    adduser -D -u 1000 -G mermaid mermaid

WORKDIR /app
COPY --from=builder /build/mermaid-ascii /usr/local/bin/mermaid-ascii
COPY --from=builder /build/templates /app/templates
RUN chown -R mermaid:mermaid /app

USER mermaid
EXPOSE 3001
ENTRYPOINT ["mermaid-ascii"]
CMD ["--help"]



