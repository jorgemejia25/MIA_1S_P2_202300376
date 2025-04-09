package auth

import (
	"sync"

	"disk.simulator.com/m/v2/internal/disk/types/structures/authentication"
)

var instance *LoggedUser
var once sync.Once

type LoggedUser struct {
	ID   string
	GID  string
	User *authentication.User
}

func GetInstance() *LoggedUser {
	once.Do(func() {
		instance = &LoggedUser{}
	})
	return instance
}

func (l *LoggedUser) SetLoggedUser(id string, gid string, user *authentication.User) {
	l.ID = id
	l.GID = gid
	l.User = user

}
