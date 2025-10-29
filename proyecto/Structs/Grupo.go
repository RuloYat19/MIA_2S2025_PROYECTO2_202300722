package Structs

import "fmt"

type Grupo struct {
	GID   string
	Tipo  string
	Grupo string
}

func NuevoGrupo(gid, grupo string) *Grupo {
	return &Grupo{gid, "G", grupo}
}

func (g *Grupo) ToString() string {
	return fmt.Sprintf("%s,%s,%s", g.GID, g.Tipo, g.Grupo)
}

func (g *Grupo) Eliminar() {
	g.GID = "0"
}
