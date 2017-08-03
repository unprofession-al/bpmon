package bpmon

type PersistenceProvider interface {
	GetOne(string) (interface{}, error)
}
