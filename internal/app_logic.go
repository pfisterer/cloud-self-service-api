package app

import (
	"github.com/farberg/cloud-self-service-api/internal/config"
	"github.com/gin-gonic/gin"
)

const AppLogicKey = "AppLogicKey"

func InjectAppLogic(app *config.AppData) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set(AppLogicKey, app)
	}
}
