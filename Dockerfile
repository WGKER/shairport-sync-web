# --------------------------
# 阶段1：从官方镜像提取 AirPlay2 程序（shairport-sync + nqptp）
# --------------------------
FROM mikebrady/shairport-sync:5.0.4 AS airplay2
# 把关键二进制文件拷出来
RUN cp /usr/local/bin/shairport-sync /tmp/ && \
    cp /usr/local/bin/nqptp /tmp/

# --------------------------
# 阶段2：构建 Python/Flask + ALSA 环境（标准 Alpine）
# --------------------------
FROM alpine:3.20

# 安装基础依赖（alsa、python、flask）
RUN apk update && apk add --no-cache \
    alsa-lib \
    alsa-utils \
    python3 \
    py3-pip \
    avahi \
    avahi-compat-libdns_sd \
    dbus \
    && python3 -m pip install --no-cache-dir flask \
    && rm -rf /var/cache/apk/*

# 从阶段1复制 AirPlay2 程序
COPY --from=airplay2 /tmp/shairport-sync /usr/local/bin/
COPY --from=airplay2 /tmp/nqptp /usr/local/bin/

# 你的 Web 代码
COPY web /app/web
WORKDIR /app/web

# 开放端口
EXPOSE 8080

# 启动脚本：先 nqptp（AirPlay2 必需）→ shairport-sync → Flask
CMD ["sh", "-c", "nqptp & shairport-sync & cd /app/web && python3 app.py"]
