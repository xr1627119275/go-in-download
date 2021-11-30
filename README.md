# 文件代理下载
1. 建议放在国外的服务器上
2. web 端口1280
3. 容器内存文件存放路径 /bin/static/files
4. docker 启动命令
```bash
    docker run \
        --name=download \
        -v /var/files:/bin/static/files \
        --privileged \
        -p 1280:80 \
        --restart=always \
        -itd \
        xr1627119275/go-in-download
```
