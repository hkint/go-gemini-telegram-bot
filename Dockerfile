FROM --platform=$BUILDPLATFORM golang:alpine AS builder
WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . ./

RUN go build -ldflags="-s -w" -v -o gemini-bot


FROM --platform=$BUILDPLATFORM alpine
WORKDIR /app

RUN apk add --no-cache tzdata

COPY --from=builder /build/gemini-bot .

ENV TZ="Asia/Shanghai"

CMD ["./gemini-bot"]
