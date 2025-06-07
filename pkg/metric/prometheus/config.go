package prometheus

import (
	"github.com/prometheus/client_golang/prometheus"
)

type MetricsCollector struct {
	RequestCounter *prometheus.CounterVec
}

var (
	instance *MetricsCollector
)

func GetMetrics() *MetricsCollector {
	return instance
}

func SetMetrics(requestCounter *prometheus.CounterVec) {
	instance = &MetricsCollector{
		RequestCounter: requestCounter,
	}
}
