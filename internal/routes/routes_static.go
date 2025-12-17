package routes

import (
	"io/fs"
	"net/http"

	"github.com/farberg/cloud-self-service-api/internal/config"
	"github.com/farberg/cloud-self-service-api/internal/generated_docs"
	"github.com/farberg/cloud-self-service-api/internal/helper"
	"github.com/gin-gonic/gin"
)

func CreateStaticFiles(group *gin.RouterGroup, app *config.AppData) *gin.RouterGroup {

	// Serve index.html
	group.GET("/", func(c *gin.Context) {
		c.Writer.Header().Set("Content-Type", "text/html; charset=utf-8")
		c.String(http.StatusOK, helper.IndexHtml)
	})

	// Swagger JSON endpoint
	group.GET("/swagger.json", func(c *gin.Context) {
		c.Writer.Header().Set("Content-Type", "application/json; charset=utf-8")
		c.String(http.StatusOK, generated_docs.SwaggerJSON)
	})

	// Serve JS client
	subFS, _ := fs.Sub(generated_docs.ClientDist, "client-dist")
	group.StaticFS("/client", http.FS(subFS))
	return group
}
