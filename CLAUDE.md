# Terraform Provider ZCC — Claude Code Guidelines

Project-specific guidance for **`terraform-provider-zcc`** (Zscaler Client Connector). Use this file together with **Agent Skills** under `.claude/skills/` and Cursor rules under `.cursor/rules/`.

This document captures the **architectural decisions** that are already baked into the provider. Treat the patterns described here as authoritative — when you find yourself reaching for a different pattern, that's a sign the conversation needs to happen in a PR description, not silently in code.

## Claude skills (use these skills for this repo)

| Skill | Path | When to use |
|-------|------|-------------|
| **plan-tf-resource** | `.claude/skills/plan-tf-resource/SKILL.md` | New or changed resources/data sources, schema work, acceptance tests, docs, examples, changelog, version bumps |
| **troubleshoot-resource** | `.claude/skills/troubleshoot-resource/SKILL.md` | Drift, API errors, import issues, `TestAcc` failures |
| **upgrade-zscaler-sdk** | `.claude/skills/upgrade-zscaler-sdk/SKILL.md` | Bumping `github.com/zscaler/zscaler-sdk-go/v3`, `go mod tidy` / vendor, documenting SDK upgrades |

**Reference example (ZCC patterns):** `.claude/skills/plan-tf-resource/examples/trusted-network-resource.md` — aligns with `internal/framework/resources/trusted_network.go` and `datasources/trusted_network.go`.

## Cursor rules (companion)

Project rules use **`.mdc`** (Markdown + YAML frontmatter: `description`, optional `globs`, `alwaysApply`). That is Cursor's native rule format; plain `.md` files here are not equivalent.

- `.cursor/rules/terraform-provider-zcc.mdc` — release/changelog/release-notes, version + Makefile, docs style, tests (`alwaysApply: true`)
- `.cursor/rules/troubleshoot-zcc-provider.mdc` — troubleshooting when editing `internal/**/*.go`
- `.cursor/rules/examples/trusted-network-resource.mdc` — ZPA-style doc template for registry pages (uses the v2 list-of-string schema)
- `.cursor/rules/ci-and-quality.mdc` — fmt, vet, staticcheck, lint, docs generation

---

## Project overview

This provider targets **Zscaler Client Connector (ZCC)** APIs via **`zscaler-sdk-go/v3`** (`zscaler/zcc/services/...`). Implementation uses the **Terraform Plugin Framework** (`terraform-plugin-framework`), not SDK v2. Authentication is **OneAPI only** (see below).

### Registered surface (current state)

`internal/framework/provider.go` is the single source of truth — verify the live list there before claiming a resource/data source exists.

**Resources** (alphabetical):

- `zcc_device_cleanup` — singleton; GET `getDeviceCleanupInfo` / PUT `setDeviceCleanupInfo`.
- `zcc_failopen_policy` — singleton; only update + read; delete removes from state only.
- `zcc_forwarding_profile` — full CRUD against `forwarding_profile`.
- `zcc_notification_template` — full CRUD; import accepts numeric id OR name.
- `zcc_trusted_network` — full CRUD against `trusted_network_v2` (v2 schema, lists of strings, `ALL`/`ANY`).
- `zcc_web_app_service` — singleton; lookup by `app_name`, update existing entry, no API delete.
- `zcc_web_privacy` — singleton; GET / PUT.
- `zcc_zia_posture` — full CRUD; import accepts numeric id OR name; data source is currently **id-only** (upstream pagination bug).

**Data sources** (alphabetical):

- `zcc_admin_roles`, `zcc_admin_user`, `zcc_company_info`, `zcc_custom_ip_apps`, `zcc_device_cleanup`, `zcc_devices`, `zcc_failopen_policy`, `zcc_forwarding_profile`, `zcc_notification_template`, `zcc_predefined_ip_apps`, `zcc_process_based_apps`, `zcc_trusted_network`, `zcc_web_app_service`, `zcc_web_privacy`, `zcc_zia_posture`.

### Parked / deregistered

The per-OS `zcc_app_profile_*` resources (`macos`, `ios`, `windows`, `linux`, `android`) and the `zcc_application_profiles` data source are **intentionally deregistered**. The underlying `/web/policy/edit` API returns `HTTP 200 { "success": "false", "id": 0 }` for undocumented field/type combinations that vary per-OS and per-UI capture, so the resources are parked under `local_dev/Backup_Config_Future/` until the API contract stabilises upstream. **Do not re-register** without first reproducing a green apply against a fresh tenant for each OS and obtaining sign-off.

## Project structure

```
internal/framework/
  provider.go                 # Registers resources & data sources (OneAPI-only schema)
  provider_configure.go       # Provider config → *client.Client (OneAPI only)
  resources/<name>.go         # resource.Resource implementations
  resources/<name>_test.go    # TestAcc* (package resources / resources_test)
  datasources/<name>.go       # datasource.DataSource implementations
  datasources/<name>_test.go  # TestAcc* for data sources
  helpers/helpers.go          # Single shared helpers package — see "Helpers" below
  acctest/acctest.go          # PreCheck + ProtoV6ProviderFactories (OneAPI env vars)
internal/client/              # *client.Client, SDK *zscaler.Service (cache disabled, see below)
version/version.go            # ProviderVersion (must match GNUmakefile build13)
docs/index.md                 # Registry landing page (ZIA-style; OneAPI only)
docs/resources/               # Registry resource docs (ZPA-style pages)
docs/data-sources/            # Registry data source docs
docs/guides/                  # release-notes.md, support, troubleshooting
examples/                     # examples/zcc_<name>/basic.tf (+ datasource.tf when applicable)
local_dev/                    # Local apply harnesses + parked app_profile_* code (NOT shipped)
```

Terraform addresses use the **`zcc`** provider prefix; resource types are `zcc_<snake>` set in each type's `Metadata` via `resp.TypeName = req.ProviderTypeName + "_<snake>"`.

---

## Authentication: OneAPI only (no legacy)

The provider authenticates **exclusively** through Zidentity OneAPI. The legacy ZCC V2 client (`zcc_client_id` / `zcc_client_secret` / `zcc_cloud` / `use_legacy_client`) has been removed from every layer — provider schema, `Config`, `NewClient`, acctest helpers, and registry docs.

### Recognised provider attributes / env vars

| Attribute / Env var | Required | Notes |
|---|---|---|
| `client_id` / `ZSCALER_CLIENT_ID` | yes | Zidentity OAuth2 client id |
| `client_secret` / `ZSCALER_CLIENT_SECRET` | one of | Conflicts with `private_key` |
| `private_key` / `ZSCALER_PRIVATE_KEY` | one of | PKCS#1 or PKCS#8 unencrypted; conflicts with `client_secret` |
| `vanity_domain` / `ZSCALER_VANITY_DOMAIN` | yes |  |
| `zscaler_cloud` / `ZSCALER_CLOUD` | optional | e.g. `beta`; required for non-prod Zidentity clouds |
| `http_proxy` / `ZSCALER_HTTP_PROXY` | optional |  |
| `max_retries`, `request_timeout`, `min_wait_seconds`, `max_wait_seconds`, `parallelism` | optional | Tuning knobs |

### Build path

`internal/client/client.go::NewClient` calls **only** `zscalerSDKV3Client`, which:

- Mirrors `terraform-provider-zia/zia/config.go` so engineers switching providers see the same shape.
- **Disables the SDK cache** (`zscaler.WithCache(false)`) — the ZCC SDK does not invalidate cached GETs across writes (PUT `/edit` does not bust cached `/listByCompany` responses), so caching would routinely return stale state on the next plan. Do not re-enable cache without first fixing that upstream.
- Builds the User-Agent via `generateUserAgent`, which is **identical** to ZIA's: `"(<os> <arch>) Terraform/<v> Provider/<v>"`. No `ZCC/zcc` suffix, no conditional branches.
- Toggles `zscaler.WithDebug(true)` when `TF_LOG` is `DEBUG`/`TRACE`.

`internal/framework/provider_configure.go` populates `client.Config` from schema attributes and falls back to the `ZSCALER_*` env vars — there is no longer any `ZCC_*` env var path.

### Acceptance tests

`internal/framework/acctest.PreCheck` enforces `TF_ACC=1`, `ZSCALER_CLIENT_ID`, `ZSCALER_VANITY_DOMAIN`, plus one of `ZSCALER_CLIENT_SECRET` or `ZSCALER_PRIVATE_KEY`. `testClientConfig` only reads OneAPI environment variables — never extend it with legacy fields.

---

## SDK client pattern (resources / data sources)

In `Configure`:

```go
c, ok := req.ProviderData.(*client.Client)
if !ok {
    resp.Diagnostics.AddError("Unexpected Configure Type",
        fmt.Sprintf("Expected *client.Client, got: %T", req.ProviderData))
    return
}
r.client = c
```

Use `r.client.Service` (or `d.client.Service`) — that's `*zscaler.Service` — for every ZCC SDK call. Use **`tflog`** for operational logs and response **diagnostics** for user-visible errors. On read, treat `*errorx.ErrorResponse` + `IsObjectNotFound()` as **remove from state** (`resp.State.RemoveResource(ctx)`).

---

## Architectural patterns

The provider currently leans on a small set of repeatable patterns. Prefer them before inventing new ones.

### 1. Full CRUD resource (reference: `trusted_network`, `notification_template`, `zia_posture`, `forwarding_profile`)

- Resource model carries only **user-configurable** fields. Server-only metadata (`createdBy`, `editedBy`, `companyId`, `guid`, ...) is **deliberately omitted** from the resource — operators read it through the matching **data source** instead.
- The id column at the TF boundary is a **string** (Terraform convention) even when the API uses an int; convert via `strconv.Atoi` / `strconv.Itoa` in expand/flatten.
- Lists are `types.List(String)` (or `Int64`) on the wire and round-tripped through `internal/framework/helpers`.
- Import supports **both** the numeric id and a human-readable key when the SDK provides a `GetByName` helper. Pattern:
  ```go
  if _, err := strconv.ParseInt(req.ID, 10, 64); err == nil {
      resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
      return
  }
  // fall back to GetByName
  ```

### 2. Singleton resource (reference: `device_cleanup`, `failopen_policy`, `web_app_service`, `web_privacy`)

The API exposes **exactly one** record that the tenant cannot create or destroy — only GET and PUT. The Terraform contract is:

- `Create`: GET the existing record → overlay the plan onto it → PUT → flatten back into state.
- `Read`: GET → flatten into state.
- `Update`: same as `Create`.
- `Delete`: **no API call** — just let the framework drop the resource from state. Log via `tflog` that the API is not destructive.
- `ImportState`: GET → write the discovered `id` into state. `web_app_service` additionally supports import by `app_name` (it falls back to `GetByName`).

The "**default-overlay**" idiom is mandatory for singletons: fetch the live record first so any tenant defaults the API populates after a "factory reset" are preserved when the plan only sets a subset of fields. Skipping the overlay produces silent regressions because the SDK marshals zero-valued Go fields with `omitempty` and the API treats absent fields as "leave alone" inconsistently.

### 3. Web policy / app_profile family (parked — DO NOT re-register without sign-off)

`/web/policy/edit` is a per-OS singleton that returns `HTTP 200 { "success": "false", "id": 0 }` when the payload contains any combination it does not like — without indicating which field is wrong. The parked implementations under `local_dev/Backup_Config_Future/` already encode the known-good payloads (`payload-macos.json`, `payload-ios.json`) and rely on per-OS `DefaultIosWebPolicy()` / `DefaultMacWebPolicy()` constructors plus careful `string`/`int` type alignment. **Before re-registering**, you must:

1. Reproduce a clean apply against a fresh tenant for each OS.
2. Confirm the SDK default constructors still match the captured payload byte-for-byte (account for new fields the upstream UI adds silently).
3. Re-run `local_dev/zcc_app_profile_*/main.tf` end-to-end.

If you re-register, the resource type names are hardcoded as `zcc_app_profile_<os>` and each resource sets `deviceType` literally (`1=iOS`, `2=Android`, `3=Windows`, `4=macOS`, `5=Linux`) — never expose `device_type` as a user-settable attribute on a per-OS resource.

---

## Helpers — `internal/framework/helpers/helpers.go`

The **only** shared helpers package. New small adapters between Plugin Framework types and SDK types belong here, not in a per-resource utility file. Current surface:

- **Bool / int / string toggles** — `BoolToInt` / `IntToBool` / `BoolToString01` / `String01ToBool`. Many ZCC fields encode on/off as the literal `"0"` / `"1"` strings; the API silently rejects `"true"` / `"false"` with `{"success":"false","id":0}`, so go through these helpers rather than `fmt.Sprint(b)`.
- **`attr.Value` extractors** — `BoolFromAttr`, `IntFromAttr`, `StringFromAttr`, `StringListFromAttr`, `IntListFromAttr`. Used inside overlay-style `Expand*` walkers that iterate a `SingleNestedAttribute`'s `Attributes()` map.
- **List ↔ slice adapters** — `StringListFromList` / `StringListValue` / `IntListFromList` / `IntListValue`. Null / unknown becomes an empty slice (never nil) so the SDK `omitempty` tag drops the field cleanly.
- **Platform name ↔ int** — `PlatformNames()`, `PlatformNameToInt`, `PlatformIntToName`. Canonical mapping `1=ios, 2=android, 3=windows, 4=macos, 5=linux`. Use `stringvalidator.OneOfCaseInsensitive(helpers.PlatformNames()...)` on any HCL platform attribute.

Helpers map **null / unknown / unexpected** inputs to the zero value of the target type — that's intentional, it pairs cleanly with `,omitempty` on the SDK side.

---

## Trusted network is v2

`zcc_trusted_network` uses the **v2** SDK package (`zscaler/zcc/services/trusted_network_v2`). On the wire:

- Criteria fields (`dns_search_domains`, `dns_server_ips`, `resolved_ips_for_hostname`, `trusted_dhcp_servers_ips`, `trusted_egress_ips`, `trusted_gateway_ips`, `trusted_subnet_ips`) are **lists of strings**. The v1 comma-separated-string surface is gone — do not re-introduce it.
- `condition_type` accepts `"ALL"` / `"ANY"` (or numeric `"0"` / `"1"`); document both.
- `hostname` and `ssid` are scalar strings, not arrays.
- The **resource** exposes only user-configurable fields. Server-side metadata (`company_id`, `created_by`, `edited_by`, `guid`) is exposed **only** on the data source.

---

## ZIA posture data source is temporarily id-only

`datasources/zia_posture.go` requires `id` and refuses name-based lookup because the upstream `/zia-posture-profiles` list endpoint mishandles pagination and silently returns a truncated set. The schema and `Read` carry an explicit `TEMPORARY:` comment. When the upstream API is fixed, re-enable name-based lookup using `GetByName`; do **not** silently remove the `Required: true` on `id` without also fixing the underlying SDK behaviour.

---

## Documentation conventions

- **`docs/index.md`** follows the **ZIA provider's** `index.md` shape (OneAPI section, env-var table, OneAPI scope, argument reference). It **only** documents OneAPI — there is no legacy path to describe. Do not add `Acceptance tests`, internal env vars, or other developer-only knobs here.
- **`docs/resources/zcc_<name>.md`** and **`docs/data-sources/zcc_<name>.md`** follow the **ZPA provider's** registry page shape. Frontmatter (`page_title`, `subcategory`, multi-line `description`), `# zcc_<name> (Resource)` or `(Data Source)`, bullet links to official + API docs, `## Example Usage`, `## Schema` → `### Required` / `### Optional` / `### Read-Only`, and `## Import` for resources that support import.
- Every attribute gets a type and a one-line description; nested blocks are documented with their sub-attributes. Where the official ZCC docs do not describe an attribute, infer from the SDK struct and the surrounding API context — and say so plainly.
- After schema changes, run `make docs` / `go generate` per `.cursor/rules/ci-and-quality.mdc`.

---

## Mandatory artifacts for user-visible changes

For any release-worthy or customer-visible change, update **all** of:

1. **`CHANGELOG.md`** — top entry; PR links to **`https://github.com/zscaler/terraform-provider-zcc/pull/<N>`**.
   ```bash
   gh pr list --repo zscaler/terraform-provider-zcc --state all --limit 1 --json number --jq '.[0].number'
   ```
2. **`docs/guides/release-notes.md`** — same entry as CHANGELOG; update `` `Last updated: v<VERSION>` ``.
3. **`version/version.go`** — `ProviderVersion`.
4. **`GNUmakefile`** — `build13` plugin path and `terraform-provider-zcc_vX.Y.Z` **must match** `ProviderVersion`.
5. **`examples/`** — create or update `examples/zcc_<name>/` for touched resources/data sources.
6. **Registry docs** — `docs/resources/` and `docs/data-sources/` in ZPA-style.
7. **Acceptance tests** — `terraform-plugin-testing` `TestAcc*` for every touched resource/data source, modelled on `terraform-provider-zpa`'s `internal/framework/resources/*_test.go`.

When changing dependencies, follow **`upgrade-zscaler-sdk`** and keep `go.mod` aligned with the **latest published** `zscaler-sdk-go` v3 release unless pinned for a known regression.

## Build, test, debug

```bash
go build ./...

# Acceptance tests (OneAPI credentials required)
TF_ACC=1 go test ./internal/framework/... -run TestAcc -timeout 120m

# Full debug logging for a manual apply
TF_LOG=DEBUG ZSCALER_SDK_VERBOSE=true ZSCALER_SDK_LOG=true terraform apply -no-color 2>&1 | tee /tmp/tf-zcc-debug.log
```

Use **`make fmtcheck`**, **`make vet`**, **`make lint`**, and **`make docs`** per `.cursor/rules/ci-and-quality.mdc`.

## Critical rules

1. Do **not** treat this repo as ZIA: no `zia/` package, no SDK v2 resource pattern, no ZIA rule-order conventions unless ZCC gains analogous APIs.
2. New work belongs in **`internal/framework`** with Framework types (`types.String`, `types.Bool`, `types.Int64`, `types.List`, `types.Object`) and plan modifiers (`UseStateForUnknown`) where appropriate.
3. **No legacy auth.** Never re-add `ZCC_*` env vars, `use_legacy_client`, or a second `NewClient` branch.
4. Singletons use the **default-overlay** pattern and a **no-op delete**. Full-CRUD resources omit server-only metadata from the resource model and expose it only on the data source.
5. Boolean toggles that the API encodes as `"0"`/`"1"` strings **must** go through `helpers.BoolToString01` / `helpers.String01ToBool` — passing `fmt.Sprint(b)` will cause silent `{"success":"false","id":0}` failures.
6. Trusted network is **v2** (lists of strings, `ALL`/`ANY`). Never re-introduce the v1 comma-separated string surface.
7. Per-OS `app_profile_*` resources are **parked** in `local_dev/Backup_Config_Future/`. Do not re-register without a documented green apply per OS.
8. Never ship resources/data sources without **docs**, **examples**, and **integration tests** when the API is testable.
9. Keep **`ProviderVersion`** and **`GNUmakefile`** `build13` versions **identical**.
10. Prefer existing ZCC resources (`trusted_network` for full CRUD, `device_cleanup` / `web_app_service` for singletons) as templates before inventing new patterns.
