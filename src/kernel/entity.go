package kernel

type EntityInterface interface {
	init()
}

type Entity struct {
	Name     string
	Timeline *Timeline
	EntityInterface
}
