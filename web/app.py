from flask import Flask, render_template, request
import os
import re
import subprocess
import time

app = Flask(__name__)

# 全局路径定义
CONF_PATH = "/etc/shairport-sync.conf"
LOG_PATH = "/var/log/shairport-sync.log"
MIXER_CMD = "amixer"

# 初始化日志文件
if not os.path.exists(LOG_PATH):
    open(LOG_PATH, "w").close()

# 读取配置文件
def read_config():
    with open(CONF_PATH, "r", encoding="utf-8") as f:
        return f.read()

# 写入配置文件
def write_config(name, password, device, mixer):
    cfg = read_config()
    cfg = re.sub(r'name\s*=\s*"[^"]*"', f'name = "{name}"', cfg)
    cfg = re.sub(r'password\s*=\s*"[^"]*"', f'password = "{password}"', cfg)
    cfg = re.sub(r'output_device\s*=\s*"[^"]*"', f'output_device = "{device}"', cfg)
    cfg = re.sub(r'mixer_control_name\s*=\s*"[^"]*"', f'mixer_control_name = "{mixer}"', cfg)
    with open(CONF_PATH, "w", encoding="utf-8") as f:
        f.write(cfg)

# 重启 shairport-sync
def restart_service():
    subprocess.run(["pkill", "-f", "shairport-sync"], check=False)
    time.sleep(1)
    # 后台启动并输出日志
    subprocess.Popen(["shairport-sync", "-d"], stdout=open(LOG_PATH, "a"), stderr=open(LOG_PATH, "a"))

# 获取当前音量
def get_volume(mixer_name):
    try:
        res = subprocess.check_output([MIXER_CMD, "get", mixer_name], text=True)
        vol = re.search(r'(\d+)%', res)
        return vol.group(1) if vol else "50"
    except:
        return "50"

# 设置音量
def set_volume(mixer_name, vol):
    try:
        subprocess.run([MIXER_CMD, "set", mixer_name, f"{vol}%"], check=False)
    except:
        pass

# 获取最新日志（最后50行）
def get_log():
    try:
        with open(LOG_PATH, "r", encoding="utf-8", errors="ignore") as f:
            lines = f.readlines()
        return "".join(lines[-50:]) if lines else "暂无日志"
    except:
        return "读取日志失败"

# 主页路由
@app.route('/', methods=['GET', 'POST'])
def index():
    msg = ""
    # 提交配置表单
    if request.method == 'POST':
        action = request.form.get("action", "")
        mixer_name = re.search(r'mixer_control_name\s*=\s*"([^"]*)"', read_config()).group(1)

        # 保存基础配置
        if action == "save_cfg":
            name = request.form['name'].strip()
            pwd = request.form['password'].strip()
            dev = request.form['device'].strip()
            mix = request.form['mixer'].strip()
            write_config(name, pwd, dev, mix)
            restart_service()
            msg = "✅ 配置已保存，服务已重启！"

        # 调节音量
        elif action == "set_vol":
            vol = request.form['volume'].strip()
            set_volume(mixer_name, vol)
            msg = f"✅ 音量已设置为 {vol}%"

    # 读取当前配置回填
    cfg = read_config()
    name = re.search(r'name\s*=\s*"([^"]*)"', cfg).group(1)
    pwd = re.search(r'password\s*=\s*"([^"]*)"', cfg).group(1)
    dev = re.search(r'output_device\s*=\s*"([^"]*)"', cfg).group(1)
    mix = re.search(r'mixer_control_name\s*=\s*"([^"]*)"', cfg).group(1)
    vol = get_volume(mix)
    log_content = get_log()

    return render_template(
        'index.html',
        msg=msg,
        name=name,
        password=pwd,
        device=dev,
        mixer=mix,
        volume=vol,
        log=log_content
    )

if __name__ == '__main__':
    app.run(host='0.0.0.0', port=8086, debug=False)
