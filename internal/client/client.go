package client

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"runtime"
	"strconv"
	"time"

	"github.com/zscaler/zscaler-sdk-go/v3/zscaler"
	"github.com/zscaler/zscaler-sdk-go/v3/zscaler/zcc"

	"github.com/zscaler/terraform-provider-zcc/version"
)

// Config contains our provider configuration values and Zscaler clients.
type Config struct {
	ClientID         string
	ClientSecret     string
	PrivateKey       string
	VanityDomain     string
	Cloud            string
	HTTPProxy        string
	RetryCount       int
	MinWait          int
	MaxWait          int
	RequestTimeout   int
	UseLegacyClient  bool
	TerraformVersion string
	ProviderVersion  string

	// Legacy SDK specific fields (ZCC env vars: ZCC_CLIENT_ID, ZCC_CLIENT_SECRET, ZCC_CLOUD)
	ZCCClientID     string
	ZCCClientSecret string
	ZCCCloud        string
}

// Client wraps the Zscaler SDK client
type Client struct {
	Service *zscaler.Service
}

// NewClient creates a new ZCC client based on the configuration
func NewClient(config *Config) (*Client, error) {
	if config.UseLegacyClient {
		return newLegacyClient(config)
	}
	return newV3Client(config)
}

// newLegacyClient creates a legacy V2 client
func newLegacyClient(config *Config) (*Client, error) {
	applyDefaults(config)

	customUserAgent := generateUserAgent(config.TerraformVersion, config.ProviderVersion)

	setters := []zcc.ConfigSetter{
		zcc.WithRateLimitMaxRetries(int32(config.RetryCount)),
		zcc.WithRateLimitMinWait(time.Duration(config.MinWait) * time.Second),
		zcc.WithRateLimitMaxWait(time.Duration(config.MaxWait) * time.Second),
		zcc.WithRequestTimeout(time.Duration(config.RequestTimeout) * time.Second),
		zcc.WithUserAgentExtra(customUserAgent),
		zcc.WithZCCClientID(config.ZCCClientID),
		zcc.WithZCCClientSecret(config.ZCCClientSecret),
		zcc.WithZCCCloud(config.ZCCCloud),
	}

	if config.HTTPProxy != "" {
		_url, err := url.Parse(config.HTTPProxy)
		if err != nil {
			return nil, fmt.Errorf("invalid proxy URL: %v", err)
		}
		setters = append(setters, zcc.WithProxyHost(_url.Hostname()))

		sPort := _url.Port()
		if sPort == "" {
			sPort = "80"
		}
		port64, err := strconv.ParseInt(sPort, 10, 32)
		if err != nil {
			return nil, fmt.Errorf("invalid proxy port: %v", err)
		}
		if port64 < 1 || port64 > 65535 {
			return nil, fmt.Errorf("invalid port number: must be between 1 and 65535, got: %d", port64)
		}
		port32 := int32(port64)
		setters = append(setters, zcc.WithProxyPort(port32))
	}

	zccCfg, err := zcc.NewConfiguration(setters...)
	if err != nil {
		return nil, fmt.Errorf("failed to create ZCC configuration: %v", err)
	}
	zccCfg.UserAgent = customUserAgent

	legacyService, err := zscaler.NewLegacyZccClient(zccCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create ZCC client: %v", err)
	}

	return &Client{
		Service: zscaler.NewService(legacyService.Client, nil),
	}, nil
}

// newV3Client creates a V3 client (ZCC uses client_id, client_secret, vanity_domain - no CustomerID required)
func newV3Client(config *Config) (*Client, error) {
	applyDefaults(config)

	customUserAgent := generateUserAgent(config.TerraformVersion, config.ProviderVersion)

	setters := []zscaler.ConfigSetter{
		zscaler.WithRateLimitMaxRetries(int32(config.RetryCount)),
		zscaler.WithRateLimitMinWait(time.Duration(config.MinWait) * time.Second),
		zscaler.WithRateLimitMaxWait(time.Duration(config.MaxWait) * time.Second),
		zscaler.WithRequestTimeout(time.Duration(config.RequestTimeout) * time.Second),
		zscaler.WithUserAgentExtra(""),
	}

	// Cache is disabled for ZCC because the SDK cache invalidation does not
	// cover cross-endpoint patterns (e.g. PUT /edit does not invalidate the
	// cached GET /listByCompany response for the same resource type).
	setters = append(setters, zscaler.WithCache(false))

	tfLog := os.Getenv("TF_LOG")
	if tfLog == "DEBUG" || tfLog == "TRACE" {
		setters = append(setters, zscaler.WithDebug(false))
		log.Println("[DEBUG] SDK debug logging enabled")
	}

	if config.HTTPProxy != "" {
		_url, err := url.Parse(config.HTTPProxy)
		if err != nil {
			return nil, err
		}
		setters = append(setters, zscaler.WithProxyHost(_url.Hostname()))

		sPort := _url.Port()
		if sPort == "" {
			sPort = "80"
		}
		port64, err := strconv.ParseInt(sPort, 10, 32)
		if err != nil {
			return nil, fmt.Errorf("invalid proxy port: %v", err)
		}
		if port64 < 1 || port64 > 65535 {
			return nil, fmt.Errorf("invalid port number: must be between 1 and 65535, got: %d", port64)
		}
		port32 := int32(port64)
		setters = append(setters, zscaler.WithProxyPort(port32))
	}

	switch {
	case config.ClientID != "" && config.ClientSecret != "" && config.VanityDomain != "":
		setters = append(setters,
			zscaler.WithClientID(config.ClientID),
			zscaler.WithClientSecret(config.ClientSecret),
			zscaler.WithVanityDomain(config.VanityDomain),
		)
		if config.Cloud != "" {
			setters = append(setters, zscaler.WithZscalerCloud(config.Cloud))
		}

	case config.ClientID != "" && config.PrivateKey != "" && config.VanityDomain != "":
		setters = append(setters,
			zscaler.WithClientID(config.ClientID),
			zscaler.WithPrivateKey(config.PrivateKey),
			zscaler.WithVanityDomain(config.VanityDomain),
		)
		if config.Cloud != "" {
			setters = append(setters, zscaler.WithZscalerCloud(config.Cloud))
		}

	default:
		return nil, fmt.Errorf("invalid authentication configuration: missing required parameters (client_id, client_secret or private_key, vanity_domain)")
	}

	configSet, err := zscaler.NewConfiguration(setters...)
	if err != nil {
		return nil, fmt.Errorf("failed to create SDK V3 configuration: %v", err)
	}

	configSet.UserAgent = customUserAgent

	v3Client, err := zscaler.NewOneAPIClient(configSet)
	if err != nil {
		return nil, fmt.Errorf("failed to create Zscaler API client: %v", err)
	}

	return &Client{
		Service: zscaler.NewService(v3Client.Client, nil),
	}, nil
}

func applyDefaults(config *Config) {
	if config.RetryCount == 0 {
		config.RetryCount = 100
	}
	if config.MinWait == 0 {
		config.MinWait = 2
	}
	if config.MaxWait == 0 {
		config.MaxWait = 10
	}
	if config.RequestTimeout == 0 {
		config.RequestTimeout = 240
	}
}

// generateUserAgent for ZCC: omit CustomerID (use "zcc" as identifier since ZCC has no customer ID)
func generateUserAgent(terraformVersion, providerVersion string) string {
	if providerVersion == "" {
		providerVersion = version.ProviderVersion
	}
	if providerVersion == "" {
		providerVersion = "dev"
	}
	if terraformVersion == "" {
		terraformVersion = "unknown"
	}
	return fmt.Sprintf("(%s %s) Terraform/%s Provider/%s ZCC/zcc",
		runtime.GOOS,
		runtime.GOARCH,
		terraformVersion,
		providerVersion,
	)
}
