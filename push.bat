@echo off
docker build -t go-in-download .
docker tag go-in-download registry.cn-hangzhou.aliyuncs.com/xrdev/go-in-download:%1
docker push registry.cn-hangzhou.aliyuncs.com/xrdev/go-in-download:%1
