# Roadmap

This document tracks planned work for goldenthread. Items are grouped by milestone, roughly ordered by priority within each.

Shipped work lives in [CHANGELOG.md](CHANGELOG.md).

---

## v0.2 — Wails Integration

Making goldenthread the standard schema companion for [Wails](https://wails.io) desktop apps.

Wails generates TypeScript bindings from bound Go structs, but provides no validation layer. goldenthread fills that gap: the same Go structs that power your Wails backend automatically produce Zod schemas for your frontend.

### `goldenthread init --wails`

> **Status: Planned.** This command is not yet implemented. The current
> `goldenthread init` is a stub. Full Wails integration is the v0.2 milestone.

Zero-friction setup for Wails projects:
- Detect `wails.json` in project root
- Default `--out` to `frontend/src/schemas/`
- Inject goldenthread as a pre-build hook in `wails.json`
- Generate initial schemas for all bound structs

### Auto-detection of Wails project layout

When `wails.json` is present:
- Infer output directory (`frontend/src/`) without requiring `--out`
- Align output alongside `wailsjs/go/models.ts` without conflict

### Wails example project

A complete example in `examples/wails/`:
- Go backend with bound structs and `gt:` validation tags
- Generated Zod schemas consumed by the frontend
- Form validation wired to the generated schemas
- Pre-build hook configured in `wails.json`

### Reference implementation: saw (Scout-and-Wave)

[scout-and-wave-web](https://github.com/blackwell-systems/scout-and-wave-web) is being migrated from a Go HTTP server + embedded React app to a native Wails desktop app. It is the first real-world production use case for goldenthread + Wails together.

The app has ~15 Go structs in `pkg/api/types.go` (`IMPLDocResponse`, `AgentStatus`, `WaveInfo`, `FileOwnershipEntry`, SSE event types, etc.) that are currently mirrored manually in `web/src/types.ts`. goldenthread will replace that manual sync with generated Zod schemas, making saw the canonical example of the integration.

Deliverables:
- `gt:` tags added to saw's Go structs
- goldenthread wired as a pre-build step in the Wails project
- Generated schemas used for frontend validation
- Case study / write-up linked from `docs/WAILS.md`

### Documentation

- `docs/WAILS.md` — end-to-end integration guide
- README "Using with Wails" section

---

## v0.3 — Emitter Targets

### OpenAPI 3.1 emitter

Generate OpenAPI schema components from the same Go structs. Useful for REST APIs that share struct definitions with a Wails or web frontend.

### JSON Schema emitter

Standard JSON Schema output as an alternative to Zod. Useful for non-TypeScript consumers.

---

## v0.4 — Type Coverage

Closing gaps in supported Go types:

- **Recursive / self-referential types** — e.g. tree nodes
- **Literal types**: fixed string/int values (standalone)
- **Fixed-length arrays** — `[N]T` tuples

> **Discriminated union types shipped early.** A Go struct with a
> `gt:"discriminator"` field and one `gt:"variant:<name>"` payload field per
> variant now compiles to a Zod `z.discriminatedUnion`, pulled forward to serve
> a visual flow-builder that consumes a Go wiring config. See
> [CHANGELOG.md](CHANGELOG.md) and [docs/TAG_SPEC.md](docs/TAG_SPEC.md).

---

## Backlog

Items without a milestone yet:

- **Watch mode** — `goldenthread watch ./models` for dev without wails dev
- **VS Code extension** — inline schema preview, error highlighting in struct tags
- **Config file** — `goldenthread.yaml` for project-level defaults (output dir, target, recursive)
- **Cross-field validation** — e.g. `confirm_password` must equal `password`
- **Custom format support** — user-defined format strings mapped to Zod refinements

---

## Not planned

- Runtime library — goldenthread is and will remain a build-time tool only
- Non-Go source languages — Go structs are the source of truth by design
