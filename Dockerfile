FROM mikebrady/shairport-sync:5.2.3

# 复制 webui 并授权
COPY webui /webui
RUN chmod +x /webui

# 复制静态图片文件夹到容器根目录（和Go代码读取路径 ./static 对应）
COPY static /static
# 赋予全局可读权限，避免图片加载失败
RUN chmod -R 755 /static

EXPOSE 8086

# 关键：自动把 webui 启动命令插入 run.sh 第五行
RUN sed -i '5i /webui &' /run.sh

# 使用官方启动脚本
CMD ["/run.sh"]
