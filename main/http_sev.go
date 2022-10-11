package main

//go:generate go run github.com/v2fly/v2ray-core/v4/common/errors/errorgen

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/v2fly/v2ray-core/v4/common/buf"
	_ "github.com/v2fly/v2ray-core/v4/main/distro/all"
)

var IDName map[string]string

func httpSev() {
	IDName = make(map[string]string)
	_ = ReadName("/var/www/html/xiaodaji/ID.conf")
	_ = ReadName("/home/wwwroot/xiaodaji/ID.conf")
	http.HandleFunc("/list", routeList)
	http.HandleFunc("/pull", routePull)
	http.HandleFunc("/isPull", routeisPull)
	http.HandleFunc("/pid", routePid)
	http.HandleFunc("/killpid", routekillpid)
	http.HandleFunc("/sign", routeRtmpSignCheck)
	fmt.Println("Listen:8887")
	http.ListenAndServe(":8887", nil)

}

func routeRtmpSignCheck(w http.ResponseWriter, r *http.Request) {
	if buf.RtmpSignCheck {
		buf.RtmpSignCheck = false
		fmt.Fprintf(w, "false")
	} else {
		buf.RtmpSignCheck = true
		fmt.Fprintf(w, "true")
	}

}

func routeisPull(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Access-Control-Allow-Origin", "*")
	id := r.FormValue("id")
	fmt.Println(id)
	if id == "" {
		fmt.Fprintf(w, "id为空")
		fmt.Println("id为空")
		return
	}

	//fmt.Println("body--------" + body)
	if strings.Contains(m7sSummary(), id) {
		fmt.Fprintf(w, "true")
	} else {
		fmt.Fprintf(w, "false")
	}
	return
}

func m7sSummary() string {
	request, err := http.NewRequest("GET", "http://127.0.0.1:8881/api/summary", nil)
	if err != nil {
		fmt.Println(err.Error())
		return ""
	}
	http_client := &http.Client{}
	response, err := http_client.Do(request)
	if err != nil {
		fmt.Println(err.Error())
		return ""
	}
	busSize := 1024 * 5
	buf := make([]byte, busSize) // any non zero value will do, try '1'.
	body := ""
	for {
		n, err := response.Body.Read(buf)
		if n == 0 && err != nil { // simplified
			break
		}
		if n < busSize {
			response.Body.Close()
		}
		//fmt.Println("buf--------" + string(buf[:n]))
		body = body + string(buf[:n])
	}
	//fmt.Println("body--------" + body)
	return body
}

func m7sRecord(streamPath string) error {
	resp, err := http.Get("http://127.0.0.1:8881/api/record/flv?streamPath=" + streamPath + "&append=false")
	if err != nil {

		// handle error

	}
	defer resp.Body.Close()
	_, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	return nil
}
func m7sStopRecord(streamPath string) error {
	resp, err := http.Get("http://127.0.0.1:8881/api/record/flv/stop?streamPath=" + streamPath)
	if err != nil {

		// handle error

	}
	defer resp.Body.Close()
	_, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	return nil
}
func routePull(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Access-Control-Allow-Origin", "*")
	link := r.FormValue("link")
	fmt.Println(link)
	if link == "" {
		fmt.Fprintf(w, "链接为空")
		fmt.Println("链接为空")
		return
	}
	fileSplit := strings.Split(link, "_")
	if len(fileSplit) != 2 {
		fmt.Fprintf(w, "链接名错误："+link)
		fmt.Println("链接名错误")
		return
	}
	rtmp := r.FormValue("r")
	if rtmp != "" {
		rtmp, _ = url.QueryUnescape(rtmp)
	} else {
		text, err := ioutil.ReadFile("./rtmpList/" + link)
		if err != nil {
			fmt.Fprintf(w, "文件打开错误"+err.Error())
			return
		}
		rtmp = string(text)
	}
	t := r.FormValue("t")
	if t == "m3u8" {
		cmd := exec.Command("./pullM3u8.sh", rtmp, fileSplit[1], fileSplit[0])
		var out bytes.Buffer
		var stderr bytes.Buffer
		cmd.Stdout = &out
		cmd.Stderr = &stderr
		err := cmd.Run()
		if err != nil {
			fmt.Println(fmt.Sprint(err) + ": " + stderr.String())
			fmt.Fprintf(w, fmt.Sprint(err)+": "+stderr.String())
			return
		}
	} else {
		cmd := exec.Command("./pull.sh", rtmp, fileSplit[1], fileSplit[0])
		var out bytes.Buffer
		var stderr bytes.Buffer
		cmd.Stdout = &out
		cmd.Stderr = &stderr
		err := cmd.Run()
		if err != nil {
			fmt.Println(fmt.Sprint(err) + ": " + stderr.String())
			fmt.Fprintf(w, fmt.Sprint(err)+": "+stderr.String())
			return
		}
		//fmt.Fprintf(w, "<script>window.location.href=\"http://www.fealive.cn/jessibuca.php?id="+fileSplit[1]+"\"</script>")
	}
	fmt.Fprintf(w, "<script>window.history.go(-1);location.reload();</script>")
	return
}

type M7sSummary struct {
	Streams []StreamsList
}
type StreamsList struct {
	StreamPath  string
	Subscribers interface{}
}

func routeList(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Access-Control-Allow-Origin", "*")
	_ = ReadName("/var/www/html/xiaodaji/ID.conf")
	_ = ReadName("/home/wwwroot/xiaodaji/ID.conf")
	action := r.FormValue("action")
	if action == "record" {
		m7sRecord(r.FormValue("path"))
	} else if action == "stopRecord" {
		m7sStopRecord(r.FormValue("path"))
	}
	nowTime := time.Now().Unix()
	pwd, _ := os.Getwd() //获取当前目录
	//获取文件或目录相关信息
	fileInfoList, err := ioutil.ReadDir(pwd + "/rtmpList/")
	if err != nil {
		log.Fatal(err)
	}
	html := ""
	list := []map[string]int64{}
	idInfo := map[string]int64{}
	for _, v := range fileInfoList {
		fileSplit := strings.Split(v.Name(), "_")
		if len(fileSplit) != 2 {
			err = os.Remove(pwd + "/rtmpList/" + v.Name())
			continue
		}
		timestamp, err := strconv.ParseInt(fileSplit[0], 10, 64)
		if err != nil {
			err = os.Remove(pwd + "/rtmpList/" + v.Name())
			continue
		}
		if nowTime-timestamp > 60*5 {
			err = os.Remove(pwd + "/rtmpList/" + v.Name())
			continue
		}
		if _, ok := idInfo[fileSplit[1]]; ok {
			if idInfo[fileSplit[1]] < timestamp {
				idInfo[fileSplit[1]] = timestamp
			} else {
				err = os.Remove(pwd + "/rtmpList/" + v.Name())
				continue
			}
		} else {
			idInfo[fileSplit[1]] = timestamp
		}

		list = append(list, map[string]int64{fileSplit[1]: idInfo[fileSplit[1]]})
	}

	var summary M7sSummary
	jsonStr := strings.Replace(m7sSummary(), "data: ", "", 1)

	err = json.Unmarshal([]byte(jsonStr), &summary)
	if err != nil {
		fmt.Printf("json解析错误 %v\n", err)
	}
	m7sList := make(map[string]int)
	for _, s := range summary.Streams {
		m7sList[s.StreamPath] = 0
		if subscribers, ok := s.Subscribers.([]interface{}); ok {
			for _, subscriber := range subscribers {
				if sub, ok := subscriber.(map[string]interface{}); ok {
					if subscriberType, ok := sub["Type"]; ok {
						if subscriberType == "FlvRecord" {
							m7sList[s.StreamPath] = 1
						}
					}
				}
			}

		}
	}
	//fmt.Println(list)
	for _, v := range list {
		for id, t := range v {
			timeLayout := "2006-01-02 15:04:05" //转化所需模板
			datetime := time.Unix(t, 0).Format(timeLayout)
			name := id
			if _, ok := IDName[name]; ok {
				name = IDName[name]
			}
			buf, err := ioutil.ReadFile("./rtmpList/" + strconv.FormatInt(t, 10) + "_" + id)
			if err != nil {
				html += "<div><span>" + name + " " + datetime + " </span>&nbsp;&nbsp;"
				html += "打开错误," + err.Error() + "</div><br/>"
				continue
			}
			rtmp := string(buf)
			rtmp = url.QueryEscape(rtmp)
			if isRecord, ok := m7sList["xdj/"+id]; !ok {
				if (time.Now().Unix() - t) < 300 {
					html += "<div><span>" + name + " " + datetime + " </span>&nbsp;&nbsp;"
					html += "<a href='http://www.fealive.cn/v2ray/pull?link=" + strconv.FormatInt(t, 10) + "_" + id + "&t=m3u8&r=" + rtmp + "'>pull m3u8</a>&nbsp;&nbsp;"
					html += "<a href='http://www.fealive.cn/v2ray/pull?link=" + strconv.FormatInt(t, 10) + "_" + id + "&r=" + rtmp + "'>pull</a>&nbsp;&nbsp;"
					html += "</div><br/>"
				}
			} else {
				html += "<div><span>" + name + " " + datetime + " </span>&nbsp;&nbsp;"
				html += "<a href='http://www.fealive.cn/v2ray/pull?link=" + strconv.FormatInt(t, 10) + "_" + id + "&t=m3u8'>pull m3u8</a>&nbsp;&nbsp;"
				if isRecord == 1 {
					//http://192.168.1.105:8881/api/record/flv/stop?streamPath=papa/danmu
					html += "<a href='http://www.fealive.cn/v2ray/list?action=stopRecord&path=xdj/" + id + "'>暂停录制</a>&nbsp;&nbsp;"
				} else {
					//http://192.168.1.105:8881/api/record/flv?streamPath=papa/danmu&append=false
					html += "<a href='http://www.fealive.cn/v2ray/list?action=record&path=xdj/" + id + "'>录制</a>&nbsp;&nbsp;"
				}
				html += "<a href='http://www.fealive.cn/jessibuca.php?id=" + id + "'>播放</a>&nbsp;&nbsp;"
				html += "</div><br/>"
			}
		}
	}
	_, _ = w.Write([]byte(html))
	return
}

func ReadName(filePth string) map[string]string {
	f, err := os.Open(filePth)
	if err != nil {
		return nil
	}
	text, err := ioutil.ReadAll(f)
	if err != nil {
		return nil
	}
	list := strings.Split(string(text), "\n")
	for _, v := range list {
		idname := strings.Split(v, ":")
		if len(idname) == 2 {
			IDName[idname[0]] = idname[1]
		}
	}
	return nil
}
func routekillpid(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Access-Control-Allow-Origin", "*")
	pid := r.FormValue("pid")
	fmt.Println(pid)
	if pid == "" {
		fmt.Fprintf(w, "pid为空")
		fmt.Println("pid为空")
		return
	}
	cmd := exec.Command("./pid.sh")
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		fmt.Println(fmt.Sprint(err) + ": " + stderr.String())
		html := "pid错误:" + stderr.String()
		fmt.Fprintf(w, html)
		return
	}
	if !strings.Contains(out.String(), pid+" rtmp://") {
		fmt.Println(out.String(), "-----------", pid+" rtmp://")
		fmt.Fprintf(w, "pid不存在")
		return
	}
	//ps -aux | grep -Eo ffmpeg
	cmd = exec.Command("kill", pid)

	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		fmt.Println(fmt.Sprint(err) + ": " + stderr.String())
		html := "kill错误:" + stderr.String()
		fmt.Fprintf(w, html)
		return
	}
	_, _ = w.Write([]byte(out.String()))
	return
}
func routePid(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Access-Control-Allow-Origin", "*")
	//
	cmd := exec.Command("./pid.sh")
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		fmt.Println(fmt.Sprint(err) + ": " + stderr.String())
		html := "pid错误:" + stderr.String()
		fmt.Fprintf(w, html)
		return
	}
	_, _ = w.Write([]byte(out.String()))
	return
}
