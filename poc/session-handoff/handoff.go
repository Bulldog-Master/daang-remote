// Package handoff — TDD red-phase stub.
//
// This file intentionally contains no behavior. The mandatory tests in
// handoff_test.go are expected to fail against this stub. The subsequent
// commit replaces this file with the real experimental implementation.
package handoff

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"sync"
	"time"
)

type Capability string

const (
	CapViewScreen        Capability = "view_screen"
	CapSendKeyboardInput Capability = "send_keyboard_input"
	CapSendPointerInput  Capability = "send_pointer_input"
	CapUseClipboard      Capability = "use_clipboard"
	CapUseAudio          Capability = "use_audio"
	CapTransferFile      Capability = "transfer_file"
)

var AllCapabilities = []Capability{
	CapViewScreen, CapSendKeyboardInput, CapSendPointerInput,
	CapUseClipboard, CapUseAudio, CapTransferFile,
}

type EventType string

const (
	EventSessionStarted     EventType = "session_started"
	EventSessionEnded       EventType = "session_ended"
	EventCapabilityDenied   EventType = "capability_denied"
	EventCapabilityRevoked  EventType = "capability_revoked"
	EventSessionInvalidated EventType = "session_invalidated"
)

type Event struct {
	Type       EventType  `json:"type"`
	SessionID  string     `json:"session_id"`
	Capability Capability `json:"capability,omitempty"`
	At         time.Time  `json:"at"`
}

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

func (a *Artifact) canonical() []byte { return nil }

func ProhibitedFieldsInArtifact(*Artifact) []string { return nil }
func ProhibitedFieldsInEvent(Event) []string        { return nil }

type Issuer struct{ mu sync.Mutex }

func NewIssuer() *Issuer                         { return &Issuer{} }
func (i *Issuer) SetClock(func() time.Time)      {}
func (i *Issuer) StartSession() (string, []byte) { return "stub", []byte{0} }
func (i *Issuer) SessionKey(string) []byte       { return nil }
func (i *Issuer) Issue(_, _, _ string, _ []Capability, _ time.Duration) (*Artifact, error) {
	return nil, errors.New("unimplemented")
}

type Validator struct{ mu sync.Mutex }

func NewValidator(string) *Validator                                     { return &Validator{} }
func (v *Validator) SetClock(func() time.Time)                           {}
func (v *Validator) InstallSession(string, []byte, string, []Capability) {}
func (v *Validator) SessionExists(string) bool                           { return false }
func (v *Validator) Events() []Event                                     { return nil }
func (v *Validator) Verify(*Artifact) error                              { return nil }
func (v *Validator) Use(*Artifact, Capability) error                     { return nil }
func (v *Validator) Revoke(string, Capability, []byte) error             { return nil }
func (v *Validator) Invalidate(string, []byte) error                     { return nil }
func (v *Validator) EndSession(string, []byte) error                     { return nil }

func randomHex(n int) string {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
func hexMac(key, msg []byte) string {
	m := hmac.New(sha256.New, key)
	m.Write(msg)
	return hex.EncodeToString(m.Sum(nil))
}
