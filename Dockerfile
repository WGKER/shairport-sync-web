FROM mikebrady/shairport-sync:5.0.4

# 复制 Actions 编译好的 webui
COPY webui /webui
RUN chmod +x /webui

EXPOSE 8086

# 启动 webui + shairport-sync
CMD ["/bin/sh", "-c", "shairport-sync && exec /webui"]
