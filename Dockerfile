FROM golang:1.20-alpine as builder
COPY . /tmp/myService
ENV GO111MODULE=on
ENV GOPROXY="https://goproxy.cn"
WORKDIR /tmp/myService/cmd/crossshare_server
RUN --mount=type=cache,target=/root/.cache/go-build go build -o crossshare_server main.go

FROM alpine
WORKDIR /usr/local/bin

EXPOSE 3007

COPY --from=builder /tmp/myService/cmd/crossshare_server/config.toml /usr/local/bin/
COPY --from=builder /tmp/myService/cmd/crossshare_server/crossshare_server /usr/local/bin/

ENTRYPOINT ["crossshare_server"]