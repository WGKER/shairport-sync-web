FROM mikebrady/shairport-sync:5.0.4

# 安装依赖
RUN apk update && \
    # 先装 alsa-lib（确保有）+ 必要依赖
    apk add --no-cache \
        python3 \
        py3-pip && \
    # 用 python3 -m pip 最稳
    python3 -m pip install --no-cache-dir flask && \
    # 清理
    rm -rf /var/cache/apk/*

# 复制文件
COPY web /app/web

# 暴露端口
EXPOSE 8086

# 启动：同时运行 shairport-sync + Web 面板
CMD ["sh", "-c", "shairport-sync & cd /app/web && python3 app.py"]
