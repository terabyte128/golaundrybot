package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

var r *gin.Engine

var stateColors = map[int]string{
	STATE_READY:              "success",
	STATE_CLAIMED:            "warning",
	STATE_RUNNING:            "primary",
	STATE_WAITING_COLLECTION: "danger",
}

func index(c *gin.Context) {
	c.HTML(http.StatusOK, "index.tmpl", gin.H{
		"stateColors": stateColors,
		"machines":    machines,
	})
}

func init() {
	r = gin.Default()
	r.LoadHTMLGlob("templates/*")
	r.GET("/", index)
}

func serveHttp() {
	r.Run()
}
