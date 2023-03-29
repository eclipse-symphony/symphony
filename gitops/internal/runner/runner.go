package runner

type Runner interface {
	GetId() string
	Start()
	Stop()
	Done() <-chan struct{}
}
