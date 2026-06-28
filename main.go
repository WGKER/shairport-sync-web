package main

import (
	"net/http"
	"os"
	"os/exec"
	"strings"
)

const (
	configFile = "/etc/shairport-sync.conf"
	staticDir  = "./static"
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

// 【完全沿用你原来正常工作的netstat检测逻辑，仅修改返回标识】
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
	// 静态图片路由，处理gif/png MIME
	staticHandler := http.StripPrefix("/static/", http.FileServer(http.Dir(staticDir)))
	fixedStatic := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, ".gif") {
			w.Header().Set("Content-Type", "image/gif")
		}
		if strings.HasSuffix(r.URL.Path, ".png") {
			w.Header().Set("Content-Type", "image/png")
		}
		w.Header().Set("Cache-Control", "max-age=1")
		staticHandler.ServeHTTP(w, r)
	})
	http.Handle("/static/", fixedStatic)

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
        h2{color:#2c3e50;margin-bottom:10px;text-align:center}
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
        /* 适配134*43小图，紧凑间距 */
        .img-wrap {
            min-height: 55px;
            display: flex;
            justify-content: center;
            align-items: center;
            margin-bottom: 3px;
        }
        .status-img {
            width: 120px;
            height: auto;
            visibility: hidden;
        }
        #statusText {
            font-size: 14px;
            color: #666;
        }
    </style>
</head>
<body>
    <div class="status-box">
        <h2>播放状态</h2>
        <div class="img-wrap" id="imgContainer">
            <img id="statusImg" class="status-img" alt="状态图">
        </div>
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
        const container = document.getElementById("imgContainer");
        let imgEl = document.getElementById("statusImg");
        const textEl = document.getElementById("statusText");
        function updateStatus(){
            fetch("/api/status")
            .then(res=>res.text())
            .then(state=>{
                // 每次重建图片DOM，重置GIF完整循环
                const newImg = document.createElement("img");
                newImg.className = "status-img";
                container.innerHTML = "";
                container.appendChild(newImg);
                imgEl = newImg;
                if(state === "playing"){
                    newImg.src = "/static/playing.gif";
                    newImg.style.visibility = "visible";
                    textEl.innerText = "正在播放";
                }else if(state === "ready"){
                    newImg.src = "/static/ready.png";
                    newImg.style.visibility = "visible";
                    textEl.innerText = "准备就绪";
                }else{
                    newImg.style.visibility = "hidden";
                    textEl.innerText = "状态加载中...";
                }
            })
            .catch(err=>{
                textEl.innerText = "状态获取失败";
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
	newName := strings.TrimPrefix(r.PostFormValue("name"), "//")
	newPassword := strings.TrimPrefix(r.PostFormValue("password"), "//")
	newDevice := strings.TrimPrefix(r.PostFormValue("device"), "//")
	newMixer := strings.TrimPrefix(r.PostFormValue("mixer"), "//")
	newVolume := strings.TrimPrefix(r.PostFormValue("volume_range_db"), "//")

	_, oldName := getConfigEx("name")
	_, oldPassword := getConfigEx("password")
	_, oldDevice := getConfigEx("output_device")
	_, oldMixer := getConfigEx("mixer_control_name")
	_, oldVolume := getConfigEx("volume_range_db")

	changed := false
	if newName != oldName || newPassword != oldPassword || newDevice != oldDevice || newMixer != oldMixer || newVolume != oldVolume {
		changed = true
	}

	if !changed {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(`
<script>alert("未检测到修改，已取消保存！");window.location.href='/'</script>
		`))
		return
	}

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
