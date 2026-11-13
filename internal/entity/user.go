package entity

import (
"fmt"
"strings"
)

type User struct {
ID       int64
Name     string
Position int
}

func New(id int64, lastName string, firstName string) User {
return User{ID: id, Name: strings.TrimSpace(strings.Join([]string{lastName, firstName}, " "))}
}

func NewFullName(id int64, fullName string) User {
return User{ID: id, Name: strings.TrimSpace(fullName)}
}

func listNumber(u User, idx int) int {
if u.Position > 0 {
return u.Position
}
return idx + 1
}

func ListToString(users []User) (result string) {
sb := strings.Builder{}
for i, u := range users {
sb.WriteString(fmt.Sprintf("%d. %s", listNumber(u, i), u.Name))
if i < len(users)-1 {
sb.WriteByte('\n')
}
}
return sb.String()
}

func ListToStringWithCurrent(users []User, currentUser int) string {
sb := strings.Builder{}
for idx, u := range users {
n := listNumber(u, idx)
if currentUser == idx {
sb.WriteString(fmt.Sprintf("%d. -> %s <-", n, u.Name))
} else {
sb.WriteString(fmt.Sprintf("%d. %s", n, u.Name))
}
if idx < len(users)-1 {
sb.WriteByte('\n')
}
}
return sb.String()
}