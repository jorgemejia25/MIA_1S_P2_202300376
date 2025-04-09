package auth

import (
	"fmt"
)

func Logout() error {
	loggedUser := GetInstance()

	if loggedUser.User == nil {
		return fmt.Errorf("error al cerrar sesión: no hay un usuario loggeado")
	}

	loggedUser.User = nil

	return nil
}
