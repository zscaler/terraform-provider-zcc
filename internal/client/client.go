// Package client wires the Terraform provider to the Zscaler SDK
// (zscaler-sdk-go/v3) OneAPI client. The package intentionally only
// supports the V3 OneAPI flow — the legacy ZCC V2 client
// (ZCC_CLIENT_ID / ZCC_CLIENT_SECRET / ZCC_CLOUD) has been removed in
// favour of the Zidentity-backed OAuth2 / private-key flow that every
// other Zscaler Terraform provider already uses (see
// terraform-provider-zia/zia/config.go::zscalerSDKV3Client).
package client

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"runtime"
	"strconv"
	"time"

	"github.com/zscaler/zscaler-sdk-go/v3/zscaler"

	"github.com/zscaler/terraform-provider-zcc/version"
)

// Config holds the OneAPI credentials and tuning knobs the provider
// surfaces in its schema (and the matching ZSCALER_* environment
// variables). The struct mirrors terraform-provider-zia/zia.Config so
// that a developer switching between the two providers sees the same
// shape of configuration.
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
	TerraformVersion string
	ProviderVersion  string
}

// Client wraps the Zscaler SDK V3 service handle. Resources and data
// sources receive a pointer to this struct via the provider's
// Configure callback.
type Client struct {
	Service *zscaler.Service
}

// NewClient builds the OneAPI V3 client from the provider Config.
// Legacy ZCC V2 authentication is no longer supported — set up the
// OneAPI credentials (client_id + client_secret OR private_key, plus
// vanity_domain) via provider attributes or the ZSCALER_* environment
// variables.
func NewClient(config *Config) (*Client, error) {
	v3Client, err := zscalerSDKV3Client(config)
	if err != nil {
		return nil, err
	}
	return &Client{
		Service: zscaler.NewService(v3Client, nil),
	}, nil
}

// zscalerSDKV3Client builds the Zscaler SDK V3 OneAPI client. The
// function mirrors terraform-provider-zia/zia/config.go's function of
// the same name so the two providers stay in lockstep on
// authentication, rate-limiting, proxy handling and debug logging.
//
// Cache is intentionally disabled here even though ZIA enables it:
// the ZCC SDK does not invalidate the cache across endpoints (e.g. a
// PUT /edit does not invalidate the cached GET /listByCompany
// response), so caching would routinely return stale state on the
// next plan.
func zscalerSDKV3Client(c *Config) (*zscaler.Client, error) {
	applyDefaults(c)

	customUserAgent := generateUserAgent(c.TerraformVersion)

	setters := []zscaler.ConfigSetter{
		zscaler.WithCache(false),
		zscaler.WithHttpClientPtr(http.DefaultClient),
		zscaler.WithRateLimitMaxRetries(int32(c.RetryCount)),
		zscaler.WithRequestTimeout(time.Duration(c.RequestTimeout) * time.Second),
		zscaler.WithRateLimitMinWait(time.Duration(c.MinWait) * time.Second),
		zscaler.WithRateLimitMaxWait(time.Duration(c.MaxWait) * time.Second),
		zscaler.WithUserAgentExtra(customUserAgent),
	}

	tfLog := os.Getenv("TF_LOG")
	if tfLog == "DEBUG" || tfLog == "TRACE" {
		setters = append(setters, zscaler.WithDebug(true))
		log.Println("[DEBUG] SDK debug logging enabled")
	}

	if c.HTTPProxy != "" {
		_url, err := url.Parse(c.HTTPProxy)
		if err != nil {
			return nil, fmt.Errorf("invalid proxy URL: %v", err)
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
	case c.ClientID != "" && c.ClientSecret != "" && c.VanityDomain != "":
		setters = append(setters,
			zscaler.WithClientID(c.ClientID),
			zscaler.WithClientSecret(c.ClientSecret),
			zscaler.WithVanityDomain(c.VanityDomain),
		)
		if c.Cloud != "" {
			setters = append(setters, zscaler.WithZscalerCloud(c.Cloud))
		}

	case c.ClientID != "" && c.PrivateKey != "" && c.VanityDomain != "":
		setters = append(setters,
			zscaler.WithClientID(c.ClientID),
			zscaler.WithPrivateKey(c.PrivateKey),
			zscaler.WithVanityDomain(c.VanityDomain),
		)
		if c.Cloud != "" {
			setters = append(setters, zscaler.WithZscalerCloud(c.Cloud))
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

	log.Println("[INFO] Successfully initialized ZCC OneAPI (V3) client")
	return v3Client.Client, nil
}

// applyDefaults backfills tuning knobs the provider schema leaves
// unset. Defaults intentionally match the SDK's own defaults
// (retry=100, min wait=2s, max wait=10s, request timeout=240s).
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

// generateUserAgent constructs the user agent string with all required details
func generateUserAgent(terraformVersion string) string {
	// Fetch the provider version dynamically from version.ProviderVersion
	providerVersion := version.ProviderVersion

	return fmt.Sprintf("(%s %s) Terraform/%s Provider/%s",
		runtime.GOOS,
		runtime.GOARCH,
		terraformVersion,
		providerVersion,
	)
}
