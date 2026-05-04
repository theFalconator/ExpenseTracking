package pages

import (
	"fmt"
	"html/template"
	"log"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/theFalconator/expenses/activity"
	"github.com/theFalconator/expenses/users"
)

type SettlementHandler struct {
	userStore     *users.Store
	activityStore *activity.Store
}

func NewSettlementHandler(u *users.Store, a *activity.Store) *SettlementHandler {
	return &SettlementHandler{
		userStore:     u,
		activityStore: a,
	}
}

type SettlmentModel struct {
	Err   error
	Users []activity.UserDto
}

func (h *SettlementHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	group, err := strconv.Atoi(r.PathValue("id"))

	if r.Method == http.MethodPost {
		err := h.activityStore.RecordSettlementForGroup(group)
		if err != nil {
			log.Fatal(err)
		}

		http.Redirect(w, r, fmt.Sprintf("/%d", group), http.StatusSeeOther)
		return
	}

	tmpl, err := template.New("settle.html").Funcs(template.FuncMap{
		"formatCents":       FormatCents,
		"formatCentsNumber": FormatCentsNumber,
		"formatTime":        FormatTime,
	}).ParseFiles("templates/settle.html")

	if err != nil {
		http.Error(w, "Could not execute template", 500)
	}

	users, err := h.userStore.ListUsersInGroup(group)
	settlment, err := h.activityStore.GetLatestSettlementForGroup(group)
	expenseList, err := h.activityStore.ListExpensesForGroup(settlment.LastExpenseId, group)

	owed := activity.ComputeAmountOwedByUserId(expenseList)
	usersById := map[int]activity.UserDto{}
	activity.ComputeDebts(owed, users, usersById)
	for i, v := range expenseList {
		// relies on returning a sorted slice -- make sure query has "order by Expense.Id" in it
		if i > 0 && v.ExpenseId == expenseList[i-1].ExpenseId {
			continue
		}

		entry := usersById[v.PaidBy]
		entry.AmountOwed += v.AmountCentsUsd
		usersById[v.PaidBy] = entry
	}

	dtos := []activity.UserDto{}
	for _, v := range usersById {
		for _, d := range v.Debts {
			v.TotalDebts += d.Amount
		}
		dtos = append(dtos, v)
	}

	model := SettlmentModel{
		Err:   err,
		Users: dtos,
	}

	err = tmpl.Execute(w, model)
	if err != nil {
		slog.Error("Error executing template", "error", err)
	}

}
