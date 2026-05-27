FROM mikebrady/shairport-sync:5.0.4

# 安装依赖
RUN apt-get update && apt-get install -y \
    python3 python3-pip alsa-utils \
    && pip3 install --no-cache-dir flask \
    && apt-get clean && rm -rf /var/lib/apt/lists/* /tmp/*

# 复制文件
COPY shairport-sync.conf /etc/shairport-sync.conf
COPY web /app/web

# 暴露端口
EXPOSE 5000/tcp 7000/tcp 8080/tcp

# 启动：同时运行 AirPlay + Web 后台
CMD ["bash", "-c", "\
    shairport-sync -d & \
    cd /app/web && python3 app.py \
"]
