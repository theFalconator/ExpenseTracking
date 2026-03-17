package pages

import (
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/theFalconator/expenses/activity"
	"github.com/theFalconator/expenses/users"
)

type AddExpenseHandler struct {
	userStore     *users.Store
	activityStore *activity.Store
}

type AddExpenseModel struct {
	Users                []users.User
	ExpenseAutocompletes []string
	Date                 string
	Description          string
	Amount               string
	Errors               map[string]error
}

func NewAddExpenseHandler(u *users.Store, a *activity.Store) *AddExpenseHandler {
	return &AddExpenseHandler{
		userStore:     u,
		activityStore: a,
	}
}

func (h *AddExpenseHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	errors := map[string]error{}
	tmpl, err := template.New("add-expense.html").ParseFiles("templates/add-expense.html")
	if err != nil {
		errors["template"] = err
	}
	group, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		errors["path"] = err
	}
	users, err := h.userStore.ListUsersInGroup(group)
	if err != nil {
		errors["users"] = err
	}
	autocompletes, err := h.activityStore.ListExpenseDescriptions()
	if err != nil {
		errors["autocompletes"] = err
	}

	if r.Method == http.MethodPost {
		r.ParseForm()
		slog.Info("Form submitted", "Paid By", r.FormValue("paidBy"), "Description", r.FormValue("desc"), "Amount", r.FormValue("amount"))

		model := AddExpenseModel{
			Users:                users,
			ExpenseAutocompletes: autocompletes,
			Date:                 r.FormValue("date"),
			Description:          r.FormValue("desc"),
			Amount:               r.FormValue("amount"),
			Errors:               errors,
		}

		paidBy, err := strconv.ParseInt(r.FormValue("paidBy"), 10, 64)
		if err != nil {
			errors["paidBy"] = err
		}

		amountF, err := strconv.ParseFloat(r.FormValue("amount"), 64)
		if err != nil {
			errors["amount"] = err
		}
		amount := int(amountF * 100)

		date, err := time.Parse("2006-01-02", r.FormValue("date"))
		if err != nil {
			errors["date"] = err
		}

		participants := []int{}
		for _, u := range users {
			if r.FormValue(fmt.Sprintf("chk-%d", u.Id)) != "" {
				participants = append(participants, u.Id)
			}
		}

		if len(participants) == 0 {
			errors["participants"] = fmt.Errorf("No participants selected")
			w.WriteHeader(http.StatusBadRequest)
			tmpl.Execute(w, model)
			return
		}

		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			tmpl.Execute(w, model)
			return
		}

		h.activityStore.AddExpense(group, int(paidBy), r.FormValue("desc"), amount, date, participants)
		http.Redirect(w, r, fmt.Sprintf("/%d", group), http.StatusSeeOther)
	} else {
		model := AddExpenseModel{
			Users:                users,
			ExpenseAutocompletes: autocompletes,
			Date:                 time.Now().Format("2006-01-02"),
			Errors:               errors,
		}
		tmpl.Execute(w, model)
	}
}
