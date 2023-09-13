FROM golang:1.19.13-alpine3.18 AS builder

WORKDIR /build

COPY ./ /build

RUN sed -i 's/1.21.0/1.21/g' /build/go.mod
RUN go mod edit -go=1.19

RUN go build -o viddl .

FROM alpine:3.18

RUN apk add --no-cache yt-dlp

COPY --from=builder /build/viddl /app/viddl

RUN chmod +x /app/viddl

HEALTHCHECK --interval=30s --timeout=30s --start-period=5s --retries=3\
    CMD curl -f http://localhost:8080 || exit 1

CMD [ "/app/viddl" ]
