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

// Poem 定义新的篇目结构，包含题目和作者
type Poem struct {
	Title  string `json:"title"`
	Author string `json:"author"`
}

// ConfigPayload 用于前后端一次性交互的“大包”结构
type ConfigPayload struct {
	Students []string `json:"students"`
	Poems    []Poem   `json:"poems"`
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
func saveJSON(filepath string, data interface{}) error {
	bytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath, bytes, 0644)
}

// ================== 3. 核心 API 路由 ==================

func main() {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	r.SetTrustedProxies(nil)
	r.Use(Cors())

	r.Use(func(c *gin.Context) {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 1024*1024)
		c.Next()
	})

	api := r.Group("/api")
	{
		// [接口 1]：拉取配置数据
		api.GET("/config", func(c *gin.Context) {
			var students []string
			var poems []Poem

			if err := loadJSON("./data/students.json", &students); err != nil {
				students = []string{}
			}
			if err := loadJSON("./data/poems.json", &poems); err != nil {
				poems = []Poem{}
			}

			c.JSON(http.StatusOK, gin.H{
				"code": 200,
				"msg":  "作战数据读取成功！",
				"data": ConfigPayload{
					Students: students,
					Poems:    poems,
				},
			})
		})

		// [接口 2]：保存配置数据
		api.POST("/config", func(c *gin.Context) {
			var payload ConfigPayload

			if err := c.ShouldBindJSON(&payload); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "接收数据失败，格式不对啊喂！"})
				return
			}

			err1 := saveJSON("./data/students.json", payload.Students)
			err2 := saveJSON("./data/poems.json", payload.Poems)

			if err1 != nil || err2 != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "覆写本地文件失败，检查下权限！"})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"code": 200,
				"msg":  "轰隆隆！目标与弹药覆写完毕！",
			})
		})
	}

	r.Static("/assets", "./dist/assets")
	r.StaticFile("/", "./dist/index.html")
	r.NoRoute(func(c *gin.Context) {
		c.File("./dist/index.html")
	})

	r.Run(":9527")
}

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
