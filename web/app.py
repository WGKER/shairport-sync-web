from flask import Flask, render_template, request
import os
import re
import subprocess
import time

app = Flask(__name__)

CONF_PATH = "/etc/shairport-sync.conf"
LOG_PATH = "/var/log/shairport-sync.log"

# 日志初始化
if not os.path.exists(LOG_PATH):
    open(LOG_PATH, "w").close()

# 读取配置
def read_config():
    with open(CONF_PATH, "r", encoding="utf-8") as f:
        return f.read()

# 写入配置（增加 mixer_control_name）
def write_config(name, password, device, mixer):
    cfg = read_config()
    cfg = re.sub(r'name\s*=\s*"[^"]*"', f'name = "{name}"', cfg)
    cfg = re.sub(r'password\s*=\s*"[^"]*"', f'password = "{password}"', cfg)
    cfg = re.sub(r'output_device\s*=\s*"[^"]*"', f'output_device = "{device}"', cfg)
    cfg = re.sub(r'mixer_control_name\s*=\s*"[^"]*"', f'mixer_control_name = "{mixer}"', cfg)
    
    with open(CONF_PATH, "w", encoding="utf-8") as f:
        f.write(cfg)

# 重启服务
def restart_service():
    subprocess.run(["pkill", "-f", "shairport-sync"], check=False)
    subprocess.run(["pkill", "-f", "nqptp"], check=False)
    time.sleep(1)
    subprocess.Popen(["nqptp"], stdout=open(LOG_PATH, "a"), stderr=open(LOG_PATH, "a"))
    subprocess.Popen(["shairport-sync"], stdout=open(LOG_PATH, "a"), stderr=open(LOG_PATH, "a"))

# 获取日志
def get_log():
    try:
        with open(LOG_PATH, "r", encoding="utf-8", errors="ignore") as f:
            lines = f.readlines()
        return "".join(lines[-50:]) if lines else "暂无日志"
    except:
        return "读取日志失败"

# 主页
@app.route('/', methods=['GET', 'POST'])
def index():
    msg = ""

    if request.method == 'POST':
        action = request.form.get("action")
        if action == "save_cfg":
            name = request.form['name'].strip()
            password = request.form['password'].strip()
            device = request.form['device'].strip()
            mixer = request.form['mixer'].strip()  # 新增
            write_config(name, password, device, mixer)
            restart_service()
            msg = "✅ 配置已保存，服务重启成功！"

    # 读取当前配置
    cfg = read_config()
    name = re.search(r'name\s*=\s*"([^"]*)"', cfg).group(1)
    pwd = re.search(r'password\s*=\s*"([^"]*)"', cfg).group(1)
    dev = re.search(r'output_device\s*=\s*"([^"]*)"', cfg).group(1)
    mixer = re.search(r'mixer_control_name\s*=\s*"([^"]*)"', cfg).group(1)  # 新增
    log_content = get_log()

    return render_template(
        "index.html",
        msg=msg,
        name=name,
        password=pwd,
        device=dev,
        mixer=mixer,
        log=log_content
    )

if __name__ == '__main__':
    app.run(host='0.0.0.0', port=8086, debug=False)
