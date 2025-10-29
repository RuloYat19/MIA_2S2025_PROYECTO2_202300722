package Structs

import "fmt"

type Usuario struct {
	UID         string
	Tipo        string
	Grupo       string
	Nombre      string
	Contrasenia string
	Estado      bool
}

func NuevoUsuario(id, grupo, nombre, contrasenia string) *Usuario {
	return &Usuario{id, "U", grupo, nombre, contrasenia, true}
}

func (u *Usuario) ToString() string {
	return fmt.Sprintf("%s,%s,%s,%s,%s", u.UID, u.Tipo, u.Grupo, u.Nombre, u.Contrasenia)
}

func (u *Usuario) Eliminar() {
	u.UID = "0"
	u.Estado = false
}
