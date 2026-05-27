# shairport-sync-web
## 为 shairport-sync 添加 web 管理面板

## 目录结构

    shairport-sync-web/
    ├── Dockerfile
    ├── start.sh
    ├── www/
    │   ├── index.html       # 你的样式页面（纯模板，由CGI填充变量）
    │   └── cgi-bin/
    │       └── main.cgi     # 统一入口CGI（读取配置、渲染页面、处理提交、输出日志）
    └── shairport-sync.conf  # 适配新字段的配置文件
