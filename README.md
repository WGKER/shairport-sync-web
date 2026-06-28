# shairport-sync-web

### 为 shairport-sync 添加 web 管理面板

  shairport-sync docker版没有web管理界面\
  自定义名称、声卡等参数需修改shairport-sync.conf配置文件，操作起来比较麻烦\
  web管理面板提供简洁快速的处理方案，无需繁琐查找，即可轻松自定义主要参数

### 更新日志
2026-06-28：\
1、播放状态添加示意图\
2、移除版本号相关\
3、页面布局及文字细节优化\

### ✏ 进度

web功能布局已完善，可用

<img width="766" height="685" alt="image" src="https://github.com/user-attachments/assets/a08fa1d0-5867-472c-8918-fdb7c5bfea64" />
<img width="767" height="686" alt="image" src="https://github.com/user-attachments/assets/eed6dcca-541d-42e8-8057-c7f72c0d71f7" />

### ✏ 使用方式一（临时）

  把webui导入已运行的shairport-sync容器\
  赋权\
  临时运行（重启失效，需修改启动脚本可自动重启）\
  访问IP:8086

### ✏ 使用方式二（持久）

  运行shairport-sync-web容器\
  已集成webui，基于shairport-sync镜像打包\
  访问IP:8086
