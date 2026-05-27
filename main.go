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

func main() {
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/save", saveHandler)
	http.ListenAndServe(":8086", nil)
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
    </style>
</head>
<body>
    <div class="card">
        <h2>Shairport Sync 管理面板</h2>
        <form method="post" action="/save">
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

	setConfig("name", name)
	setConfig("password", password)
	setConfig("output_device", device)
	setConfig("mixer_control_name", mixer)

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
func setConfig(key, val string) {
	data, _ := os.ReadFile(configFile)
	lines := strings.Split(string(data), "\n")
	for i, line := range lines {
		lineTrim := strings.TrimSpace(line)
		if strings.HasPrefix(lineTrim, key+" = ") {
			lines[i] = key + ` = ` + val + `;`
		}
	}
	os.WriteFile(configFile, []byte(strings.Join(lines, "\n")), 0644)
}
