package main

import (
	"net/http"
	"os"
	"os/exec"
	"strings"
)

const (
	configFile = "/etc/shairport-sync.conf"
)

func getShairportSyncVersion() string {
	cmd := exec.Command("shairport-sync", "-V")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "unknown"
	}
	line := strings.TrimSpace(string(out))
	parts := strings.SplitN(line, "-", 2)
	if len(parts) > 0 {
		return parts[0]
	}
	return "unknown"
}

// 终极实时状态检测（秒切播放/停止，无延迟）
func getPlayStatus() string {
	statusFiles := []string{
		"/tmp/shairport-sync.status",
		"/var/run/shairport-sync.status",
		"/run/shairport-sync.status",
	}

	for _, f := range statusFiles {
		if _, err := os.Stat(f); err == nil {
			data, _ := os.ReadFile(f)
			s := strings.ToLower(string(data))
			if strings.Contains(s, "playing") || strings.Contains(s, "connected") || strings.Contains(s, "paused") {
				return "Playing..."
			}
		}
	}

	cmd := exec.Command("pgrep", "-x", "shairport-sync")
	if err := cmd.Run(); err != nil {
		return "服务未运行"
	}

	return "Waiting for playback..."
}

func main() {
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/save", saveHandler)
	http.HandleFunc("/api/status", statusHandler)
	http.ListenAndServe(":8086", nil)
}

func statusHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Write([]byte(getPlayStatus()))
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	name := getConfig("name")
	password := getConfig("password")
	device := getConfig("output_device")
	mixer := getConfig("mixer_control_name")
	volumeRange := getConfig("volume_range_db")
	version := getShairportSyncVersion()

	html := `
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Shairport Sync 管理面板</title>
    <style>
        *{margin:0;padding:0;box-sizing:border-box;font-family:Microsoft Yahei,sans-serif}
        body{background:#f5f7fa;padding:20px;max-width:800px;margin:0 auto}
        .card{background:#fff;border-radius:10px;padding:25px;margin-bottom:20px;box-shadow:0 2px 8px #e0e0e0}
        h2{color:#2c3e50;margin-bottom:20px;text-align:center}
        label{display:block;margin:15px 0 5px;color:#34495e}
        input{width:100%;padding:10px;border:1px solid #ddd;border-radius:6px}
        .btn{
            margin-top:18px;
            padding:10px 24px;
            background:#3498db;
            color:#fff;
            border:none;
            border-radius:6px;
            cursor:pointer;
            display:block;
            margin-left:auto;
            margin-right:auto;
        }
		.version {
            position: fixed;
            bottom: 15px;
            left: 0;
            right: 0;
            text-align: center;
            font-size: 13px;
            color: #999;
        }
        .status-box {
            background: #fff;
            border-radius: 10px;
            box-shadow: 0 2px 8px #e0e0e0;
            padding:25px;
            margin-bottom:20px;
            text-align:center;
        }
        #statusText {
            font-size: 24px;
            color: #999;
        }
    </style>
</head>
<body>
    <div class="status-box">
        <h2>当前播放状态</h2>
        <div id="statusText">状态加载中...</div>
    </div>

    <div class="card">
        <h2>Shairport Sync 管理面板</h2>
        <form method="post" action="/save" onsubmit="return confirm('确定要保存并重启吗？\n重启后配置才会生效！')">
            <label>设备名称</label>
            <input type="text" name="name" value="`+name+`" placeholder="( AirPlay 名称 )">
            <label>连接密码</label>
			<input type="text" name="password" value="`+password+`" placeholder="( AirPlay 1 Only )">
            <label>声卡设备</label>
            <input type="text" name="device" value="`+device+`" placeholder="( hw:0、hw:1 等声卡序号 )">
            <label>混音器名</label>
            <input type="text" name="mixer" value="`+mixer+`" placeholder="( PCM、Master 等 )">
			<label>音量范围</label>
            <input type="text" name="volume_range_db" value="`+volumeRange+`" placeholder="( 例如：30，Range is 30 to 150 dB )">
            <button class="btn" type="submit">保存并重启生效</button>
        </form>
    </div>

	<div class="version">Shairport Sync 版本：`+version+`</div>

    <script>
        function updateStatus(){
            fetch("/api/status")
            .then(res=>res.text())
            .then(text=>{
                document.getElementById("statusText").innerText = text;
            })
        }
        updateStatus();
        setInterval(updateStatus, 1000);
    </script>
</body>
</html>`

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(html))
}

func saveHandler(w http.ResponseWriter, r *http.Request) {
	name := r.PostFormValue("name")
	password := r.PostFormValue("password")
	device := r.PostFormValue("device")
	mixer := r.PostFormValue("mixer")
	volumeRange := r.PostFormValue("volume_range_db")

	setConfig("name", name)
	setConfig("password", password)
	setConfig("output_device", device)
	setConfig("mixer_control_name", mixer)
	setConfig("volume_range_db", volumeRange)

	exec.Command("pkill", "shairport-sync").Run()
	exec.Command("shairport-sync", "-d").Run()

	http.Redirect(w, r, "/", 302)
}

func getConfig(key string) string {
	data, _ := os.ReadFile(configFile)
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			continue
		}
		if strings.HasPrefix(line, key+" = ") {
			val := strings.TrimPrefix(line, key+" = ")
			if idx := strings.Index(val, ";"); idx != -1 {
				val = val[:idx]
			}
			val = strings.TrimSpace(val)
			val = strings.Trim(val, `"`)
			val = strings.Trim(val, `'`)
			return val
		}
	}
	return ""
}

func setConfig(key, val string) {
	data, _ := os.ReadFile(configFile)
	lines := strings.Split(string(data), "\n")

	prefix := key + " = "

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if !strings.HasPrefix(trimmed, prefix) {
			continue
		}

		commentIdx := strings.Index(line, ";")
		var suffix string
		if commentIdx != -1 {
			suffix = line[commentIdx:]
		} else {
			suffix = ""
		}

		var newLine string
		if key == "volume_range_db" {
			newLine = key + " = " + val + suffix
		} else {
			newLine = key + " = \"" + val + "\"" + suffix
		}

		lines[i] = newLine
		break
	}
	os.WriteFile(configFile, []byte(strings.Join(lines, "\n")), 0644)
}
