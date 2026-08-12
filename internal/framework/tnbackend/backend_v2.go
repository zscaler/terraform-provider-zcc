package tnbackend

import (
	"context"
	"fmt"
	"strconv"

	"github.com/zscaler/zscaler-sdk-go/v3/zscaler"
	"github.com/zscaler/zscaler-sdk-go/v3/zscaler/zcc/services/trusted_network_v2"
)

// v2Backend is a thin passthrough to the trusted_network_v2 SDK package;
// the only translation is the string id at the Terraform boundary.
type v2Backend struct {
	svc *zscaler.Service
}

func (b *v2Backend) Version() string       { return VersionV2 }
func (b *v2Backend) SupportsGetByID() bool { return true }

func (b *v2Backend) Get(ctx context.Context, id string) (*trusted_network_v2.TrustedNetworkV2, error) {
	numericID, err := parseID(id)
	if err != nil {
		return nil, err
	}
	return trusted_network_v2.Get(ctx, b.svc, numericID)
}

// GetByName lists with the server-side NAME keyword filter (a substring
// search) and resolves the result through pickByName rather than the
// SDK's GetByName, which only accepts exact full-name matches — an
// unambiguous partial name must resolve too.
func (b *v2Backend) GetByName(ctx context.Context, name string) (*trusted_network_v2.TrustedNetworkV2, error) {
	networks, err := trusted_network_v2.GetAll(ctx, b.svc, &trusted_network_v2.GetAllFilterOptions{
		Keyword: name,
		Type:    trusted_network_v2.FilterTypeName,
	})
	if err != nil {
		return nil, err
	}
	candidates := make([]*trusted_network_v2.TrustedNetworkV2, len(networks))
	for i := range networks {
		candidates[i] = &networks[i]
	}
	return pickByName(name, candidates)
}

func (b *v2Backend) Create(ctx context.Context, net *trusted_network_v2.TrustedNetworkV2) (*trusted_network_v2.TrustedNetworkV2, error) {
	created, _, err := trusted_network_v2.Create(ctx, b.svc, net)
	return created, err
}

func (b *v2Backend) Update(ctx context.Context, id string, net *trusted_network_v2.TrustedNetworkV2) (*trusted_network_v2.TrustedNetworkV2, error) {
	numericID, err := parseID(id)
	if err != nil {
		return nil, err
	}
	net.ID = numericID
	updated, _, err := trusted_network_v2.Update(ctx, b.svc, numericID, net)
	return updated, err
}

func (b *v2Backend) Delete(ctx context.Context, id string) error {
	numericID, err := parseID(id)
	if err != nil {
		return err
	}
	_, err = trusted_network_v2.Delete(ctx, b.svc, numericID)
	return err
}

func parseID(id string) (int, error) {
	numericID, err := strconv.Atoi(id)
	if err != nil {
		return 0, fmt.Errorf("trusted network id %q is not a valid integer: %w", id, err)
	}
	return numericID, nil
}
