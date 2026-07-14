package handoff

// This file is test-only (build-tagged by the _test.go suffix): its
// symbols are compiled into the handoff test binary and are NOT part of
// the production package surface.
//
// The partial-compromise tests need read access to Session A's
// session-local material to model an adversary that has learned it.
// StartSession already returns that material to its caller, so tests
// obtain skA/skB directly from StartSession and do not need a broader
// accessor.
//
// testSessionKey exists only as a documented, package-internal, test-only
// path to per-session material — used when a test receives an *Issuer but
// not the original StartSession return values. It is deliberately kept
// inside a _test.go file so production callers cannot link against it.
func testSessionKey(i *Issuer, sessionID string) []byte {
	i.mu.Lock()
	defer i.mu.Unlock()
	sk, ok := i.sessionKey[sessionID]
	if !ok {
		return nil
	}
	return append([]byte(nil), sk...)
}
