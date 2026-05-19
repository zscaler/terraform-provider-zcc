// Package acctest provides acceptance-test helpers for the ZCC Terraform provider.
package acctest

import (
	"os"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"

	"github.com/zscaler/terraform-provider-zcc/internal/client"
	framework "github.com/zscaler/terraform-provider-zcc/internal/framework"
	"github.com/zscaler/terraform-provider-zcc/version"
)

// ProtoV6ProviderFactories wires the Framework provider for terraform-plugin-testing.
var ProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"zcc": providerserver.NewProtocol6WithError(framework.New("test")),
}

// PreCheck skips or fails fast when acceptance tests cannot authenticate.
func PreCheck(t *testing.T) {
	t.Helper()
	if os.Getenv("TF_ACC") == "" {
		t.Fatal("TF_ACC must be set for acceptance tests")
	}

	legacy := strings.EqualFold(os.Getenv("ZSCALER_USE_LEGACY_CLIENT"), "true")
	if legacy {
		if os.Getenv("ZCC_CLIENT_ID") == "" || os.Getenv("ZCC_CLIENT_SECRET") == "" {
			t.Fatal("ZCC_CLIENT_ID and ZCC_CLIENT_SECRET must be set when ZSCALER_USE_LEGACY_CLIENT is true")
		}
		return
	}

	if os.Getenv("ZSCALER_CLIENT_ID") == "" || os.Getenv("ZSCALER_VANITY_DOMAIN") == "" {
		t.Fatal("ZSCALER_CLIENT_ID and ZSCALER_VANITY_DOMAIN must be set for OneAPI acceptance tests")
	}
	if os.Getenv("ZSCALER_CLIENT_SECRET") == "" && os.Getenv("ZSCALER_PRIVATE_KEY") == "" {
		t.Fatal("ZSCALER_CLIENT_SECRET or ZSCALER_PRIVATE_KEY must be set for OneAPI acceptance tests")
	}
}

// MustClient builds a ZCC API client using the same environment variables as the provider.
func MustClient(t *testing.T) *client.Client {
	t.Helper()
	cfg := ClientConfigFromEnv()
	c, err := client.NewClient(cfg)
	if err != nil {
		t.Fatalf("failed to create ZCC client: %v", err)
	}
	return c
}

// ClientConfigFromEnv mirrors provider configuration from environment variables.
func ClientConfigFromEnv() *client.Config {
	return testClientConfig()
}

func testClientConfig() *client.Config {
	return &client.Config{
		ClientID:         os.Getenv("ZSCALER_CLIENT_ID"),
		ClientSecret:     os.Getenv("ZSCALER_CLIENT_SECRET"),
		PrivateKey:       os.Getenv("ZSCALER_PRIVATE_KEY"),
		VanityDomain:     os.Getenv("ZSCALER_VANITY_DOMAIN"),
		Cloud:            os.Getenv("ZSCALER_CLOUD"),
		ZCCClientID:      os.Getenv("ZCC_CLIENT_ID"),
		ZCCClientSecret:  os.Getenv("ZCC_CLIENT_SECRET"),
		ZCCCloud:         os.Getenv("ZCC_CLOUD"),
		UseLegacyClient:  strings.EqualFold(os.Getenv("ZSCALER_USE_LEGACY_CLIENT"), "true"),
		HTTPProxy:        os.Getenv("ZSCALER_HTTP_PROXY"),
		TerraformVersion: os.Getenv("TF_ACC_TERRAFORM_VERSION"),
		ProviderVersion:  version.ProviderVersion,
	}
}
