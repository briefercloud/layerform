package terraform

type Client interface {
	Init(dir string) error
	Apply(dir string, state *State) (*State, error)
	Destroy(dir string, state *State, target ...string) (*State, error)
}
