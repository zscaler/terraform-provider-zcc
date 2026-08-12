package tnbackend

import (
	"fmt"
	"strconv"
	"strings"

	tnv1 "github.com/zscaler/zscaler-sdk-go/v3/zscaler/zcc/services/trusted_network"
	"github.com/zscaler/zscaler-sdk-go/v3/zscaler/zcc/services/trusted_network_v2"
)

// v1 numeric conditionType values. The v1 API encodes the match policy
// as an int where v2 uses "ALL"/"ANY".
//
// The mapping below (1=ALL, 2=ANY) is derived from correlating
// operator-created records across a v1-only tenant (conditionType: 1)
// and a v2-enabled tenant where the same-named networks report "ALL",
// and matches the {1,2} values used by the SDK's v1 unit fixtures. It
// has not yet been confirmed against a tenant that serves both endpoint
// generations — if it ever proves wrong, this constant pair is the only
// place to fix. HCL authors can always bypass the mapping entirely by
// setting condition_type to the raw numeric string ("1", "2"), which is
// passed to the v1 API verbatim.
const (
	v1ConditionTypeAll = 1
	v1ConditionTypeAny = 2
)

// conditionTypeToV1 maps the canonical condition_type string to the v1
// numeric encoding. Numeric strings are passed through verbatim as an
// escape hatch.
func conditionTypeToV1(s string) (int, error) {
	trimmed := strings.TrimSpace(s)
	switch strings.ToUpper(trimmed) {
	case "ALL":
		return v1ConditionTypeAll, nil
	case "ANY":
		return v1ConditionTypeAny, nil
	}
	if n, err := strconv.Atoi(trimmed); err == nil {
		return n, nil
	}
	return 0, fmt.Errorf("invalid condition_type %q: must be ALL, ANY, or a numeric value", s)
}

// conditionTypeFromV1 maps the v1 numeric encoding back to the canonical
// string. Unrecognised values round-trip as their decimal string, which
// the schema already accepts as input.
func conditionTypeFromV1(n int) string {
	switch n {
	case v1ConditionTypeAll:
		return "ALL"
	case v1ConditionTypeAny:
		return "ANY"
	}
	return strconv.Itoa(n)
}

// splitCSV materialises a v1 comma-separated criteria string into the
// canonical list form. Whitespace around items is trimmed (the v1 API
// returns values like "10.0.0.5, 10.0.0.6") and empty segments are
// dropped, so "" becomes an empty — never nil — slice, matching the
// helpers.StringListValue semantics the flatten path relies on.
func splitCSV(s string) []string {
	out := []string{}
	for _, part := range strings.Split(s, ",") {
		if part = strings.TrimSpace(part); part != "" {
			out = append(out, part)
		}
	}
	return out
}

// joinCSV renders the canonical list form as the v1 comma-separated
// string. Items are trimmed and empty items dropped so the two
// representations round-trip cleanly.
func joinCSV(items []string) string {
	return strings.Join(splitCSV(strings.Join(items, ",")), ",")
}

// v1ToCanonical converts a v1 wire record into the canonical v2 struct.
// Fields v1 does not carry (zpaId) stay zero-valued.
func v1ToCanonical(in *tnv1.TrustedNetwork) (*trusted_network_v2.TrustedNetworkV2, error) {
	id := 0
	if in.ID != "" {
		parsed, err := strconv.Atoi(in.ID)
		if err != nil {
			return nil, fmt.Errorf("v1 trusted network id %q is not numeric: %w", in.ID, err)
		}
		id = parsed
	}
	// companyId is informational; tolerate the empty string some v1
	// responses carry.
	companyID, _ := strconv.Atoi(in.CompanyID)

	return &trusted_network_v2.TrustedNetworkV2{
		ID:                     id,
		CompanyID:              companyID,
		Active:                 in.Active,
		ConditionType:          conditionTypeFromV1(in.ConditionType),
		Name:                   in.NetworkName,
		NetworkName:            in.NetworkName,
		CreatedBy:              in.CreatedBy,
		EditedBy:               in.EditedBy,
		Guid:                   in.Guid,
		Hostname:               in.Hostnames,
		SSID:                   in.Ssids,
		DNSSearchDomains:       splitCSV(in.DnsSearchDomains),
		DNSServerIPs:           splitCSV(in.DnsServers),
		ResolvedIPsForHostname: splitCSV(in.ResolvedIpsForHostname),
		TrustedDhcpServersIPs:  splitCSV(in.TrustedDhcpServers),
		TrustedEgressIPs:       splitCSV(in.TrustedEgressIps),
		TrustedGatewayIPs:      splitCSV(in.TrustedGateways),
		TrustedSubnetIPs:       splitCSV(in.TrustedSubnets),
	}, nil
}

// canonicalToV1 converts the canonical v2 struct into the v1 wire form.
// The v2 model carries both name and networkName; v1 only has
// networkName, so name wins when both are set.
func canonicalToV1(in *trusted_network_v2.TrustedNetworkV2) (*tnv1.TrustedNetwork, error) {
	conditionType, err := conditionTypeToV1(in.ConditionType)
	if err != nil {
		return nil, err
	}

	networkName := in.Name
	if networkName == "" {
		networkName = in.NetworkName
	}

	out := &tnv1.TrustedNetwork{
		Active:                 in.Active,
		ConditionType:          conditionType,
		NetworkName:            networkName,
		Hostnames:              in.Hostname,
		Ssids:                  in.SSID,
		DnsSearchDomains:       joinCSV(in.DNSSearchDomains),
		DnsServers:             joinCSV(in.DNSServerIPs),
		ResolvedIpsForHostname: joinCSV(in.ResolvedIPsForHostname),
		TrustedDhcpServers:     joinCSV(in.TrustedDhcpServersIPs),
		TrustedEgressIps:       joinCSV(in.TrustedEgressIPs),
		TrustedGateways:        joinCSV(in.TrustedGatewayIPs),
		TrustedSubnets:         joinCSV(in.TrustedSubnetIPs),
	}
	if in.ID != 0 {
		out.ID = strconv.Itoa(in.ID)
	}
	return out, nil
}
