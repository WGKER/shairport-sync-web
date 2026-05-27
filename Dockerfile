# 阶段1：用标准 alpine 装 python（干净、无错）
FROM alpine:3.20 AS python
RUN apk add --no-cache python3 py3-pip \
    && python3 -m pip install --no-cache-dir flask

# 阶段2：直接用官方 AirPlay2 原版镜像（不动它）
FROM mikebrady/shairport-sync:5.0.4

# 只把 python 复制进来（不执行任何 apk，永不报错）
COPY --from=python /usr/bin/python3 /usr/bin/
COPY --from=python /usr/lib/python3 /usr/lib/python3
COPY --from=python /usr/lib/libpython3.so* /usr/lib/

# 复制你的 web 管理页面
COPY web /app/web
WORKDIR /app/web

# 暴露 web 端口
EXPOSE 8086

# 你要的启动命令（完全不变）
CMD ["sh", "-c", "\
    python3 app.py & \
    exec /docker-entrypoint.sh shairport-sync \
"]
