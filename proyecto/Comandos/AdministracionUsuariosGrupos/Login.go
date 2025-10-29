package Comandos

import (
	globales "Proyecto/Globales"
	"Proyecto/Structs"
	"encoding/binary"
	"fmt"
	"os"
	"strings"
)

/*type LOGIN struct {
	usuario     string
	contrasenia string
	ID          string
}*/

func Login(parametros []string) string {
	fmt.Println("\n======= LOGIN =======")

	var salida1 = ""

	//Datos comando
	//Obligatorios
	var usuario string
	var contrasenia string
	var id string

	//Otras variables
	paramC := true
	usuarioInit := false
	contraseniaInit := false
	idInit := false

	for _, parametro := range parametros[1:] {
		tmp2 := strings.TrimRight(parametro, " ")
		tmp := strings.Split(tmp2, "=")

		if len(tmp) != 2 {
			salida1 += "LOGIN Error: Valor desconocido del parámetro" + tmp[0] + ".\n"
			fmt.Println("LOGIN Error: Valor desconocido del parámetro", tmp[0])
			paramC = false
			break
		}

		if strings.ToLower(tmp[0]) == "user" {
			if strings.ToLower(tmp[1]) == "" || strings.ToLower(tmp[1]) == " " {
				fmt.Println("LOGIN Error: El parámetro de -user es nulo o no tiene nada.")
				salida1 += "LOGIN Error: El parámetro de -user es nulo o no tiene nada.\n"
				return salida1
			}
			usuarioInit = true
			usuario = tmp[1]
		} else if strings.ToLower(tmp[0]) == "pass" {
			if strings.ToLower(tmp[1]) == "" || strings.ToLower(tmp[1]) == " " {
				fmt.Println("LOGIN Error: El parámetro de -pass es nulo o no tiene nada.")
				salida1 += "LOGIN Error: El parámetro de -pass es nulo o no tiene nada.\n"
				return salida1
			}
			contraseniaInit = true
			contrasenia = tmp[1]
		} else if strings.ToLower(tmp[0]) == "id" {
			if strings.ToLower(tmp[1]) == "" || strings.ToLower(tmp[1]) == " " {
				fmt.Println("LOGIN Error: El parámetro de -id es nulo o no tiene nada.")
				salida1 += "LOGIN Error: El parámetro de -id es nulo o no tiene nada.\n"
				return salida1
			}
			idInit = true
			id = tmp[1]
		} else {
			salida1 += "LOGIN Error: Parámetro desconocido" + tmp[0] + ".\n"
			fmt.Println("LOGIN Error: Parámetro desconocido: ", tmp[0])
			paramC = false
			break
		}
	}

	if paramC && usuarioInit && contraseniaInit && idInit {
		err := iniciarSesion(usuario, contrasenia, id)

		if err != nil {
			salida1 += "LOGIN Error: Hubo problemas logueando al usuario " + usuario + ".\n"
			fmt.Printf("LOGIN Error: Hubo problemas logueando al usuario '%s'.", usuario)
		} else {
			salida1 += "El inició sesión ha sido exitoso con el usuario " + usuario + ".\n"
		}
	}
	fmt.Println("\n======FIN LOGIN======")
	return salida1
}

func iniciarSesion(usuarioIS string, contraseniaIS string, idIS string) error {
	fmt.Printf("Intentando iniciar sesión con ID: %s, Usuario: %s\n", idIS, usuarioIS)

	if globales.UsuarioActual != nil && globales.UsuarioActual.Estado {
		return fmt.Errorf("ya hay un usuario logueado, debe cerrar sesión primero")
	}

	fmt.Println("Particiones montadas:")
	for id, ruta := range globales.ParticionesMontadas {
		fmt.Printf("ID: %s | Path: %s\n", id, ruta)
	}

	_, ruta, err := globales.ObtenerParticionMontadas(idIS)

	if err != nil {
		return fmt.Errorf("no se puede encontrar la partición: %v", err)
	}

	fmt.Printf("Partición montada en: %s\n", ruta)

	_, superbloque, _, err := globales.ObtenerParticionMontadaRep(idIS)

	if err != nil {
		return fmt.Errorf("no se pudo cargar el Superbloque: %v", err)
	}

	fmt.Println("Superbloque cargado correctamente")

	archivo, err := os.Open(ruta)

	if err != nil {
		return fmt.Errorf("no se puede abrir el archivo de partición: %v", err)
	}

	defer archivo.Close()

	var usuariosInodo Structs.Inodo
	inodoOffset := int64(superbloque.S_inode_start + int32(binary.Size(usuariosInodo)))

	fmt.Printf("Leyendo inodo users.txt en la posición: %d\n", inodoOffset)

	err = usuariosInodo.Decodificar(archivo, inodoOffset)

	if err != nil {
		return fmt.Errorf("error leyendo inodo de users.txt: %v", err)
	}

	fmt.Println("Inodo de users.txt leído correctamente")

	usuariosInodo.ActualizarAtime()
	usuariosInodo.ImprimirInodo()

	var contenido string

	for _, indiceBloque := range usuariosInodo.I_block {
		if indiceBloque == -1 {
			continue
		}

		bloqueOffset := int64(superbloque.S_block_start + indiceBloque*int32(binary.Size(Structs.BloqueFile{})))
		fmt.Printf("Leyendo bloque en la posición: %d (índice de bloque: %d)\n", bloqueOffset, indiceBloque)

		var bloqueFile Structs.BloqueFile

		err = bloqueFile.Decodificar(archivo, bloqueOffset)
		if err != nil {
			return fmt.Errorf("error leyendo bloque de users.txt: %v", err)
		}

		contenido += string(bloqueFile.B_contenido[:])
	}

	fmt.Println("\nContenido total de users.txt:")
	fmt.Println(contenido)

	encontrado := false
	lineas := strings.Split(strings.TrimSpace(contenido), "\n")

	for _, linea := range lineas {
		if linea == "" {
			continue
		}

		datos := strings.Split(linea, ",")
		if len(datos) == 5 && datos[1] == "U" {
			usuario := Structs.NuevoUsuario(datos[0], datos[2], datos[3], datos[4])

			if usuario.Nombre == usuarioIS && usuario.Contrasenia == contraseniaIS {
				encontrado = true
				globales.UsuarioActual = usuario
				globales.UsuarioActual.Estado = true
				fmt.Printf("Bienvenido %s, inicio de sesión exitoso.\n", usuario.Nombre) // Mensaje importante para el usuario

				globales.UsuarioActual.UID = idIS
				break
			}
		}
	}

	if !encontrado {
		return fmt.Errorf("usuario o contraseña incorrectos")
	}

	globales.ParticionInicioSesion = idIS

	return nil
}
