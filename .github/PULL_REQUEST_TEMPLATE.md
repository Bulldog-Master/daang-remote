<!--
Every Daang Remote pull request must comply with Foundry 1.

Authoritative governance:
https://github.com/Bulldog-Master/foundry

Sections may not be deleted.
If a section does not apply, explain why.
A failed mandatory gate blocks merge by default.
-->

## Change packet

### Problem

What single concrete problem does this change solve?

### Declared scope

List every file, component, endpoint, layer, or behaviour this change is permitted to modify.

Anything outside this list is out of scope.

### Out of scope

What related work is deliberately excluded?

### Acceptance criteria

List observable pass/fail conditions.

### Layers touched

Check all that apply:

- [ ] Frontend
- [ ] Bridge
- [ ] Backend
- [ ] Storage
- [ ] Authentication
- [ ] Transport
- [ ] Security
- [ ] Cryptography
- [ ] Privacy and metadata
- [ ] Update system
- [ ] Deployment
- [ ] Monetization
- [ ] Documentation only

### Evidence and assumptions

What is known?

What is assumed?

What evidence supports the change?

If evidence is thin, say so honestly.

### Privacy and metadata impact

State explicitly:

- identifiers introduced, changed, or removed;
- metadata collected, stored, transmitted, inferred, or exposed;
- sessions, devices, identities, routes, endpoints, or activities that may become linkable;
- who can perform that linkage;
- under what conditions;
- retention period;
- unavoidable linkability;
- engineering telemetry, build metadata, debug tools, source maps, environment identifiers, or other engineering-side data that could enter a shipped artifact.

### Security and cryptographic impact

State explicitly:

- trust boundaries changed;
- authentication or authorization changes;
- secrets or key-lifecycle changes;
- cryptographic primitives or protocols introduced;
- replay, integrity, randomness, validation, or secure-failure concerns;
- post-quantum claims;
- custom cryptography, if any;
- dependencies introduced.

### Testing and evidence plan

What tests, measurements, manual checks, or review artifacts demonstrate that the acceptance criteria are satisfied?

### Rollback

How is the change reversed?

If rollback is partial, expensive, or impossible, state that clearly.

## Mandatory gate outcomes

Allowed outcomes:

- `Pass`
- `Fail`
- `Pass with conditions`
- `N/A — <specific reason>`

The evaluator must be an independent source.

Naming the same person under two role labels does not satisfy independence.

The founder remains the final approval authority, but founder approval does not substitute for independent evaluation when the founder produced the change.

### Architecture

- **Evaluator:**
- **Outcome:**
- **Findings and evidence:**
- **Conditions:**
- **Possible misses or limitations:**
- **Value added:**

### Security and Cryptography

- **Evaluator:**
- **Outcome:**
- **Findings and evidence:**
- **Conditions:**
- **Possible misses or limitations:**
- **Value added:**

### Privacy and Metadata

- **Evaluator:**
- **Outcome:**
- **Findings and evidence:**
- **Engineering-to-product identity leakage checked:** Yes | No | N/A
- **Conditions:**
- **Possible misses or limitations:**
- **Value added:**

### Quality and Verification

- **Evaluator:**
- **Outcome:**
- **Findings and evidence:**
- **Conditions:**
- **Possible misses or limitations:**
- **Value added:**

## Failed-gate disposition

A failed mandatory gate blocks merge by default.

- **Failed gates:** None, or list each failed gate.
- **Founder override:** None, or identify the explicit recorded override.
- **Accepted risk:**
- **Rationale for proceeding:**
- **Required follow-up:**
- **Expiry or reassessment condition:**

An override does not convert a failed gate into a pass.

The gate remains recorded as failed.

## Gate-value review

For this change:

- Did each gate add value?
- Did any gate produce immaterial findings?
- Did any gate appear likely to miss a defect within its stated scope?
- Did any gate create friction disproportionate to its value?
- Does any gate require recalibration, narrower scope, a different evaluator, or demotion?

## Human approval checklist

- [ ] The change remained inside declared scope.
- [ ] Acceptance criteria are satisfied.
- [ ] AI-generated work received human review.
- [ ] Each gate record names an independent evaluating source.
- [ ] All mandatory gate outcomes are recorded.
- [ ] Any failed gate has an explicit recorded founder override.
- [ ] Privacy and metadata exposure is documented honestly.
- [ ] No unsupported cryptographic, privacy, unlinkability, or post-quantum claim is made.
- [ ] The founder has approved this pull request for merge.
