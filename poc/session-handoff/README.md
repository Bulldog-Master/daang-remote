# Session Handoff Contract PoC

**EXPERIMENTAL — evidence-generating code only.**

This module is *not* the Daang Remote backend, *not* the production Control
Plane, *not* the production Data Plane, *not* the permanent security core,
*not* the final token format, *not* the final cryptographic design, and
*not* the permanent Go architecture.

It exists to answer one bounded question:

> Can Daang Remote implement the ADR-0003 and ADR-0004 Control Plane →
> Data Plane handoff properties in a small, local, deterministic Go proof
> of concept without passing long-lived identity across the plane seam,
> creating transferable bearer authority, allowing one compromised session
> to gain authority over another session, or silently weakening the
> accepted ADR contracts?

Go was selected as the *experimental* substrate for this bounded PoC only.
Go is not selected as Daang Remote's permanent product language. Rust, Go,
and a split Go/Rust architecture remain open after the evidence review.

## Layout

```
poc/session-handoff/
├── go.mod
├── README.md
├── handoff.go        // mock Issuer (Control Plane) + Validator (Data Plane)
└── handoff_test.go   // 30 mandatory tests + concurrency + fuzz seed
```

The evidence report lives at `docs/evidence/session-handoff-poc.md`.

## Run

```
go test ./...
go test -race ./...
go test -run=^$ -fuzz=FuzzVerifyMalformed -fuzztime=10s ./...
```

## Dependencies

None outside the Go standard library.
