package main

import (
	"testing"

	"github.com/alecthomas/assert"
	tb "gopkg.in/tucnak/telebot.v2"
)

func U(id int, f, l, u string) *tb.User {
	return &tb.User{ID: id, FirstName: f, LastName: l, Username: u}
}

func TestUserNames(t *testing.T) {
	for user, names := range map[tb.User][]string{
		{ID: 4}:                                {"#4"},
		{ID: 4, FirstName: "a"}:                {"a"},
		{ID: 4, LastName: "b"}:                 {"b"},
		{ID: 4, FirstName: "a", LastName: "b"}: {"a", "a b"},
		{ID: 4, Username: "c"}:                 {"@c"},
		*U(4, "a", "b", "c"):                   {"a", "a b", "@c"},
	} {
		assert.Equal(t, names, UserNames(&user))
	}
}

func TestChooseName(t *testing.T) {
	for name, users := range map[string][]*tb.User{
		"a":   {U(1, "a", "b", "c")},
		"a b": {U(1, "a", "b", "c"), U(2, "a", "e", "f")},
		"@c":  {U(1, "a", "b", "c"), U(2, "a", "b", "f")},
		"g h": {U(1, "g", "h", ""), U(2, "g", "h", "")},
	} {
		assert.Equal(t, name, ChooseName(users[0], users))
	}
}
