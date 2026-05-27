FROM mikebrady/shairport-sync:5.0.4

# 安装依赖
RUN apk update && \
    apk add --no-cache \
    python3 \
    py3-pip \
    alsa-utils && \
    pip3 install --no-cache-dir flask

# 复制文件
COPY web /app/web

# 暴露端口
EXPOSE 8086

# 启动：同时运行 shairport-sync + Web 面板
CMD ["sh", "-c", "shairport-sync & cd /app/web && python3 app.py"]
