#!/bin/sh
set -e

mkdir -p /var/log
touch /var/log/shairport-sync.log

# 启动内置HTTP服务，根目录/www，CGI目录/www/cgi-bin，端口8080
httpd -h /www -p 8086 -c /www/cgi-bin

# 前台运行主程序，保持容器存活
exec shairport-sync -P /var/run/shairport-sync.pid
