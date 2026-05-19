package main

import (
	"context"
	"flag"
	"log"

	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	framework "github.com/zscaler/terraform-provider-zcc/internal/framework"
	"github.com/zscaler/terraform-provider-zcc/version"
)

func main() {
	var debug bool

	flag.BoolVar(&debug, "debug", false, "Start provider in debug mode.")
	flag.Parse()

	logFlags := log.Flags()
	logFlags = logFlags &^ (log.Ldate | log.Ltime)
	log.SetFlags(logFlags)

	opts := providerserver.ServeOpts{
		Address: "registry.terraform.io/zscaler/zcc",
		Debug:   debug,
	}

	err := providerserver.Serve(context.Background(), func() provider.Provider {
		return framework.New(version.ProviderVersion)
	}, opts)
	if err != nil {
		log.Fatal(err.Error())
	}
}
