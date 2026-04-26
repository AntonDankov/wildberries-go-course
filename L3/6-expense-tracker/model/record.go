package model

import (
	"time"
)

type RecordType int

const (
	Expense RecordType = iota
	Income
	RecordTypeEnd
)

type RecordCategory int

const (
	Electronics RecordCategory = iota
	Food
	Delivery
	Taxes
	RecordCategoryEnd
)

type Record struct {
	ID       int64
	Type     RecordType
	Category RecordCategory
	Amount   float64
	Date     time.Time
}
