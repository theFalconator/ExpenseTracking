package pages

import (
	"fmt"
	"time"
)

// TODO: Why do I have two of these? Couldn't i just do <span>${{ value
// | formatCentsNumber }}</span>
// and then remove this fn?
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
