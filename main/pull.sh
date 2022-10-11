#!/bin/sh
# 启动脚本
#
#by pgf@fealive.com
#s1641823648_337f924daf?timestamp=1610013078&secret=a26c21b27d4e25dc825716fd9728fe47

nohup ./pushM7s.sh $1 $2 $3> /dev/null 2>./log/pull.error & echo $! >> ./log/pull.pid