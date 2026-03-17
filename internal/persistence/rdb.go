package persistence

type Snapshotter interface {
	Save(path string) error
	Load(path string) error
}
