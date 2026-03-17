package pages

import (
	"encoding/csv"
	"errors"
	"fmt"
	"html/template"
	"log/slog"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/theFalconator/expenses/users"
)

type Row struct {
	TransactionDate time.Time
	PostDate        time.Time
	Description     string
	Category        string
	Amount          int
	Memo            string
}

func parseUSDCents(value string) (int, error) {
	if value == "" {
		return 0, nil
	}
	fValue, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid value: %w", err)
	}

	cents := int(math.Round(fValue * 100))

	return cents, nil
}

type Bank int

const (
	Chase Bank = iota
	CapitalOne
)

func parseRow(r []string, bank Bank) (Row, error) {
	if len(r) != 7 {
		return Row{}, errors.New("expected seven values")
	}

	chaseLayout := "01/02/2006"
	capitalOneLayout := "2006-01-02"

	dateLayout := chaseLayout
	if bank == CapitalOne {
		dateLayout = capitalOneLayout
	}

	tranDate, err := time.Parse(dateLayout, r[0])
	if err != nil {
		slog.Error("invalid transaction date", "error", err)
		return Row{}, err
	}

	postDate, err := time.Parse(dateLayout, r[1])
	if err != nil {
		slog.Error("invalid post date", "error", err)
		return Row{}, err
	}

	cents, err := parseUSDCents(r[5])
	if err != nil {
		slog.Error("invalid amount", "error", err)
		return Row{}, err
	}

	descriptionCol := 2
	if bank == CapitalOne {
		descriptionCol = 3
	}

	categoryCol := 3
	if bank == CapitalOne {
		categoryCol = 4
	}

	return Row{
		TransactionDate: tranDate,
		PostDate:        postDate,
		Description:     r[descriptionCol],
		Category:        r[categoryCol],
		Amount:          cents,
		Memo:            "",
	}, nil
}

type ImportModel struct {
	GroupId int
	Users   []users.User
}

type ImportHandler struct {
	userStore *users.Store
}

func NewImportHandler(u *users.Store) *ImportHandler {
	return &ImportHandler{userStore: u}
}

func (h *ImportHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		r.ParseMultipartForm(10 << 20)
		file, _, err := r.FormFile("myFile")
		if err != nil {
			http.Error(w, "File upload failed", http.StatusBadRequest)
			return
		}
		defer file.Close()

		reader := csv.NewReader(file)
		records, err := reader.ReadAll()
		if err != nil {
			http.Error(w, "could not read file", http.StatusBadRequest)
			return
		}

		bank := Chase
		var rows []Row
		for i, r := range records {
			if i == 0 {
				slog.Info("Headers are", "headers", r)
				if r[5] == "Amount" {
					bank = Chase
				} else if r[5] == "Debit" {
					bank = CapitalOne
					slog.Info("Setting bank to capital one", "bank", bank)
				} else {
					http.Error(w, "unsupported csv format", http.StatusBadRequest)
					return
				}
				// we have headers, skip the first Row
				continue
			}
			row, err := parseRow(r, bank)
			if err != nil {
				slog.Error("Could not read row", "index", i)
				continue
			}

			// Chase has debits and credits in one column, may need to flip sign
			if bank == Chase {
				if row.Amount > 0 {
					continue
				} else {
					// flip it -- we need these to show up as positive amounts for our app
					row.Amount *= -1
				}
			}

			if row.Amount == 0 {
				continue
			}
			rows = append(rows, row)
		}

		tmpl, err := template.New("selection.html").Funcs(template.FuncMap{
			"formatCents":       FormatCents,
			"formatTime":        FormatTime,
			"formatCentsNumber": FormatCentsNumber,
		}).ParseFiles("templates/selection.html")

		userId, _ := strconv.Atoi(r.FormValue("user"))

		group, _ := strconv.Atoi(r.PathValue("id"))
		model := UploadSelectionModel{
			rows,
			len(rows),
			userId,
			group,
		}

		tmpl.Execute(w, model)
		return
	} else {

		errors := map[string]error{}
		tmpl, err := template.New("import.html").ParseFiles("templates/import.html")

		group, err := strconv.Atoi(r.PathValue("id"))
		if err != nil {
			errors["path"] = err
		}
		usersInGroup, err := h.userStore.ListUsersInGroup(group)

		groupId := r.PathValue("id")
		num, err := strconv.Atoi(groupId)
		if err != nil {
			http.Error(w, "invalid group id", http.StatusBadRequest)
			return
		}
		model := ImportModel{
			num,
			usersInGroup,
		}

		tmpl.Execute(w, model)
	}

}
