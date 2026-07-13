# ADR-0002: Control Plane / Data Plane Separation

- Status: Proposed
- Date: 2026-07-13
- Foundry generation: 1
- Product: Daang Remote
- Supersedes: none
- Superseded by: none
- Related: ADR-0001 (Foundry, model-neutral boundaries); Daang Remote product charter

## Context

Daang Remote is a remote-access product whose charter requires
metadata minimization, resistance to unnecessary correlation, prevention of
avoidable linkage, explicit disclosure of unavoidable linkability,
cryptographic agility, an evidence-supported post-quantum posture, platform
independence, and infrastructure replaceability.

Remote access, unlike a batch messaging application, combines two very
different traffic profiles inside one product:

1. **Rare, small, latency-tolerant, security-critical exchanges**
   used to find peers, negotiate a session, and prove identity.
2. **Continuous, high-throughput, latency-sensitive streams**
   used to actually operate a remote device (screen, input, audio).

Choosing a single transport for both classes forces one profile to be
punished for the other's requirements. The charter is explicit that the
latency-sensitive interactive path must be evaluated separately and must
not be forced through a mechanism only because that mechanism is
associated with the product's privacy goals.

This ADR does not select a transport. It defines the **planes** the product
is architected around, enumerates the responsibilities that belong to
each, and specifies the properties any mechanism selected in a later ADR
must be evaluated against.

## Decision

Daang Remote is architected as **two logically distinct planes** with
independently selectable transports and cryptographic mechanisms.

### Control Plane

The Control Plane owns:

- discovery (locating a peer without revealing its network position);
- rendezvous (bringing two peers into a shared coordination context);
- session negotiation (parameters, versions, transport selection);
- authentication negotiation (which method, which credentials, which policy);
- identity exchange (long-lived identifiers, device identity, revocation state);
- relay coordination (selection, rotation, failover of any intermediary);
- metadata protection (limiting what an observer or infrastructure operator
  learns about who is talking to whom, when, and how often);
- capability negotiation (features, codecs, permissions, quotas).

Control Plane traffic is characterized as: **infrequent, small, tolerant of
seconds of added latency, intolerant of metadata leakage, and the primary
target of correlation attacks.**

### Data Plane

The Data Plane owns:

- interactive screen updates;
- keyboard input;
- mouse and pointer input;
- clipboard synchronization;
- audio;
- future video streams;
- future bulk file transfer.

Data Plane traffic is characterized as: **continuous, high-volume, extremely
sensitive to added latency and jitter, protected by a session context that
was already established on the Control Plane, and correlatable simply by
its volume and timing regardless of what carries it.**

### Separation invariants

The two planes are logically separate even if a later ADR decides they
share physical machinery. Specifically:

1. Control Plane mechanisms may not be chosen to satisfy Data Plane
   latency requirements, and Data Plane mechanisms may not be chosen to
   satisfy Control Plane metadata requirements.
2. Neither plane may leak identifiers or session state to the other in a
   form that re-links what the other was designed to unlink.
3. Any shared component (identity, key material, relay pool) must be
   analyzed as a correlation surface and documented in the ADR that
   introduces the sharing.

## Responsibility matrix

Each responsibility is evaluated on eleven axes. The matrix does not
select a mechanism; it defines the property surface any candidate
mechanism must be scored against in a later ADR.

Axes:

- **LAT** — latency sensitivity (low / medium / high)
- **BW** — bandwidth profile (small burst / small steady / large steady / bulk)
- **META** — metadata exposure risk if unprotected (low / medium / high)
- **CORR** — correlation risk across sessions or devices (low / medium / high)
- **PRIV** — required privacy posture (baseline / strong / hardened)
- **UNLINK** — required unlinkability (per-session / per-device / per-identity / none)
- **CRYPTO** — required cryptographic protection (integrity / confidentiality / both / mutual auth + both)
- **PQ** — post-quantum posture required at Foundry 1 (none / hybrid handshake / hybrid handshake + PQ-safe long-lived credentials)
- **cMixx** — plausible measurable value from cMixx (yes / conditional / no)
- **xxDK** — plausible measurable value from xxDK (yes / conditional / no)
- **ALT** — whether a differently optimized transport is likely preferable

### Control Plane

| Responsibility          | LAT | BW           | META | CORR | PRIV     | UNLINK        | CRYPTO             | PQ                              | cMixx        | xxDK        | ALT |
|-------------------------|-----|--------------|------|------|----------|---------------|--------------------|---------------------------------|--------------|-------------|-----|
| Discovery               | low | small burst  | high | high | hardened | per-identity  | integrity + conf.  | hybrid + PQ long-lived creds    | conditional  | conditional | no  |
| Rendezvous              | low | small burst  | high | high | hardened | per-session   | mutual auth + both | hybrid + PQ long-lived creds    | conditional  | conditional | no  |
| Session negotiation     | low | small burst  | med  | med  | strong   | per-session   | mutual auth + both | hybrid handshake                | conditional  | conditional | conditional |
| Authentication negot.   | low | small burst  | high | high | hardened | per-identity  | mutual auth + both | hybrid + PQ long-lived creds    | conditional  | yes         | conditional |
| Identity exchange       | low | small burst  | high | high | hardened | per-identity  | mutual auth + both | hybrid + PQ long-lived creds    | conditional  | yes         | conditional |
| Relay coordination      | med | small steady | high | high | hardened | per-session   | mutual auth + both | hybrid handshake                | conditional  | conditional | conditional |
| Metadata protection     | low | small steady | high | high | hardened | per-session   | both               | hybrid handshake                | yes          | conditional | no  |
| Capability negotiation  | low | small burst  | med  | low  | strong   | per-session   | mutual auth + both | hybrid handshake                | conditional  | conditional | conditional |

### Data Plane

| Responsibility        | LAT  | BW           | META | CORR | PRIV     | UNLINK       | CRYPTO             | PQ                       | cMixx | xxDK        | ALT |
|-----------------------|------|--------------|------|------|----------|--------------|--------------------|--------------------------|-------|-------------|-----|
| Screen updates        | high | large steady | med  | high | strong   | per-session  | mutual auth + both | hybrid handshake         | no    | conditional | yes |
| Keyboard input        | high | small steady | med  | high | strong   | per-session  | mutual auth + both | hybrid handshake         | no    | conditional | yes |
| Mouse / pointer input | high | small steady | med  | high | strong   | per-session  | mutual auth + both | hybrid handshake         | no    | conditional | yes |
| Clipboard             | med  | small burst  | med  | med  | strong   | per-session  | mutual auth + both | hybrid handshake         | no    | conditional | yes |
| Audio                 | high | large steady | med  | high | strong   | per-session  | mutual auth + both | hybrid handshake         | no    | conditional | yes |
| Future video          | high | bulk         | med  | high | strong   | per-session  | mutual auth + both | hybrid handshake         | no    | conditional | yes |
| Future file transfer  | med  | bulk         | med  | med  | strong   | per-session  | mutual auth + both | hybrid handshake         | no    | conditional | yes |

## Objective evaluation of candidate mechanisms

The charter does not preselect a transport; the following is an objective
evaluation against the axes above. No mechanism is chosen here.

### cMixx

- **Design goal.** Metadata-resistant messaging via mixing: high-latency,
  anonymity-set-based unlinkability between sender and receiver.
- **Fit for Control Plane.** Plausible value where the observable fact of
  a rendezvous, discovery lookup, or coordination message would itself
  leak actionable metadata. Latency tolerance of Control Plane traffic is
  compatible with a mix's added delay. Value is highest for
  metadata-protection and discovery/rendezvous responsibilities, and
  conditional elsewhere on the Control Plane depending on whether the
  same anonymity set actually contains meaningfully diverse traffic at
  Daang Remote's usage scale.
- **Fit for Data Plane.** No. A mix network's core mechanism is added
  latency and reordering across a batching window. Interactive screen,
  input, and audio traffic have hard latency and jitter budgets that the
  mechanism is designed to violate. Attempting to force interactive
  traffic through a mix either destroys interactivity or shrinks the
  anonymity set to the point that the privacy claim is no longer real.
- **Risks.** Anonymity set collapse at low participation; correlation via
  volume and timing at network edges even inside the mix; long-tail
  liveness dependency on the mix operator population.

### xxDK

- **Design goal.** A development kit that packages identity, end-to-end
  encryption, and messaging primitives on top of the underlying mix.
- **Fit for Control Plane.** Plausible value for identity exchange and
  authentication negotiation, where its identity primitives and E2E
  guarantees may reduce custom cryptography and provide a coherent
  post-quantum posture if its documented primitives meet the required PQ
  posture (a matter for a later ADR with evidence). Value for other
  Control Plane responsibilities is conditional on whether its API shape
  matches Daang Remote's rendezvous and relay-coordination needs without
  forcing the interactive path through the mix.
- **Fit for Data Plane.** Conditional. If xxDK exposes a bearer-only,
  non-mix, post-quantum-capable channel derived from an already-mixed
  handshake, it could carry Data Plane traffic; if using xxDK requires
  routing Data Plane traffic through the mix, it is disqualified for the
  Data Plane by the same reasoning as cMixx.
- **Risks.** Coupling identity, cryptography, and transport into a single
  SDK reduces cryptographic agility unless the SDK itself is agile;
  ecosystem and long-term maintenance concentration risk.

### Separately optimized encrypted transport for the Data Plane

- **Design goal.** A modern, well-analyzed transport (representative
  examples for future evaluation, not selections: QUIC/TLS 1.3 with a
  hybrid PQ key exchange, Noise-framework channels, WireGuard-style
  tunnels) established under a session context authenticated on the
  Control Plane.
- **Fit for Data Plane.** High. Directly targets the latency, jitter,
  loss-recovery, and bandwidth profile of interactive remote access; can
  carry a hybrid post-quantum handshake today; keeps cryptographic
  agility by allowing the primitive suite to be swapped without changing
  the Control Plane.
- **Metadata cost.** Traffic-analysis-visible: an observer at either
  endpoint's network sees the existence, direction, and volume of a
  continuous session. This is an **unavoidable linkability** of
  interactive remote access, must be disclosed as such per the charter,
  and cannot be honestly hidden by tunneling the same traffic through a
  mix at Daang Remote's expected scale.
- **Fit for Control Plane.** Not selected; a separately optimized
  encrypted transport does not by itself provide the metadata protection
  the Control Plane requires.

### Should the Data Plane traverse the same privacy transport as the Control Plane?

Evaluated on the charter's own terms: **no, not by default.** The
Control Plane's metadata protection requirement and the Data Plane's
latency requirement are in direct tension inside a single mechanism.
Forcing the interactive path through a metadata-resistant transport
either destroys the property the product is trying to deliver
(usability good enough to replace a commercial remote-access tool) or
degrades the mix's own anonymity set so severely that the metadata
protection becomes theatre. Coupling them also removes cryptographic
agility on the Data Plane, because the Data Plane's primitives become
whatever the mix exposes. A later ADR may still choose to co-locate
them if it produces evidence that both properties survive; the burden
of that evidence is on the co-location proposal, not on separation.

## Consequences

### Positive

- Each plane can be evaluated, measured, and later replaced independently.
- Latency and metadata trade-offs are made explicit rather than hidden
  inside a monolithic transport choice.
- Cryptographic agility is preserved: primitives on either plane can be
  rotated without rewriting the other.
- Post-quantum posture can be advanced per plane, on evidence, at
  different times.
- Unavoidable linkability on the Data Plane can be disclosed honestly
  without contaminating the Control Plane's stronger unlinkability claims.

### Negative

- Two transports and two threat models must be maintained.
- A binding between the two planes (session handoff) becomes a new
  correlation surface that every later ADR touching that binding must
  analyze.
- Product complexity is higher than a single-transport design.

### Neutral

- No transport, library, language, framework, identity system, relay
  system, or cryptographic primitive is selected by this ADR.

## Alternatives considered

1. **Single privacy transport for both planes.** Rejected as default;
   see the "Should the Data Plane traverse the same privacy transport"
   evaluation above.
2. **Single conventional encrypted transport for both planes.** Rejected
   as default; it does not satisfy the Control Plane's metadata
   protection requirement.
3. **Three or more planes** (e.g. splitting identity from control).
   Deferred. May be revisited if a later ADR shows that identity or
   relay coordination has materially different properties from the rest
   of the Control Plane at implementation time.

## Non-goals

This ADR does not:

- select a control-plane transport;
- select a data-plane transport;
- select or reject cMixx;
- select or reject xxDK;
- specify identity, authentication, key-management, storage, deployment,
  or update mechanisms;
- specify a language, framework, or runtime;
- describe APIs, wire formats, or data schemas;
- describe user experience.

## Verification

This ADR is verifiable as an architectural artifact by:

1. Confirming both planes and all listed responsibilities are covered by
   the responsibility matrix.
2. Confirming no mechanism is selected — every mechanism reference is
   evaluated on the eleven axes above, not chosen.
3. Confirming the separation invariants are stated and testable against
   every future ADR that proposes shared components.
4. Confirming the axes trace back to the charter's required properties
   (metadata minimization, correlation resistance, unlinkability
   disclosure, cryptographic agility, PQ posture, replaceability).

## Rollback

This decision is reversed by a superseding ADR that either:

- reunifies the planes onto a single transport with evidence that both
  the Control Plane's metadata property and the Data Plane's latency
  property survive that unification; or
- further subdivides the planes with evidence that a two-plane model is
  insufficient.

Until such an ADR is accepted, this decision governs.
