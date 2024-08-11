# 基础镜像
FROM ubuntu:latest
LABEL authors="crazyfrank"

# 把编译后的镜像打包到这个位置 放到工作目录
COPY main /usr/local/bin
WORKDIR /usr/local/bin

ENTRYPOINT ["/usr/local/bin/main"]