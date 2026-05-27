FROM mikebrady/shairport-sync:latest
LABEL maintainer="shairport-sync-web"

# 复制文件
COPY ./www /www
COPY ./start.sh /start.sh

# 添加执行权限
RUN chmod +x /start.sh \
    && chmod +x /www/cgi-bin/main.cgi

# 端口：5000(AirPlay) 8080(Web管理)
EXPOSE 8086

CMD ["/start.sh"]
