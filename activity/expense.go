package activity

import (
	"time"
)

type Expense struct {
	Id             int
	Description    string
	AmountCentsUsd int
	PaidBy         int
	CreatedOn      time.Time
	GroupId        int
}

type ExpenseRow struct {
	ExpenseId      int
	Description    string
	AmountCentsUsd int
	PaidBy         int
	CreatedOn      time.Time
	ParticipantId  int
}

type Settlement struct {
	Id            int
	LastExpenseId int
	CreatedOn     time.Time
}
