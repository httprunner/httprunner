FROM golang:1.17 as builder
ENV GOPROXY=https://goproxy.cn,https://goproxy.io,direct
ENV GO111MODULE=on
ENV GOCACHE=/go/pkg/.cache/go-build
ENV CGO_ENABLED=0

WORKDIR /work
ADD . .
RUN make build

FROM alpine:3.6 as alpine
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.ustc.edu.cn/g' /etc/apk/repositories && \
    apk update && \
    apk add -U --no-cache ca-certificates tzdata

FROM alpine:3.6
ENV TZ="Asia/Shanghai"

RUN ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && \
    echo $TZ > /etc/timezone

COPY --from=alpine /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=alpine /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /work/output/hrp /bin/hrp

CMD ["hrp"]
