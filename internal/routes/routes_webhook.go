package routes

import (
	"errors"
	"net/http"
	"strings"

	"github.com/farberg/cloud-self-service-api/internal/auth"
	"github.com/farberg/cloud-self-service-api/internal/config"
	"github.com/farberg/cloud-self-service-api/internal/helper"
	"github.com/gin-gonic/gin"
)

func CreateWebhookApiGroup(group *gin.RouterGroup, app *config.AppData) *gin.RouterGroup {
	group.POST("/dns-policy", webhookFunc(app))

	return group
}

func verifyApiKey(c *gin.Context, apiKey string) error {
	// Get the Authorization header
	authHeader := c.GetHeader("Authorization")

	// Remove the bearer prefix from the Authorization header (if present)
	const bearerPrefix = "Bearer "
	tokenString, ok := strings.CutPrefix(authHeader, bearerPrefix)
	if !ok {
		return errors.New("missing or invalid Authorization Bearer header")
	}

	// Check if token is an API key (starts with your prefix)
	if tokenString != apiKey {
		return errors.New("invalid API key provided in Authorization header")
	}

	return nil
}

func webhookFunc(app *config.AppData) gin.HandlerFunc {
	return func(c *gin.Context) {
		app.Log.Debug("Received webhook DNS policy request")

		// Get the Authorization header
		err := verifyApiKey(c, app.Config.DnsPolicyConfig.WebhookApiKey)
		if err != nil {
			app.Log.Warnf("Webhook API key verification failed: %v", err)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}

		// Extract JSON body and bind to UserClaimsRequest struct
		var userClaimsReq auth.UserClaims
		if err := c.ShouldBindJSON(&userClaimsReq); err != nil {
			app.Log.Warnf("Failed to bind JSON body: %v", err)
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
			return
		}
		app.Log.Debugf("Received user claims: %+v", userClaimsReq)

		// Get user rules
		rules, err := listUserRules(app, &userClaimsReq, false /* is_super_admin */)
		if err != nil {
			// Return error response
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve rules"})
			return
		}

		// Prepare data for pattern replacement
		userDnsLabel := helper.DnsMakeCompliant(userClaimsReq.Email)

		// Iterate over the rules create responses
		zones := make([]ZoneResponse, 0)
		for _, rule := range rules {
			zone := strings.ReplaceAll(rule.ZonePattern, "%u", userDnsLabel)
			zones = append(zones, ZoneResponse{
				Zone:    zone,
				ZoneSOA: rule.ZoneSoa,
			})
		}

		// Return the zones as JSON response
		c.JSON(http.StatusOK, zones)
	}
}
