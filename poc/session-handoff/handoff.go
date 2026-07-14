// Package handoff is the Session Handoff Contract PoC for Daang Remote.
//
// EXPERIMENTAL. This code is not the Daang Remote backend, not the production
// Control Plane, not the production Data Plane, not the permanent security
// core, not the final token format, not the final cryptographic design, and
// not the permanent Go architecture. It exists only to test whether the
// ADR-0003 and ADR-0004 handoff contract survives contact with executable,
// adversarial tests inside a small deterministic single-process construction.
//
// The cryptographic operations here (HMAC-SHA256 over a canonical JSON
// encoding, random nonces from crypto/rand, per-session subkeys derived from
// a global issuer key by domain-separated HMAC) are experimental, non-
// production, replaceable, and selected only to make the architectural
// property testable. They are not evidence of production security and not
// evidence of post-quantum security.
package handoff

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"sort"
	"sync"
	"time"
)

// Capability is one of the independent capability types the PoC models.
// No capability implies another. Deny by default.
type Capability string

const (
	CapViewScreen        Capability = "view_screen"
	CapSendKeyboardInput Capability = "send_keyboard_input"
	CapSendPointerInput  Capability = "send_pointer_input"
	CapUseClipboard      Capability = "use_clipboard"
	CapUseAudio          Capability = "use_audio"
	CapTransferFile      Capability = "transfer_file"
)

// AllCapabilities lists every capability the PoC models.
var AllCapabilities = []Capability{
	CapViewScreen,
	CapSendKeyboardInput,
	CapSendPointerInput,
	CapUseClipboard,
	CapUseAudio,
	CapTransferFile,
}

// EventType is a bounded lifecycle event on the return flow.
type EventType string

const (
	EventSessionStarted     EventType = "session_started"
	EventSessionEnded       EventType = "session_ended"
	EventCapabilityDenied   EventType = "capability_denied"
	EventCapabilityRevoked  EventType = "capability_revoked"
	EventSessionInvalidated EventType = "session_invalidated"
)

// Event carries only bounded lifecycle information scoped to one session.
// It contains no user identity, no account identity, no durable device
// identifier, no session content, and no content-derived material.
//
// This PoC does not solve event timing, order, frequency, or cross-session
// pattern metadata; those remain out of scope.
type Event struct {
	Type       EventType  `json:"type"`
	SessionID  string     `json:"session_id"`
	Capability Capability `json:"capability,omitempty"`
	At         time.Time  `json:"at"`
}

// Artifact is the session-scoped handoff artifact carried from the mock
// Control Plane to the mock Data Plane.
//
// It contains only the minimum fields required to test the accepted
// handoff contract. It intentionally does NOT carry user identity, account
// identity, email, username, phone number, durable device identifier,
// machine hostname, long-lived route or relay identifier, Control Plane
// internal correlation key, reusable login credential, password,
// authentication secret, or production key material.
type Artifact struct {
	Version       int          `json:"version"`
	SessionID     string       `json:"session_id"`
	Recipient     string       `json:"recipient"`
	Purpose       string       `json:"purpose"`
	Capabilities  []Capability `json:"capabilities"`
	IssuedAt      int64        `json:"issued_at_unix_nano"`
	ExpiresAt     int64        `json:"expires_at_unix_nano"`
	Nonce         string       `json:"nonce"`
	TargetCap     string       `json:"target_capability_handle"`
	RecipientBind string       `json:"recipient_binding"`
	Signature     string       `json:"signature"`
}

// canonical returns the deterministic byte sequence signed by the
// experimental HMAC. Signature and RecipientBind are excluded from the
// signature input (RecipientBind is bound separately).
func (a *Artifact) canonical() []byte {
	caps := append([]Capability(nil), a.Capabilities...)
	sort.Slice(caps, func(i, j int) bool { return caps[i] < caps[j] })
	aux := struct {
		V   int          `json:"v"`
		SID string       `json:"sid"`
		R   string       `json:"r"`
		P   string       `json:"p"`
		C   []Capability `json:"c"`
		IA  int64        `json:"ia"`
		EA  int64        `json:"ea"`
		N   string       `json:"n"`
		T   string       `json:"t"`
	}{a.Version, a.SessionID, a.Recipient, a.Purpose, caps, a.IssuedAt, a.ExpiresAt, a.Nonce, a.TargetCap}
	b, err := json.Marshal(aux)
	if err != nil {
		panic(fmt.Errorf("canonical marshal: %w", err))
	}
	return b
}

// prohibitedFieldSubstrings is the set of substrings that must never appear
// in an artifact's JSON field names or values. It backs the artifact-
// representation test (mandatory test 21).
var prohibitedFieldSubstrings = []string{
	"user_id", "account", "email", "username", "phone",
	"device_id", "hostname", "route_id", "relay_id",
	"correlation", "password", "credential", "auth_secret",
	"private_key",
}

// ProhibitedFieldsInArtifact returns the list of prohibited long-lived
// identity fields that appear in the artifact's JSON representation. An
// empty result means the artifact respects the identity-minimization rule.
func ProhibitedFieldsInArtifact(a *Artifact) []string {
	return scanProhibited(a)
}

// ProhibitedFieldsInEvent returns prohibited fields present in the event's
// JSON representation. Return-flow events must remain bounded to lifecycle
// state.
func ProhibitedFieldsInEvent(e Event) []string {
	return scanProhibited(e)
}

func scanProhibited(v interface{}) []string {
	b, err := json.Marshal(v)
	if err != nil {
		return []string{"unmarshalable"}
	}
	var m interface{}
	if err := json.Unmarshal(b, &m); err != nil {
		return []string{"unparseable"}
	}
	var hits []string
	walkJSON(m, func(key string) {
		for _, sub := range prohibitedFieldSubstrings {
			if containsFold(key, sub) {
				hits = append(hits, key)
			}
		}
	})
	// also check known field-name blocklist against reflected struct tags
	rt := reflect.TypeOf(v)
	if rt.Kind() == reflect.Ptr {
		rt = rt.Elem()
	}
	if rt.Kind() == reflect.Struct {
		for i := 0; i < rt.NumField(); i++ {
			name := rt.Field(i).Name
			for _, sub := range prohibitedFieldSubstrings {
				if containsFold(name, sub) {
					hits = append(hits, name)
				}
			}
		}
	}
	return hits
}

func walkJSON(v interface{}, keyFn func(string)) {
	switch t := v.(type) {
	case map[string]interface{}:
		for k, sub := range t {
			keyFn(k)
			walkJSON(sub, keyFn)
		}
	case []interface{}:
		for _, sub := range t {
			walkJSON(sub, keyFn)
		}
	}
}

func containsFold(s, sub string) bool {
	if len(sub) == 0 {
		return false
	}
	ls, lsub := lower(s), lower(sub)
	for i := 0; i+len(lsub) <= len(ls); i++ {
		if ls[i:i+len(lsub)] == lsub {
			return true
		}
	}
	return false
}

func lower(s string) string {
	b := []byte(s)
	for i, c := range b {
		if c >= 'A' && c <= 'Z' {
			b[i] = c + 32
		}
	}
	return string(b)
}

// Issuer models the Control Plane. It holds a global issuer key and derives
// per-session subkeys by domain-separated HMAC. The issuer key is the only
// piece of global authority; per-session material is scoped to its session.
type Issuer struct {
	mu         sync.Mutex
	issuerKey  []byte
	sessionKey map[string][]byte
	clock      func() time.Time
}

// NewIssuer allocates an Issuer with a fresh random 32-byte issuer key.
func NewIssuer() *Issuer {
	k := make([]byte, 32)
	if _, err := rand.Read(k); err != nil {
		panic(err)
	}
	return &Issuer{
		issuerKey:  k,
		sessionKey: map[string][]byte{},
		clock:      time.Now,
	}
}

// SetClock overrides the clock (for deterministic tests).
func (i *Issuer) SetClock(fn func() time.Time) {
	i.mu.Lock()
	defer i.mu.Unlock()
	i.clock = fn
}

// StartSession registers a new session and returns its opaque id and the
// per-session key that must be delivered out-of-band to the Data Plane
// validator. This PoC does not model the out-of-band delivery.
func (i *Issuer) StartSession() (sessionID string, sessionKey []byte) {
	i.mu.Lock()
	defer i.mu.Unlock()
	sessionID = randomHex(16)
	mac := hmac.New(sha256.New, i.issuerKey)
	mac.Write([]byte("dhr-poc/session-key/v1|"))
	mac.Write([]byte(sessionID))
	sessionKey = mac.Sum(nil)
	i.sessionKey[sessionID] = sessionKey
	return sessionID, sessionKey
}

// Issue produces a signed handoff artifact for the given session.
func (i *Issuer) Issue(sessionID, recipient, purpose string, caps []Capability, ttl time.Duration) (*Artifact, error) {
	i.mu.Lock()
	sk, ok := i.sessionKey[sessionID]
	clock := i.clock
	i.mu.Unlock()
	if !ok {
		return nil, errors.New("handoff: unknown session")
	}
	if recipient == "" {
		return nil, errors.New("handoff: recipient required")
	}
	if purpose == "" {
		return nil, errors.New("handoff: purpose required")
	}
	if ttl <= 0 {
		return nil, errors.New("handoff: ttl must be positive")
	}
	now := clock()
	a := &Artifact{
		Version:       1,
		SessionID:     sessionID,
		Recipient:     recipient,
		Purpose:       purpose,
		Capabilities:  append([]Capability(nil), caps...),
		IssuedAt:      now.UnixNano(),
		ExpiresAt:     now.Add(ttl).UnixNano(),
		Nonce:         randomHex(16),
		TargetCap:     randomHex(8),
		RecipientBind: hexMac(sk, []byte("dhr-poc/recipient-bind/v1|"+recipient)),
	}
	a.Signature = hexMac(sk, a.canonical())
	return a, nil
}

// NOTE: no accessor is provided for the per-session key. Tests obtain
// session-local material via StartSession's return value (see
// export_test.go for the test-only helper used by partial-compromise
// tests). Publishing a production-compilable accessor for session-local
// material would weaken the boundary the PoC is designed to test.

func hexMac(key, msg []byte) string {
	m := hmac.New(sha256.New, key)
	m.Write(msg)
	return hex.EncodeToString(m.Sum(nil))
}

func randomHex(n int) string {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return hex.EncodeToString(b)
}

// sessionState is the Data Plane's independent per-session state.
type sessionState struct {
	key         []byte
	recipient   string
	granted     map[Capability]bool
	revoked     map[Capability]bool
	invalidated bool
	seenNonce   map[string]bool
}

// Validator models the Data Plane. It performs every check independently
// of the Issuer.
type Validator struct {
	mu       sync.Mutex
	self     string
	sessions map[string]*sessionState
	events   []Event
	clock    func() time.Time
}

// NewValidator returns a validator identifying itself by role/recipient name.
func NewValidator(self string) *Validator {
	return &Validator{
		self:     self,
		sessions: map[string]*sessionState{},
		clock:    time.Now,
	}
}

// SetClock overrides the clock (for deterministic tests).
func (v *Validator) SetClock(fn func() time.Time) {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.clock = fn
}

// InstallSession is the out-of-band delivery of session-local material and
// the initial granted capability set. This PoC does not model transport of
// this data.
func (v *Validator) InstallSession(sessionID string, key []byte, recipient string, granted []Capability) {
	v.mu.Lock()
	defer v.mu.Unlock()
	g := map[Capability]bool{}
	for _, c := range granted {
		g[c] = true
	}
	v.sessions[sessionID] = &sessionState{
		key:       append([]byte(nil), key...),
		recipient: recipient,
		granted:   g,
		revoked:   map[Capability]bool{},
		seenNonce: map[string]bool{},
	}
	v.events = append(v.events, Event{
		Type: EventSessionStarted, SessionID: sessionID, At: v.clock(),
	})
}

// SessionExists reports whether the validator has state for a session id.
func (v *Validator) SessionExists(sessionID string) bool {
	v.mu.Lock()
	defer v.mu.Unlock()
	_, ok := v.sessions[sessionID]
	return ok
}

// Events returns a copy of the recorded return-flow events.
func (v *Validator) Events() []Event {
	v.mu.Lock()
	defer v.mu.Unlock()
	out := make([]Event, len(v.events))
	copy(out, v.events)
	return out
}

// Verify performs every structural, authenticity, session, and expiry
// check without consuming the nonce. It exists so callers can inspect
// artifact validity independently of capability use.
func (v *Validator) Verify(a *Artifact) error {
	v.mu.Lock()
	defer v.mu.Unlock()
	_, err := v.verifyLocked(a)
	return err
}

func (v *Validator) verifyLocked(a *Artifact) (*sessionState, error) {
	if a == nil {
		return nil, errors.New("handoff: nil artifact")
	}
	if a.Version != 1 {
		return nil, errors.New("handoff: unsupported version")
	}
	if a.Recipient != v.self {
		return nil, errors.New("handoff: recipient mismatch")
	}
	st, ok := v.sessions[a.SessionID]
	if !ok {
		return nil, errors.New("handoff: unknown session")
	}
	if st.invalidated {
		return nil, errors.New("handoff: session invalidated")
	}
	if a.Recipient != st.recipient {
		return nil, errors.New("handoff: recipient/session mismatch")
	}
	now := v.clock().UnixNano()
	// small clock skew tolerance forward
	if a.IssuedAt > now+int64(5*time.Second) {
		return nil, errors.New("handoff: issued in future")
	}
	if a.ExpiresAt <= now {
		return nil, errors.New("handoff: expired")
	}
	expectSig := hexMac(st.key, a.canonical())
	if !hmac.Equal([]byte(expectSig), []byte(a.Signature)) {
		return nil, errors.New("handoff: bad signature")
	}
	expectBind := hexMac(st.key, []byte("dhr-poc/recipient-bind/v1|"+a.Recipient))
	if !hmac.Equal([]byte(expectBind), []byte(a.RecipientBind)) {
		return nil, errors.New("handoff: bad recipient binding")
	}
	return st, nil
}

// Use asks the Data Plane to exercise `cap` under the authority of `a`.
// It runs every check, consumes the nonce, enforces revocation, and
// enforces that the artifact and installed session both grant `cap`.
//
// Event semantics (ADR-0004 bounded return-flow contract):
// capability_denied is emitted ONLY when the artifact has passed every
// authenticity, integrity, binding, freshness, replay and session check,
// but the requested capability is absent from the artifact, absent from
// the installed session grant set, or has been revoked. Validation
// failures (bad signature, wrong recipient, unknown/invalidated session,
// expired or future-dated artifact, replayed nonce, malformed artifact)
// fail closed and do NOT emit a return-flow event: the event stream must
// not become a probing oracle for malformed or hostile inputs.
func (v *Validator) Use(a *Artifact, cap Capability) error {
	v.mu.Lock()
	defer v.mu.Unlock()
	st, err := v.verifyLocked(a)
	if err != nil {
		return err
	}
	if st.seenNonce[a.Nonce] {
		return errors.New("handoff: replay")
	}
	artifactGrants := false
	for _, c := range a.Capabilities {
		if c == cap {
			artifactGrants = true
			break
		}
	}
	if !artifactGrants || !st.granted[cap] || st.revoked[cap] {
		v.events = append(v.events, Event{
			Type: EventCapabilityDenied, SessionID: a.SessionID,
			Capability: cap, At: v.clock(),
		})
		return errors.New("handoff: capability denied")
	}
	st.seenNonce[a.Nonce] = true
	return nil
}

// Revoke removes one capability from a session. It requires the caller to
// present the session-local material (proof) for that session. Exposure of
// another session's key does not authorize this operation.
func (v *Validator) Revoke(sessionID string, cap Capability, proof []byte) error {
	v.mu.Lock()
	defer v.mu.Unlock()
	st, ok := v.sessions[sessionID]
	if !ok {
		return errors.New("handoff: unknown session")
	}
	if len(proof) == 0 || !hmac.Equal(proof, st.key) {
		return errors.New("handoff: bad proof")
	}
	st.revoked[cap] = true
	v.events = append(v.events, Event{
		Type: EventCapabilityRevoked, SessionID: sessionID,
		Capability: cap, At: v.clock(),
	})
	return nil
}

// Invalidate disables every capability on a session. It requires the
// caller to present the session-local material (proof) for that session.
func (v *Validator) Invalidate(sessionID string, proof []byte) error {
	v.mu.Lock()
	defer v.mu.Unlock()
	st, ok := v.sessions[sessionID]
	if !ok {
		return errors.New("handoff: unknown session")
	}
	if len(proof) == 0 || !hmac.Equal(proof, st.key) {
		return errors.New("handoff: bad proof")
	}
	st.invalidated = true
	v.events = append(v.events, Event{
		Type: EventSessionInvalidated, SessionID: sessionID, At: v.clock(),
	})
	return nil
}

// EndSession records a bounded session_ended event.
func (v *Validator) EndSession(sessionID string, proof []byte) error {
	v.mu.Lock()
	defer v.mu.Unlock()
	st, ok := v.sessions[sessionID]
	if !ok {
		return errors.New("handoff: unknown session")
	}
	if len(proof) == 0 || !hmac.Equal(proof, st.key) {
		return errors.New("handoff: bad proof")
	}
	st.invalidated = true
	v.events = append(v.events, Event{
		Type: EventSessionEnded, SessionID: sessionID, At: v.clock(),
	})
	return nil
}
