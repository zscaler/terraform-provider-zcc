package tnbackend

import (
	"context"
	"errors"
	"strings"

	"github.com/zscaler/zscaler-sdk-go/v3/zscaler"
	tnv1 "github.com/zscaler/zscaler-sdk-go/v3/zscaler/zcc/services/trusted_network"
	"github.com/zscaler/zscaler-sdk-go/v3/zscaler/zcc/services/trusted_network_v2"
)

// v1Backend serves tenants that only expose the legacy
// /zcc/papi/public/v1/webTrustedNetwork endpoints. The v1 API has no
// GET-by-id — reads are paginated scans of /listByCompany — and its
// create/edit responses carry only {"success","errorCode"}, so the SDK
// resolves mutations by re-listing (create additionally retries the
// name lookup to absorb the API's eventual consistency). That makes
// network names effectively required and unique per tenant on v1.
type v1Backend struct {
	svc *zscaler.Service
}

const v1PageSize = 1000

func (b *v1Backend) Version() string { return VersionV1 }

// SupportsGetByID is false: reads are list scans, so the pre-delete GET
// of the Get-then-Delete pattern is skipped (list-based-GET exception).
func (b *v1Backend) SupportsGetByID() bool { return false }

func (b *v1Backend) Get(ctx context.Context, id string) (*trusted_network_v2.TrustedNetworkV2, error) {
	return b.find(ctx, func(n *tnv1.TrustedNetwork) bool { return n.ID == id })
}

// GetByName scans the listing and resolves through the same pickByName
// policy as the v2 backend: exact case-insensitive match first, then a
// single unambiguous partial match, and a loud failure when several
// networks match.
func (b *v1Backend) GetByName(ctx context.Context, name string) (*trusted_network_v2.TrustedNetworkV2, error) {
	var candidates []*trusted_network_v2.TrustedNetworkV2
	page := 1
	pageSize := v1PageSize
	for {
		res, _, err := tnv1.GetMultipleTrustedNetworks(ctx, b.svc, "", "", &page, &pageSize)
		if err != nil {
			return nil, err
		}
		for i := range res.TrustedNetworkContracts {
			n := &res.TrustedNetworkContracts[i]
			if !strings.EqualFold(n.NetworkName, name) && !containsFold(n.NetworkName, name) {
				continue
			}
			canonical, err := v1ToCanonical(n)
			if err != nil {
				return nil, err
			}
			candidates = append(candidates, canonical)
		}
		if len(res.TrustedNetworkContracts) < pageSize {
			break
		}
		page++
	}
	return pickByName(name, candidates)
}

// find pages through /v1/webTrustedNetwork/listByCompany until match
// hits or the listing is exhausted, returning ErrNotFound in the latter
// case so callers get the same not-found signal as a structured 404.
func (b *v1Backend) find(ctx context.Context, match func(*tnv1.TrustedNetwork) bool) (*trusted_network_v2.TrustedNetworkV2, error) {
	page := 1
	pageSize := v1PageSize
	for {
		res, _, err := tnv1.GetMultipleTrustedNetworks(ctx, b.svc, "", "", &page, &pageSize)
		if err != nil {
			return nil, err
		}
		for i := range res.TrustedNetworkContracts {
			if match(&res.TrustedNetworkContracts[i]) {
				return v1ToCanonical(&res.TrustedNetworkContracts[i])
			}
		}
		if len(res.TrustedNetworkContracts) < pageSize {
			return nil, ErrNotFound
		}
		page++
	}
}

func (b *v1Backend) Create(ctx context.Context, net *trusted_network_v2.TrustedNetworkV2) (*trusted_network_v2.TrustedNetworkV2, error) {
	payload, err := canonicalToV1(net)
	if err != nil {
		return nil, err
	}
	if payload.NetworkName == "" {
		// The v1 create response has no id; the SDK resolves the new
		// record by name, so an unnamed network could never be found
		// again. Fail before touching the API.
		return nil, errors.New("`name` must be set: this tenant is served by the v1 trusted-network API, whose create call is resolved by network name")
	}
	created, _, err := tnv1.CreateTrustedNetwork(ctx, b.svc, payload)
	if err != nil {
		return nil, err
	}
	return v1ToCanonical(created)
}

func (b *v1Backend) Update(ctx context.Context, id string, net *trusted_network_v2.TrustedNetworkV2) (*trusted_network_v2.TrustedNetworkV2, error) {
	payload, err := canonicalToV1(net)
	if err != nil {
		return nil, err
	}
	payload.ID = id
	updated, _, err := tnv1.UpdateTrustedNetwork(ctx, b.svc, payload)
	if err != nil {
		return nil, err
	}
	return v1ToCanonical(updated)
}

func (b *v1Backend) Delete(ctx context.Context, id string) error {
	_, err := tnv1.DeleteTrustedNetwork(ctx, b.svc, id)
	return err
}
