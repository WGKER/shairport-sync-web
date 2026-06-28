FROM mikebrady/shairport-sync:5.0.4

# 复制 webui 并授权
COPY webui /webui
RUN chmod +x /webui

EXPOSE 8086

# 关键：自动把 webui 启动命令插入 run.sh 第五行
RUN sed -i '5i /webui &' /run.sh

# 使用官方启动脚本
CMD ["/run.sh"]
