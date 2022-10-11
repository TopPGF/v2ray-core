//go:build !confonly
// +build !confonly

package buf

//go:generate go run github.com/v2fly/v2ray-core/v4/common/errors/errorgen

import (
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/v2fly/v2ray-core/v4/common/errors"
)

var RtmpID = map[string]int64{}
var RtmpSignCheck = true
var RtmpHost = ""

// UpdateActivity is a CopyOption to update activity on each data copy operation.
func RmptCutHandler() CopyOption {
	return func(handler *copyHandler) {
		handler.onData = append(handler.onData, func(buffer MultiBuffer) {
			if err := rmptCut(buffer); err != nil {

			}
		})
	}
}
func copyRtmpInternal(reader Reader, writer Writer, handler *copyHandler) error {
	for {
		buffer, err := reader.ReadMultiBuffer()
		if !buffer.IsEmpty() {
			for _, handler := range handler.onData {
				handler(buffer)
			}
			if err = rmptCut(buffer); err != nil {
				return readError{err}
			}
			if werr := writer.WriteMultiBuffer(buffer); werr != nil {
				return writeError{werr}
			}
		}

		if err != nil {
			return readError{err}
		}
	}
}

// Copy dumps all payload from reader to writer or stops when an error occurs. It returns nil when EOF.
func CopyRtmp(reader Reader, writer Writer, options ...CopyOption) error {
	var handler copyHandler
	for _, option := range options {
		option(&handler)
	}
	err := copyRtmpInternal(reader, writer, &handler)
	if err != nil && errors.Cause(err) != io.EOF {
		return err
	}
	return nil
}

func rmptCut(buffer MultiBuffer) error {
	// for _, b := range buffer {
	// 	b.Bytes()
	// }
	//return nil
	//fmt.Println(RtmpID)
	str := buffer.String()
	//fmt.Println("rmptCut-------------" + TrimHiddenCharacter(str))
	isHas := false
	//RtmpHost := "rtmp://pull.cscynet.com/live"
	//str := buffer.String()
	//fmt.Println("rmptCut-------------" + str)
	//-----FCSubscribe�connect?�applivetcUrlrtmp://pull.cscynet.com/livefpad
	// if strings.Contains(str, "token") {
	// 	fmt.Println("----token:" + TrimHiddenCharacter(str))
	// }

	if strings.Contains(str, "rtmp://") {
		//isHas = true
		str = strings.Split(TrimHiddenCharacter(str), "rtmp://")[1]
		RtmpHost = "rtmp://" + strings.Split(str, "fpad")[0]
		fmt.Println("----RtmpHost:" + RtmpHost)
	}
	if strings.Contains(str, "Subscribe") {
		//isHas = true
		str = TrimHiddenCharacter(str)
		fmt.Println("-----Subscribe----------:" + str)

		//类别 1:小妲己 2:小猫
		rtmpType := 0
		if strings.Contains(str, "_checkbw") {
			rtmpInfo := strings.Split(str, "@")
			if len(rtmpInfo) < 3 {
				return nil
			}
			str = rtmpInfo[2]
			rtmpType = 1
		} else {
			rtmpInfo := strings.Split(str, "@")
			str = rtmpInfo[len(rtmpInfo)-1]
			rtmpType = 2
		}
		if str == "" {
			fmt.Println("-----------Subscribe解析为空:")
			return nil
		}
		timestamp := int64(0)
		liveID := ""
		rtmpUrl := ""
		switch rtmpType {
		case 0:
			rtmpUrl = RtmpHost + "/" + str
			fmt.Println("-----------rtmpUrl:" + rtmpUrl)
			return nil
		case 1:
			rtmpUrl = RtmpHost + "/" + strings.Split(str, "_checkbw")[0]
			fmt.Println("-----------rtmpUrl:" + rtmpUrl)
			reg := regexp.MustCompile(`(s[0-9]+)_.*([^0-9a-zA-Z]16[0-9]{8})`)
			submatch := reg.FindAllSubmatch([]byte(rtmpUrl), -1)
			if len(submatch) < 1 || len(submatch[0]) < 3 {
				reg = regexp.MustCompile(`(s[0-9]+)_.*([^0-9a-zA-Z]6[a-fA-F0-9]{7})`)
				submatch = reg.FindAllSubmatch([]byte(rtmpUrl), -1)
				if len(submatch) < 1 || len(submatch[0]) < 3 {
					fmt.Println("-----------rtmpUrl正则错误:" + rtmpUrl)
					return nil
				}
			}
			timestampStr := string(submatch[0][2])
			//去一个字符
			timestampStr = timestampStr[1:]
			var err error
			if IsNum(timestampStr) {
				timestamp, err = strconv.ParseInt(timestampStr, 10, 64)
				if err != nil {
					fmt.Println("-----------字符串转换成整数失败:" + timestampStr)
					return nil
				}
			} else {
				timestamp, err = Hex2Dec(timestampStr)
				if err != nil {
					fmt.Println("-----------转换成整数时间戳失败:" + timestampStr)
					return nil
				}
			}
			liveID = string(submatch[0][1])
		case 2:
			//去一个字符
			rtmpUrl = RtmpHost + "/" + str[1:]
			//rtmpUrl = "rtmp://pull28iah0p1t5.changjiangjin.com/live/155920736_efd4939b284430afd13d82c1165aeb68?token=577301d6ad18ad897ce2edd6978d3d96&t=1665475642"
			fmt.Println("-----------rtmpUrl:" + rtmpUrl)
			reg := regexp.MustCompile(`([0-9]+)_.*([^0-9a-zA-Z]16[0-9]{8})`)
			submatch := reg.FindAllSubmatch([]byte(rtmpUrl), -1)

			if len(submatch) < 1 || len(submatch[0]) < 3 {
				fmt.Println("-----------rtmpUrl正则错误:" + rtmpUrl)
				return nil
			}
			//xm的时间戳是过期时间戳
			//timestampStr = string(submatch[0][2])
			//去一个字符
			//timestampStr = timestampStr[1:]
			timestamp = time.Now().Unix()
			liveID = "xm" + string(submatch[0][1])
		}

		fmt.Println("-----------liveID:" + liveID)
		if _, ok := RtmpID[liveID]; !ok {
			RtmpID[liveID] = timestamp
			isHas = true
		} else {
			if timestamp-RtmpID[liveID] > 20 {
				RtmpID[liveID] = timestamp
				isHas = true
			}
		}
		//fmt.Println(RtmpID)
		if isHas {
			f, err := os.Create("./rtmpList/" + strconv.FormatInt(timestamp, 10) + "_" + liveID)
			defer f.Close()
			if err != nil {
				return err
			} else {
				if _, err = f.Write([]byte(rtmpUrl)); err != nil {
					return err
				}
			}
		}
		// cmd := exec.Command("./pull.sh", rtmpUrl, liveID, strconv.FormatInt(timestamp, 10))
		// var out bytes.Buffer
		// var stderr bytes.Buffer
		// cmd.Stdout = &out
		// cmd.Stderr = &stderr
		// err := cmd.Run()
		// if err != nil {
		// 	fmt.Println(fmt.Sprint(err) + ": " + stderr.String())
		// }
	}

	if isHas && RtmpSignCheck {
		return errors.New("return")
	}
	return nil
}

//Hex2Dec 16转10进制
func Hex2Dec(val string) (int64, error) {
	n, err := strconv.ParseUint(val, 16, 64)
	return int64(n), err
}

//IsNum 判断是否数值
func IsNum(s string) bool {
	_, err := strconv.ParseFloat(s, 64)
	return err == nil
}

//TrimHiddenCharacter 清除不可见字符
func TrimHiddenCharacter(originStr string) string {
	dstRunes := []byte("")
	for _, c := range []byte(originStr) {
		if c >= 0 && c <= 31 {
			continue
		}
		if c >= 127 {
			continue
		}
		dstRunes = append(dstRunes, c)
	}
	return string(dstRunes)
}
