# Terraform Provider ZCC — Claude Code Guidelines

Project-specific guidance for **`terraform-provider-zcc`** (Zscaler Client Connector). Use this file together with **Agent Skills** under `.claude/skills/` and Cursor rules under `.cursor/rules/`.

## Claude skills (use these skills for this repo)

| Skill | Path | When to use |
|-------|------|-------------|
| **plan-tf-resource** | `.claude/skills/plan-tf-resource/SKILL.md` | New or changed resources/data sources, schema work, acceptance tests, docs, examples, changelog, version bumps |
| **troubleshoot-resource** | `.claude/skills/troubleshoot-resource/SKILL.md` | Drift, API errors, import issues, `TestAcc` failures |
| **upgrade-zscaler-sdk** | `.claude/skills/upgrade-zscaler-sdk/SKILL.md` | Bumping `github.com/zscaler/zscaler-sdk-go/v3`, `go mod tidy` / vendor, documenting SDK upgrades |

**Reference example (ZCC patterns):** `.claude/skills/plan-tf-resource/examples/trusted-network-resource.md` — aligns with `internal/framework/resources/trusted_network.go` and `datasources/trusted_network.go`.

## Cursor rules (companion)

Project rules use **`.mdc`** (Markdown + YAML frontmatter: `description`, optional `globs`, `alwaysApply`). That is Cursor’s native rule format; plain `.md` files here are not equivalent.

- `.cursor/rules/terraform-provider-zcc.mdc` — release/changelog/release-notes, version + Makefile, docs style, tests (`alwaysApply: true`)
- `.cursor/rules/troubleshoot-zcc-provider.mdc` — troubleshooting when editing `internal/**/*.go`
- `.cursor/rules/examples/trusted-network-resource.mdc` — ZPA-style doc template for registry pages
- `.cursor/rules/ci-and-quality.mdc` — fmt, vet, staticcheck, lint, docs generation

## Project overview

This provider targets **Zscaler Client Connector (ZCC)** APIs via **`zscaler-sdk-go/v3`** (`zscaler/zcc/services/...`). Implementation uses the **Terraform Plugin Framework** (`terraform-plugin-framework`), not SDK v2.

## Project structure

```
internal/framework/
  provider.go                 # Registers resources & data sources
  provider_configure.go       # Provider config → client
  resources/<name>.go         # resource.Resource implementations
  datasources/<name>.go       # datasource.DataSource implementations
internal/client/              # *client.Client, SDK *zscaler.Service
version/version.go            # ProviderVersion (must match GNUmakefile build13)
docs/resources/               # Registry resource docs (ZPA-style pages)
docs/data-sources/            # Registry data source docs
docs/guides/                  # release-notes.md, troubleshooting, etc.
examples/                     # examples/zcc_<name>/basic.tf, datasource.tf
```

Terraform addresses use the **`zcc`** provider prefix; resource types are `zcc_<name>` (set in each type’s `Metadata`).

## SDK client pattern

Resources and data sources receive the configured client in **`Configure`**:

```go
c, ok := req.ProviderData.(*client.Client)
// handle !ok with diagnostics
service := c.Service // *zscaler.Service — pass to ZCC SDK calls
```

Use **`tflog`** for operational logs and response **diagnostics** for user-visible errors. On read, if the API returns not found (`errorx.ErrorResponse` / `IsObjectNotFound()`), remove the resource from state as existing resources do.

## Mandatory artifacts for user-visible changes

For any release-worthy or customer-visible change, update:

1. **`CHANGELOG.md`** — top entry; PR links to **`https://github.com/zscaler/terraform-provider-zcc/pull/<N>`** (next PR: `gh pr list --repo zscaler/terraform-provider-zcc --state all --limit 1 --json number`)
2. **`docs/guides/release-notes.md`** — same entry as CHANGELOG; update ``Last updated: v<VERSION>``
3. **`version/version.go`** — `ProviderVersion`
4. **`GNUmakefile`** — `build13` plugin path and `terraform-provider-zcc_vX.Y.Z` **must match** `ProviderVersion`
5. **`examples/`** — create or update `examples/zcc_<name>/` for touched resources/data sources
6. **Registry docs** — `docs/resources/` and `docs/data-sources/` in **ZPA-style** (see `.cursor/rules/examples/trusted-network-resource.mdc` and [ZPA resource examples](https://registry.terraform.io/providers/zscaler/zpa/latest/docs/resources/zpa_app_connector_group))
7. **Acceptance tests** — every resource and data source: `terraform-plugin-testing` `TestAcc*` tests (mirror **terraform-provider-zpa** `internal/framework/resources/*_test.go`)

When changing dependencies, follow **`upgrade-zscaler-sdk`** and keep `go.mod` aligned with the **latest published** `zscaler-sdk-go` v3 release unless pinned for a known regression.

## Build and test

```bash
go build ./...

# Acceptance tests (credentials required)
TF_ACC=1 go test ./internal/framework/... -run TestAcc -timeout 120m

# Debug logging
TF_LOG=DEBUG ZSCALER_SDK_VERBOSE=true ZSCALER_SDK_LOG=true terraform apply -no-color 2>&1 | tee /tmp/tf-zcc-debug.log
```

Use **`make fmtcheck`**, **`make vet`**, **`make lint`**, and **`make docs`** per `.cursor/rules/ci-and-quality.mdc`.

## Critical rules

1. Do **not** treat this repo as ZIA: no `zia/` package, no SDK v2 resource pattern, no ZIA rule-order conventions unless ZCC gains analogous APIs.
2. New work belongs in **`internal/framework`** with Framework types (`types.String`, schema attributes, plan modifiers as needed).
3. Never ship resources/data sources without **docs**, **examples**, and **integration tests** when the API is testable.
4. Keep **`ProviderVersion`** and **`GNUmakefile`** `build13` versions **identical**.
5. Prefer existing ZCC resources (e.g. **trusted_network**) as templates before inventing new patterns.
