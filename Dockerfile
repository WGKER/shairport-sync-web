FROM mikebrady/shairport-sync:5.0.4

# 复制 webui 并授权
COPY webui /webui
RUN chmod +x /webui

EXPOSE 8086

# ✅ 正确启动：后台webui + 前台shairport-sync（永不重启）
CMD ["/bin/sh", "-c", "/webui & exec shairport-sync"]
