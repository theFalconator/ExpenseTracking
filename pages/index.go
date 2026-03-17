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
	Users       []UserDto
	Expenses    []ExpenseDto
	Err         error
}

type DebtDto struct {
	Owes   string
	Amount int
}

type UserDto struct {
	Id         int
	Name       string
	AmountOwed int
	TotalPaid  int
	Debts      []DebtDto
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

type Pair struct {
	Owed   int
	OwedBy int
}

func computeAmountOwedByUserId(expenses []activity.ExpenseRow) map[Pair]int {
	numParticipants := map[int]int{}
	dict := map[Pair]int{}

	for _, e := range expenses {
		numParticipants[e.ExpenseId] += 1
	}

	for _, e := range expenses {
		if e.PaidBy == e.ParticipantId {
			continue
		}

		normal := Pair{Owed: e.PaidBy, OwedBy: e.ParticipantId}
		inverse := Pair{OwedBy: e.PaidBy, Owed: e.ParticipantId}

		dict[normal] += e.AmountCentsUsd / numParticipants[e.ExpenseId]
		if dict[inverse] > dict[normal] {
			dict[inverse] -= dict[normal]
			dict[normal] = 0
		} else if dict[normal] > dict[inverse] {
			dict[normal] -= dict[inverse]
			dict[inverse] = 0
		}

		if dict[normal] == 0 {
			delete(dict, normal)
		}

		if dict[inverse] == 0 {
			delete(dict, inverse)
		}
	}

	return dict
}

func computeDebts(owed map[Pair]int, userList []users.User, usersById map[int]UserDto) {
	for i := range userList {
		u := userList[i]

		usersById[u.Id] = UserDto{
			Id:         u.Id,
			Name:       u.Name,
			AmountOwed: 0,
			TotalPaid:  0,
			Debts:      []DebtDto{},
		}
	}

	for k, v := range owed {
		user := usersById[k.OwedBy]

		user.Debts = append(user.Debts, DebtDto{
			Owes:   usersById[k.Owed].Name,
			Amount: v,
		})

		usersById[k.OwedBy] = user
	}
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
			Users:       []UserDto{},
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
	owed := computeAmountOwedByUserId(expenseList)

	usersById := map[int]UserDto{}
	computeDebts(owed, userList, usersById)

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

	dtos := []UserDto{}
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
