# 直接用官方原版（自带AirPlay2，最稳定）
FROM mikebrady/shairport-sync:5.0.4

# 只安装 Python + Flask（极简，不装任何多余依赖）
RUN apk update && apk add --no-cache python3 py3-pip \
    && pip3 install --no-cache-dir flask \
    && rm -rf /var/cache/apk/*

# 复制你的Web管理页面
COPY web /app/web
WORKDIR /app/web

# 只暴露Web端口
EXPOSE 8086

# 启动：官方原命令 + 启动Web面板
CMD ["sh", "-c", "\
    python3 app.py & \
    exec /docker-entrypoint.sh shairport-sync \
"]
