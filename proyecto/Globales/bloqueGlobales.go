package globales

import (
	"Proyecto/Structs"
	"fmt"
	"os"
	"strings"
)

func BuscarEnElArchivoDeUsuario(archivo *os.File, superbloque *Structs.Superbloque, inodo *Structs.Inodo, nombre, entityType string) (string, error) {
	contenido, err := LeerBloquesArchivo(archivo, superbloque, inodo)
	if err != nil {
		return "", err
	}

	// Se usa la función auxiliar para buscar la línea
	linea, _, err := BuscarLineaEnElArchivoDeUsuarios(contenido, nombre, entityType)
	if err != nil {
		return "", err
	}

	return linea, nil
}

func LeerBloquesArchivo(archivo *os.File, superbloque *Structs.Superbloque, inodo *Structs.Inodo) (string, error) {
	var contenido string

	for _, blockIndex := range inodo.I_block {
		if blockIndex == -1 {
			break
		}

		blockOffset := int64(superbloque.S_block_start + blockIndex*int32(superbloque.S_block_s))
		var bloqueFile Structs.BloqueFile

		// Se lee el bloque desde el archivo
		err := bloqueFile.Decodificar(archivo, blockOffset)
		if err != nil {
			return "", fmt.Errorf("error leyendo bloque %d: %w", blockIndex, err)
		}

		// Se concatena el contenido del bloque al resultado total
		contenido += string(bloqueFile.B_contenido[:])
	}

	// Se actualiza el tiempo de último acceso
	inodo.ActualizarAtime()

	return strings.TrimRight(contenido, "\x00"), nil
}

// Se busca una línea en el archivo users.txt según nombre y tipo
func BuscarLineaEnElArchivoDeUsuarios(contenido string, nombre, entityType string) (string, int, error) {
	// Dividir el contenido en líneas
	lineas := strings.Split(contenido, "\n")

	for i, linea := range lineas {
		campos := strings.Split(linea, ",")
		if len(campos) < 3 {
			continue
		}

		// Determinar si es un grupo o un usuario según el entityType
		if entityType == "G" && len(campos) == 3 {
			grupo := Structs.NuevoGrupo(campos[0], campos[2]) // Crear instancia de Group
			if grupo.Tipo == entityType && grupo.Grupo == nombre {
				return grupo.ToString(), i, nil // Devolver la línea y el índice
			}
		} else if entityType == "U" && len(campos) == 5 {
			usuario := Structs.NuevoUsuario(campos[0], campos[2], campos[3], campos[4]) // Crear instancia de User
			if usuario.Tipo == entityType && usuario.Nombre == nombre {
				return usuario.ToString(), i, nil // Devolver la línea y el índice
			}
		}
	}

	return "", -1, fmt.Errorf("%s '%s' no encontrado en users.txt", entityType, nombre)
}

// Se añade una entrada al archivo users.txt ya sea grupo o usuario usando la depuración
func AgregarEntradaAlArchivoDeUserstxt(archivo *os.File, superbloque *Structs.Superbloque, inodo *Structs.Inodo, entrada, nombre, tipoEntidad string) error {
	// Se lee el contenido actual de users.txt
	contenidoActual, err := LeerBloquesArchivo(archivo, superbloque, inodo)
	if err != nil {
		return fmt.Errorf("error leyendo blocks de users.txt: %w", err)
	}

	// Se verifica si el grupo o usuario ya existe
	_, _, err = BuscarLineaEnElArchivoDeUsuarios(contenidoActual, nombre, tipoEntidad)
	if err == nil {
		// Si ya existe, no se crea el grupo o usuario, se retorna sin hacer nada
		fmt.Printf("La entidad '%s' de nombre '%s' ya existe en users.txt\n", tipoEntidad, nombre)
		return nil
	}

	fmt.Println("=== Escribiendo nuevo contenido en users.txt ===")
	fmt.Println(entrada)

	// Se escribe solo la nueva entrada al final de los bloques
	err = EscribirBloquesEnUserstxt(archivo, superbloque, inodo, entrada+"\n") // Solo el nuevo grupo
	if err != nil {
		fmt.Printf("Error agregando entrada a users.txt: %v\n", err)
		return err
	}

	// Se muestra el estado del inodo después de la modificación
	fmt.Println("\n=== Estado del inodo después de la modificación ===")
	superbloque.ImprimirInodos(archivo.Name())

	// Se muestra el estado de los bloques después de la modificación
	fmt.Println("\n=== Estado de los bloques después de la modificación ===")
	superbloque.ImprimirBloques(archivo.Name())

	return nil
}

// Se escribe la info en los bloques de los usuarios
func EscribirBloquesEnUserstxt(archivo *os.File, superbloque *Structs.Superbloque, inodo *Structs.Inodo, nuevoContenido string) error {
	// Se lee el contenido actual de los bloques asignados al inodo
	contenidoExistente, err := LeerBloquesArchivo(archivo, superbloque, inodo)

	if err != nil {
		return fmt.Errorf("error leyendo contenido existente de users.txt: %w", err)
	}

	// Se combina el contenido existente con el nuevo contenido
	contenidoTotal := contenidoExistente + nuevoContenido

	// Se divide el contenido total en bloques de tamaño BlockSize
	bloques, err := Structs.DividirContenido(contenidoTotal)
	if err != nil {
		return fmt.Errorf("error al dividir el contenido en bloques: %w", err)
	}

	// Variable para mantener el índice del bloque
	indice := 0

	// Se itera sobre los bloques generados y se escribe en los bloques del inodo
	for _, bloque := range bloques {
		// Se verifica si el índice excede la capacidad del array I_block
		if indice >= len(inodo.I_block) {
			return fmt.Errorf("se alcanzó el límite máximo de bloques del inodo")
		}

		// Si el bloque actual en el inodo está vacío, asignar uno nuevo
		if inodo.I_block[indice] == -1 {
			indiceNuevoBloque, err := superbloque.AsignarNuevoBloque(archivo, inodo, indice)
			if err != nil {
				return fmt.Errorf("error asignando nuevo bloque: %w", err)
			}
			inodo.I_block[indice] = indiceNuevoBloque
		}

		// Se calcula el offset del bloque en el archivo
		blockOffset := int64(superbloque.S_block_start + inodo.I_block[indice]*int32(superbloque.S_block_s))

		// Se escribe el contenido del bloque en la partición
		err = bloque.Codificar(archivo, blockOffset)
		if err != nil {
			return fmt.Errorf("error escribiendo el bloque %d: %w", inodo.I_block[indice], err)
		}

		// Se mueve al siguiente bloque
		indice++
	}

	// Se actualiza el tamaño del archivo en el inodo (I_s)
	nuevoTamano := len(contenidoTotal)
	inodo.I_s = int32(nuevoTamano)

	// Se actualizan los tiempos de modificación y cambio
	inodo.ActualizarMTime()
	inodo.ActualizarCTime()

	return nil
}

func InsertarEnElArchivoDeUsuarios(archivo *os.File, superbloque *Structs.Superbloque, inodo *Structs.Inodo, entrada string) error {
	// Se lee el contenido actual de los bloques asignados al inodo
	contenidoActual, err := LeerBloquesArchivo(archivo, superbloque, inodo)
	if err != nil {
		return fmt.Errorf("error leyendo el contenido de users.txt: %w", err)
	}

	// Se eliminan las líneas vacías o con espacios innecesarios del contenido actual
	lineas := strings.Split(strings.TrimSpace(contenidoActual), "\n")

	// Se obtiene el grupo desde la nueva entrada
	partesEntry := strings.Split(entrada, ",")

	if len(partesEntry) < 4 { // Se esperan los argumentos del Usuario que son: UID, U, Grupo, Usuario, Contraseña
		return fmt.Errorf("entrada de usuario inválida: %s", entrada)
	}

	grupoUsuario := partesEntry[2]

	// Se busca el ID del grupo correspondiente en el contenido actual
	var idGrupo string
	var nuevoContenido []string
	usuarioInsertado := false

	// Se recorren las líneas de `users.txt` para encontrar el grupo correspondiente
	for _, linea := range lineas {
		partes := strings.Split(linea, ",")
		// Se agrega la línea actual al nuevo contenido
		nuevoContenido = append(nuevoContenido, strings.TrimSpace(linea))

		// Si se encuentra el grupo correcto
		if len(partes) > 2 && partes[1] == "G" && partes[2] == grupoUsuario {
			idGrupo = partes[0] // Se obtiene el ID del grupo

			// Se inserta el usuario justo después del grupo si no se ha insertado ya
			if idGrupo != "" && !usuarioInsertado {
				usuarioConGrupo := fmt.Sprintf("%s,U,%s,%s,%s", idGrupo, partesEntry[2], partesEntry[3], partesEntry[4])
				nuevoContenido = append(nuevoContenido, usuarioConGrupo)
				usuarioInsertado = true
				//break
			}
		}
	}

	// Se verifica si el grupo fue encontrado
	if idGrupo == "" {
		fmt.Printf("Error agregando al usuario ya que el grupo '%s' no existe.\n", grupoUsuario)
		return fmt.Errorf("el grupo '%s' no existe", grupoUsuario)
	}

	// Se combinan todas las líneas en un solo contenido para escribirlo en el archivo, donde se eliminan las posibles líneas en blanco
	contenidoCombinado := strings.Join(nuevoContenido, "\n") + "\n"
	fmt.Println("=== Escribiendo nuevo contenido en users.txt ===")
	fmt.Println(contenidoCombinado)

	// Se limpian los bloques asignados al archivo
	for _, indiceBloque := range inodo.I_block {
		if indiceBloque == -1 {
			break
		}

		bloqueOffset := int64(superbloque.S_block_start + indiceBloque*superbloque.S_block_s)
		var bloqueFile Structs.BloqueFile

		// Se limpia el contenido del bloque
		bloqueFile.LimpiarContenido()

		// Se escribe el bloque vacío de nuevo
		err = bloqueFile.Codificar(archivo, bloqueOffset)
		if err != nil {
			return fmt.Errorf("error escribiendo bloque limpio %d: %w", indiceBloque, err)
		}
	}

	// Se reescribe todo el contenido línea por línea
	err = EscribirBloquesEnUserstxt(archivo, superbloque, inodo, contenidoCombinado)
	if err != nil {
		return fmt.Errorf("error escribiendo el nuevo contenido en users.txt: %w", err)
	}

	// Se actualiza el tamaño del archivo (I_s)
	inodo.I_s = int32(len(contenidoCombinado))

	// Se actualizan los tiempos de modificación y cambio
	inodo.ActualizarMTime()
	inodo.ActualizarCTime()

	return nil
}
