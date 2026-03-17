package users

import (
	"database/sql"
)

type User struct {
	Id   int
	Name string
}

type Store struct {
	db *sql.DB
}

func (s *Store) ListUsersInGroup(group int) ([]User, error) {

	var ux []User
	rows, err := s.db.Query("select id, name from Users u inner join users_groups ug on u.id = ug.user_id where ug.group_id = $1", group)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var u User
		if err := rows.Scan(&u.Id, &u.Name); err != nil {
			return nil, err
		}

		ux = append(ux, u)
	}

	return ux, nil
}

func NewStore(db *sql.DB) *Store {
	return &Store{
		db: db,
	}
}
