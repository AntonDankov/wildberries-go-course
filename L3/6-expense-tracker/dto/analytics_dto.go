package dto

import (
	"strconv"
	"time"
	"wildberries-go-course/L3-6/model"

	"github.com/wb-go/wbf/ginext"
)

type AnalyticsFilter struct {
	From     time.Time             `json:"from"`
	To       time.Time             `json:"to"`
	Type     *model.RecordType     `json:"type"`
	Category *model.RecordCategory `json:"category"`
}

type Analytics struct {
	Sum        float64 `json:"sum"`
	Average    float64 `json:"average"`
	Count      int64   `json:"count"`
	Median     float64 `json:"median"`
	Percentile float64 `json:"percentile"`
}

func ParseFilter(c *ginext.Context) *AnalyticsFilter {
	filter := &AnalyticsFilter{}

	if fromStr := c.Query("from"); fromStr != "" {
		if from, err := time.Parse("2006-01-02", fromStr); err == nil {
			filter.From = from
		}
	}

	if toStr := c.Query("to"); toStr != "" {
		if to, err := time.Parse("2006-01-02", toStr); err == nil {
			filter.To = to
		}
	}

	if typeStr := c.Query("type"); typeStr != "" {
		if typeVal, err := strconv.Atoi(typeStr); err == nil {
			recordType := model.RecordType(typeVal)
			filter.Type = &recordType
		}
	}

	if categoryStr := c.Query("category"); categoryStr != "" {
		if categoryVal, err := strconv.Atoi(categoryStr); err == nil {
			recordCategory := model.RecordCategory(categoryVal)
			filter.Category = &recordCategory
		}
	}

	return filter
}
