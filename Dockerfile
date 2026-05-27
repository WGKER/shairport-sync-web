FROM mikebrady/shairport-sync:5.0.4

# 安装依赖
RUN apt-get update && apt-get install -y \
    python3 python3-pip alsa-utils \
    && pip3 install --no-cache-dir flask \
    && apt-get clean && rm -rf /var/lib/apt/lists/* /tmp/*

# 复制文件
COPY web /app/web

# 暴露端口
EXPOSE 8086

# 启动：同时运行 shairport-sync + Web 面板
CMD ["sh", "-c", "shairport-sync & cd /app/web && python3 app.py"]
