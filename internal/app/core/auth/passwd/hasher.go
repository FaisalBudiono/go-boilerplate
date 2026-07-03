package passwd

type Hasher interface {
	Hash(password string) (string, error)
	Verify(password, hash string) (match bool, err error)
}
