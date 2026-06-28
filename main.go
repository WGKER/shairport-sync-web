package main

import (
	"net/http"
	"os"
	"os/exec"
	"strings"
)

const (
	configFile = "/etc/shairport-sync.conf"
	staticDir  = "./static" // 静态图片存放目录
)

// 从 shairport-sync -V 自动获取版本号
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

// 播放状态检测：返回固定标识字符串供前端判断
func getPlayStatus() string {
	cmd := exec.Command("sh", "-c", "netstat -anp | grep shairport | grep ESTABLISHED | wc -l")
	out, _ := cmd.CombinedOutput()
	count := strings.TrimSpace(string(out))

	if count != "0" {
		return "playing"
	}
	return "ready"
}

func main() {
	// 静态文件路由，映射/static/访问静态文件夹图片
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(staticDir))))
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/save", saveHandler)
	http.HandleFunc("/api/status", statusHandler)
	http.ListenAndServe(":8086", nil)
}

func statusHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Write([]byte(getPlayStatus()))
}

// 读取配置并返回：是否为//注释行 + 实际值
func getConfigEx(key string) (bool, string) {
	data, _ := os.ReadFile(configFile)
	searchStr := key + " = "

	for _, line := range strings.Split(string(data), "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		isComment := false
		cleanLine := trimmed

		if strings.HasPrefix(trimmed, "//") {
			isComment = true
			cleanLine = strings.TrimSpace(trimmed[2:])
		} else if strings.HasPrefix(trimmed, "#") || strings.HasPrefix(trimmed, ";") {
			cleanLine = strings.TrimSpace(trimmed[1:])
		}

		if strings.HasPrefix(cleanLine, searchStr) {
			val := strings.TrimPrefix(cleanLine, searchStr)

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
			return isComment, val
		}
	}
	return false, ""
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	// 读取配置
	isNameComment, name := getConfigEx("name")
	isPwdComment, password := getConfigEx("password")
	isDevComment, device := getConfigEx("output_device")
	isMixerComment, mixer := getConfigEx("mixer_control_name")
	isVolComment, volumeRange := getConfigEx("volume_range_db")

	// 如果是//注释，前面加//
	displayName := name
	if isNameComment {
		displayName = "//" + name
	}
	displayPwd := password
	if isPwdComment {
		displayPwd = "//" + password
	}
	displayDevice := device
	if isDevComment {
		displayDevice = "//" + device
	}
	displayMixer := mixer
	if isMixerComment {
		displayMixer = "//" + mixer
	}
	displayVol := volumeRange
	if isVolComment {
		displayVol = "//" + volumeRange
	}

	version := getShairportSyncVersion()

html := `
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>ShairportSync设置</title>
    <style>
        *{margin:0;padding:0;box-sizing:border-box;font-family:Microsoft Yahei,sans-serif}
        body{background:#f5f7fa;padding:20px;max-width:800px;margin:0 auto}
        .card{background:#fff;border-radius:10px;padding:25px;margin-bottom:20px;box-shadow:0 2px 8px #e0e0e0}
        h2{color:#2c3e50;margin-bottom:20px;text-align:center}
        label{display:block;margin:15px 0 5px;color:#34495e}
        input{width:100%;padding:10px;border:1px solid #ddd;border-radius:6px}
        .gray { color: #999 !important; }
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
        .status-img {
            width: 120px;
            height: auto;
            margin-bottom: 12px;
        }
        #statusText {
            font-size: 16px;
            color: #333;
        }
    </style>
</head>
<body>
    <div class="status-box">
        <h2>播放状态</h2>
        <!-- 初始src为空，加载中不显示图片 -->
        <img id="statusImg" class="status-img" src="" alt="状态图">
        <!-- 状态文字 -->
        <div id="statusText">状态加载中...</div>
    </div>
    
    <div class="card">
        <h2>配置管理</h2>
        <form method="post" action="/save" onsubmit="return confirm('确定要保存并重启吗？\n重启后配置才会生效！')">
            <label>设备名称</label>
            <input type="text" name="name" value="`+displayName+`" class="`+ifTrue(isNameComment, "gray")+`" placeholder="( AirPlay 名称 )">
            
            <label>连接密码</label>
			<input type="text" name="password" value="`+displayPwd+`" class="`+ifTrue(isPwdComment, "gray")+`" placeholder="( AirPlay 1 Only )">
            
            <label>声卡设备</label>
            <input type="text" name="device" value="`+displayDevice+`" class="`+ifTrue(isDevComment, "gray")+`" placeholder="( hw:0、hw:1 等声卡序号 )">
            
            <label>混音器名</label>
            <input type="text" name="mixer" value="`+displayMixer+`" class="`+ifTrue(isMixerComment, "gray")+`" placeholder="( PCM、Master 等 )">
            
			<label>音量范围</label>
            <input type="text" name="volume_range_db" value="`+displayVol+`" class="`+ifTrue(isVolComment, "gray")+`" placeholder="( 例如：30，Range is 30 to 150 dB )">
            
            <button class="btn" type="submit">保存并重启生效</button>
        </form>
    </div>
	<div class="version">Shairport Sync 版本：`+version+`</div>

    <script>
        const imgEl = document.getElementById("statusImg");
        const textEl = document.getElementById("statusText");
        function updateStatus(){
            fetch("/api/status")
            .then(res=>res.text())
            .then(state=>{
                if(state === "playing"){
                    imgEl.src = "/static/playing.gif";
                    textEl.innerText = "正在播放";
                }else if(state === "ready"){
                    imgEl.src = "/static/ready.png";
                    textEl.innerText = "准备就绪";
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

func ifTrue(cond bool, s string) string {
	if cond {
		return s
	}
	return ""
}

// 优化：未修改则提示并取消保存
func saveHandler(w http.ResponseWriter, r *http.Request) {
	// 读取表单提交的值
	newName := strings.TrimPrefix(r.PostFormValue("name"), "//")
	newPassword := strings.TrimPrefix(r.PostFormValue("password"), "//")
	newDevice := strings.TrimPrefix(r.PostFormValue("device"), "//")
	newMixer := strings.TrimPrefix(r.PostFormValue("mixer"), "//")
	newVolume := strings.TrimPrefix(r.PostFormValue("volume_range_db"), "//")

	// 读取原有配置
	_, oldName := getConfigEx("name")
	_, oldPassword := getConfigEx("password")
	_, oldDevice := getConfigEx("output_device")
	_, oldMixer := getConfigEx("mixer_control_name")
	_, oldVolume := getConfigEx("volume_range_db")

	// 检查是否有修改
	changed := false
	if newName != oldName || newPassword != oldPassword || newDevice != oldDevice || newMixer != oldMixer || newVolume != oldVolume {
		changed = true
	}

	// 未修改 → 弹出提示并返回
	if !changed {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(`
<script>alert("未检测到修改，已取消保存！");window.location.href='/'</script>
		`))
		return
	}

	// 有修改才执行保存
	if newName != oldName {
		setConfig("name", newName)
	}
	if newPassword != oldPassword {
		setConfig("password", newPassword)
	}
	if newDevice != oldDevice {
		setConfig("output_device", newDevice)
	}
	if newMixer != oldMixer {
		setConfig("mixer_control_name", newMixer)
	}
	if newVolume != oldVolume {
		setConfig("volume_range_db", newVolume)
	}

	// 重启服务
	exec.Command("pkill", "shairport-sync").Run()
	exec.Command("shairport-sync", "-d").Run()

	// 返回主页
	http.Redirect(w, r, "/", 302)
}

// 写入配置：自动取消 // 注释
func setConfig(key, val string) {
	data, _ := os.ReadFile(configFile)
	lines := strings.Split(string(data), "\n")
	prefix := key + " = "

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		originalLine := line

		cleanLine := trimmed
		if strings.HasPrefix(trimmed, "//") {
			cleanLine = strings.TrimSpace(trimmed[2:])
		} else if strings.HasPrefix(trimmed, "#") || strings.HasPrefix(trimmed, ";") {
			cleanLine = strings.TrimSpace(trimmed[1:])
		}

		if strings.HasPrefix(cleanLine, prefix) {
			indent := ""
			if len(originalLine) > len(trimmed) {
				indent = originalLine[:len(originalLine)-len(trimmed)]
			}

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

			var newLine string
			if key == "volume_range_db" {
				newLine = key + " = " + val
			} else {
				newLine = key + " = \"" + val + "\""
			}
			if commentPart != "" {
				newLine += " " + commentPart
			}

			lines[i] = indent + newLine
			break
		}
	}
	os.WriteFile(configFile, []byte(strings.Join(lines, "\n")), 0644)
}
