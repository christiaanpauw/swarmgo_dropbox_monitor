package main

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/dropbox"
)

func main() {
	router := gin.Default()

	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{
			"title": "Dropbox Monitor",
		})
	})

	router.GET("/list-folders", func(c *gin.Context) {
		folders := dropbox.GetFolders() // Assume this function returns a slice of folder names
		c.HTML(http.StatusOK, "folders.html", gin.H{
			"title":   "Dropbox Folders",
			"folders": folders,
		})
	})

	router.GET("/last-changed", func(c *gin.Context) {
		folders := dropbox.GetLastChangedFolders() // Assume this function returns a slice of folder names and dates
		c.HTML(http.StatusOK, "last_changed.html", gin.H{
			"title":   "Last Changed Dates",
			"folders": folders,
		})
	})

	router.GET("/changes-last-24-hours", func(c *gin.Context) {
		folders := dropbox.GetChangesLast24Hours() // Assume this function returns a slice of folder names and dates
		c.HTML(http.StatusOK, "changes_24_hours.html", gin.H{
			"title":   "Changes in Last 24 Hours",
			"folders": folders,
		})
	})

	router.LoadHTMLGlob("templates/*")
	router.Run(":8080")
}