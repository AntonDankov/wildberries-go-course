package entity

import (
	"fmt"
	"runtime/debug"
	"sync"
)

type MetricType uint8

const (
	MetricType_Gauge MetricType = iota
	MetricType_Counter
)

type Metric struct {
	Name  string
	Help  string
	Type  MetricType
	Value float64
}

func (t MetricType) String() string {
	switch t {
	case MetricType_Gauge:
		return "gauge"
	case MetricType_Counter:
		return "counter"
	default:
		return "untyped"
	}
}

func (metric Metric) String() string {
	return fmt.Sprintf(
		"# HELP %s %s\n"+
			"# TYPE %s %s\n"+
			"%s %g\n\n",
		metric.Name, metric.Help,
		metric.Name, metric.Type,
		metric.Name, metric.Value,
	)
}

var (
	gcMutex             sync.RWMutex
	currentGCPercentage int
)

func SetGCPercentage(percentage int) {
	gcMutex.Lock()
	defer gcMutex.Unlock()

	currentGCPercentage = percentage
	debug.SetGCPercent(percentage)
}

func GetGCPercentage() int {
	gcMutex.RLock()
	defer gcMutex.RUnlock()

	return currentGCPercentage
}
