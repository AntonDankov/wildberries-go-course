package handler

import (
	"fmt"
	"net/http"
	"runtime"
	"strings"
	"time"
	"wildberries-go-course/L4-4/entity"

	"github.com/wb-go/wbf/ginext"
)

type ChangeGCPercentageRequest struct {
	Percentage int `json:"percentage"`
}

func ChangeGarbageCollectorPercentage() ginext.HandlerFunc {
	return func(c *ginext.Context) {

		var changeRequestData ChangeGCPercentageRequest
		if err := c.ShouldBindJSON(&changeRequestData); err != nil {
			addJSONWithError(c, http.StatusBadRequest, fmt.Errorf("not valid request"))
			return
		}

		if changeRequestData.Percentage == 0 || changeRequestData.Percentage < -1 {
			addJSONWithError(c, http.StatusBadRequest,
				fmt.Errorf("percent must be -1 (disable GC) or a positive integer"),
			)
			return
		}
		entity.SetGCPercentage(changeRequestData.Percentage)
		c.String(http.StatusOK, "")
	}
}

func GetMemoryMetrics() ginext.HandlerFunc {
	return func(c *ginext.Context) {
		var memoryStats runtime.MemStats
		runtime.ReadMemStats(&memoryStats)
		var timeSinceLastGC float64
		if memoryStats.NumGC > 0 {
			timeSinceLastGC = float64(time.Now().UnixNano()-int64(memoryStats.LastGC)) / 1e9
		}

		totalAllocationsMetric := entity.Metric{
			Name:  "go_allocations_amount",
			Help:  "Amount of allocations",
			Type:  entity.MetricType_Counter,
			Value: float64(memoryStats.Mallocs),
		}

		garbageCollectionAmountMetric := entity.Metric{
			Name:  "go_garbage_collection_amount",
			Help:  "Amount of garbage collections was made",
			Type:  entity.MetricType_Counter,
			Value: float64(memoryStats.NumGC),
		}
		usedMemoryMetric := entity.Metric{
			Name:  "go_used_memory",
			Help:  "Amount of memory allocated by gc in bytes",
			Type:  entity.MetricType_Gauge,
			Value: float64(memoryStats.Alloc),
		}

		lastTimeGarbageCollectednMetric := entity.Metric{
			Name:  "go_last_time_garbage_collected",
			Help:  "Last time when the garbage collection was performed",
			Type:  entity.MetricType_Gauge,
			Value: float64(timeSinceLastGC),
		}

		totalPauseTimeInGCMetric := entity.Metric{
			Name:  "go_total_pause_time_gc",
			Help:  "Total pause time by garbage collector working in seconds",
			Type:  entity.MetricType_Gauge,
			Value: float64(memoryStats.PauseTotalNs) / 1e9,
		}
		currentGCPercentageMetric := entity.Metric{
			Name:  "go_gc_percentage",
			Help:  "Current Garbage collectior target in percentage",
			Type:  entity.MetricType_Gauge,
			Value: float64(entity.GetGCPercentage()),
		}

		metrics := []entity.Metric{totalAllocationsMetric, garbageCollectionAmountMetric, usedMemoryMetric,
			lastTimeGarbageCollectednMetric, totalPauseTimeInGCMetric, currentGCPercentageMetric}

		var sb strings.Builder
		for _, m := range metrics {
			sb.WriteString(m.String())
		}

		c.Header("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
		c.String(http.StatusOK, sb.String())

	}
}

func addJSONWithError(c *ginext.Context, httpCode int, err error) {
	c.JSON(httpCode, ginext.H{
		"error": err.Error(),
	})
}
