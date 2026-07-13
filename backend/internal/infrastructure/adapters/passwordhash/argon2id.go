package passwordhash

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"math"
	"strings"

	"golang.org/x/crypto/argon2"
	gobcrypt "golang.org/x/crypto/bcrypt"

	"github.com/els/backend/internal/domain/shared/vo"
)

const (
	argon2idID  = "argon2id"
	argon2Pref  = "$" + argon2idID + "$"
	bcrypt2aPre = "$2a$"
	bcrypt2bPre = "$2b$"
	bcrypt2yPre = "$2y$"

	maxPasswordInputLen = 1024

	maxBcryptInputLen = 72
)

var (
	ErrPasswordTooLong  = errors.New("password exceeds max length")
	ErrUnknownHashFmt   = errors.New("unknown password hash format")
	ErrInvalidArgonHash = errors.New("invalid argon2id hash format")
	ErrUnsupportedArgon = errors.New("unsupported argon2 algorithm version")
	ErrInvalidParams    = errors.New("invalid argon2id parameters")
)

type Argon2idParams struct {
	Memory  uint32
	Time    uint32
	Threads uint8
	SaltLen uint32
	KeyLen  uint32
}

func DefaultArgon2idParams() Argon2idParams {
	return Argon2idParams{
		Memory:  64 * 1024,
		Time:    3,
		Threads: 2,
		SaltLen: 16,
		KeyLen:  32,
	}
}

func (p Argon2idParams) Validate() error {
	var errs []error
	if p.Memory < 8*1024 {
		errs = append(errs, fmt.Errorf("argon2id memory must be >= 8192 KiB, got %d", p.Memory))
	}
	if p.Time < 1 {
		errs = append(errs, errors.New("argon2id time must be >= 1"))
	}
	if p.Threads < 1 {
		errs = append(errs, errors.New("argon2id threads must be >= 1"))
	}
	if p.SaltLen < 8 {
		errs = append(errs, errors.New("argon2id salt length must be >= 8"))
	}
	if p.KeyLen < 16 {
		errs = append(errs, errors.New("argon2id key length must be >= 16"))
	}
	if len(errs) > 0 {
		return errors.Join(append([]error{ErrInvalidParams}, errs...)...)
	}
	return nil
}

type Hasher struct {
	params Argon2idParams
}

func NewArgon2id(p Argon2idParams) *Hasher {
	if err := p.Validate(); err != nil {
		p = DefaultArgon2idParams()
	}
	return &Hasher{params: p}
}

func (h *Hasher) Hash(plain string) (vo.PasswordHash, error) {
	if len(plain) > maxPasswordInputLen {
		return vo.PasswordHash{}, ErrPasswordTooLong
	}
	salt := make([]byte, h.params.SaltLen)
	if _, err := rand.Read(salt); err != nil {
		return vo.PasswordHash{}, fmt.Errorf("read salt: %w", err)
	}
	key := argon2.IDKey([]byte(plain), salt, h.params.Time, h.params.Memory, h.params.Threads, h.params.KeyLen)
	encoded := fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version,
		h.params.Memory, h.params.Time, h.params.Threads,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(key),
	)
	return vo.NewPasswordHash(encoded)
}

func (h *Hasher) Verify(hash vo.PasswordHash, plain string) error {
	if len(plain) > maxPasswordInputLen {
		return ErrPasswordTooLong
	}
	s := hash.String()
	switch {
	case strings.HasPrefix(s, argon2Pref):
		return verifyArgon2id(s, plain)
	case strings.HasPrefix(s, bcrypt2aPre), strings.HasPrefix(s, bcrypt2bPre), strings.HasPrefix(s, bcrypt2yPre):
		if len(plain) > maxBcryptInputLen {
			return ErrPasswordTooLong
		}
		return gobcrypt.CompareHashAndPassword([]byte(s), []byte(plain))
	default:
		return ErrUnknownHashFmt
	}
}

func verifyArgon2id(encoded, plain string) error {
	parts := strings.Split(encoded, "$")
	if len(parts) != 6 {
		return ErrInvalidArgonHash
	}
	if parts[1] != argon2idID {
		return ErrInvalidArgonHash
	}
	var version int
	if _, err := fmt.Sscanf(parts[2], "v=%d", &version); err != nil {
		return ErrInvalidArgonHash
	}
	if version != argon2.Version {
		return ErrUnsupportedArgon
	}
	var memory, time uint32
	var threads uint8
	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memory, &time, &threads); err != nil {
		return ErrInvalidArgonHash
	}
	salt, err := base64.RawStdEncoding.Strict().DecodeString(parts[4])
	if err != nil {
		return ErrInvalidArgonHash
	}
	want, err := base64.RawStdEncoding.Strict().DecodeString(parts[5])
	if err != nil {
		return ErrInvalidArgonHash
	}
	if len(want) == 0 || len(want) > math.MaxUint32 {
		return ErrInvalidArgonHash
	}
	wantLen := uint32(len(want)) // #nosec G115 -- bounded by MaxUint32 check above
	got := argon2.IDKey([]byte(plain), salt, time, memory, threads, wantLen)
	if subtle.ConstantTimeCompare(want, got) != 1 {
		return errors.New("argon2id: mismatch")
	}
	return nil
}
