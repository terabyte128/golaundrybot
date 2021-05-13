package main

import (
	"errors"
	"log"
	"net/http"
	"os"

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
		"roommates":   roommates,
	})
}

func decodeParams(c *gin.Context) (machine *LaundryMachine, roommate *Roommate) {
	roommateName := c.PostForm("roommate")

	if roommateName == "" {
		c.AbortWithError(http.StatusBadRequest, errors.New("roommate is required"))
		return
	}

	var selectedRoommate *Roommate

	for _, roommate := range roommates {
		if roommate.Name == roommateName {
			selectedRoommate = roommate
		}
	}

	if selectedRoommate == nil {
		c.AbortWithError(http.StatusNotFound, errors.New("roommate does not exist"))
	}

	machineName := c.PostForm("machine")

	if machineName == "" {
		c.AbortWithError(http.StatusBadRequest, errors.New("machine is required"))
		return
	}

	var selectedMachine *LaundryMachine
	var ok bool

	if selectedMachine, ok = machines[machineName]; !ok {
		c.AbortWithError(http.StatusNotFound, errors.New("machine does not exist"))
	}

	return selectedMachine, selectedRoommate
}

func claim(c *gin.Context) {
	machine, roommate := decodeParams(c)
	machine.Claim(roommate)
	c.Redirect(http.StatusSeeOther, "/")
}

func collect(c *gin.Context) {
	machineName := c.PostForm("machine")

	if machine, ok := machines[machineName]; ok {
		machine.MarkCollected()
		c.Redirect(http.StatusSeeOther, "/")
		return
	}

	c.AbortWithError(http.StatusBadRequest, errors.New("machine not found"))
}

func cancel(c *gin.Context) {
	machineName := c.PostForm("machine")

	if machine, ok := machines[machineName]; ok {
		machine.Unclaim()
		c.Redirect(http.StatusSeeOther, "/")
		return
	}

	c.AbortWithError(http.StatusBadRequest, errors.New("machine not found"))
}

func init() {
	r = gin.Default()
	r.LoadHTMLGlob("templates/*")
	r.GET("/", index)
	r.POST("/claim", claim)
	r.POST("/collect", collect)
	r.POST("/cancel", cancel)
}

func serveHttp() {
	addr, ok := os.LookupEnv("HTTP_ADDR")
	if !ok {
		log.Fatal("env variable HTTP_ADDR not found")
	}

	r.Run(addr)
}
