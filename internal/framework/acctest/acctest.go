// Package acctest provides acceptance-test helpers for the ZCC Terraform provider.
package acctest

import (
	"os"
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

// PreCheck skips or fails fast when acceptance tests cannot authenticate
// against OneAPI. The legacy ZCC V2 client has been removed; tests now
// require ZSCALER_CLIENT_ID + ZSCALER_VANITY_DOMAIN + either
// ZSCALER_CLIENT_SECRET or ZSCALER_PRIVATE_KEY.
func PreCheck(t *testing.T) {
	t.Helper()
	if os.Getenv("TF_ACC") == "" {
		t.Fatal("TF_ACC must be set for acceptance tests")
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
		HTTPProxy:        os.Getenv("ZSCALER_HTTP_PROXY"),
		TerraformVersion: os.Getenv("TF_ACC_TERRAFORM_VERSION"),
		ProviderVersion:  version.ProviderVersion,
	}
}
