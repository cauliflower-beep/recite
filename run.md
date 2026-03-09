```shell
# 后台运行并将日志输出到poem.log（方便排查问题） 
nohup ./poem > poem.log 2>&1 &

# 查看服务是否运行
# 方法1：查看8888端口是否被监听（最直接）
ss -tulpn | grep 8888
# 正常输出示例：LISTEN 0      128        0.0.0.0:8888    0.0.0.0:*    users:(("poem",pid=12345,fd=3))

# 方法2：直接查找poem进程
ps -ef | grep poem | grep -v grep
# 正常输出示例：root     12345     1  0 10:00 ?        00:00:01 ./poem

# 关闭服务
# 方法1：通过端口号精准关闭（推荐）
# 第一步：查找8888端口对应的poem进程ID（PID）
PID=$(lsof -i:8888 -t)
# 第二步：杀死该进程
if [ -n "$PID" ]; then
  kill -9 $PID
  echo "poem服务已关闭，PID: $PID"
else
  echo "未找到监听8888端口的poem服务"
fi

# 方法2：直接通过进程名关闭
pkill -f "./poem"
# 或更精准（避免误杀含poem字符的其他进程）
ps -ef | grep "./poem" | grep -v grep | awk '{print $2}' | xargs kill -9 2>/dev/null

# 确认服务已关闭
# 方法1：检查8888端口是否还有监听
ss -tulpn | grep 8888
# 无任何输出 = 端口已释放，服务已关闭

# 方法2：检查poem进程是否存在
ps -ef | grep poem | grep -v grep
# 无./poem相关进程输出 = 服务已关闭
```



