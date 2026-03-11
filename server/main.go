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

// ================== 1. 数据模型定义 ==================

// Book 定义前端新的“教材->篇目”嵌套结构
type Book struct {
	Title string   `json:"title"`
	Poems []string `json:"poems"`
}

// ConfigPayload 用于前后端一次性交互的“大包”结构
type ConfigPayload struct {
	Students []string `json:"students"`
	Arsenal  []Book   `json:"arsenal"`
}

// ================== 2. 文件读写工具函数 ==================

// loadJSON 读取文件并解析到结构体
func loadJSON(filepath string, dest interface{}) error {
	bytes, err := os.ReadFile(filepath)
	if err != nil {
		return err
	}
	return json.Unmarshal(bytes, dest)
}

// saveJSON 将结构体格式化为 JSON 并覆写到本地文件
// 用 MarshalIndent 可以让存下来的 json 文件带有缩进，方便你朋友直接用记事本看
func saveJSON(filepath string, data interface{}) error {
	bytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	// 0644 是文件权限（读写权限）
	return os.WriteFile(filepath, bytes, 0644)
}

// ================== 3. 核心 API 路由 ==================

func main() {
	// 切换到生产模式
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	// 消除警告：告诉 Gin 不存在代理，直接取远程 IP 即可
	r.SetTrustedProxies(nil)

	r.Use(Cors())

	// 限制请求体大小为 1MB
	r.Use(func(c *gin.Context) {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 1024*1024)
		c.Next()
	})

	api := r.Group("/api")
	{
		// [接口 1]：拉取统筹室配置数据
		api.GET("/config", func(c *gin.Context) {
			var students []string
			var arsenal []Book

			// 分别读取两个文件（如果有报错，可能是文件不存在或格式错，先给个空数组）
			if err := loadJSON("data/students.json", &students); err != nil {
				students = []string{}
			}
			// 注意：这里我们把读取古诗的文件也当做 arsenal 的存储目标
			if err := loadJSON("data/poems.json", &arsenal); err != nil {
				arsenal = []Book{}
			}

			// 组装成大包返回给前端
			c.JSON(http.StatusOK, gin.H{
				"code": 200,
				"msg":  "作战数据读取成功！",
				"data": ConfigPayload{
					Students: students,
					Arsenal:  arsenal,
				},
			})
		})

		// [接口 2]：保存统筹室配置数据（覆盖写入）
		api.POST("/config", func(c *gin.Context) {
			var payload ConfigPayload

			// 1. 接收前端传过来的 JSON 数据
			if err := c.ShouldBindJSON(&payload); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "接收数据失败，格式不对啊喂！"})
				return
			}

			// 2. 将数据拆分，分别覆写到本地文件
			err1 := saveJSON("data/students.json", payload.Students)
			err2 := saveJSON("data/poems.json", payload.Arsenal)

			if err1 != nil || err2 != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "覆写本地文件失败，检查下权限！"})
				return
			}

			// 3. 成功响应
			c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "轰隆隆！目标与弹药覆写完毕！"})
		})
	}

	// 静态文件托管与兜底 (保持不变)
	r.Static("/assets", "./dist/assets")
	r.StaticFile("/", "./dist/index.html")
	r.NoRoute(func(c *gin.Context) {
		c.File("./dist/index.html")
	})

	r.Run(":9527")
}

// Cors 跨域中间件 (保持不变)
func Cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, DELETE")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}
