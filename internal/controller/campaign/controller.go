package campaign

import (
	"main/internal/campaign/service"
	"sync"
)

type Controller struct {
	service service.Service
}

var (
	syneOnce sync.Once
	ctrl     *Controller
)

func NewController(service *service.Service) *Controller {
	syneOnce.Do(func() {
		ctrl = &Controller{
			service: *service,
		}
	})

	return ctrl
}

type Interface interface {
}
