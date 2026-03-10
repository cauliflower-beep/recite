/**
 * @Author: LiuShuXin
 * @Description: main file
 * @File:  main
 * Software: Goland
 * @Date: 2026/3/9 17:46
 */

package main

import (
	"encoding/json"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

// ================== 第 2 步：数据模型与读取逻辑 ==================

// Poem 定义古诗的数据结构（字段首字母大写，后面跟上 json 标签，方便解析）
type Poem struct {
	Title   string `json:"title"`
	Author  string `json:"author"`
	Rarity  string `json:"rarity"`
	Content string `json:"content"`
}

// loadJSON 封装一个通用的读取文件并解析JSON的函数
// 每次调用它，都会去读最新的文件，这就是我们说的“热更新”魔法！
func loadJSON(filepath string, dest any) error {
	bytes, err := os.ReadFile(filepath)
	if err != nil {
		return err
	}
	return json.Unmarshal(bytes, dest)
}

// ================== 第 3 步：API 接口搭建 ==================

func main() {
	// 1. 初始化 Gin 引擎
	r := gin.Default()

	// 【老G贴心小插件】：跨域中间件
	// 为什么加这个？因为后面开发前端时，Vue跑在 5173 端口，Go跑在 8080 端口。
	// 浏览器会有跨域限制（CORS），加上这段代码，咱们本地联调就畅通无阻了！
	r.Use(Cors())

	// 2. 注册 /api 路由组
	api := r.Group("/api")
	{
		// 接口 1：获取学生名单 (GET /api/students)
		api.GET("/students", func(c *gin.Context) {
			var students []string
			// 每次请求都去读一次文件
			if err := loadJSON("data/students.json", &students); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "读取学生名单失败: " + err.Error()})
				return
			}
			// 成功返回
			c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "success", "data": students})
		})

		// 接口 2：获取古诗数据 (GET /api/poems)
		api.GET("/poems", func(c *gin.Context) {
			var poems []Poem
			// 每次请求都去读一次文件
			if err := loadJSON("data/poems.json", &poems); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "读取古诗数据失败: " + err.Error()})
				return
			}
			// 成功返回
			c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "success", "data": poems})
		})
	}

	// 将打包好的前端文件（dist）设为静态目录
	r.Static("/assets", "./dist/assets")   // 处理 JS/CSS 等资源
	r.StaticFile("/", "./dist/index.html") // 处理首页

	// 2. 核心：处理 SPA 路由（前端路由刷新 404 问题）
	// 如果浏览器访问了一个不存在的路由，统统返回 index.html
	// 让 Vue Router 接管剩下的逻辑
	r.NoRoute(func(c *gin.Context) {
		//path := c.Request.URL.Path
		// 如果请求的不是 API，则直接返回 index.html
		c.File("./dist/index.html")
	})

	r.Run(":8080") // 监听 8080 端口
}

// Cors 一个极简的跨域中间件
func Cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*") // 允许所有源
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		// 放行 OPTIONS 预检请求
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}
