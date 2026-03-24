查看端口服务：

`sudo ss -tulpn | grep :9527`

干掉服务：

` sudo kill -9 141`

授权：

`chmod +x recite`

启动服务器：

` nohup ./recite > output.log 2>&1 &`