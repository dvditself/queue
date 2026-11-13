package entity

import (
"fmt"
"strings"
)

type User struct {
ID   int64
Name string
}

func New(id int64, lastName string, firstName string) User {
return User{ID: id, Name: strings.TrimSpace(strings.Join([]string{lastName, firstName}, " "))}
}

func NewFullName(id int64, fullName string) User {
return User{ID: id, Name: strings.TrimSpace(fullName)}
}

// ListToString — список с номерами: "1. Иванов Иван\n2. Петров Пётр"
func ListToString(users []User) (result string) {
sb := strings.Builder{}
for i, user := range users {
sb.WriteString(fmt.Sprintf("%d. %s", i+1, user.Name))
if i < len(users)-1 {
sb.WriteByte('\n')
}
}
return sb.String()
}

// ListToStringWithCurrent — то же, но текущий выделен стрелками.
// Номер сохраняется, чтобы не путать порядок.
func ListToStringWithCurrent(users []User, currentUser int) string {
sb := strings.Builder{}
for idx, user := range users {
if currentUser == idx {
sb.WriteString(fmt.Sprintf("%d. -> %s <-", idx+1, user.Name))
} else {
sb.WriteString(fmt.Sprintf("%d. %s", idx+1, user.Name))
}
if idx < len(users)-1 {
sb.WriteByte('\n')
}
}
return sb.String()
}