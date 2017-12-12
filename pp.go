package bpmon

type PersistenceProvider interface {
	GetOne([]string, string, []string, string) (map[string]interface{}, error)
	GetAll([]string, string, []string, string) ([]map[string]interface{}, error)
}
