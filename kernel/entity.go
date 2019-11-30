package kernel

type EntityInterface interface {
	init()
}

type Entity struct {
	name     string
	timeline *Timeline
	EntityInterface
}
