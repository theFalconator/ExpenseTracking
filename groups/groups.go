package groups

import (
	"database/sql"
	"log/slog"
)

type Group struct {
	Id   int
	Name string
}

type Store struct {
	db *sql.DB
}

func (s *Store) ListGroups() ([]Group, error) {
	var gs []Group
	rows, err := s.db.Query("select id, name from groups")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var g Group
		if err := rows.Scan(&g.Id, &g.Name); err != nil {
			return nil, err
		}

		gs = append(gs, g)
	}

	slog.Debug("found group list", "groups", gs)

	return gs, nil
}

func NewStore(db *sql.DB) *Store {
	return &Store{
		db: db,
	}
}
