#docker run --name download  -itd -p 1280:1280 -v /var/files:/bin/static/files  --restart=always --privileged=true registry.cn-hangzhou.aliyuncs.com/xrdev/go-in-download:$1
# Preparing the build environment
FROM golang:1.15-alpine AS builder

WORKDIR $GOPATH/src/goindownload
COPY . $GOPATH/src/goindownload
RUN GOPROXY="https://goproxy.io" GO111MODULE=on go build .

# Preparing the final image
FROM alpine:3.12.1
EXPOSE 1280

COPY --from=builder /go/src/goindownload/goIndownload /bin/goIndownload
COPY --from=builder /go/src/goindownload/static/ /bin/static/

WORKDIR /bin
ENTRYPOINT [ "/bin/goIndownload" ]