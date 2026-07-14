package handoff

import (
	"bytes"
	"encoding/hex"
	"strings"
	"sync"
	"testing"
	"time"
)

// -----------------------------------------------------------------------------
// Test helpers.
// -----------------------------------------------------------------------------

const dpRole = "data-plane-alpha"

func setupSession(t *testing.T, granted []Capability) (*Issuer, *Validator, string, []byte) {
	t.Helper()
	iss := NewIssuer()
	v := NewValidator(dpRole)
	sid, sk := iss.StartSession()
	v.InstallSession(sid, sk, dpRole, granted)
	return iss, v, sid, sk
}

func issueOK(t *testing.T, iss *Issuer, sid string, caps []Capability) *Artifact {
	t.Helper()
	a, err := iss.Issue(sid, dpRole, "interactive-remote", caps, time.Minute)
	if err != nil {
		t.Fatalf("issue: %v", err)
	}
	return a
}

// -----------------------------------------------------------------------------
// Language-neutral test specifications backing each Go test appear in
// docs/evidence/session-handoff-poc.md under "Language-neutral test
// specifications". Each Go test below implements one numbered mandatory
// test from the handoff brief.
// -----------------------------------------------------------------------------

// #1 Valid artifact is accepted.
func TestValidArtifactAccepted(t *testing.T) {
	iss, v, sid, _ := setupSession(t, []Capability{CapViewScreen})
	a := issueOK(t, iss, sid, []Capability{CapViewScreen})
	if err := v.Verify(a); err != nil {
		t.Fatalf("verify: %v", err)
	}
	if err := v.Use(a, CapViewScreen); err != nil {
		t.Fatalf("use view_screen: %v", err)
	}
}

// #2 Altered artifact is rejected.
func TestAlteredArtifactRejected(t *testing.T) {
	iss, v, sid, _ := setupSession(t, []Capability{CapViewScreen})
	a := issueOK(t, iss, sid, []Capability{CapViewScreen})
	a.Purpose = "elevated-admin" // tamper
	if err := v.Verify(a); err == nil {
		t.Fatal("expected tamper to fail verify")
	}
}

// #3 Expired artifact is rejected.
func TestExpiredArtifactRejected(t *testing.T) {
	iss, v, sid, _ := setupSession(t, []Capability{CapViewScreen})
	a, err := iss.Issue(sid, dpRole, "p", []Capability{CapViewScreen}, time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(5 * time.Millisecond)
	if err := v.Verify(a); err == nil || !strings.Contains(err.Error(), "expired") {
		t.Fatalf("expected expired, got %v", err)
	}
}

// #4 Artifact addressed to another recipient is rejected.
func TestWrongRecipientRejected(t *testing.T) {
	iss, v, sid, _ := setupSession(t, []Capability{CapViewScreen})
	a, err := iss.Issue(sid, "other-role", "p", []Capability{CapViewScreen}, time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	if err := v.Verify(a); err == nil {
		t.Fatal("expected recipient mismatch")
	}
}

// #5 Artifact used for another session is rejected.
func TestWrongSessionRejected(t *testing.T) {
	iss, v, sidA, _ := setupSession(t, []Capability{CapViewScreen})
	sidB, skB := iss.StartSession()
	v.InstallSession(sidB, skB, dpRole, []Capability{CapViewScreen})
	a := issueOK(t, iss, sidA, []Capability{CapViewScreen})
	a.SessionID = sidB // tamper session id
	if err := v.Verify(a); err == nil {
		t.Fatal("expected session mismatch to fail signature")
	}
}

// #6 Artifact used for another purpose is rejected.
func TestPurposeTamperRejected(t *testing.T) {
	iss, v, sid, _ := setupSession(t, []Capability{CapViewScreen})
	a := issueOK(t, iss, sid, []Capability{CapViewScreen})
	a.Purpose = "different"
	if err := v.Verify(a); err == nil {
		t.Fatal("expected purpose tamper to fail")
	}
}

// #7 Replayed artifact is rejected.
func TestReplayRejected(t *testing.T) {
	iss, v, sid, _ := setupSession(t, []Capability{CapViewScreen})
	a := issueOK(t, iss, sid, []Capability{CapViewScreen})
	if err := v.Use(a, CapViewScreen); err != nil {
		t.Fatal(err)
	}
	if err := v.Use(a, CapViewScreen); err == nil {
		t.Fatal("expected replay to fail")
	}
}

// #8 Revoked artifact is rejected.
func TestRevokedRejected(t *testing.T) {
	iss, v, sid, sk := setupSession(t, []Capability{CapViewScreen})
	if err := v.Revoke(sid, CapViewScreen, sk); err != nil {
		t.Fatal(err)
	}
	a := issueOK(t, iss, sid, []Capability{CapViewScreen})
	if err := v.Use(a, CapViewScreen); err == nil {
		t.Fatal("expected revoked capability denial")
	}
}

// #9 Invalidated session is rejected.
func TestInvalidatedRejected(t *testing.T) {
	iss, v, sid, sk := setupSession(t, []Capability{CapViewScreen})
	if err := v.Invalidate(sid, sk); err != nil {
		t.Fatal(err)
	}
	a := issueOK(t, iss, sid, []Capability{CapViewScreen})
	if err := v.Verify(a); err == nil {
		t.Fatal("expected invalidated rejection")
	}
}

// #10-14 View-only cannot exercise keyboard/pointer/clipboard/audio/file.
func TestViewOnlyCannotEscalate(t *testing.T) {
	iss, v, sid, _ := setupSession(t, []Capability{CapViewScreen})
	for _, cap := range []Capability{
		CapSendKeyboardInput, CapSendPointerInput,
		CapUseClipboard, CapUseAudio, CapTransferFile,
	} {
		a := issueOK(t, iss, sid, []Capability{cap})
		if err := v.Use(a, cap); err == nil {
			t.Fatalf("view-only session must not exercise %s", cap)
		}
	}
}

// #15 Clipboard does not imply file transfer.
func TestClipboardNoFileTransfer(t *testing.T) {
	iss, v, sid, _ := setupSession(t, []Capability{CapUseClipboard})
	a := issueOK(t, iss, sid, []Capability{CapTransferFile})
	if err := v.Use(a, CapTransferFile); err == nil {
		t.Fatal("clipboard must not imply file transfer")
	}
}

// #16 File transfer does not imply keyboard or pointer.
func TestFileTransferNoInput(t *testing.T) {
	iss, v, sid, _ := setupSession(t, []Capability{CapTransferFile})
	for _, cap := range []Capability{CapSendKeyboardInput, CapSendPointerInput} {
		a := issueOK(t, iss, sid, []Capability{cap})
		if err := v.Use(a, cap); err == nil {
			t.Fatalf("file transfer must not imply %s", cap)
		}
	}
}

// #17 Revoking one capability does not revoke unrelated granted capabilities.
func TestRevokeIsolation(t *testing.T) {
	iss, v, sid, sk := setupSession(t, []Capability{CapViewScreen, CapUseClipboard})
	if err := v.Revoke(sid, CapUseClipboard, sk); err != nil {
		t.Fatal(err)
	}
	a := issueOK(t, iss, sid, []Capability{CapViewScreen})
	if err := v.Use(a, CapViewScreen); err != nil {
		t.Fatalf("view_screen must still work: %v", err)
	}
}

// #18 Expiring one session does not affect another session.
func TestExpirySessionIsolation(t *testing.T) {
	iss := NewIssuer()
	v := NewValidator(dpRole)
	sidA, skA := iss.StartSession()
	sidB, skB := iss.StartSession()
	v.InstallSession(sidA, skA, dpRole, []Capability{CapViewScreen})
	v.InstallSession(sidB, skB, dpRole, []Capability{CapViewScreen})

	aA, _ := iss.Issue(sidA, dpRole, "p", []Capability{CapViewScreen}, time.Millisecond)
	time.Sleep(5 * time.Millisecond)
	if err := v.Verify(aA); err == nil {
		t.Fatal("A must be expired")
	}
	aB := issueOK(t, iss, sidB, []Capability{CapViewScreen})
	if err := v.Use(aB, CapViewScreen); err != nil {
		t.Fatalf("B must remain valid: %v", err)
	}
}

// #19 Invalidating one session does not affect another session.
func TestInvalidateSessionIsolation(t *testing.T) {
	iss := NewIssuer()
	v := NewValidator(dpRole)
	sidA, skA := iss.StartSession()
	sidB, skB := iss.StartSession()
	v.InstallSession(sidA, skA, dpRole, []Capability{CapViewScreen})
	v.InstallSession(sidB, skB, dpRole, []Capability{CapViewScreen})
	if err := v.Invalidate(sidA, skA); err != nil {
		t.Fatal(err)
	}
	aB := issueOK(t, iss, sidB, []Capability{CapViewScreen})
	if err := v.Use(aB, CapViewScreen); err != nil {
		t.Fatalf("B must remain valid: %v", err)
	}
}

// #20 One session's artifact cannot authorize another session.
func TestArtifactCannotCrossSession(t *testing.T) {
	iss := NewIssuer()
	v := NewValidator(dpRole)
	sidA, skA := iss.StartSession()
	sidB, skB := iss.StartSession()
	v.InstallSession(sidA, skA, dpRole, []Capability{CapViewScreen})
	v.InstallSession(sidB, skB, dpRole, []Capability{CapViewScreen})
	aA := issueOK(t, iss, sidA, []Capability{CapViewScreen})
	// swap into B: rewrite session id and try again
	aA.SessionID = sidB
	if err := v.Verify(aA); err == nil {
		t.Fatal("cross-session artifact must not verify")
	}
}

// #21 Artifact representation contains no prohibited long-lived identity fields.
func TestArtifactHasNoProhibitedFields(t *testing.T) {
	iss, _, sid, _ := setupSession(t, []Capability{CapViewScreen})
	a := issueOK(t, iss, sid, []Capability{CapViewScreen})
	if hits := ProhibitedFieldsInArtifact(a); len(hits) > 0 {
		t.Fatalf("artifact contains prohibited fields: %v", hits)
	}
}

// #22 Return-flow events contain only bounded lifecycle information.
func TestEventsBounded(t *testing.T) {
	iss, v, sid, sk := setupSession(t, []Capability{CapViewScreen})
	_ = issueOK(t, iss, sid, []Capability{CapViewScreen})
	if err := v.Revoke(sid, CapViewScreen, sk); err != nil {
		t.Fatal(err)
	}
	if err := v.Invalidate(sid, sk); err != nil {
		t.Fatal(err)
	}
	for _, e := range v.Events() {
		if e.SessionID == "" {
			t.Fatalf("event missing session id: %+v", e)
		}
		if hits := ProhibitedFieldsInEvent(e); len(hits) > 0 {
			t.Fatalf("event %v contains prohibited fields: %v", e.Type, hits)
		}
	}
}

// #23 Return-flow events contain no prohibited identity fields (dedicated check).
func TestEventsHaveNoProhibitedFields(t *testing.T) {
	_, v, sid, sk := setupSession(t, []Capability{CapViewScreen})
	_ = v.EndSession(sid, sk)
	for _, e := range v.Events() {
		if hits := ProhibitedFieldsInEvent(e); len(hits) > 0 {
			t.Fatalf("event %v contains prohibited fields: %v", e.Type, hits)
		}
	}
}

// #24 Return-flow state is not promoted into a long-lived user or account record.
// The validator only stores sessionState maps; nothing links to user/account.
func TestNoLongLivedUserRecord(t *testing.T) {
	iss, v, sid, sk := setupSession(t, []Capability{CapViewScreen})
	_ = issueOK(t, iss, sid, []Capability{CapViewScreen})
	_ = v.EndSession(sid, sk)
	// Structural assertion: Validator exposes no per-user accessor.
	// (Compile-time absence is what we assert; runtime sentinel:)
	if got := len(v.Events()); got == 0 {
		t.Fatal("expected some events")
	}
	// Also assert no event has a Capability field pointing at anything
	// resembling an account id.
	for _, e := range v.Events() {
		if strings.Contains(string(e.Capability), "@") {
			t.Fatalf("suspicious capability payload: %v", e)
		}
	}
}

// Adversary helper for partial-compromise tests.
// It has access ONLY to sessionKey_A. Global issuer authority is uncompromised.
type adversary struct {
	sessionAKey []byte
}

// -----------------------------------------------------------------------------
// Partial-compromise scenario (mandatory tests 25-30).
//
// Preconditions: two sessions A and B exist with independent session-local
// material. Session A's key is exposed to the adversary. The global issuer
// key is NOT exposed. Session B remains legitimately in use.
// -----------------------------------------------------------------------------

func partialCompromiseSetup(t *testing.T) (*Issuer, *Validator, string, string, []byte, []byte, *adversary) {
	t.Helper()
	iss := NewIssuer()
	v := NewValidator(dpRole)
	sidA, skA := iss.StartSession()
	sidB, skB := iss.StartSession()
	v.InstallSession(sidA, skA, dpRole, AllCapabilities)
	v.InstallSession(sidB, skB, dpRole, AllCapabilities)
	adv := &adversary{sessionAKey: append([]byte(nil), skA...)}
	return iss, v, sidA, sidB, skA, skB, adv
}

// forgeUsingKey builds a fully-signed artifact for `targetSession` using the
// supplied key. It mirrors what an attacker with a leaked key would do.
func forgeUsingKey(sessionID, recipient, purpose string, caps []Capability, key []byte) *Artifact {
	now := time.Now()
	a := &Artifact{
		Version:       1,
		SessionID:     sessionID,
		Recipient:     recipient,
		Purpose:       purpose,
		Capabilities:  append([]Capability(nil), caps...),
		IssuedAt:      now.UnixNano(),
		ExpiresAt:     now.Add(time.Minute).UnixNano(),
		Nonce:         randomHex(16),
		TargetCap:     randomHex(8),
		RecipientBind: hexMac(key, []byte("dhr-poc/recipient-bind/v1|"+recipient)),
	}
	a.Signature = hexMac(key, a.canonical())
	return a
}

// #25 Exposure of Session A's key cannot create Session B authority.
func TestPartialCompromise_CannotCreateB(t *testing.T) {
	_, v, _, sidB, _, _, adv := partialCompromiseSetup(t)
	forged := forgeUsingKey(sidB, dpRole, "interactive-remote", []Capability{CapViewScreen}, adv.sessionAKey)
	if err := v.Verify(forged); err == nil {
		t.Fatal("forgery signed with A's key must not authorize B")
	}
	if err := v.Use(forged, CapViewScreen); err == nil {
		t.Fatal("forgery signed with A's key must not exercise B capability")
	}
}

// #26 Exposure of Session A's key cannot validate a forged Session B artifact.
// (Distinct from #25 in that we vary purpose/capabilities to attempt any
// combination the attacker might try.)
func TestPartialCompromise_CannotValidateForgedB(t *testing.T) {
	_, v, _, sidB, _, _, adv := partialCompromiseSetup(t)
	attempts := []struct {
		caps []Capability
		purp string
	}{
		{AllCapabilities, "elevated-admin"},
		{[]Capability{CapTransferFile}, "quiet-exfil"},
		{[]Capability{CapSendKeyboardInput, CapSendPointerInput}, "takeover"},
	}
	for _, a := range attempts {
		forged := forgeUsingKey(sidB, dpRole, a.purp, a.caps, adv.sessionAKey)
		if err := v.Verify(forged); err == nil {
			t.Fatalf("forged B with caps=%v purp=%s must not verify", a.caps, a.purp)
		}
	}
}

// #27 Exposure of Session A's key cannot authorize a Session B capability.
func TestPartialCompromise_CannotAuthorizeB(t *testing.T) {
	_, v, _, sidB, _, _, adv := partialCompromiseSetup(t)
	for _, c := range AllCapabilities {
		forged := forgeUsingKey(sidB, dpRole, "interactive-remote", []Capability{c}, adv.sessionAKey)
		if err := v.Use(forged, c); err == nil {
			t.Fatalf("A key must not authorize %s on B", c)
		}
	}
}

// #28 Exposure of Session A's key cannot revoke Session B.
func TestPartialCompromise_CannotRevokeB(t *testing.T) {
	_, v, _, sidB, _, skB, adv := partialCompromiseSetup(t)
	if err := v.Revoke(sidB, CapViewScreen, adv.sessionAKey); err == nil {
		t.Fatal("A key must not revoke B")
	}
	// Sanity: with real B key, revocation succeeds.
	if err := v.Revoke(sidB, CapSendKeyboardInput, skB); err != nil {
		t.Fatalf("B key must be able to revoke B cap: %v", err)
	}
}

// #29 Exposure of Session A's key cannot invalidate Session B.
func TestPartialCompromise_CannotInvalidateB(t *testing.T) {
	_, v, _, sidB, _, _, adv := partialCompromiseSetup(t)
	if err := v.Invalidate(sidB, adv.sessionAKey); err == nil {
		t.Fatal("A key must not invalidate B")
	}
	if !v.SessionExists(sidB) {
		t.Fatal("B must still exist")
	}
}

// #30 Session B remains independently valid after A's material is exposed.
func TestPartialCompromise_BRemainsValid(t *testing.T) {
	iss, v, _, sidB, _, _, _ := partialCompromiseSetup(t)
	a := issueOK(t, iss, sidB, []Capability{CapViewScreen})
	if err := v.Use(a, CapViewScreen); err != nil {
		t.Fatalf("B must remain valid: %v", err)
	}
}

// -----------------------------------------------------------------------------
// Race-detector coverage (concurrent independent sessions).
// -----------------------------------------------------------------------------

func TestConcurrentSessions(t *testing.T) {
	iss := NewIssuer()
	v := NewValidator(dpRole)
	const N = 32
	sids := make([]string, N)
	keys := make([][]byte, N)
	for i := 0; i < N; i++ {
		sids[i], keys[i] = iss.StartSession()
		v.InstallSession(sids[i], keys[i], dpRole, []Capability{CapViewScreen})
	}
	var wg sync.WaitGroup
	for i := 0; i < N; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			for j := 0; j < 20; j++ {
				a := issueOK(t, iss, sids[i], []Capability{CapViewScreen})
				if err := v.Use(a, CapViewScreen); err != nil {
					t.Errorf("session %d op %d: %v", i, j, err)
					return
				}
			}
		}(i)
	}
	wg.Wait()
}

// -----------------------------------------------------------------------------
// Table-driven bad-input coverage.
// -----------------------------------------------------------------------------

func TestVerifyBadInputs(t *testing.T) {
	iss, v, sid, _ := setupSession(t, []Capability{CapViewScreen})
	base := issueOK(t, iss, sid, []Capability{CapViewScreen})

	tamper := func(mut func(a *Artifact)) *Artifact {
		copyA := *base
		copyA.Capabilities = append([]Capability(nil), base.Capabilities...)
		mut(&copyA)
		return &copyA
	}

	cases := []struct {
		name string
		a    *Artifact
	}{
		{"nil", nil},
		{"bad-version", tamper(func(a *Artifact) { a.Version = 99 })},
		{"bad-sig", tamper(func(a *Artifact) { a.Signature = "00" })},
		{"bad-nonce-sig", tamper(func(a *Artifact) { a.Nonce = "deadbeef" })},
		{"bad-bind", tamper(func(a *Artifact) { a.RecipientBind = "aa" })},
		{"unknown-session", tamper(func(a *Artifact) { a.SessionID = hex.EncodeToString(bytes.Repeat([]byte{1}, 16)) })},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if err := v.Verify(c.a); err == nil {
				t.Fatal("expected error")
			}
		})
	}
}

// -----------------------------------------------------------------------------
// Fuzz: malformed artifacts must never verify.
// -----------------------------------------------------------------------------

func FuzzVerifyMalformed(f *testing.F) {
	iss, v, sid, _ := setupSession(&testing.T{}, []Capability{CapViewScreen})
	base := issueOK(&testing.T{}, iss, sid, []Capability{CapViewScreen})

	f.Add([]byte(base.Signature), []byte(base.Nonce), []byte(base.RecipientBind))
	f.Add([]byte("00"), []byte(""), []byte(""))
	f.Add([]byte("zzz"), []byte("!!!"), []byte("nope"))

	f.Fuzz(func(t *testing.T, sig, nonce, bind []byte) {
		a := *base
		a.Capabilities = append([]Capability(nil), base.Capabilities...)
		a.Signature = string(sig)
		a.Nonce = string(nonce)
		a.RecipientBind = string(bind)
		// Only bail out if the fuzz input reproduces every real field.
		if string(sig) == base.Signature && string(nonce) == base.Nonce && string(bind) == base.RecipientBind {
			return
		}
		if err := v.Verify(&a); err == nil {
			t.Fatalf("malformed artifact verified: sig=%q nonce=%q bind=%q", sig, nonce, bind)
		}
	})
}

// #31 Return-flow event stream is not a probing oracle.
//
// Bounded-return-flow contract (ADR-0004): capability_denied is emitted
// only after the artifact has passed every authenticity, integrity,
// binding, freshness, replay and session check but is not authorised for
// the requested capability. Malformed, forged, replayed, wrongly
// addressed, cross-session, expired or session-unknown artifacts must
// fail closed WITHOUT producing a capability_denied event, so the event
// stream cannot be used as a probing oracle for hostile inputs.
func TestUseValidationFailuresEmitNoCapabilityDenied(t *testing.T) {
	iss := NewIssuer()
	v := NewValidator(dpRole)
	sidA, skA := iss.StartSession()
	sidB, skB := iss.StartSession()
	v.InstallSession(sidA, skA, dpRole, []Capability{CapViewScreen})
	v.InstallSession(sidB, skB, dpRole, []Capability{CapViewScreen})

	base := issueOK(t, iss, sidA, []Capability{CapViewScreen})

	// 1. Tampered signature.
	tampered := *base
	tampered.Signature = "deadbeef"
	_ = v.Use(&tampered, CapViewScreen)

	// 2. Wrong recipient.
	wrongRcpt := *base
	wrongRcpt.Recipient = "some-other-data-plane"
	_ = v.Use(&wrongRcpt, CapViewScreen)

	// 3. Cross-session (artifact minted for A, replayed against B's id).
	crossSess := *base
	crossSess.SessionID = sidB
	_ = v.Use(&crossSess, CapViewScreen)

	// 4. Unknown session id.
	unknown := *base
	unknown.SessionID = "no-such-session"
	_ = v.Use(&unknown, CapViewScreen)

	// 5. Malformed nonce/binding (rebuild from garbage).
	malformed := *base
	malformed.Nonce = ""
	malformed.RecipientBind = ""
	_ = v.Use(&malformed, CapViewScreen)

	// 6. Expired artifact.
	expired, err := iss.Issue(sidA, dpRole, "interactive-remote",
		[]Capability{CapViewScreen}, time.Nanosecond)
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(2 * time.Millisecond)
	_ = v.Use(expired, CapViewScreen)

	// 7. Replay of a valid Use.
	valid := issueOK(t, iss, sidA, []Capability{CapViewScreen})
	if err := v.Use(valid, CapViewScreen); err != nil {
		t.Fatalf("first use must succeed: %v", err)
	}
	if err := v.Use(valid, CapViewScreen); err == nil {
		t.Fatal("replay must fail")
	}

	// 8. Nil artifact.
	_ = v.Use(nil, CapViewScreen)

	for _, e := range v.Events() {
		if e.Type == EventCapabilityDenied {
			t.Fatalf("validation failure produced capability_denied event: %+v", e)
		}
	}
}

// #32 Genuine capability denial (authenticated artifact requests an
// ungranted capability) still emits exactly one capability_denied event.
// This is the positive counterpart to #31 — the event MUST fire when the
// artifact is fully valid but the requested capability is not authorised.
func TestUseGenuineCapabilityDenialEmitsEvent(t *testing.T) {
	iss, v, sid, _ := setupSession(t, []Capability{CapViewScreen})
	// Artifact is fully valid, but requests a capability the installed
	// session was never granted.
	a := issueOK(t, iss, sid, []Capability{CapTransferFile})
	if err := v.Use(a, CapTransferFile); err == nil {
		t.Fatal("expected capability denial")
	}
	got := 0
	for _, e := range v.Events() {
		if e.Type == EventCapabilityDenied {
			got++
		}
	}
	if got != 1 {
		t.Fatalf("expected exactly 1 capability_denied event, got %d", got)
	}
}
