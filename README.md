# shairport-sync-web

### 为 shairport-sync 添加 web 管理面板
    shairport-sync docker版没有web管理界面
    自定义名称、声卡等参数需修改shairport-sync.conf配置文件，操作起来比较麻烦
    web管理面板提供简洁快速的处理方案，无需繁琐查找，即可轻松自定义主要参数

### 进度
    web功能布局已完善，可用
    
### 使用方式一（临时）
    把webui导入已运行的shairport-sync容器
    赋权
    临时运行（重启失效，需修改启动脚本可自动重启）
    访问IP:8086

### 使用方式二（持久）
    运行shairport-sync-web容器
    已集成webui，基于shairport-sync镜像打包
    访问IP:8086
