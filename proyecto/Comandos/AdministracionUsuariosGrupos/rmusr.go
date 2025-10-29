package Comandos

import (
	globales "Proyecto/Globales"
	"Proyecto/Structs"
	"encoding/binary"
	"fmt"
	"os"
	"strings"
)

type RMUSR struct {
	Usuario string
}

func Rmusr(parametros []string) string {
	fmt.Println("\n======= RMUSR =======")

	var salida1 = ""

	// Se inicializa la estructura RMUSR
	rmusr := &RMUSR{}

	// Otras variables
	paramC := true
	usuarioInit := false

	// Se asignan los parámetros a la estructura
	for _, parametro := range parametros[1:] {
		tmp2 := strings.TrimRight(parametro, " ")
		tmp := strings.Split(tmp2, "=")

		if len(tmp) != 2 {
			salida1 += "RMGRP Error: Valor desconocido del parámetro " + tmp[0] + ".\n"
			fmt.Println("RMGRP Error: Valor desconocido del parámetro", tmp[0])
			paramC = false
			break
		}

		if strings.ToLower(tmp[0]) == "user" {
			usuarioInit = true
			rmusr.Usuario = tmp[1]
		} else {
			salida1 += "RMGRP Error: Parámetro desconocido: " + tmp[0] + ".\n"
			fmt.Println("RMUSR Error: Parametro desconocido: ", tmp[0])
			paramC = false
			break
		}
	}

	// Se elimina el usuario con los datos asignados
	if paramC && usuarioInit {
		err := eliminarUsuario(rmusr)
		if err != nil {
			salida1 += "RMGRP Error: Hubo problemas al eliminar el usuario " + rmusr.Usuario + ".\n"
		} else {
			salida1 += "El usuario " + rmusr.Usuario + " fue eliminado exitosamente.\n"
		}
	}

	return salida1
}

func eliminarUsuario(rmusr *RMUSR) error {
	fmt.Println(rmusr.Usuario)

	// Se verifica si hay alguna sesión activa
	if !globales.HaIniciadoSesion() {
		return fmt.Errorf("no hay ninguna sesión activa")
	}

	// Se verifica si el usuario es el root
	if globales.UsuarioActual.Nombre != "root" {
		return fmt.Errorf("solo el usuario root puede ejecutar este comando")
	}

	// Se verifica que la partición esté montada en el sistema
	_, ruta, err := globales.ObtenerParticionMontadas(globales.UsuarioActual.UID)
	if err != nil {
		return fmt.Errorf("no se puede encontrar la partición montada: %v", err)
	}

	// Se abre el archivo donde se encuentra la partición montada
	archivo, err := os.OpenFile(ruta, os.O_RDWR, 0755)

	if err != nil {
		return fmt.Errorf("no se puede abrir el archivo de la partición: %v", err)
	}

	defer archivo.Close()

	// Se carga el Superblock y la Partición
	mbr, superbloque, _, err := globales.ObtenerParticionMontadaRep(globales.UsuarioActual.UID) //Id de la particion del usuario actual

	if err != nil {
		return fmt.Errorf("no se pudo cargar el Superblock: %v", err)
	}

	// Se obtiene la partición asociada al id
	particion, err := mbr.ObtenerParticionPorID(globales.UsuarioActual.UID)
	if err != nil {
		return fmt.Errorf("no se pudo obtener la partición: %v", err)
	}

	// Se lee el inodo de users.txt
	var usuariosInodo Structs.Inodo

	// Se calcula el offset del inodo de users.txt que esta en el inodo 1
	inodoOffset := int64(superbloque.S_inode_start + int32(binary.Size(usuariosInodo)))

	// Se decodifica el inodo de users.txt
	err = usuariosInodo.Decodificar(archivo, inodoOffset)

	if err != nil {
		return fmt.Errorf("error leyendo el inodo de users.txt: %v", err)
	}

	// Se verifica si el usuario ya existe en users.txt
	_, err = globales.BuscarEnElArchivoDeUsuario(archivo, superbloque, &usuariosInodo, rmusr.Usuario, "U")

	if err != nil {
		fmt.Printf("El usuario '%s' no existe\n", rmusr.Usuario)
		return err
	}

	// Se marca el usuario como eliminado
	err = ActualizarEstadoUsuario(archivo, superbloque, &usuariosInodo, rmusr.Usuario)

	if err != nil {
		fmt.Printf("Error eliminando el usuario '%s': %v\n", rmusr.Usuario, err)
		return err
	}

	// Se actualiza el inodo de users.txt
	err = usuariosInodo.Codificar(archivo, inodoOffset)

	if err != nil {
		fmt.Printf("Error actualizando inodo de users.txt: %v\n", err)
		return fmt.Errorf("error actualizando inodo de users.txt: %v", err)
	}

	// Se guarda el Superblock utilizando el Part_start como el offset
	err = superbloque.Codificar(archivo, int64(particion.Start))

	if err != nil {
		fmt.Printf("Error guardando el Superblock: %v\n", err)
		return err
	} else {
		fmt.Println("Superbloque guardado correctamente")
	}

	// Se muestran los mensajes donde confirman que se eliminó el usuario
	superbloque.ImprimirSuperbloque()
	fmt.Println("------")
	fmt.Printf("Usuario '%s' eliminado exitosamente.\n", rmusr.Usuario)
	fmt.Println("\nBloques:")
	superbloque.ImprimirBloques(archivo.Name())
	fmt.Println("\nInodos:")
	superbloque.ImprimirInodos(archivo.Name())

	return nil
}

// Se cambia el estado de un usuario a eliminado (ID=0) y actualiza el archivo
func ActualizarEstadoUsuario(archivo *os.File, superbloque *Structs.Superbloque, usuariosInodo *Structs.Inodo, nombreUsuario string) error {
	// Lee el contenido actual de users.txt
	contenido, err := globales.LeerBloquesArchivo(archivo, superbloque, usuariosInodo)

	if err != nil {
		return fmt.Errorf("error leyendo el contenido de users.txt: %v", err)
	}

	// Separa las líneas del archivo
	lineas := strings.Split(contenido, "\n")
	modificado := false

	// Recorre las líneas para buscar y modificar el estado del usuario
	for i, linea := range lineas {
		linea = strings.TrimSpace(linea) // Elimina espacios en blanco adicionales
		if linea == "" {
			continue
		}

		// Crea un objeto User a partir de la línea
		usuario := crearUsuarioDesdeLinea(linea)

		// Verifica si es el usuario que queremos eliminar
		if usuario != nil && usuario.Nombre == nombreUsuario {
			// Elimina el usuario (cambiar ID a "0")
			usuario.Eliminar()

			// Actualiza la línea en el archivo
			lineas[i] = usuario.ToString()
			modificado = true
			break // Una vez se encuentra, se modifica el usuario y se puede salir
		}
	}

	// Si no se encontró el usuario, retornar error
	if !modificado {
		return fmt.Errorf("usuario '%s' no encontrado en users.txt", nombreUsuario)
	}

	// Limpia y actualiza las líneas antes de escribir
	contenidoActualizado := limpiarYActualizarContenido(lineas)

	// Escribe los cambios al archivo
	return escribirCambiosEnArchivo(archivo, superbloque, usuariosInodo, contenidoActualizado)
}

// Crea un objeto User a partir de una línea del archivo
func crearUsuarioDesdeLinea(linea string) *Structs.Usuario {
	partes := strings.Split(linea, ",")
	if len(partes) >= 5 && partes[1] == "U" {
		return Structs.NuevoUsuario(partes[0], partes[2], partes[3], partes[4])
	}
	return nil
}

// Elimina líneas vacías y devuelve el contenido actualizado como string
func limpiarYActualizarContenido(lineas []string) string {
	var contenidoActualizado []string
	for _, linea := range lineas {
		if strings.TrimSpace(linea) != "" {
			contenidoActualizado = append(contenidoActualizado, linea)
		}
	}
	return strings.Join(contenidoActualizado, "\n") + "\n"
}

// Limpia los bloques y escribe el contenido actualizado en el archivo
func escribirCambiosEnArchivo(archivo *os.File, superbloque *Structs.Superbloque, usuariosInodo *Structs.Inodo, contenido string) error {
	// Limpia los bloques asignados al archivo antes de escribir
	for _, indiceBloque := range usuariosInodo.I_block {
		if indiceBloque == -1 {
			break
		}

		bloqueOffset := int64(superbloque.S_block_start + indiceBloque*superbloque.S_block_s)
		var bloqueFile Structs.BloqueFile

		// Limpia el contenido del bloque
		bloqueFile.LimpiarContenido()

		// Escribe el bloque vacío de nuevo
		err := bloqueFile.Codificar(archivo, bloqueOffset)
		if err != nil {
			return fmt.Errorf("error escribiendo bloque limpio %d: %w", indiceBloque, err)
		}
	}

	// Reescribe todo el contenido en los bloques después de limpiar
	err := globales.EscribirBloquesEnUserstxt(archivo, superbloque, usuariosInodo, contenido)

	if err != nil {
		return fmt.Errorf("error guardando los cambios en users.txt: %v", err)
	}

	return nil
}
