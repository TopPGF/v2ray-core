#!/bin/bash
# 启动脚本
#
#by pgf@fealive.com
#s1641823648_337f924daf?timestamp=1610013078&secret=a26c21b27d4e25dc825716fd9728fe47
#s1132831914_64a77b1a52?txSecret=aedd89b79f61cc6e73b3e9e890feb565&txTime=6035ce21
#rtmp://pull.reralv.com/live/s1296643175_8aedbefa3a?txSecret=1d555651ad2ddcff263d9e047b84fdc4&txTime=603f46a8
if [ ! -d "rtmpList" ]; then
 echo "文件夹rtmpList不存在,创建"
 mkdir rtmpList
fi
echo ${1} > ./rtmpList/${3}_${2}

#isPull=`curl http://192.168.1.105:8887/isPull?id=xdj/${2}`

#if [ $isPull == "false" ];then
    #/home/pgf/app/ffmpeg-4.3.1 -re -i ${1} -acodec copy -vcodec copy -f rtsp rtsp://127.0.0.1:554/xdj/${2}
#fi


