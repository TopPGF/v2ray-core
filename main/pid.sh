#!/bin/sh
# 启动脚本
#
#by pgf@fealive.com
#s1641823648_337f924daf?timestamp=1610013078&secret=a26c21b27d4e25dc825716fd9728fe47
#ps -ef | grep "name" | grep -v grep | awk '{print $2}'
pidInfo=`ps  -f -C ffmpeg-5.1 | awk '{print $2,$11}'`
echo $pidInfo
