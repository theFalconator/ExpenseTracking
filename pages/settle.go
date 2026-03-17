package pages

import (
	"fmt"
	"html/template"
	"log"
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
	Users []UserDto
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
		log.Fatal(err)
	}

	users, err := h.userStore.ListUsersInGroup(group)
	settlment, err := h.activityStore.GetLatestSettlementForGroup(group)
	expenseList, err := h.activityStore.ListExpensesForGroup(settlment.LastExpenseId, group)

	owed := computeAmountOwedByUserId(expenseList)
	usersById := map[int]UserDto{}
	computeDebts(owed, users, usersById)
	for i, v := range expenseList {
		// relies on returning a sorted slice -- make sure query has "order by Expense.Id" in it
		if i > 0 && v.ExpenseId == expenseList[i-1].ExpenseId {
			continue
		}

		entry := usersById[v.PaidBy]
		entry.AmountOwed += v.AmountCentsUsd
		usersById[v.PaidBy] = entry
	}

	dtos := []UserDto{}
	for _, v := range usersById {
		dtos = append(dtos, v)
	}

	model := SettlmentModel{
		Err:   err,
		Users: dtos,
	}

	tmpl.Execute(w, model)
}
