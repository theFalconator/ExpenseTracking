package pages

import (
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/theFalconator/expenses/activity"
	"github.com/theFalconator/expenses/users"
)

type UploadSelectionModel struct {
	Rows    []Row
	Count   int
	PaidBy  int
	GroupId int
}

type UploadHandler struct {
	store     *activity.Store
	userStore *users.Store
}

func NewUploadHandler(store *activity.Store, userStore *users.Store) *UploadHandler {
	return &UploadHandler{store: store, userStore: userStore}
}

func (h *UploadHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		r.ParseForm()
		slog.Info("parsed form", "paid by", r.FormValue("paidBy"), "count", r.FormValue("count"))

		selectedExpenses := []activity.Expense{}

		group, _ := strconv.Atoi(r.FormValue("groupId"))
		count, _ := strconv.Atoi(r.FormValue("count"))
		paidBy, _ := strconv.Atoi(r.FormValue("paidBy"))
		for i := 0; i < count; i++ {
			key := fmt.Sprintf("selected-%d", i)
			if r.FormValue(key) != "on" {
				continue
			}

			layout := "2006-01-02 15:04:05 -0700 MST"
			dateStr := r.FormValue(fmt.Sprintf("date-%d", i))
			date, err := time.Parse(layout, dateStr)
			if err != nil {
				slog.Error("Could not parse date", "date", dateStr)
			}
			amount, _ := strconv.Atoi(r.FormValue(fmt.Sprintf("amount-%d", i)))
			expense := activity.Expense{
				Id:             0,
				Description:    r.FormValue(fmt.Sprintf("description-%d", i)),
				AmountCentsUsd: amount,
				PaidBy:         paidBy,
				CreatedOn:      date,
			}
			selectedExpenses = append(selectedExpenses, expense)
		}

		participants, err := h.userStore.ListUsersInGroup(group)
		if err != nil {
			http.Error(w, "could not list participants in group", http.StatusInternalServerError)
			return
		}

		participantIds := []int{}
		for _, p := range participants {
			participantIds = append(participantIds, p.Id)
		}

		for _, e := range selectedExpenses {
			err := h.store.AddExpense(group, paidBy, e.Description, e.AmountCentsUsd, e.CreatedOn, participantIds)
			if err != nil {
				http.Error(w, "failed to add expense", http.StatusInternalServerError)
				return
			}
		}

		http.Redirect(w, r, fmt.Sprintf("/%d", group), http.StatusSeeOther)
	}
}
