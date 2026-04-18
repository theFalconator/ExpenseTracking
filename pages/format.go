package pages

import (
	"fmt"
	"time"
)

func FormatCents(cents int) string {
	dollars := cents / 100
	remainingCents := cents % 100
	return fmt.Sprintf("$%d.%02d", dollars, remainingCents)
}

func FormatCentsNumber(cents int) string {
	dollars := cents / 100
	remainingCents := cents % 100
	return fmt.Sprintf("%d.%02d", dollars, remainingCents)
}

func FormatTime(t time.Time) string {
	return t.Format("January 2, 2006")
}
