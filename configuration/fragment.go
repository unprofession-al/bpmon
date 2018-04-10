package configuration

type Fragment interface {
	Validate() (error, []string)
}
