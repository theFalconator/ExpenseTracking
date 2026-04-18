package pages

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/theFalconator/expenses/activity"
	"github.com/theFalconator/expenses/groups"
	"github.com/theFalconator/expenses/users"
)

type IndexPageHandler struct {
	userStore     *users.Store
	activityStore *activity.Store
	groupStore    *groups.Store
}

type IndexPageModel struct {
	ActiveGroup int
	Groups      []groups.Group
	Users       []activity.UserDto
	Expenses    []ExpenseDto
	Err         error
}


type ExpenseDto struct {
	Id             int
	Description    string
	AmountCentsUsd int
	PaidBy         int
	PaidByName     string
	CreatedOn      time.Time
}

func NewIndexPageHandler(s *users.Store, a *activity.Store, g *groups.Store) *IndexPageHandler {
	return &IndexPageHandler{
		userStore:     s,
		activityStore: a,
		groupStore:    g,
	}
}

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


func (h *IndexPageHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.New("index.html").Funcs(template.FuncMap{
		"formatCents":       FormatCents,
		"formatTime":        FormatTime,
		"formatCentsNumber": FormatCentsNumber,
	}).ParseFiles("templates/index.html")

	if err != nil {
		log.Fatal(err)
	}

	groupList, err := h.groupStore.ListGroups()
	// failure is okay here, just means we don't have a route param and 0 should be set as ActiveGroup
	activeGroup, _ := strconv.Atoi(r.PathValue("id"))

	if activeGroup == 0 {
		model := IndexPageModel{
			ActiveGroup: activeGroup,
			Expenses:    []ExpenseDto{},
			Err:         err,
			Groups:      groupList,
			Users:       []activity.UserDto{},
		}

		tmpl.Execute(w, model)
		return
	}

	userList, err := h.userStore.ListUsersInGroup(activeGroup)
	if err != nil {
		model := IndexPageModel{
			ActiveGroup: activeGroup,
			Err:         err,
		}

		tmpl.Execute(w, model)
		return

	}
	settlment, err := h.activityStore.GetLatestSettlementForGroup(activeGroup)
	expenseList, err := h.activityStore.ListExpensesForGroup(settlment.LastExpenseId, activeGroup)
	owed := activity.ComputeAmountOwedByUserId(expenseList)

	usersById := map[int]activity.UserDto{}
	activity.ComputeDebts(owed, userList, usersById)

	expenses := []ExpenseDto{}
	for i, v := range expenseList {
		// relies on returning a sorted slice -- make sure query has "order by Expense.Id" in it
		if i > 0 && v.ExpenseId == expenseList[i-1].ExpenseId {
			continue
		}
		expenses = append(expenses, ExpenseDto{
			Id:             v.ExpenseId,
			Description:    v.Description,
			AmountCentsUsd: v.AmountCentsUsd,
			PaidBy:         v.PaidBy,
			PaidByName:     usersById[v.PaidBy].Name,
			CreatedOn:      v.CreatedOn,
		})

		entry := usersById[v.PaidBy]
		entry.AmountOwed += v.AmountCentsUsd
		usersById[v.PaidBy] = entry
	}

	dtos := []activity.UserDto{}
	for _, v := range usersById {
		dtos = append(dtos, v)
	}

	sort.Slice(expenses, func(a,b int) bool {
		return expenses[a].CreatedOn.After(expenses[b].CreatedOn)
	})

	model := IndexPageModel{
		ActiveGroup: activeGroup,
		Groups:      groupList,
		Expenses:    expenses,
		Users:       dtos,
		Err:         err,
	}

	tmpl.Execute(w, model)
}
