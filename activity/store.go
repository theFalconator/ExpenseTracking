package activity

import (
	"database/sql"
	"log/slog"
	"time"
)

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{
		db: db,
	}
}

func (s *Store) RecordSettlementForGroup(group int) error {
	_, err := s.db.Exec("insert into Settlements (group_id, last_expense_id, created_on) values ($1, (select max(id) from Expenses where group_id = $1), current_timestamp)", group)
	return err
}

func (s *Store) GetLatestSettlementForGroup(g int) (Settlement, error) {
	var stmt Settlement

	query := "select id, last_expense_id, created_on from Settlements where group_id = $1 order by created_on desc limit 1"
	if err := s.db.QueryRow(query, g).Scan(&stmt.Id, &stmt.LastExpenseId, &stmt.CreatedOn); err != nil {
		return stmt, err
	}

	return stmt, nil
}

func (s *Store) ListExpensesForGroup(from int, group int) ([]ExpenseRow, error) {
	query := `select e.id
	, e.description
	, e.amount_cents_usd
	, e.paid_by
	, e.created_on
	, u.id
	from expenses e
 inner join users_expenses ue ON ue.expense_id = e.id
 inner join users u on ue.user_id = u.id
 where e.id > $1 and e.group_id = $2
 order by ue.expense_id`

	var expenses []ExpenseRow

	rows, err := s.db.Query(query, from, group)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var e ExpenseRow
		if err := rows.Scan(&e.ExpenseId, &e.Description, &e.AmountCentsUsd, &e.PaidBy, &e.CreatedOn, &e.ParticipantId); err != nil {
			return nil, err
		}

		expenses = append(expenses, e)
	}

	return expenses, nil
}

func (s *Store) ListExpenseDescriptions() ([]string, error) {
	var descriptions []string
	rows, err := s.db.Query("select distinct description from Expenses")

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var s string
		if err := rows.Scan(&s); err != nil {
			return nil, err
		}

		descriptions = append(descriptions, s)
	}

	return descriptions, nil
}

func (s *Store) AddExpense(group, paidBy int, description string, cents int, date time.Time, participants []int) error {
	id := 0
	command := "insert into Expenses (description, amount_cents_usd, paid_by, created_on, group_id) values ($1, $2, $3, $4, $5) returning id"
	err := s.db.QueryRow(command, description, cents, paidBy, date, group).Scan(&id)

	if err != nil {
		slog.Error("Could not insert into expenses", "error", err)
		return err
	}

	for _, participant := range participants {
		_, err := s.db.Exec("insert into users_expenses (user_id, expense_id) values ($1, $2)", participant, id)
		if err != nil {
			slog.Error("Could not insert into users_expenses", "error", err, "user id", participant, "expense_id", id)
			return err
		}
	}

	return err
}
