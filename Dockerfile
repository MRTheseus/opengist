FROM golang:1.25-alpine AS builder

WORKDIR /app
COPY . .
RUN export GOPROXY=https://goproxy.cn,direct && go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o closegist ./cmd

FROM alpine:latest

RUN apk add --no-cache tzdata
ENV TZ=Asia/Shanghai

WORKDIR /app
COPY --from=builder /app/closegist .
COPY --from=builder /app/cmd/web/templates ./web/templates

EXPOSE 6157
CMD ["./closegist"]
