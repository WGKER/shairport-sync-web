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

// 从 shairport-sync -V 自动获取版本号
func getShairportSyncVersion() string {
	cmd := exec.Command("shairport-sync", "-V")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "unknown"
	}
// 自动去掉 - 后面的所有内容，只保留主版本号
	line := strings.TrimSpace(string(out))
	parts := strings.SplitN(line, "-", 2)
	if len(parts) > 0 {
		return parts[0]
	}

	return "unknown"
}

// ✅ 官方原生必生效的状态检测（所有 shairport 通用）
func getPlayStatus() string {
	// 检测是否有客户端连接（AirPlay 连接 = 正在播放/暂停）
	cmd := exec.Command("sh", "-c", "netstat -anp | grep shairport | grep ESTABLISHED | wc -l")
	out, _ := cmd.CombinedOutput()
	count := strings.TrimSpace(string(out))

	if count != "0" {
		return "PLAYING..."
	}
	return "WAITING FOR PLAYBACK..."
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
	volumeRange := getConfig("volume_range_db") // 新增
	version := getShairportSyncVersion() // 自动获取

html := `
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>设置管理面板</title>
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
            font-size: 13px;
            color: #999;
        }
		/* 红色缓慢闪烁动画 */
        @keyframes red-blink {
            0% { color: #ff0000; }
            50% { color: #cc0000; }
            100% { color: #ff0000; }
        }
        .blink {
            animation: red-blink 1.5s infinite ease-in-out;
        }
    </style>
</head>
<body>
    <div class="status-box">
        <h2>当前播放状态</h2>
        <div id="statusText">状态加载中...</div>
    </div>
    
    <div class="card">
        <h2>设置管理面板</h2>
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
	<div class="version">Shairport Sync 版本： `+version+`</div>

    <script>
        function updateStatus(){
            fetch("/api/status")
            .then(res=>res.text())
            .then(text=>{
                const el = document.getElementById("statusText");
                el.innerText = text;
                // 播放时：红色闪烁，其他状态：灰色
                if(text === "PLAYING..."){
                    el.classList.add("blink");
                }else{
                    el.classList.remove("blink");
                }
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

// 读取配置，自动忽略 ; 开头的注释
func getConfig(key string) string {
	data, _ := os.ReadFile(configFile)
	for _, line := range strings.Split(string(data), "\n") {
		// 去掉前后空格
		line = strings.TrimSpace(line)

		// 忽略注释行
		if strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			continue
		}

		// 匹配 key = value
		if strings.HasPrefix(line, key+" = ") {
			val := strings.TrimPrefix(line, key+" = ")

			// 遇到 ; 就截断，忽略后面的注释
			if idx := strings.Index(val, ";"); idx != -1 {
				val = val[:idx]
			}

			// 去掉引号
			val = strings.TrimSpace(val)
			val = strings.Trim(val, `"`)
			val = strings.Trim(val, `'`)
			return val
		}
	}
	return ""
}

// 写入配置
// 重点：volume_range_db 不加双引号
// 优化后的读取：支持 // # ; 注释行，都能读到值
func getConfig(key string) string {
	data, _ := os.ReadFile(configFile)
	searchStr := key + " = "

	for _, line := range strings.Split(string(data), "\n") {
		trimmed := strings.TrimSpace(line)

		// 跳过空行
		if trimmed == "" {
			continue
		}

		// 去掉行首注释符号 // # ; 后再检查
		cleanLine := trimmed
		if strings.HasPrefix(trimmed, "//") {
			cleanLine = strings.TrimSpace(trimmed[2:])
		} else if strings.HasPrefix(trimmed, "#") || strings.HasPrefix(trimmed, ";") {
			cleanLine = strings.TrimSpace(trimmed[1:])
		}

		// 匹配配置项
		if strings.HasPrefix(cleanLine, searchStr) {
			val := strings.TrimPrefix(cleanLine, searchStr)

			// 去掉行尾注释
			if idx := strings.Index(val, ";"); idx != -1 {
				val = val[:idx]
			}
			if idx := strings.Index(val, "#"); idx != -1 {
				val = val[:idx]
			}
			if idx := strings.Index(val, "//"); idx != -1 {
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

// 优化后的写入：自动取消 // 注释，保存后变成有效配置行
func setConfig(key, val string) {
	data, _ := os.ReadFile(configFile)
	lines := strings.Split(string(data), "\n")
	prefix := key + " = "

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		originalLine := line

		// 解析是否为注释行
		isComment := false
		cleanLine := trimmed
		if strings.HasPrefix(trimmed, "//") {
			cleanLine = strings.TrimSpace(trimmed[2:])
			isComment = true
		} else if strings.HasPrefix(trimmed, "#") || strings.HasPrefix(trimmed, ";") {
			cleanLine = strings.TrimSpace(trimmed[1:])
			isComment = true
		}

		// 找到目标配置项
		if strings.HasPrefix(cleanLine, prefix) {
			// 保留原始缩进
			indent := ""
			if len(originalLine) > len(trimmed) {
				indent = originalLine[:len(originalLine)-len(trimmed)]
			}

			// 保留行尾注释
			commentPart := ""
			commentIdx := -1
			if idx := strings.Index(cleanLine, ";"); idx != -1 {
				commentIdx = idx
			} else if idx := strings.Index(cleanLine, "#"); idx != -1 {
				commentIdx = idx
			} else if idx := strings.Index(cleanLine, "//"); idx != -1 {
				commentIdx = idx
			}
			if commentIdx != -1 {
				commentPart = strings.TrimSpace(cleanLine[commentIdx:])
			}

			// 构建新行：如果是 // 注释，自动取消注释
			var newLine string
			if key == "volume_range_db" {
				newLine = key + " = " + val
			} else {
				newLine = key + " = \"" + val + "\""
			}
			if commentPart != "" {
				newLine += " " + commentPart
			}

			// 恢复缩进
			lines[i] = indent + newLine
			break
	}
	os.WriteFile(configFile, []byte(strings.Join(lines, "\n")), 0644)
}
