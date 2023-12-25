FROM --platform=$BUILDPLATFORM golang:alpine AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . ./

RUN go build -ldflags="-s -w" -v -o bot main.go


FROM --platform=$BUILDPLATFORM alpine
WORKDIR /app

RUN apk add --no-cache tzdata

COPY --from=builder /app/bot .

ENV TZ="Asia/Shanghai"

ENTRYPOINT ["./bot"]