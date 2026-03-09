/**
 * @Author: LiuShuXin
 * @Description:
 * @File:  main
 * Software: Goland
 * @Date: 2026/2/10 16:03
 */

package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	// 1. 创建 Gin 实例
	// set release mode for production
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	// 2. 加载 templates 目录下的所有 HTML 文件
	// 这样 Gin 就会认识这些文件，后续可以直接渲染
	r.LoadHTMLGlob("templates/*")

	// --- 路由配置 ---

	// 首页：做一个简单的导航列表，方便点击
	r.GET("/", func(c *gin.Context) {
		htmlContent := `
		<html>
		<head>
			<meta charset="utf-8">
			<title>诗词抽查 - 预览版</title>
			<style>
				body { font-family: sans-serif; padding: 50px; text-align: center; background: #f0f2f5; }
				h1 { color: #333; }
				.link-box { margin-top: 30px; }
				a { 
					display: inline-block; 
					margin: 10px; 
					padding: 15px 30px; 
					background: #007bff; 
					color: white; 
					text-decoration: none; 
					border-radius: 8px; 
					font-size: 18px;
					transition: 0.3s;
				}
				a:hover { background: #0056b3; transform: scale(1.05); }
			</style>
		</head>
		<body>
			<h1>🎒 诗词抽背 - 预览版</h1>
			<p>请点击下方按钮预览不同风格的版本：</p>
			<div class="link-box">
				<a href="/v1" target="_blank">版本一：v1.html</a>
				<a href="/v2" target="_blank">版本二：v2.html</a>
			</div>
		</body>
		</html>
		`
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(htmlContent))
	})

	// 版本 1 的路由
	r.GET("/v1", func(c *gin.Context) {
		// 这里的 "v1.html" 必须和你 templates 目录下的文件名一致
		c.HTML(http.StatusOK, "v1.html", nil)
	})

	// 版本 2 的路由
	r.GET("/v2", func(c *gin.Context) {
		c.HTML(http.StatusOK, "v2.html", nil)
	})

	// 3. 启动服务，监听 8080 端口
	// 如果你的阿里云有其他服务占用了 8080，可以改成 8090 或其他
	r.Run(":8888")
}
