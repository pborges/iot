package process

type OnMessageFn func(MetaData) error

type MetaData struct {
	Process string
	Key     string
	Value   interface{}
}
