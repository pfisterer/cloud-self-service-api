package config

import (
	"fmt"

	"github.com/farberg/cloud-self-service-api/internal/helper"
	"github.com/farberg/cloud-self-service-api/internal/storage"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
)

type AppData struct {
	Config  AppConfig
	Storage *storage.Storage
	Logger  *zap.Logger
	Log     *zap.SugaredLogger
}

type StorageConfig struct {
	// The type of database to use
	DbType string `json:"db_type" validate:"oneof=sqlite postgres mysql"`
	// The connection string for the database (using GORM format)
	DbConnectionString string `json:"db_connection_string" validate:"required"`
	// Flag to indicate if dummy data should be added (for development/testing)
	AddDummyData bool `json:"add_dummy_data"`
}

type WebServerConfig struct {
	// The OIDC issuer URL for authentication
	OIDCIssuerURL string `json:"oidc_issuer_url" validate:"required_if=AuthProvider oidc,url"`
	// The OIDC client ID for authentication
	OIDCClientID string `json:"oidc_client_id" validate:"required_if=AuthProvider oidc"`
	// The bind string for the Gin web server (e.g., ":8082")
	GinBindString string `json:"gin_bind_string" validate:"required"`
	// The base URL for the web server (e.g., "http://localhost:8083")
	WebserverBaseUrl string `json:"webserver_base_url" validate:"required,url"`
	// The TTL (in hours) for API tokens
	ApiTokenTTLHours int `json:"api_token_ttl_hours"`
}

type AppConfig struct {
	Storage         StorageConfig   `json:"storage_config"`
	WebServer       WebServerConfig `json:"webserver_config"`
	DnsPolicyConfig DnsPolicyConfig `json:"dns_policy_config"`
	// Flag indicating if the application is running in development mode
	DevMode bool `json:"dev_mode"`
}

type DnsPolicyConfig struct {
	SuperAdminEmails map[string]struct{} `json:"super_admin_emails"`
	WebhookApiKey    string              `json:"webhook_api_key"`
}

func GetAppConfigFromEnvironment() (AppConfig, error) {

	appConfig := AppConfig{
		DnsPolicyConfig: DnsPolicyConfig{
			SuperAdminEmails: helper.GetEnvStringSet("DNS_POLICY_SUPERADMIN_EMAILS", map[string]struct{}{}, ",", true),
			WebhookApiKey:    helper.GetEnvString("DNS_POLICY_WEBHOOK_API_KEY", ""),
		},
		Storage: StorageConfig{
			DbType:             helper.GetEnvString("DB_TYPE", "sqlite"),
			DbConnectionString: helper.GetEnvString("DB_CONNECTION_STRING", "file::memory:?cache=shared"),
			AddDummyData:       helper.GetEnvBool("DEV_STORAGE_ADD_DUMMY_DATA", false),
		},

		WebServer: WebServerConfig{
			GinBindString:    helper.GetEnvString("API_BIND", ":8083"),
			WebserverBaseUrl: helper.GetEnvString("API_BASE_URL", "http://localhost:8083"),
			OIDCIssuerURL:    helper.GetEnvString("OIDC_ISSUER_URL", ""),
			OIDCClientID:     helper.GetEnvString("OIDC_CLIENT_ID", ""),
			ApiTokenTTLHours: helper.GetEnvInt("API_TOKEN_TTL_HOURS", 24*365),
		},
		DevMode: helper.GetEnvString("API_MODE", "production") == "development",
	}

	err := appConfig.Validate()
	return appConfig, err
}

func (config *AppConfig) Validate() error {
	validate := validator.New(validator.WithRequiredStructEnabled())

	if err := validate.Struct(config); err != nil {

		// This part converts the generic error into a list of specific errors
		if validationErrors, ok := err.(validator.ValidationErrors); ok {

			// Format the errors for better readability
			return fmt.Errorf("configuration validation failed: %s", formatValidationErrors(validationErrors))
		}
		return err // Return other types of errors if any
	}
	return nil
}

// Helper to format validation errors
func formatValidationErrors(errs validator.ValidationErrors) string {
	var errorMessages string
	for _, e := range errs {
		errorMessages += fmt.Sprintf(
			"\n - Field '%s' failed on the '%s' tag (Value: '%v')",
			e.Field(),
			e.Tag(),
			e.Value(),
		)
	}
	return errorMessages
}
