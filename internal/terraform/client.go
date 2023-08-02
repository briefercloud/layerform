package terraform

type Client interface {
	Init(dir string) error
	Apply(dir string, state []byte) ([]byte, error)
	Destroy(dir string, state []byte) ([]byte, error)
}
