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

// 沿用原有正常工作的netstat检测逻辑
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
	http.HandleFunc("/save-confirm", saveConfirmHandler)
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
	// 新增service-type读取
	isSvcTypeComment, serviceType := getConfigEx("service_type")
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
	// service-type页面展示值，默认auto
	displaySvcType := serviceType
	if isSvcTypeComment {
		displaySvcType = "//" + serviceType
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

html := `
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>ShairportSync管理面板</title>
    <style>
        *{margin:0;padding:0;box-sizing:border-box;font-family:Microsoft Yahei,sans-serif}
        body{background:#f5f7fa;padding:20px;max-width:800px;margin:0 auto}
        .card{background:#fff;border-radius:10px;padding:25px;margin-bottom:20px;box-shadow:0 2px 8px #e0e0e0}
        h2{color:#2c3e50;margin-bottom:5px;text-align:center}
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
        <div id="statusText">（状态加载中...）</div>
    </div>
    
    <div class="card">
        <h2>参数配置</h2>
        <form method="post" action="/save">
            <label>设备名称</label>
            <input type="text" name="name" value="`+displayName+`" class="`+ifTrue(isNameComment, "gray")+`" placeholder="( AirPlay 名称 )">
            
            <label>连接密码</label>
			<input type="text" name="password" value="`+displayPwd+`" class="`+ifTrue(isPwdComment, "gray")+`" placeholder="( AirPlay 1 & 2 )">

            <!-- 新增service_type输入框，放在密码下方 -->
            <label>服务类型</label>
			<input type="text" name="service_type" value="`+displaySvcType+`" class="`+ifTrue(isSvcTypeComment, "gray")+`" placeholder="( auto(默认) / classic / airplay2 )">
            
            <label>声卡设备</label>
            <input type="text" name="output_device" value="`+displayDevice+`" class="`+ifTrue(isDevComment, "gray")+`" placeholder="( hw:0、hw:1 等声卡序号 )">
            
            <label>混音器名</label>
            <input type="text" name="mixer_control_name" value="`+displayMixer+`" class="`+ifTrue(isMixerComment, "gray")+`" placeholder="( PCM、Master 等 )">
            
			<label>音量范围</label>
            <input type="text" name="volume_range_db" value="`+displayVol+`" class="`+ifTrue(isVolComment, "gray")+`" placeholder="( 例如：30，Range is 30 to 150 dB )">
            
            <button class="btn" type="submit">保存并重启生效</button>
        </form>
    </div>

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
                    textEl.innerText = "（正在播放）";
                }else if(state === "ready"){
                    newImg.src = "/static/ready.png";
                    newImg.style.visibility = "visible";
                    textEl.innerText = "（准备就绪）";
                }else{
                    newImg.style.visibility = "hidden";
                    textEl.innerText = "（状态加载中...）";
                }
            })
            .catch(err=>{
                textEl.innerText = "（状态获取失败）";
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

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	// 增加POST限制
	if r.Method != http.MethodPost {
		w.Write([]byte(`<script>alert("非法请求");history.back()</script>`))
		return
	}
	
	// 统一清洗：先去//再清首尾空格，和saveConfirm保持一致
	newName := strings.TrimSpace(strings.TrimPrefix(r.PostFormValue("name"), "//"))
	newPassword := strings.TrimSpace(strings.TrimPrefix(r.PostFormValue("password"), "//"))
	newServiceType := strings.TrimSpace(strings.TrimPrefix(r.PostFormValue("service_type"), "//"))
	newDevice := strings.TrimSpace(strings.TrimPrefix(r.PostFormValue("output_device"), "//"))
	newMixer := strings.TrimSpace(strings.TrimPrefix(r.PostFormValue("mixer_control_name"), "//"))
	newVolume := strings.TrimSpace(strings.TrimPrefix(r.PostFormValue("volume_range_db"), "//"))

	_, oldName := getConfigEx("name")
	_, oldPassword := getConfigEx("password")
	_, oldServiceType := getConfigEx("service_type")
	_, oldDevice := getConfigEx("output_device")
	_, oldMixer := getConfigEx("mixer_control_name")
	_, oldVolume := getConfigEx("volume_range_db")

	changed := false
	if newName != oldName || newPassword != oldPassword || newServiceType != oldServiceType || newDevice != oldDevice || newMixer != oldMixer || newVolume != oldVolume {
		changed = true
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if !changed {
		// 分支1：无修改弹窗，直接返回首页
		w.Write([]byte(`
<script>
alert("没有修改，保存取消！");
</script>
		`))
		return
	}

	// 分支2：存在修改，仅输出弹窗+隐藏表单，【不执行写入、不重启】
	w.Write([]byte(`
<form id="confirmForm" action="/save-confirm" method="POST" style="display:none">
	<input name="name" value="` + newName + `">
	<input name="password" value="` + newPassword + `">
	<input name="service_type" value="` + newServiceType + `">
	<input name="output_device" value="` + newDevice + `">
	<input name="mixer_control_name" value="` + newMixer + `">
	<input name="volume_range_db" value="` + newVolume + `">
</form>
<script>
if (confirm("确认修改并重启服务生效？")){
	document.getElementById("confirmForm").submit();
}
</script>`))
}

// 新增 /save-confirm 处理确认后的写入、重启（放在同文件内）
func saveConfirmHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	// 限制仅POST提交
	if r.Method != http.MethodPost {
		w.Write([]byte(`<script>alert("非法请求方式");history.back()</script>`))
		return
	}
	
// 读取并清理输入
	newName := strings.TrimSpace(strings.TrimPrefix(r.PostFormValue("name"), "//"))
	newPassword := strings.TrimSpace(strings.TrimPrefix(r.PostFormValue("password"), "//"))
	newServiceType := strings.TrimSpace(strings.TrimPrefix(r.PostFormValue("service_type"), "//"))
	newDevice := strings.TrimSpace(strings.TrimPrefix(r.PostFormValue("output_device"), "//"))
	newMixer := strings.TrimSpace(strings.TrimPrefix(r.PostFormValue("mixer_control_name"), "//"))
	newVolume := strings.TrimSpace(strings.TrimPrefix(r.PostFormValue("volume_range_db"), "//"))

// 获取旧配置
	_, oldName := getConfigEx("name")
	_, oldPassword := getConfigEx("password")
	_, oldServiceType := getConfigEx("service_type")
	_, oldDevice := getConfigEx("output_device")
	_, oldMixer := getConfigEx("mixer_control_name")
	_, oldVolume := getConfigEx("volume_range_db")

	// 仅变更项写入
	if newName != oldName {
		setConfig("name", newName)
	}
	if newPassword != oldPassword {
		setConfig("password", newPassword)
	}
	if newServiceType != oldServiceType {
		setConfig("service_type", newServiceType)
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

// 异步启停，不阻塞页面响应
go func() {
	_ = exec.Command("pkill", "shairport-sync").Run()
}()
go func() {
	cmd := exec.Command("shairport-sync", "-d")
	_ = cmd.Start()
}()

	// 页面延时刷新
w.Write([]byte(`
<script>
setTimeout(()=>window.location.reload(), 3000);
</script>
`))

// 写入配置：自动取消 // 注释
func setConfig(key, val string) {
	data, _ := os.ReadFile(configFile)
	lines := strings.Split(string(data), "\n")
	prefix := key + " = "

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		originalLine := line

		cleanLine := trimmed
		isDoubleSlashComment := false
		if strings.HasPrefix(trimmed, "//") {
			isDoubleSlashComment = true
			cleanLine = strings.TrimSpace(trimmed[2:])
		} else if strings.HasPrefix(trimmed, "#") || strings.HasPrefix(trimmed, ";") {
			cleanLine = strings.TrimSpace(trimmed[1:])
		}

		if strings.HasPrefix(cleanLine, prefix) {
			indent := ""
			if len(originalLine) > len(trimmed) {
				indent = originalLine[:len(originalLine)-len(trimmed)]
			}

			// 原先是//注释行，强制替换缩进为1个Tab，保证对齐
			if isDoubleSlashComment {
				indent = "	"
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
			// volume_range_db 不加引号
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
