package activity

import "github.com/theFalconator/expenses/users"

type Pair struct {
	Owed   int
	OwedBy int
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
	TotalDebts int
}

func ComputeAmountOwedByUserId(expenses []ExpenseRow) map[Pair]int {
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

func ComputeDebts(owed map[Pair]int, userList []users.User, usersById map[int]UserDto) {
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
