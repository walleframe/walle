package app

type NoopService struct {
}

func (svc *NoopService) Name() string {
	return "noop"
}

func (svc *NoopService) Init(Stoper) (err error) {
	return
}

func (svc *NoopService) Start(Stoper) (err error) {
	return
}

func (svc *NoopService) Stop() {
	return
}

func (svc *NoopService) Finish() {
	return
}
