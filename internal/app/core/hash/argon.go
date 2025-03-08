package hash

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"github.com/ztrue/tracerr"
	"golang.org/x/crypto/argon2"
)

var (
	ErrInvalidHash         = errors.New("the encoded hash is not in the correct format")
	ErrIncompatibleVersion = errors.New("incompatible version of argon2")
)

type argonParams struct {
	memory      uint32
	iterations  uint32
	parallelism uint8
	saltLength  uint32
	keyLength   uint32
}

func NewArgon() *argonParams {
	return &argonParams{
		memory:      64 * 1024,
		iterations:  3,
		parallelism: 2,
		saltLength:  16,
		keyLength:   32,
	}
}

func (ap *argonParams) Generate(password string) (string, error) {
	// Generate a cryptographically secure random salt.
	salt, err := generateRandomBytes(ap.saltLength)
	if err != nil {
		return "", err
	}

	// Pass the plaintext password, salt and parameters to the argon2.IDKey
	// function. This will generate a hash of the password using the Argon2id
	// variant.
	hash := argon2.IDKey(
		[]byte(password),
		salt,
		ap.iterations,
		ap.memory,
		ap.parallelism,
		ap.keyLength,
	)

	// Base64 encode the salt and hashed password.
	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	// Return a string using the standard encoded hash representation.
	encodedHash := fmt.Sprintf(
		"$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version,
		ap.memory,
		ap.iterations,
		ap.parallelism,
		b64Salt,
		b64Hash,
	)

	return encodedHash, nil
}

func (ap *argonParams) Verify(password, encodedHash string) (match bool, err error) {
	// Extract the parameters, salt and derived key from the encoded password
	// hash.
	p, salt, hash, err := decodeHash(encodedHash)
	if err != nil {
		return false, err
	}

	// Derive the key from the other password using the same parameters.
	otherHash := argon2.IDKey(
		[]byte(password),
		salt,
		p.iterations,
		p.memory,
		p.parallelism,
		p.keyLength,
	)

	// Check that the contents of the hashed passwords are identical. Note
	// that we are using the subtle.ConstantTimeCompare() function for this
	// to help prevent timing attacks.
	if subtle.ConstantTimeCompare(hash, otherHash) == 1 {
		return true, nil
	}

	return false, nil
}

func decodeHash(encodedHash string) (p *argonParams, salt, hash []byte, err error) {
	vals := strings.Split(encodedHash, "$")
	if len(vals) != 6 {
		return nil, nil, nil, tracerr.Wrap(ErrInvalidHash)
	}

	var version int
	_, err = fmt.Sscanf(vals[2], "v=%d", &version)
	if err != nil {
		return nil, nil, nil, tracerr.Wrap(err)
	}
	if version != argon2.Version {
		return nil, nil, nil, tracerr.Wrap(ErrIncompatibleVersion)
	}

	p = &argonParams{}
	_, err = fmt.Sscanf(vals[3], "m=%d,t=%d,p=%d", &p.memory, &p.iterations, &p.parallelism)
	if err != nil {
		return nil, nil, nil, tracerr.Wrap(err)
	}

	salt, err = base64.RawStdEncoding.Strict().DecodeString(vals[4])
	if err != nil {
		return nil, nil, nil, tracerr.Wrap(err)
	}
	p.saltLength = uint32(len(salt))

	hash, err = base64.RawStdEncoding.Strict().DecodeString(vals[5])
	if err != nil {
		return nil, nil, nil, tracerr.Wrap(err)
	}
	p.keyLength = uint32(len(hash))

	return p, salt, hash, nil
}

func generateRandomBytes(n uint32) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return nil, tracerr.Wrap(err)
	}

	return b, nil
}
