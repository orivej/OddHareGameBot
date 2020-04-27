package bot

import (
	"fmt"

	tb "gopkg.in/tucnak/telebot.v2"
)

func UserNames(user *tb.User) (names []string) {
	name := func(s string) { names = append(names, s) }
	if user.FirstName != "" {
		name(user.FirstName)
	} else if user.LastName != "" {
		name(user.LastName)
	}
	if user.FirstName != "" && user.LastName != "" {
		name(user.FirstName + " " + user.LastName)
	}
	if user.Username != "" {
		name("@" + user.Username)
	}
	if len(names) == 0 {
		name(fmt.Sprint("#", user.ID))
	}
	return
}

func ChooseName(user *tb.User, others []*tb.User) (name string) {
Names:
	for _, name = range UserNames(user) {
		for _, other := range others {
			if user.ID != other.ID {
				for _, name2 := range UserNames(other) {
					if name == name2 {
						continue Names
					}
				}
			}
		}
		break
	}
	return
}
