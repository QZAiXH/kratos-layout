ARG APP=helloworld

FROM golang:1.25 AS builder
ARG APP

COPY . /src
WORKDIR /src

RUN GOPROXY=https://goproxy.cn make build APP=${APP}

FROM debian:stable-slim
ARG APP

RUN apt-get update && apt-get install -y --no-install-recommends \
		ca-certificates  \
        netbase \
        && rm -rf /var/lib/apt/lists/ \
        && apt-get autoremove -y && apt-get autoclean -y

COPY --from=builder /src/bin/${APP} /app/server

WORKDIR /app

EXPOSE 8000
EXPOSE 9000
VOLUME /data/conf

CMD ["./server", "-conf", "/data/conf"]
