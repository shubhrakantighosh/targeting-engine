package delivery

import (
	"main/internal/delivery/service"
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
