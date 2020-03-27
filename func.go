package iot

type Function struct {
	Name string
	Args []FunctionArg
}
type FunctionArg struct {
	Name string
	Type string
}
