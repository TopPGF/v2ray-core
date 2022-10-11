// +build !confonly

package inbound

//go:generate go run github.com/v2fly/v2ray-core/v4/common/errors/errorgen

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/v2fly/v2ray-core/v4/common/buf"
	"github.com/v2fly/v2ray-core/v4/common/errors"
)

var RtmpID = map[string]int64{}

func rmptCut(buffer buf.MultiBuffer) error {
	// for _, b := range buffer {
	// 	b.Bytes()
	// }

	//fmt.Println(RtmpID)
	str := buffer.String()
	//fmt.Println("rmptCut-------------" + TrimHiddenCharacter(str))
	isHas := false
	host := "rtmp://pull.cscynet.com/live"
	//str := buffer.String()
	//fmt.Println("rmptCut-------------" + str)
	//-----FCSubscribe�connect?�applivetcUrlrtmp://pull.cscynet.com/livefpad
	if strings.Contains(str, "applivetcUrlrtmp://") {
		//isHas = true
		fmt.Println("-----applivetcUrlrtmp" + TrimHiddenCharacter(str))
		str = strings.Split(TrimHiddenCharacter(str), "applivetcUrl")[1]
		host = strings.Split(str, "fpad")[0]
	}
	if strings.Contains(str, "createStream") {
		//isHas = true
		fmt.Println("-----createStream" + TrimHiddenCharacter(str))
		str = strings.Split(TrimHiddenCharacter(str), "@")[2]
		rtmpUrl := host + "/" + strings.Split(str, "_checkbw")[0]
		fmt.Println("-----------rtmpUrl:" + rtmpUrl)
		reg := regexp.MustCompile(`(s[0-9]+)_.*([^0-9a-zA-Z]16[0-9]{8})`)
		submatch := reg.FindAllSubmatch([]byte(rtmpUrl), -1)
		if len(submatch) < 1 || len(submatch[0]) < 3 {
			reg = regexp.MustCompile(`(s[0-9]+)_.*([^0-9a-zA-Z]6[a-fA-F0-9]{7})`)
			submatch = reg.FindAllSubmatch([]byte(rtmpUrl), -1)
			if len(submatch) < 1 || len(submatch[0]) < 3 {
				fmt.Println("-----------rtmpUrl正则错误:" + rtmpUrl)
				return errors.New("-----------rtmpUrl正则错误:" + rtmpUrl)
			}
		}

		timestampStr := string(submatch[0][2])
		timestampStr = timestampStr[1:]
		timestamp := int64(0)
		var err error
		if IsNum(timestampStr) {
			timestamp, err = strconv.ParseInt(timestampStr, 10, 64)
			if err != nil {
				fmt.Println("-----------字符串转换成整数失败:" + timestampStr)
				return errors.New("-----------字符串转换成整数失败:" + timestampStr)
			}
		} else {
			timestamp, err = Hex2Dec(timestampStr)
			if err != nil {
				fmt.Println("-----------转换成整数时间戳失败:" + timestampStr)
				return errors.New("-----------转换成整数时间戳失败:" + timestampStr)
			}
		}
		liveID := string(submatch[0][1])
		fmt.Println("-----------liveID:" + liveID)
		if _, ok := RtmpID[liveID]; !ok {
			RtmpID[liveID] = timestamp
			isHas = true
		} else {
			if timestamp-RtmpID[liveID] > 30 {
				RtmpID[liveID] = timestamp
				isHas = true
			}
		}
		fmt.Println(RtmpID)
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
			return errors.New("return")
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

	if isHas {
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
