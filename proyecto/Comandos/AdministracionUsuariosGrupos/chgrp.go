package Comandos

import (
	globales "Proyecto/Globales"
	"Proyecto/Structs"
	"encoding/binary"
	"fmt"
	"os"
	"strings"
)

type CHGRP struct {
	Usuario string
	Grupo   string
}

func Chgrp(parametros []string) string {
	fmt.Println("\n======= CHGRP =======")

	var salida1 = ""

	// Se inicializa la estructura CHGRP
	chgrp := &CHGRP{}

	// Otras variables
	paramC := true
	usuarioInit := false
	grupoInit := false

	// Se asignan los parámetros a la estructura
	for _, parametro := range parametros[1:] {
		tmp2 := strings.TrimRight(parametro, " ")
		tmp := strings.Split(tmp2, "=")

		if len(tmp) != 2 {
			salida1 += "CHGRP Error: Valor desconocido del parámetro " + tmp[0] + ".\n"
			fmt.Println("CHGRP Error: Valor desconocido del parámetro", tmp[0])
			paramC = false
			break
		}

		if strings.ToLower(tmp[0]) == "user" {
			usuarioInit = true
			chgrp.Usuario = tmp[1]
		} else if strings.ToLower(tmp[0]) == "grp" {
			grupoInit = true
			chgrp.Grupo = tmp[1]
		} else {
			salida1 += "CHGRP Error: Parámetro desconocido " + tmp[0] + ".\n"
			fmt.Println("CHGRP Error: Parametro desconocido: ", tmp[0])
			paramC = false
			break
		}
	}

	// Se cambia al usuario a otro grupo con los datos asignados
	if paramC && usuarioInit && grupoInit {
		err := cambiarGrupoAlUsuario(chgrp)
		if err != nil {
			salida1 += "CHGRP Error: Hubo problemas al cambiar al usuario " + chgrp.Usuario + " al grupo " + chgrp.Grupo + ".\n"
		} else {
			salida1 += "El usuario " + chgrp.Usuario + " fue cambiado al grupo " + chgrp.Grupo + " exitosamente.\n"
		}
	}

	fmt.Println("\n======FIN CHGRP======")
	return salida1
}

func cambiarGrupoAlUsuario(chgrp *CHGRP) error {
	fmt.Println(chgrp.Usuario, " ", chgrp.Grupo)

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

	// Se cambia el grupo del usuario
	err = realizarElCambioDeGrupo(archivo, superbloque, &usuariosInodo, chgrp.Usuario, chgrp.Grupo)
	if err != nil {
		return fmt.Errorf("error cambiando el grupo del usuario '%s': %v", chgrp.Usuario, err)
	}

	// Se guardar el superbloque
	err = superbloque.Codificar(archivo, int64(particion.Start))

	if err != nil {
		return fmt.Errorf("error guardando el superbloque: %v", err)
	}

	// Se muestran mensajes donde confirman que se cambió al usuario de grupo
	fmt.Printf("El grupo del usuario '%s' ha sido cambiado exitosamente a '%s'\n", chgrp.Usuario, chgrp.Grupo)
	fmt.Println("\nInodos")
	superbloque.ImprimirInodos(archivo.Name())
	fmt.Println("\nBloques")
	superbloque.ImprimirBloques(archivo.Name())
	fmt.Println("=====================================================")

	return nil
}

// Se cambia el grupo de un usuario en el archivo users.txt y reorganiza el contenido
func realizarElCambioDeGrupo(archivo *os.File, superbloque *Structs.Superbloque, usuariosInodo *Structs.Inodo, userName, nuevoGrupo string) error {
	// Se lee el contenido actual de users.txt
	contenidoActual, err := globales.LeerBloquesArchivo(archivo, superbloque, usuariosInodo)
	if err != nil {
		return fmt.Errorf("error leyendo el contenido de users.txt: %w", err)
	}

	// Se eliminan las líneas vacías o con espacios innecesarios del contenido actual
	lineas := strings.Split(strings.TrimSpace(contenidoActual), "\n")
	var nuevoContenido []string
	var usuarioModificado bool
	var grupoEncontrado bool

	// Se procesa el contenido del archivo y convierte a objetos de Usuario y Grupo
	var usuarios []Structs.Usuario
	var grupos []Structs.Grupo

	// Se separa usuarios y grupos
	for _, linea := range lineas {
		partes := strings.Split(linea, ",")
		if len(partes) < 3 {
			continue
		}

		// Se identifica si es un grupo o un usuario
		tipo := strings.TrimSpace(partes[1])
		if tipo == "G" {
			// Se crea un objeto de tipo Grupo
			grupo := Structs.NuevoGrupo(partes[0], partes[2])
			grupos = append(grupos, *grupo)
		} else if tipo == "U" && len(partes) >= 5 {
			// Se crea un objeto de tipo Usuario
			usuario := Structs.NuevoUsuario(partes[0], partes[2], partes[3], partes[4])
			usuarios = append(usuarios, *usuario)
		}
	}

	// Se comprueba si el nuevo grupo existe y no está eliminado
	var nuevoIDGrupo string
	for _, grupo := range grupos {
		if grupo.Grupo == nuevoGrupo && grupo.GID != "0" {
			nuevoIDGrupo = grupo.GID
			grupoEncontrado = true
			break
		}
	}

	if !grupoEncontrado {
		fmt.Printf("El grupo '%s' no existe o está eliminado\n", nuevoGrupo)
		return fmt.Errorf("el grupo '%s' no existe o está eliminado", nuevoGrupo)
	}

	// Se modifica el grupo del usuario si existe
	for i, usuario := range usuarios {
		if usuario.Nombre == userName && usuario.UID != "0" { // Verifica que el usuario no esté eliminado
			// Se cambia el grupo del usuario y se actualiza su ID al ID del nuevo grupo
			fmt.Printf("Cambiando el grupo del usuario '%s' al grupo '%s' (ID grupo: %s)\n", usuario.Nombre, nuevoGrupo, nuevoIDGrupo)
			usuarios[i].Grupo = nuevoGrupo
			usuarios[i].UID = nuevoIDGrupo // Se cambia el ID del usuario al ID del grupo destino
			fmt.Printf("Nuevo estado del usuario: %s\n", usuarios[i].ToString())
			usuarioModificado = true
		}
	}

	if !usuarioModificado {
		return fmt.Errorf("el usuario '%s' no existe o está eliminado", userName)
	}

	// Se reorganiza el contenido para agrupar los usuarios bajo sus grupos correspondientes
	for _, grupo := range grupos {
		nuevoContenido = append(nuevoContenido, grupo.ToString()) // Agrega grupo al contenido

		// Se agrega los usuarios asociados al grupo
		for _, usuario := range usuarios {
			if usuario.Grupo == grupo.Grupo {
				nuevoContenido = append(nuevoContenido, usuario.ToString())
			}
		}
	}

	// Se limpia los bloques asignados antes de escribir el nuevo contenido
	for _, indiceBloque := range usuariosInodo.I_block {
		if indiceBloque == -1 {
			break
		}

		blockOffset := int64(superbloque.S_block_start + indiceBloque*superbloque.S_block_s)
		var bloqueFile Structs.BloqueFile

		// Se limpia el contenido del bloque
		bloqueFile.LimpiarContenido()

		// Se escribe el bloque vacío de nuevo
		err = bloqueFile.Codificar(archivo, blockOffset)
		if err != nil {
			return fmt.Errorf("error escribiendo bloque limpio %d: %w", indiceBloque, err)
		}
	}

	// Se reescribe el contenido agrupado en los bloques de `users.txt`
	// Donde se asegura que el contenido se divida correctamente entre los bloques
	err = EscribirContenidoEnLosBloques(archivo, superbloque, usuariosInodo, nuevoContenido)
	if err != nil {
		return fmt.Errorf("error guardando los cambios en users.txt: %v", err)
	}

	// Se actualiza el tamaño del archivo (i_s)
	usuariosInodo.I_s = int32(len(strings.Join(nuevoContenido, "\n")))

	// Se actualiza tiempos de modificación y cambio
	usuariosInodo.ActualizarMTime()
	usuariosInodo.ActualizarCTime()

	// Se guarda el inodo actualizado en el archivo
	inodeOffset := int64(superbloque.S_inode_start + int32(binary.Size(*usuariosInodo)))

	err = usuariosInodo.Codificar(archivo, inodeOffset)
	if err != nil {
		return fmt.Errorf("error actualizando inodo de users.txt: %w", err)
	}

	return nil
}

func EscribirContenidoEnLosBloques(archivo *os.File, superbloque *Structs.Superbloque, usuariosInodo *Structs.Inodo, contenido []string) error {
	// Se convierte el contenido en una cadena
	contenidoFinal := strings.Join(contenido, "\n") + "\n"
	data := []byte(contenidoFinal)

	// Se asigna el tamaño máximo del bloque
	blockSize := int(superbloque.S_block_s)

	// Se escribe el contenido por bloques
	for i, indiceBloque := range usuariosInodo.I_block {
		if indiceBloque == -1 {
			break
		}

		// Se divide los datos en bloques de tamaño máximo
		start := i * blockSize
		if start >= len(data) {
			break
		}

		end := start + blockSize
		if end > len(data) {
			end = len(data)
		}

		chunk := data[start:end]

		// Se crea un bloque con el contenido correspondiente
		var bloqueFile Structs.BloqueFile
		copy(bloqueFile.B_contenido[:], chunk)

		//Se calcula la posición física en el disco
		posicionBloque := int64(superbloque.S_block_start) + (int64(indiceBloque) * int64(blockSize))

		_, err := archivo.Seek(posicionBloque, 0)
		if err != nil {
			return fmt.Errorf("error buscando la posición del bloque %d: %w", indiceBloque, err)
		}

		//Se escribe el bloque en el disco
		err = binary.Write(archivo, binary.LittleEndian, &bloqueFile)
		if err != nil {
			return fmt.Errorf("error escribiendo el bloque %d: %w", indiceBloque, err)
		}

		// Se muestra el bloque escrito para depuración
		fmt.Printf("Escribiendo bloque %d: %s\n", indiceBloque, string(bloqueFile.B_contenido[:]))
	}

	return nil
}
