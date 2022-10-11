#!/bin/sh
# 启动脚本
#
#by pgf@fealive.com
#s1641823648_337f924daf?timestamp=1610013078&secret=a26c21b27d4e25dc825716fd9728fe47

LIVE=$1
echo "<div>Url:${LIVE}</div>"
echo "<div>ID:$2</div>"
echo "<div>time:$3</div>"
echo "Url:${LIVE}" >> ./log/pull.log
echo "ID:$2" >> ./log/pull.log
echo "time:$3" >> ./log/pull.log
cd /home/wwwroot/temp/share/xiaodaji/
nohup ./cut.sh $1 $2 $3 >> ./log/pull.log 2>./log/pull.error & echo $! >> ./log/pull.pid
