# build go app
FROM golang:1.14-alpine3.11 as builder

RUN echo "http://mirrors.aliyun.com/alpine/v3.11/main" > /etc/apk/repositories \
 && echo "http://mirrors.aliyun.com/alpine/v3.11/community" >> /etc/apk/repositories

WORKDIR /var

COPY . /var

RUN go mod vendor && go build -ldflags "-s -w" -o main main.go

# build main container
FROM alpine:3.11 as runtime

RUN echo "http://mirrors.aliyun.com/alpine/v3.11/main" > /etc/apk/repositories \
 && echo "http://mirrors.aliyun.com/alpine/v3.11/community" >> /etc/apk/repositories

RUN apk add --no-cache tzdata ca-certificates bash btrfs-progs \
 && cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime \
 && apk del tzdata

WORKDIR /bin

COPY --from=builder /var/main /bin/main

RUN chmod +x /bin/main && rm -Rf /var/cache/apk/*

EXPOSE 4222

CMD ["/bin/main"]