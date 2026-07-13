package passwordhash

import (
	"errors"
	"strings"
	"testing"

	gobcrypt "golang.org/x/crypto/bcrypt"

	"github.com/els/backend/internal/domain/shared/vo"
)

func testParams() Argon2idParams {
	return Argon2idParams{Memory: 8 * 1024, Time: 1, Threads: 1, SaltLen: 16, KeyLen: 32}
}

func TestArgon2id_HashAndVerify_RoundTrip(t *testing.T) {
	t.Parallel()
	h := NewArgon2id(testParams())
	hash, err := h.Hash("Str0ng!Pass")
	if err != nil {
		t.Fatalf("Hash: %v", err)
	}
	if !strings.HasPrefix(hash.String(), "$argon2id$") {
		t.Fatalf("expected argon2id prefix, got %q", hash.String())
	}
	if err := h.Verify(hash, "Str0ng!Pass"); err != nil {
		t.Fatalf("Verify: %v", err)
	}
}

func TestArgon2id_VerifyRejectsWrongPassword(t *testing.T) {
	t.Parallel()
	h := NewArgon2id(testParams())
	hash, err := h.Hash("good")
	if err != nil {
		t.Fatal(err)
	}
	if err := h.Verify(hash, "bad"); err == nil {
		t.Fatal("Verify: must reject wrong password")
	}
}

func TestArgon2id_RejectsOverMaxLen(t *testing.T) {
	t.Parallel()
	h := NewArgon2id(testParams())
	if _, err := h.Hash(strings.Repeat("a", maxPasswordInputLen+1)); !errors.Is(err, ErrPasswordTooLong) {
		t.Fatalf("Hash: expected ErrPasswordTooLong, got %v", err)
	}
	hash, err := h.Hash("ok")
	if err != nil {
		t.Fatal(err)
	}
	if err := h.Verify(hash, strings.Repeat("a", maxPasswordInputLen+1)); !errors.Is(err, ErrPasswordTooLong) {
		t.Fatalf("Verify: expected ErrPasswordTooLong, got %v", err)
	}
}

func TestArgon2id_VerifyBcryptLegacyHash(t *testing.T) {
	t.Parallel()
	bc, err := gobcrypt.GenerateFromPassword([]byte("Str0ng!Pass"), gobcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}
	hash, err := vo.NewPasswordHash(string(bc))
	if err != nil {
		t.Fatal(err)
	}
	h := NewArgon2id(testParams())
	if err := h.Verify(hash, "Str0ng!Pass"); err != nil {
		t.Fatalf("Verify(bcrypt legacy): %v", err)
	}
	if err := h.Verify(hash, "wrong"); err == nil {
		t.Fatal("Verify(bcrypt legacy) must reject wrong password")
	}
}

func TestArgon2id_VerifyRejectsBcryptOver72Bytes(t *testing.T) {
	t.Parallel()
	bc, err := gobcrypt.GenerateFromPassword([]byte("ok"), gobcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}
	hash, err := vo.NewPasswordHash(string(bc))
	if err != nil {
		t.Fatal(err)
	}
	h := NewArgon2id(testParams())
	if err := h.Verify(hash, strings.Repeat("a", 73)); !errors.Is(err, ErrPasswordTooLong) {
		t.Fatalf("expected ErrPasswordTooLong, got %v", err)
	}
}

func TestArgon2id_VerifyUnknownFormat(t *testing.T) {
	t.Parallel()
	hash, err := vo.NewPasswordHash("not-a-real-hash")
	if err != nil {
		t.Fatal(err)
	}
	h := NewArgon2id(testParams())
	if err := h.Verify(hash, "x"); !errors.Is(err, ErrUnknownHashFmt) {
		t.Fatalf("expected ErrUnknownHashFmt, got %v", err)
	}
}

func TestArgon2idParams_Validate(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		p    Argon2idParams
		ok   bool
	}{
		{"defaults", DefaultArgon2idParams(), true},
		{"low memory", Argon2idParams{Memory: 1024, Time: 1, Threads: 1, SaltLen: 16, KeyLen: 32}, false},
		{"zero time", Argon2idParams{Memory: 64 * 1024, Time: 0, Threads: 1, SaltLen: 16, KeyLen: 32}, false},
		{"short salt", Argon2idParams{Memory: 64 * 1024, Time: 1, Threads: 1, SaltLen: 4, KeyLen: 32}, false},
		{"short key", Argon2idParams{Memory: 64 * 1024, Time: 1, Threads: 1, SaltLen: 16, KeyLen: 8}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.p.Validate()
			if tc.ok && err != nil {
				t.Fatalf("expected ok, got %v", err)
			}
			if !tc.ok && err == nil {
				t.Fatal("expected error")
			}
		})
	}
}
