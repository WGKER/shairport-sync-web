#!/bin/sh
# 配置与日志路径
CFG="/etc/shairport-sync.conf"
LOG="/var/log/shairport-sync.log"
PID_FILE="/var/run/shairport-sync.pid"
TMP_HTML="/tmp/tmp_page.html"

# 1. 初始化日志文件
mkdir -p /var/log
touch "$LOG"

# 2. 读取配置项
get_cfg() {
    NAME=$(grep '^name = ' "$CFG" | cut -d'"' -f2)
    PWD=$(grep '^password = ' "$CFG" | cut -d'"' -f2)
    DEV=$(grep '^output_device = ' "$CFG" | sed -E 's/.*"(.*)".*/\1/')
    MIXER=$(grep '^mixer_name = ' "$CFG" | sed -E 's/.*"(.*)".*/\1/')
}

# 3. 读取日志（取最后200行）
get_log() {
    if [ -s "$LOG" ]; then
        tail -n 200 "$LOG"
    else
        echo "暂无运行日志"
    fi
}

# 4. 渲染模板：替换 {{ 变量 }}
render_tpl() {
    local MSG="$1"
    local NAME="$2"
    local PWD="$3"
    local DEV="$4"
    local MIXER="$5"
    local LOG_TXT="$6"

    # 复制原始页面到临时文件
    cp /www/index.html "$TMP_HTML"

    # 变量替换
    sed -i "s|{{ msg }}|${MSG}|g" "$TMP_HTML"
    sed -i "s|{{ name }}|${NAME}|g" "$TMP_HTML"
    sed -i "s|{{ password }}|${PWD}|g" "$TMP_HTML"
    sed -i "s|{{ device }}|${DEV}|g" "$TMP_HTML"
    sed -i "s|{{ mixer }}|${MIXER}|g" "$TMP_HTML"
    sed -i "s|{{ log }}|${LOG_TXT}|g" "$TMP_HTML"

    # 输出最终页面
    cat "$TMP_HTML"
    rm -f "$TMP_HTML"
}

# 5. 保存配置 + 重启服务
save_and_restart() {
    local NEW_NAME="$1"
    local NEW_PWD="$2"
    local NEW_DEV="$3"
    local NEW_MIXER="$4"

    # 写入配置
    sed -i "s/^name = .*/name = \"${NEW_NAME}\"/" "$CFG"
    sed -i "s/^password = .*/password = \"${NEW_PWD}\"/" "$CFG"
    sed -i "s/^output_device = .*/output_device = \"${NEW_DEV}\"/" "$CFG"
    sed -i "s/^mixer_name = .*/mixer_name = \"${NEW_MIXER}\"/" "$CFG"

    # 停止旧进程
    if [ -f "$PID_FILE" ]; then
        PID=$(cat "$PID_FILE")
        kill "$PID" 2>/dev/null
        wait "$PID" 2>/dev/null
        rm -f "$PID_FILE"
    fi

    # 后台启动服务
    shairport-sync -P "$PID_FILE" &
    echo "配置已保存，服务重启成功！"
}

# ========== CGI 入口逻辑 ==========
# 输出HTTP头
echo "Content-Type: text/html; charset=utf-8"
echo ""

MSG=""
# 处理POST 表单提交
if [ "$REQUEST_METHOD" = "POST" ]; then
    POST_DATA=$(cat)
    # 解析表单参数
    ACTION=$(echo "$POST_DATA" | awk -F'&' '{for(i=1;i<=NF;i++){print $i}}' | grep 'action=' | cut -d'=' -f2)
    if [ "$ACTION" = "save_cfg" ]; then
        NAME=$(echo "$POST_DATA" | awk -F'&' '{for(i=1;i<=NF;i++){print $i}}' | grep 'name=' | cut -d'=' -f2)
        PWD=$(echo "$POST_DATA" | awk -F'&' '{for(i=1;i<=NF;i++){print $i}}' | grep 'password=' | cut -d'=' -f2)
        DEV=$(echo "$POST_DATA" | awk -F'&' '{for(i=1;i<=NF;i++){print $i}}' | grep 'device=' | cut -d'=' -f2)
        MIXER=$(echo "$POST_DATA" | awk -F'&' '{for(i=1;i<=NF;i++){print $i}}' | grep 'mixer=' | cut -d'=' -f2)

        # URL解码（简易处理空格/常规字符）
        NAME=$(echo "$NAME" | sed 's/+/ /g')
        PWD=$(echo "$PWD" | sed 's/+/ /g')
        DEV=$(echo "$DEV" | sed 's/+/ /g')
        MIXER=$(echo "$MIXER" | sed 's/+/ /g')

        # 执行保存重启
        MSG=$(save_and_restart "$NAME" "$PWD" "$DEV" "$MIXER")
    fi
fi

# 读取当前配置 & 日志
get_cfg
LOG_TXT=$(get_log)

# 渲染并输出页面
render_tpl "$MSG" "$NAME" "$PWD" "$DEV" "$MIXER" "$LOG_TXT"
