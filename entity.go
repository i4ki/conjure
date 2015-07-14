package conjure

type Entity interface {
	Pull() error
	Create() error
	Run() error
	Start() error
	Stop() error
	WaitOK() error
}
