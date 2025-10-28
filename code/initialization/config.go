package initialization

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/spf13/pflag"

	"github.com/spf13/viper"
)

type Config struct {
	// Indicates whether the configuration has been initialized.
	Initialized                bool
	FeishuBaseUrl              string
	FeishuAppId                string
	FeishuAppSecret            string
	FeishuAppEncryptKey        string
	FeishuAppVerificationToken string
	FeishuBotName              string
	OpenaiApiKeys              []string
	HttpPort                   int
	HttpsPort                  int
	UseHttps                   bool
	CertFile                   string
	KeyFile                    string
	OpenaiApiUrl               string
	OpenaiModel                string
	OpenAIHttpClientTimeOut    int
	OpenaiMaxTokens            int
	HttpProxy                  string
	AzureOn                    bool
	AzureApiVersion            string
	AzureDeploymentName        string
	AzureResourceName          string
	AzureOpenaiToken           string
	StreamMode                 bool
}

var (
	cfg    = pflag.StringP("config", "c", "./config.yaml", "apiserver config file path.")
	config *Config
	once   sync.Once
)

func GetConfig() *Config {
	once.Do(func() {
		config = LoadConfig(*cfg)
		config.Initialized = true
	})

	return config
}

func LoadConfig(cfg string) *Config {
	// Read config file if it exists (optional for Railway)
	viper.SetConfigFile(cfg)
	if err := viper.ReadInConfig(); err != nil {
		// Config file is optional, especially on Railway
		fmt.Printf("Warning: Could not read config file %s: %v (using env vars)\n", cfg, err)
	}
	viper.AutomaticEnv()

	// Railway compatibility: Try HTTP_PORT first, then PORT, then default to 9000
	// Railway automatically sets PORT, but HTTP_PORT can override it
	var httpPort int

	// Try reading from environment variables directly (Railway sets these)
	if portStr := os.Getenv("HTTP_PORT"); portStr != "" {
		if p, err := strconv.Atoi(portStr); err == nil {
			httpPort = p
			fmt.Printf("Using HTTP_PORT from env: %d\n", httpPort)
		}
	}

	// If HTTP_PORT not set, try PORT (Railway default)
	if httpPort == 0 {
		if portStr := os.Getenv("PORT"); portStr != "" {
			if p, err := strconv.Atoi(portStr); err == nil {
				httpPort = p
				fmt.Printf("Using PORT from env: %d\n", httpPort)
			} else {
				fmt.Printf("Warning: PORT env var '%s' is not a valid integer, using default\n", portStr)
			}
		}
	}

	// If still not set, try Viper (config file)
	if httpPort == 0 {
		if viper.IsSet("HTTP_PORT") {
			httpPort = getViperIntValue("HTTP_PORT", 0)
		} else if viper.IsSet("PORT") {
			httpPort = getViperIntValue("PORT", 0)
		}
	}

	// Final fallback to default
	if httpPort == 0 {
		httpPort = 9000
		fmt.Printf("Using default port: %d\n", httpPort)
	}

	config := &Config{
		FeishuBaseUrl:              getViperStringValue("BASE_URL", ""),
		FeishuAppId:                getViperStringValue("APP_ID", ""),
		FeishuAppSecret:            getViperStringValue("APP_SECRET", ""),
		FeishuAppEncryptKey:        getViperStringValue("APP_ENCRYPT_KEY", ""),
		FeishuAppVerificationToken: getViperStringValue("APP_VERIFICATION_TOKEN", ""),
		FeishuBotName:              getViperStringValue("BOT_NAME", ""),
		OpenaiApiKeys:              getViperStringArray("OPENAI_KEY", []string{""}),
		OpenaiModel:                getViperStringValue("OPENAI_MODEL", "gpt-3.5-turbo"),
		OpenAIHttpClientTimeOut:    getViperIntValue("OPENAI_HTTP_CLIENT_TIMEOUT", 550),
		OpenaiMaxTokens:            getViperIntValue("OPENAI_MAX_TOKENS", 2000),
		HttpPort:                   httpPort,
		HttpsPort:                  getViperIntValue("HTTPS_PORT", 9001),
		UseHttps:                   getViperBoolValue("USE_HTTPS", false),
		CertFile:                   getViperStringValue("CERT_FILE", "cert.pem"),
		KeyFile:                    getViperStringValue("KEY_FILE", "key.pem"),
		OpenaiApiUrl:               getViperStringValue("API_URL", "https://api.openai.com"),
		HttpProxy:                  getViperStringValue("HTTP_PROXY", ""),
		AzureOn:                    getViperBoolValue("AZURE_ON", false),
		AzureApiVersion:            getViperStringValue("AZURE_API_VERSION", "2023-03-15-preview"),
		AzureDeploymentName:        getViperStringValue("AZURE_DEPLOYMENT_NAME", ""),
		AzureResourceName:          getViperStringValue("AZURE_RESOURCE_NAME", ""),
		AzureOpenaiToken:           getViperStringValue("AZURE_OPENAI_TOKEN", ""),
		StreamMode:                 getViperBoolValue("STREAM_MODE", false),
	}

	return config
}

func getViperStringValue(key string, defaultValue string) string {
	value := viper.GetString(key)
	if value == "" {
		return defaultValue
	}
	return value
}

//OPENAI_KEY: sk-xxx,sk-xxx,sk-xxx
//result:[sk-xxx sk-xxx sk-xxx]
func getViperStringArray(key string, defaultValue []string) []string {
	value := viper.GetString(key)
	if value == "" {
		return defaultValue
	}
	raw := strings.Split(value, ",")
	return filterFormatKey(raw)
}

func getViperIntValue(key string, defaultValue int) int {
	value := viper.GetString(key)
	if value == "" {
		return defaultValue
	}
	intValue, err := strconv.Atoi(value)
	if err != nil {
		fmt.Printf("Invalid value for %s, using default value %d\n", key, defaultValue)
		return defaultValue
	}
	return intValue
}

func getViperBoolValue(key string, defaultValue bool) bool {
	value := viper.GetString(key)
	if value == "" {
		return defaultValue
	}
	boolValue, err := strconv.ParseBool(value)
	if err != nil {
		fmt.Printf("Invalid value for %s, using default value %v\n", key, defaultValue)
		return defaultValue
	}
	return boolValue
}

func (config *Config) GetCertFile() string {
	if config.CertFile == "" {
		return "cert.pem"
	}
	if _, err := os.Stat(config.CertFile); err != nil {
		fmt.Printf("Certificate file %s does not exist, using default file cert.pem\n", config.CertFile)
		return "cert.pem"
	}
	return config.CertFile
}

func (config *Config) GetKeyFile() string {
	if config.KeyFile == "" {
		return "key.pem"
	}
	if _, err := os.Stat(config.KeyFile); err != nil {
		fmt.Printf("Key file %s does not exist, using default file key.pem\n", config.KeyFile)
		return "key.pem"
	}
	return config.KeyFile
}

// Filter keys starting with "sk-"
func filterFormatKey(keys []string) []string {
	var result []string
	for _, key := range keys {
		if strings.HasPrefix(key, "sk-") || strings.HasPrefix(key,
			"fk") || strings.HasPrefix(key, "fastgpt") {
			result = append(result, key)
		}
	}
	return result

}
