// Package tnbackend abstracts the two generations of the ZCC trusted
// network API behind a single interface so the zcc_trusted_network
// resource and data source keep working on tenants where the newer
// /zcc/papi/public/v2/trusted-networks endpoints are not yet enabled.
//
// The canonical wire model is the v2 struct
// (trusted_network_v2.TrustedNetworkV2) — the Terraform schema keeps its
// v2 shape (lists of strings, ALL/ANY, scalar hostname/ssid) on every
// tenant. The v1 backend converts at the edge:
//
//   - list criteria  <-> comma-separated strings
//   - int id         <-> string id
//   - hostname/ssid  <-> hostnames/ssids (verbatim passthrough)
//   - conditionType  <-> ALL/ANY (see convert.go for the numeric mapping)
//
// Backend selection is fully automatic — nothing to configure: the first
// trusted-network operation probes the v2 list endpoint once and caches
// the verdict for the lifetime of the provider's SDK service handle, so a
// tenant that gains the v2 endpoints simply starts using them on its next
// Terraform run.
package tnbackend

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/zscaler/zscaler-sdk-go/v3/zscaler"
	"github.com/zscaler/zscaler-sdk-go/v3/zscaler/errorx"
	"github.com/zscaler/zscaler-sdk-go/v3/zscaler/zcc/services/common"
	"github.com/zscaler/zscaler-sdk-go/v3/zscaler/zcc/services/trusted_network_v2"
)

// Version identifiers reported by Backend.Version.
const (
	VersionV1 = "v1"
	VersionV2 = "v2"
)

// v2Endpoint mirrors the unexported trustedNetworkEndpointV2 constant in
// the SDK's trusted_network_v2 package; it is only used by the
// availability probe, which needs a single-page list call the SDK does
// not expose.
const v2Endpoint = "/zcc/papi/public/v2/trusted-networks"

// ErrNotFound is returned by the v1 backend when a record is absent from
// the tenant. The v1 API has no GET-by-id endpoint — absence is detected
// by scanning the list — so there is no HTTP 404 to classify. IsNotFound
// folds this sentinel and the structured errorx 404 check into one
// predicate for callers.
var ErrNotFound = errors.New("trusted network not found")

// IsNotFound reports whether err means "the trusted network does not
// exist upstream", regardless of which backend produced it. It wraps the
// mandatory errorx.ErrorResponse + IsObjectNotFound() classification (see
// CLAUDE.md, "404 handling") and adds the v1 list-scan sentinel.
func IsNotFound(err error) bool {
	if errors.Is(err, ErrNotFound) {
		return true
	}
	var respErr *errorx.ErrorResponse
	return errors.As(err, &respErr) && respErr.IsObjectNotFound()
}

// Backend is the version-neutral trusted network API surface. All
// methods speak the canonical v2 struct; ids are strings, matching the
// Terraform boundary.
type Backend interface {
	// Version returns VersionV1 or VersionV2.
	Version() string
	// SupportsGetByID is false when reads are list scans (v1). Callers
	// use it to skip the pre-delete GET, per the list-based-GET
	// exception to the Get-then-Delete pattern.
	SupportsGetByID() bool
	Get(ctx context.Context, id string) (*trusted_network_v2.TrustedNetworkV2, error)
	GetByName(ctx context.Context, name string) (*trusted_network_v2.TrustedNetworkV2, error)
	Create(ctx context.Context, net *trusted_network_v2.TrustedNetworkV2) (*trusted_network_v2.TrustedNetworkV2, error)
	Update(ctx context.Context, id string, net *trusted_network_v2.TrustedNetworkV2) (*trusted_network_v2.TrustedNetworkV2, error)
	Delete(ctx context.Context, id string) error
}

// pickByName selects a record from name-lookup candidates:
//
//  1. an exact match wins (case-insensitive), matching the historical
//     GetByName semantics;
//  2. otherwise a single candidate whose name contains the lookup string
//     is accepted — so `terraform import ... TrustedNetwork02` resolves
//     "BD-TrustedNetwork02" when it is the only network matching;
//  3. several partial matches fail loudly with the candidate names —
//     never silently pick one;
//  4. no match at all returns ErrNotFound.
func pickByName(name string, candidates []*trusted_network_v2.TrustedNetworkV2) (*trusted_network_v2.TrustedNetworkV2, error) {
	var partial []*trusted_network_v2.TrustedNetworkV2
	for _, c := range candidates {
		if strings.EqualFold(c.Name, name) {
			return c, nil
		}
		if containsFold(c.Name, name) {
			partial = append(partial, c)
		}
	}
	switch len(partial) {
	case 0:
		return nil, fmt.Errorf("no trusted network found with name %q: %w", name, ErrNotFound)
	case 1:
		return partial[0], nil
	}
	names := make([]string, len(partial))
	for i, c := range partial {
		names[i] = c.Name
	}
	return nil, fmt.Errorf("multiple trusted networks match name %q (%s): use the exact name or the numeric id", name, strings.Join(names, ", "))
}

func containsFold(haystack, needle string) bool {
	return strings.Contains(strings.ToLower(haystack), strings.ToLower(needle))
}

// autoCache memoises the auto-detection verdict per SDK service handle.
// The provider builds one *zscaler.Service per Configure, so entries are
// bounded by the number of provider configurations in the process.
var autoCache sync.Map // *zscaler.Service -> Backend

// For returns the backend serving this tenant, probing the v2 list
// endpoint on first use and caching the verdict — the caller never
// chooses a version.
func For(ctx context.Context, svc *zscaler.Service) (Backend, error) {
	if cached, ok := autoCache.Load(svc); ok {
		return cached.(Backend), nil
	}

	available, err := probeV2(ctx, svc)
	if err != nil {
		return nil, fmt.Errorf("unable to detect the trusted-network API version for this tenant: %w", err)
	}

	var backend Backend
	if available {
		backend = &v2Backend{svc: svc}
	} else {
		backend = &v1Backend{svc: svc}
	}
	tflog.Info(ctx, "Auto-detected ZCC trusted-network API version", map[string]any{
		"version": backend.Version(),
	})

	actual, _ := autoCache.LoadOrStore(svc, backend)
	return actual.(Backend), nil
}

// probeV2 issues a single-page GET of the v2 list endpoint. A 200 —
// even with zero items — proves the endpoint exists. The list endpoint
// is used deliberately: a 404 from GET-by-id legitimately means "record
// deleted" (the drift signal), while a 4xx on the list can only mean the
// tenant does not serve v2.
func probeV2(ctx context.Context, svc *zscaler.Service) (bool, error) {
	_, err := common.ReadPageV2[trusted_network_v2.TrustedNetworkV2](ctx, svc.Client, v2Endpoint, common.QueryParamsV2{PerPage: 1})
	if err == nil {
		return true, nil
	}

	var respErr *errorx.ErrorResponse
	if errors.As(err, &respErr) && v2EndpointUnavailable(respErr) {
		return false, nil
	}
	// Transport failures, 5xx, auth errors on the token flow, … — do not
	// silently pick a version off an ambiguous signal.
	return false, err
}

// v2EndpointUnavailable classifies a failed probe response. Endpoint-level
// 4xx / 501 statuses mean the tenant does not (yet) serve v2; anything
// else is treated as a genuine error by the caller.
func v2EndpointUnavailable(respErr *errorx.ErrorResponse) bool {
	if respErr.IsObjectNotFound() {
		return true
	}
	// IsObjectNotFound requires a non-nil Response; also honour the
	// parsed body's structured id when only that survived.
	if respErr.Parsed != nil && respErr.Parsed.ID == "resource.not.found" {
		return true
	}
	status := 0
	if respErr.Parsed != nil {
		status = respErr.Parsed.Status
	}
	if status == 0 && respErr.Response != nil {
		status = respErr.Response.StatusCode
	}
	switch status {
	case http.StatusBadRequest, http.StatusForbidden, http.StatusNotFound,
		http.StatusMethodNotAllowed, http.StatusNotImplemented:
		return true
	}
	return false
}
