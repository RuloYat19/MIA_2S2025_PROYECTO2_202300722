package Comandos

import (
	globales "Proyecto/Globales"
	"Proyecto/Structs"
	"fmt"
)

type LOGOUT struct{}

func Logout(parametros []string) string {
	fmt.Println("\n======= LOGOUT =======")

	var salida1 = ""

	if len(parametros) > 1 {
		salida1 += "LOGOUT Error: El comando Logout no acepta parámetros.\n"
		fmt.Println("LOGOUT Error: El comando Logout no acepta parámetros.")
		return salida1
	}

	err := cerrarSesion()
	if err != nil {
		salida1 += "LOGOUT Error: Hubo problemas cerrando la sesión.\n"
		fmt.Println("LOGOUT Error: Hubo problemas cerrando la sesión.")
		return salida1
	} else {
		salida1 += "El cierre de sesión ha sido exitoso.\n"
	}

	fmt.Println("\n======FIN LOGOUT======")
	return salida1
}

func cerrarSesion() error {
	// Verificar si hay alguna sesión activa
	if globales.UsuarioActual == nil || !globales.UsuarioActual.Estado {
		return fmt.Errorf("no hay ninguna sesión activa")
	}

	// Mensaje para informar el cierre de sesión
	fmt.Printf("Cerrando sesión de usuario: %s\n", globales.UsuarioActual.Nombre)

	// Se reinicia la estructura del usuario actual
	globales.UsuarioActual = &Structs.Usuario{}
	globales.ParticionInicioSesion = ""

	// Mensaje de cierre éxitoso de sesión
	fmt.Println("Sesión cerrada correctamente.")

	return nil
}
