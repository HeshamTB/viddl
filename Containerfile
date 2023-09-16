FROM golang:1.19.13-alpine3.18 AS builder

WORKDIR /build

COPY ./ /build

RUN ls -lah
RUN sed -i 's/1.21.1/1.21/g' /build/go.mod
RUN sed -i 's/1.21.1/1.21/g' /build/go.work
RUN sed -i 's/1.21.1/1.21/g' /build/apkupdater/go.mod
RUN go mod edit -go=1.19

RUN go build -o viddl .

FROM alpine:3.18

RUN apk add --no-cache yt-dlp

COPY --from=builder /build/viddl /app/viddl

RUN chmod +x /app/viddl

CMD [ "/app/viddl" ]
