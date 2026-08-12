package tnbackend

import (
	"context"
	"net/http"
	"reflect"
	"testing"

	"github.com/zscaler/zscaler-sdk-go/v3/zscaler/errorx"
	tnv1 "github.com/zscaler/zscaler-sdk-go/v3/zscaler/zcc/services/trusted_network"
	"github.com/zscaler/zscaler-sdk-go/v3/zscaler/zcc/services/trusted_network_v2"
)

func TestConditionTypeRoundTrip(t *testing.T) {
	toV1 := []struct {
		in      string
		want    int
		wantErr bool
	}{
		{"ALL", v1ConditionTypeAll, false},
		{"all", v1ConditionTypeAll, false},
		{" Any ", v1ConditionTypeAny, false},
		{"ANY", v1ConditionTypeAny, false},
		// Numeric escape hatch is passed through verbatim.
		{"0", 0, false},
		{"1", 1, false},
		{"2", 2, false},
		{"", 0, true},
		{"SOME", 0, true},
	}
	for _, tc := range toV1 {
		got, err := conditionTypeToV1(tc.in)
		if tc.wantErr {
			if err == nil {
				t.Errorf("conditionTypeToV1(%q): expected error, got %d", tc.in, got)
			}
			continue
		}
		if err != nil {
			t.Errorf("conditionTypeToV1(%q): unexpected error: %v", tc.in, err)
			continue
		}
		if got != tc.want {
			t.Errorf("conditionTypeToV1(%q) = %d, want %d", tc.in, got, tc.want)
		}
	}

	fromV1 := []struct {
		in   int
		want string
	}{
		{v1ConditionTypeAll, "ALL"},
		{v1ConditionTypeAny, "ANY"},
		// Unknown values round-trip as their decimal string.
		{0, "0"},
		{7, "7"},
	}
	for _, tc := range fromV1 {
		if got := conditionTypeFromV1(tc.in); got != tc.want {
			t.Errorf("conditionTypeFromV1(%d) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestSplitJoinCSV(t *testing.T) {
	split := []struct {
		in   string
		want []string
	}{
		{"", []string{}},
		{"   ", []string{}},
		{"8.8.8.8", []string{"8.8.8.8"}},
		// The v1 API returns space-padded lists like "10.0.0.5, 10.0.0.6".
		{"10.0.0.5, 10.0.0.6", []string{"10.0.0.5", "10.0.0.6"}},
		{"a,,b,", []string{"a", "b"}},
	}
	for _, tc := range split {
		got := splitCSV(tc.in)
		if !reflect.DeepEqual(got, tc.want) {
			t.Errorf("splitCSV(%q) = %#v, want %#v", tc.in, got, tc.want)
		}
		if got == nil {
			t.Errorf("splitCSV(%q) returned nil; must always be a non-nil slice", tc.in)
		}
	}

	join := []struct {
		in   []string
		want string
	}{
		{nil, ""},
		{[]string{}, ""},
		{[]string{"8.8.8.8"}, "8.8.8.8"},
		{[]string{" 10.0.0.5 ", "10.0.0.6"}, "10.0.0.5,10.0.0.6"},
		{[]string{"a", "", "b"}, "a,b"},
	}
	for _, tc := range join {
		if got := joinCSV(tc.in); got != tc.want {
			t.Errorf("joinCSV(%#v) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestV1ToCanonical(t *testing.T) {
	in := &tnv1.TrustedNetwork{
		ID:                     "13381",
		CompanyID:              "4543",
		CreatedBy:              "453111",
		EditedBy:               "453111",
		Guid:                   "ca624ef9-703b-45ed-8c6c-6b6217f2b355",
		NetworkName:            "BD-TrustedNetwork01",
		ConditionType:          v1ConditionTypeAll,
		DnsServers:             "8.8.8.8, 4.4.2.2",
		DnsSearchDomains:       "acme.com",
		Hostnames:              "server1.acme.com",
		Ssids:                  "CorpWiFi",
		ResolvedIpsForHostname: "",
		TrustedSubnets:         "10.0.0.20/24",
		TrustedGateways:        "10.0.0.1",
		TrustedDhcpServers:     "",
		TrustedEgressIps:       "10.0.0.20",
		Active:                 true,
	}

	got, err := v1ToCanonical(in)
	if err != nil {
		t.Fatalf("v1ToCanonical: unexpected error: %v", err)
	}

	if got.ID != 13381 {
		t.Errorf("ID = %d, want 13381", got.ID)
	}
	if got.CompanyID != 4543 {
		t.Errorf("CompanyID = %d, want 4543", got.CompanyID)
	}
	if got.Name != "BD-TrustedNetwork01" || got.NetworkName != "BD-TrustedNetwork01" {
		t.Errorf("Name/NetworkName = %q/%q, want both BD-TrustedNetwork01", got.Name, got.NetworkName)
	}
	if got.ConditionType != "ALL" {
		t.Errorf("ConditionType = %q, want ALL", got.ConditionType)
	}
	if got.Hostname != "server1.acme.com" {
		t.Errorf("Hostname = %q, want server1.acme.com", got.Hostname)
	}
	if got.SSID != "CorpWiFi" {
		t.Errorf("SSID = %q, want CorpWiFi", got.SSID)
	}
	if !reflect.DeepEqual(got.DNSServerIPs, []string{"8.8.8.8", "4.4.2.2"}) {
		t.Errorf("DNSServerIPs = %#v", got.DNSServerIPs)
	}
	if !reflect.DeepEqual(got.ResolvedIPsForHostname, []string{}) {
		t.Errorf("ResolvedIPsForHostname = %#v, want empty non-nil slice", got.ResolvedIPsForHostname)
	}
	if got.ZPAID != "" {
		t.Errorf("ZPAID = %q, want empty (v1 does not carry zpaId)", got.ZPAID)
	}
	if !got.Active {
		t.Error("Active = false, want true")
	}

	if _, err := v1ToCanonical(&tnv1.TrustedNetwork{ID: "not-a-number"}); err == nil {
		t.Error("v1ToCanonical with non-numeric id: expected error")
	}
}

func TestCanonicalToV1(t *testing.T) {
	in := &trusted_network_v2.TrustedNetworkV2{
		ID:                    69389,
		Name:                  "BD-TrustedNetwork02",
		ConditionType:         "ALL",
		Active:                true,
		Hostname:              "server1.acme.com",
		SSID:                  "CorpWiFi",
		DNSServerIPs:          []string{"8.8.8.8", "4.4.2.2"},
		DNSSearchDomains:      []string{"acme.com"},
		TrustedSubnetIPs:      []string{"10.0.0.20/24"},
		TrustedGatewayIPs:     []string{"10.0.0.1"},
		TrustedDhcpServersIPs: []string{},
		TrustedEgressIPs:      nil,
	}

	got, err := canonicalToV1(in)
	if err != nil {
		t.Fatalf("canonicalToV1: unexpected error: %v", err)
	}
	if got.ID != "69389" {
		t.Errorf("ID = %q, want 69389", got.ID)
	}
	if got.NetworkName != "BD-TrustedNetwork02" {
		t.Errorf("NetworkName = %q", got.NetworkName)
	}
	if got.ConditionType != v1ConditionTypeAll {
		t.Errorf("ConditionType = %d, want %d", got.ConditionType, v1ConditionTypeAll)
	}
	if got.DnsServers != "8.8.8.8,4.4.2.2" {
		t.Errorf("DnsServers = %q", got.DnsServers)
	}
	if got.TrustedDhcpServers != "" || got.TrustedEgressIps != "" {
		t.Errorf("empty lists must serialize as empty strings, got %q / %q", got.TrustedDhcpServers, got.TrustedEgressIps)
	}
	if got.Hostnames != "server1.acme.com" || got.Ssids != "CorpWiFi" {
		t.Errorf("Hostnames/Ssids = %q/%q", got.Hostnames, got.Ssids)
	}

	// name falls back to networkName when only the latter is set.
	fallback, err := canonicalToV1(&trusted_network_v2.TrustedNetworkV2{NetworkName: "OnlyNetworkName", ConditionType: "ANY"})
	if err != nil {
		t.Fatalf("canonicalToV1 fallback: unexpected error: %v", err)
	}
	if fallback.NetworkName != "OnlyNetworkName" {
		t.Errorf("NetworkName fallback = %q, want OnlyNetworkName", fallback.NetworkName)
	}
	if fallback.ID != "" {
		t.Errorf("zero id must stay empty, got %q", fallback.ID)
	}

	if _, err := canonicalToV1(&trusted_network_v2.TrustedNetworkV2{ConditionType: "NEITHER"}); err == nil {
		t.Error("canonicalToV1 with invalid condition type: expected error")
	}
}

func TestRoundTripCanonicalV1Canonical(t *testing.T) {
	orig := &trusted_network_v2.TrustedNetworkV2{
		ID:                     42,
		Name:                   "RoundTrip",
		ConditionType:          "ANY",
		Active:                 true,
		Hostname:               "h.example.com",
		SSID:                   "Wifi",
		DNSSearchDomains:       []string{"a.com", "b.com"},
		DNSServerIPs:           []string{"1.1.1.1"},
		ResolvedIPsForHostname: []string{},
		TrustedDhcpServersIPs:  []string{},
		TrustedEgressIPs:       []string{},
		TrustedGatewayIPs:      []string{},
		TrustedSubnetIPs:       []string{"192.0.2.0/24"},
	}

	v1form, err := canonicalToV1(orig)
	if err != nil {
		t.Fatalf("canonicalToV1: %v", err)
	}
	back, err := v1ToCanonical(v1form)
	if err != nil {
		t.Fatalf("v1ToCanonical: %v", err)
	}

	back.NetworkName = "" // canonical from the resource never sets NetworkName
	back.CompanyID = 0
	if !reflect.DeepEqual(orig, back) {
		t.Errorf("round trip mismatch:\n orig: %#v\n back: %#v", orig, back)
	}
}

func TestV2EndpointUnavailable(t *testing.T) {
	byParsedStatus := func(status int) *errorx.ErrorResponse {
		return &errorx.ErrorResponse{Parsed: &errorx.ParsedAPIError{Status: status}}
	}
	byResponseStatus := func(status int) *errorx.ErrorResponse {
		return &errorx.ErrorResponse{Response: &http.Response{StatusCode: status}}
	}

	unavailable := []*errorx.ErrorResponse{
		byParsedStatus(http.StatusNotFound),
		byParsedStatus(http.StatusBadRequest),
		byParsedStatus(http.StatusForbidden),
		byParsedStatus(http.StatusMethodNotAllowed),
		byParsedStatus(http.StatusNotImplemented),
		byResponseStatus(http.StatusNotFound),
		{Parsed: &errorx.ParsedAPIError{ID: "resource.not.found"}},
	}
	for _, e := range unavailable {
		if !v2EndpointUnavailable(e) {
			t.Errorf("v2EndpointUnavailable(%v) = false, want true", e.Parsed)
		}
	}

	available := []*errorx.ErrorResponse{
		byParsedStatus(http.StatusInternalServerError),
		byParsedStatus(http.StatusBadGateway),
		byResponseStatus(http.StatusServiceUnavailable),
		{},
	}
	for _, e := range available {
		if v2EndpointUnavailable(e) {
			t.Errorf("v2EndpointUnavailable(status=%v) = true, want false", e.Parsed)
		}
	}
}

func TestV1CreateRequiresName(t *testing.T) {
	b := &v1Backend{svc: nil} // the name check must fire before any API use
	_, err := b.Create(context.Background(), &trusted_network_v2.TrustedNetworkV2{ConditionType: "ALL"})
	if err == nil {
		t.Fatal("expected error creating an unnamed trusted network on v1")
	}
}

func TestPickByName(t *testing.T) {
	nets := func(names ...string) []*trusted_network_v2.TrustedNetworkV2 {
		out := make([]*trusted_network_v2.TrustedNetworkV2, len(names))
		for i, n := range names {
			out[i] = &trusted_network_v2.TrustedNetworkV2{ID: i + 1, Name: n}
		}
		return out
	}

	// Exact match wins, case-insensitively.
	got, err := pickByName("bd-trustednetwork02", nets("BD-TrustedNetwork02", "BD-TrustedNetwork02-Backup"))
	if err != nil {
		t.Fatalf("exact match: unexpected error: %v", err)
	}
	if got.Name != "BD-TrustedNetwork02" {
		t.Errorf("exact match picked %q", got.Name)
	}

	// A single unambiguous partial match resolves (the import-by-name
	// case from the field: "TrustedNetwork02" -> "BD-TrustedNetwork02").
	got, err = pickByName("TrustedNetwork02", nets("BD-TrustedNetwork02"))
	if err != nil {
		t.Fatalf("partial match: unexpected error: %v", err)
	}
	if got.Name != "BD-TrustedNetwork02" {
		t.Errorf("partial match picked %q", got.Name)
	}

	// Multiple partial matches fail loudly, never silently pick one.
	if _, err = pickByName("TrustedNetwork", nets("BD-TrustedNetwork01", "BD-TrustedNetwork02")); err == nil {
		t.Error("ambiguous partial match: expected error")
	} else if IsNotFound(err) {
		t.Error("ambiguous partial match must not classify as not-found")
	}

	// No match at all is ErrNotFound.
	if _, err = pickByName("Nope", nets("BD-TrustedNetwork01")); !IsNotFound(err) {
		t.Errorf("no match: expected ErrNotFound, got %v", err)
	}
	if _, err = pickByName("Anything", nil); !IsNotFound(err) {
		t.Errorf("empty candidates: expected ErrNotFound, got %v", err)
	}
}

func TestIsNotFound(t *testing.T) {
	if !IsNotFound(ErrNotFound) {
		t.Error("IsNotFound(ErrNotFound) = false")
	}
	notFound := &errorx.ErrorResponse{Response: &http.Response{StatusCode: http.StatusNotFound}}
	if !IsNotFound(notFound) {
		t.Error("IsNotFound(structured 404) = false")
	}
	other := &errorx.ErrorResponse{Parsed: &errorx.ParsedAPIError{Status: http.StatusInternalServerError}}
	if IsNotFound(other) {
		t.Error("IsNotFound(500) = true")
	}
}
