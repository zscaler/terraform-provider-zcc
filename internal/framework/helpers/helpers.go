// Package helpers hosts the boundary adapters between the Terraform
// Plugin Framework value types (types.Bool, types.Int64, types.String,
// types.List, …) and the ZCC SDK's native Go representations (int,
// string, []string, []int, "0"/"1" strings, etc.). It is the canonical
// home for the small, sharply-scoped utilities that every resource and
// data source in this provider reuses, so the conversion logic stays in
// one place and the per-resource files focus on schema and CRUD wiring.
//
// What lives here:
//
//   - Bool / int / string toggle conversions ("0"/"1" ↔ bool, int ↔
//     bool) — many ZCC APIs encode on/off flags as the literal strings
//     "0" / "1", and silently reject "true" / "false".
//   - Attribute extractors (BoolFromAttr, IntFromAttr, StringFromAttr)
//     used inside overlay-style Expand* functions that walk a
//     SingleNestedAttribute's Attributes() map.
//   - List ↔ Go slice adapters for the trusted-network V2 family and
//     anything else that exposes List(String) / List(Int64) attributes.
//   - The platform / deviceType name ↔ int mapping that's shared across
//     the app_profile_* resources and zia_posture (1=iOS, 2=Android,
//     3=Windows, 4=macOS, 5=Linux).
//
// All helpers map null / unknown / unexpected inputs to the zero value
// of the target type. That's deliberate: it lets resources hand a plan
// model straight through to the helper without first checking
// IsNull()/IsUnknown(), and it pairs naturally with the SDK's
// `,omitempty` JSON tags.
package helpers

import (
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// =============================================================================
// Bool / int / string toggle conversions
// =============================================================================

// BoolToInt converts a Plugin Framework types.Bool into 1 (true) or 0
// (false). Null / unknown values map to 0, which is the safe default
// for every ZCC field that follows the 0/1 toggle convention.
func BoolToInt(b types.Bool) int {
	if b.ValueBool() {
		return 1
	}
	return 0
}

// IntToBool converts a Go int into a Plugin Framework types.Bool. Any
// non-zero value is treated as true.
func IntToBool(i int) types.Bool {
	return types.BoolValue(i != 0)
}

// BoolToString01 converts a Plugin Framework types.Bool into the
// literal "1" / "0" strings the ZCC API expects on the wire for many
// fields (e.g. macPolicy.cacheSystemProxy, policyExtension.useV8JsEngine).
// The API rejects "true" / "false" with a silent
// {"success":"false","id":0} response, so callers MUST use this helper
// rather than fmt.Sprint(b).
func BoolToString01(b types.Bool) string {
	if b.ValueBool() {
		return "1"
	}
	return "0"
}

// String01ToBool converts the literal "0"/"1" strings the ZCC API uses
// for many fields into a Plugin Framework types.Bool. To be forgiving
// of older policies created via the SDK v2 era (which sometimes stored
// "true"/"false"), the helper also accepts those tokens; anything else
// maps to false.
func String01ToBool(s string) types.Bool {
	return types.BoolValue(s == "1" || strings.EqualFold(s, "true"))
}

// =============================================================================
// attr.Value extractors (used inside overlay-style Expand*)
// =============================================================================

// BoolFromAttr extracts a Go bool out of an attr.Value, returning false
// for null / unknown / non-Bool inputs. Used inside overlay-style
// Expand* functions that walk a SingleNestedAttribute's Attributes()
// map and need to bridge the framework's attr.Value into a typed bool
// for the SDK.
func BoolFromAttr(v attr.Value) bool {
	if v == nil || v.IsNull() || v.IsUnknown() {
		return false
	}
	if b, ok := v.(types.Bool); ok {
		return b.ValueBool()
	}
	return false
}

// IntFromAttr extracts a Go int out of an attr.Value, returning 0 for
// null / unknown / non-Int64 inputs.
func IntFromAttr(v attr.Value) int {
	if v == nil || v.IsNull() || v.IsUnknown() {
		return 0
	}
	if i, ok := v.(types.Int64); ok {
		return int(i.ValueInt64())
	}
	return 0
}

// StringFromAttr extracts a Go string out of an attr.Value, returning
// the empty string for null / unknown / non-String inputs.
func StringFromAttr(v attr.Value) string {
	if v == nil || v.IsNull() || v.IsUnknown() {
		return ""
	}
	if s, ok := v.(types.String); ok {
		return s.ValueString()
	}
	return ""
}

// =============================================================================
// List ↔ Go slice adapters
// =============================================================================

// StringListFromList materialises a types.List of String values into a
// Go []string. Null / unknown lists become an empty slice.
func StringListFromList(l types.List) []string {
	if l.IsNull() || l.IsUnknown() {
		return []string{}
	}
	elems := l.Elements()
	out := make([]string, 0, len(elems))
	for _, e := range elems {
		if v, ok := e.(types.String); ok {
			out = append(out, v.ValueString())
		}
	}
	return out
}

// StringListFromAttr is the attr.Value variant of StringListFromList.
// Useful inside overlay-style Expand* functions that walk a
// SingleNestedAttribute's Attributes() map.
func StringListFromAttr(v attr.Value) []string {
	if l, ok := v.(types.List); ok {
		return StringListFromList(l)
	}
	return []string{}
}

// StringListValue wraps a Go []string into a types.List of String for
// state flattening. A nil slice produces an empty list (NOT a null
// list), which is the right behaviour for Optional+Computed attributes
// that should always be known after apply.
func StringListValue(in []string) types.List {
	vals := make([]attr.Value, 0, len(in))
	for _, v := range in {
		vals = append(vals, types.StringValue(v))
	}
	out, _ := types.ListValue(types.StringType, vals)
	return out
}

// IntListFromList materialises a types.List of Int64 values into a Go
// []int. Null / unknown lists become an empty slice.
func IntListFromList(l types.List) []int {
	if l.IsNull() || l.IsUnknown() {
		return []int{}
	}
	elems := l.Elements()
	out := make([]int, 0, len(elems))
	for _, e := range elems {
		if v, ok := e.(types.Int64); ok {
			out = append(out, int(v.ValueInt64()))
		}
	}
	return out
}

// IntListFromAttr is the attr.Value variant of IntListFromList.
func IntListFromAttr(v attr.Value) []int {
	if l, ok := v.(types.List); ok {
		return IntListFromList(l)
	}
	return []int{}
}

// IntListValue wraps a Go []int into a types.List of Int64 for state
// flattening.
func IntListValue(in []int) types.List {
	vals := make([]attr.Value, 0, len(in))
	for _, v := range in {
		vals = append(vals, types.Int64Value(int64(v)))
	}
	out, _ := types.ListValue(types.Int64Type, vals)
	return out
}

// =============================================================================
// Platform / deviceType name ↔ int
//
// Many ZCC APIs identify a target operating system through an integer
// platform / deviceType field:
//
//	1 = iOS
//	2 = Android
//	3 = Windows
//	4 = macOS
//	5 = Linux
//
// Exposing that raw integer in HCL is hostile — operators have to
// remember the mapping (and the mapping is the same across the whole
// product, so it shouldn't be re-derived per-resource). These helpers
// let resources keep a human-readable string on the HCL surface
// (`platform = "macos"`) and translate to/from the SDK integer at the
// expand / flatten boundary.
// =============================================================================

// PlatformNames is the canonical list of platform names accepted on the
// HCL surface. Order matches the numeric value (index+1 == int code).
// Exported as a function (not a var) to keep the slice immutable from
// the caller's perspective — every call returns a fresh copy.
func PlatformNames() []string {
	return []string{"ios", "android", "windows", "macos", "linux"}
}

// PlatformNameToInt converts a HCL-friendly platform name into the
// integer code the ZCC API expects on the wire. Matching is
// case-insensitive ("macOS", "MACOS", "macos" all map to 4). Unknown /
// empty input maps to 0, which the SDK then drops via `omitempty`; the
// caller is responsible for validating "required" semantics through a
// schema validator (`stringvalidator.OneOfCaseInsensitive(...)`).
func PlatformNameToInt(name string) int {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "ios":
		return 1
	case "android":
		return 2
	case "windows":
		return 3
	case "macos":
		return 4
	case "linux":
		return 5
	default:
		return 0
	}
}

// PlatformIntToName converts a ZCC platform integer back to its HCL
// name. Unknown codes (including 0) yield an empty string so flatten
// can decide whether to emit types.StringNull() or surface the value.
func PlatformIntToName(i int) string {
	switch i {
	case 1:
		return "ios"
	case 2:
		return "android"
	case 3:
		return "windows"
	case 4:
		return "macos"
	case 5:
		return "linux"
	default:
		return ""
	}
}
