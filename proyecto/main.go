package main

import (
	AA "Proyecto/Comandos/AdministracionArchivos"
	ACAP "Proyecto/Comandos/AdministracionCarpetasArchivosPermisos"
	AD "Proyecto/Comandos/AdministracionDiscos"
	AR "Proyecto/Comandos/AdministracionReportes"
	AUG "Proyecto/Comandos/AdministracionUsuariosGrupos"
	globales "Proyecto/Globales"
	"Proyecto/Structs"
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/rs/cors"
)

var (
	bandera string
)

type Entrada struct {
	Text string `json:"text"`
}

type Salida struct {
	Text string `json:"textsalida"`
}

type StatusResponse struct {
	Message string `json:"message"`
	Type    string `json:"type"`
}

type LoginRequest struct {
	NombreUsuario string `json:"nombreUsuario"`
	Contrasenia   string `json:"contrasenia"`
	IdUsuario     string `json:"idUsuario"`
}

type LogoutRequest struct {
	Comando string `json:"comando"`
}

type Discos struct {
	Nombre      string      `json:"nombre"`
	Tamaño      int64       `json:"tamaño"`
	Fit         string      `json:"fit"`
	Particiones []Particion `json:"particiones"`
}

type Particion struct {
	Nombre    string `json:"nombre"`
	Id        string `json:"id"`
	Tamaño    int64  `json:"tamaño"`
	Fit       string `json:"fit"`
	Tipo      string `json:"tipo"`
	Estado    string `json:"estado"`
	IsMounted bool   `json:"ismounted"`
}

type Respuesta struct {
	DiscosCreados []Discos `json:"discoscreados"`
	Error         string   `json:"error,omitempty"`
}

type PartInicioSesionRequest struct {
	IdParticion string `json:"idParticion"`
}

type DirectoryTreeResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Tree    interface{} `json:"tree,omitempty"`
	Type    string      `json:"type,omitempty"`
}

func main() {
	//Endpoints
	http.HandleFunc("/analizar", getCadenaAnalizar)
	http.HandleFunc("/validarInicioSesion", validarInicioSesion)
	http.HandleFunc("/validarCerrarSesion", validarCerrarSesion)
	http.HandleFunc("/obtenerDiscos", obtenerDiscos)
	http.HandleFunc("/validarParticionEnInicioSesion", validarParticionEnInicioSesion)
	http.HandleFunc("/directory-tree", controladorSistemaArchivos)

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://127.0.0.1:5173", "http://localhost:5173"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
	})

	handler := c.Handler(http.DefaultServeMux)

	fmt.Println("Servidor del backend escuchando en http://localhost:8080")

	http.ListenAndServe(":8080", handler)
}

func getCadenaAnalizar(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "application/json")

	var respuesta string

	var status StatusResponse
	var salida Salida

	if r.Method == http.MethodPost {
		var entrada Entrada

		if err := json.NewDecoder(r.Body).Decode(&entrada); err != nil {
			http.Error(w, "Error al decodificar JSON", http.StatusBadRequest)

			status = StatusResponse{Message: "Error al decodificar JSON", Type: "unsuccess"}
			json.NewEncoder(w).Encode(status)
			return
		}

		//creo un lector de bufer para el archivo
		lector := bufio.NewScanner(strings.NewReader(entrada.Text))
		//leer la entrada linea por linea
		for lector.Scan() {
			//Elimina los saltos de linea
			if lector.Text() != "" {
				//Divido por # para ignorar todo lo que este a la derecha del mismo
				linea := strings.Split(lector.Text(), "#") //lector.Text() retorna la linea leida
				if len(linea[0]) != 0 {
					//fmt.Println("\n*********************************************************************************************")
					fmt.Println("Comando en ejecucion: ", linea[0])
					//respuesta += "***************************************************************************************************************************\n"
					respuesta += "Comando en ejecucion: " + linea[0] + "\n"
					respuesta += analizar(linea[0]) + "\n"
				}
				//Comentarios
				if len(linea) > 1 && linea[1] != "" {
					fmt.Println("#" + linea[1] + "\n")
					respuesta += "#" + linea[1] + "\n"
				}
			}
		}

		//fmt.Println("Cadena recibida ", entrada.Text)
		w.WriteHeader(http.StatusOK)

		salida = Salida{Text: respuesta}
		json.NewEncoder(w).Encode(salida)
	} else {
		status = StatusResponse{Message: "Metodo no permitido", Type: "unsuccess"}
		json.NewEncoder(w).Encode(status)
	}
}

func analizar(entrada string) string {
	respuesta := ""

	if strings.Contains(entrada, "#") {
		fmt.Println(entrada)
	} else if bandera == "true" || bandera == "" {
		tmp := strings.TrimRight(entrada, " ")
		parametros := strings.Split(tmp, " -")

		if strings.ToLower(parametros[0]) == "mkdisk" {
			if len(parametros) > 1 {
				respuesta += AD.Mkdisk(parametros)
			} else {
				fmt.Println("MKDISK ERROR: parametros no validos para el comando.")
				respuesta += "MKDISK ERROR: parametros no validos para el comando."
			}

		} else if strings.ToLower(parametros[0]) == "rmdisk" {
			if len(parametros) > 0 {
				respuesta += AD.Rmdisk(parametros)
			} else {
				fmt.Println("RMDISK ERROR: parametros no validos para el comando.")
				respuesta += "RMDISK ERROR: parametros no validos para el comando."
			}
		} else if strings.ToLower(parametros[0]) == "fdisk" {
			if len(parametros) > 1 {
				respuesta += AD.Fdisk(parametros)
			} else {
				fmt.Println("FDISK ERROR: parametros no validos para el comando.")
				respuesta += "FDISK ERROR: parametros no validos para el comando."
			}
		} else if strings.ToLower(parametros[0]) == "mount" {
			if len(parametros) > 1 {
				respuesta += AD.Mount(parametros)
			} else {
				fmt.Println("MOUNT ERROR: parametros no validos para el comando.")
				respuesta += "MOUNT ERROR: parametros no validos para el comando."
			}
		} else if strings.ToLower(parametros[0]) == "mounted" {
			if len(parametros) == 1 {
				respuesta += AD.Mounted(parametros)
			} else {
				fmt.Println("MOUNTED ERROR: parametros no validos para el comando.")
				respuesta += "MOUNTED ERROR: parametros no validos para el comando."
			}
		} else if strings.ToLower(parametros[0]) == "unmount" {
			if len(parametros) == 2 {
				respuesta += AD.Unmount(parametros)
			} else {
				fmt.Println("UNMOUNT ERROR: parametros no validos para el comando.")
				respuesta += "UNMOUNT ERROR: parametros no validos para el comando."
			}
		} else if strings.ToLower(parametros[0]) == "mkfs" {
			if len(parametros) > 0 {
				respuesta += AA.Mkfs(parametros)
			} else {
				fmt.Println("MKFS ERROR: parametros no validos para el comando.")
				respuesta += "MKFS ERROR: parametros no validos para el comando."
			}
		} else if strings.ToLower(parametros[0]) == "login" {
			if len(parametros) > 2 {
				respuesta += AUG.Login(parametros)
			} else {
				fmt.Println("LOGIN ERROR: parametros no validos para el comando.")
				respuesta += "LOGIN ERROR: parametros no validos para el comando."
			}
		} else if strings.ToLower(parametros[0]) == "logout" {
			if len(parametros) == 1 {
				respuesta += AUG.Logout(parametros)
			} else {
				fmt.Println("LOGOUT ERROR: parametros no validos para el comando.")
				respuesta += "LOGOUT ERROR: parametros no validos para el comando."
			}
		} else if strings.ToLower(parametros[0]) == "mkgrp" {
			if len(parametros) > 1 {
				respuesta += AUG.Mkgrp(parametros)
			} else {
				fmt.Println("MKGRP ERROR: parametros no validos para el comando.")
				respuesta += "MKGRP ERROR: parametros no validos para el comando."
			}
		} else if strings.ToLower(parametros[0]) == "rmgrp" {
			if len(parametros) > 1 {
				respuesta += AUG.Rmgrp(parametros)
			} else {
				fmt.Println("RMGRP ERROR: parametros no validos para el comando.")
				respuesta += "RMGRP ERROR: parametros no validos para el comando."
			}
		} else if strings.ToLower(parametros[0]) == "mkusr" {
			if len(parametros) > 3 {
				respuesta += AUG.Mkusr(parametros)
			} else {
				fmt.Println("MKUSR ERROR: parametros no validos para el comando.")
				respuesta += "MKUSR ERROR: parametros no validos para el comando."
			}
		} else if strings.ToLower(parametros[0]) == "rmusr" {
			if len(parametros) > 1 {
				respuesta += AUG.Rmusr(parametros)
			} else {
				fmt.Println("RMUSR ERROR: parametros no validos para el comando.")
				respuesta += "RMUSR ERROR: parametros no validos para el comando."
			}
		} else if strings.ToLower(parametros[0]) == "chgrp" {
			if len(parametros) > 2 {
				respuesta += AUG.Chgrp(parametros)
			} else {
				fmt.Println("CHGRP ERROR: parametros no validos para el comando.")
				respuesta += "CHGRP ERROR: parametros no validos para el comando."
			}
		} else if strings.ToLower(parametros[0]) == "mkfile" {
			if len(parametros) > 1 {
				respuesta += ACAP.Mkfile(parametros)
			} else {
				fmt.Println("MKFILE ERROR: parametros no validos para el comando.")
				respuesta += "MKFILE ERROR: parametros no validos para el comando."
			}
		} else if strings.ToLower(parametros[0]) == "mkdir" {
			if len(parametros) > 1 {
				respuesta += ACAP.Mkdir(parametros)
			} else {
				fmt.Println("MKDIR ERROR: parametros no validos para el comando.")
				respuesta += "MKDIR ERROR: parametros no validos para el comando."
			}
		} else if strings.ToLower(parametros[0]) == "cat" {
			if len(parametros) > 1 {
				respuesta += AA.Cat(parametros)
			} else {
				fmt.Println("CAT ERROR: parametros no validos para el comando.")
				respuesta += "CAT ERROR: parametros no validos para el comando."
			}
		} else if strings.ToLower(parametros[0]) == "rep" {
			if len(parametros) > 2 {
				respuesta += AR.Rep(parametros)
			} else {
				fmt.Println("REP ERROR: parametros no validos para el comando.")
				respuesta += "REP ERROR: parametros no validos para el comando."
			}
		} else if strings.ToLower(parametros[0]) == "remove" {
			if len(parametros) > 0 {
				respuesta += ACAP.Remove(parametros)
			} else {
				fmt.Println("REMOVE ERROR: parametros no validos para el comando.")
				respuesta += "REMOVE ERROR: parametros no validos para el comando."
			}
		} else if strings.ToLower(parametros[0]) == "edit" {
			if len(parametros) > 0 {
				respuesta += ACAP.Edit(parametros)
			} else {
				fmt.Println("EDIT ERROR: parametros no validos para el comando.")
				respuesta += "EDIT ERROR: parametros no validos para el comando."
			}
		} else if strings.ToLower(parametros[0]) == "rename" {
			if len(parametros) > 0 {
				respuesta += ACAP.Rename(parametros)
			} else {
				fmt.Println("RENAME ERROR: parametros no validos para el comando.")
				respuesta += "RENAME ERROR: parametros no validos para el comando."
			}
		} else if strings.ToLower(parametros[0]) == "copy" {
			if len(parametros) > 1 {
				respuesta += ACAP.Copy(parametros)
			} else {
				fmt.Println("COPY ERROR: parametros no validos para el comando.")
				respuesta += "COPY ERROR: parametros no validos para el comando."
			}
		} else if strings.ToLower(parametros[0]) == "find" {
			if len(parametros) > 1 {
				respuesta += ACAP.Find(parametros)
			} else {
				fmt.Println("FIND ERROR: parametros no validos para el comando.")
				respuesta += "FIND ERROR: parametros no validos para el comando."
			}
		} else if strings.ToLower(parametros[0]) == "move" {
			if len(parametros) > 1 {
				respuesta += ACAP.Move(parametros)
			} else {
				fmt.Println("MOVE ERROR: parametros no validos para el comando.")
				respuesta += "MOVE ERROR: parametros no validos para el comando."
			}
		} else if strings.ToLower(parametros[0]) == "chown" {
			if len(parametros) > 1 {
				respuesta += ACAP.Chown(parametros)
			} else {
				fmt.Println("CHOWN ERROR: parametros no validos para el comando.")
				respuesta += "CHOWN ERROR: parametros no validos para el comando."
			}
		} else if strings.ToLower(parametros[0]) == "chmod" {
			if len(parametros) > 1 {
				respuesta += ACAP.Chmod(parametros)
			} else {
				fmt.Println("CHMOD ERROR: parametros no validos para el comando.")
				respuesta += "CHMOD ERROR: parametros no validos para el comando."
			}
		} else if strings.ToLower(parametros[0]) == "journaling" {
			if len(parametros) > 1 {
				respuesta += AA.Journaling(parametros)
			} else {
				fmt.Println("JOURNALING ERROR: parametros no validos para el comando.")
				respuesta += "JOURNALING ERROR: parametros no validos para el comando."
			}
		} else if strings.ToLower(parametros[0]) == "loss" {
			if len(parametros) > 1 {
				respuesta += AA.Loss(parametros)
				bandera = "false"
			} else {
				fmt.Println("LOSS ERROR: parametros no validos para el comando.")
				respuesta += "LOSS ERROR: parametros no validos para el comando."
			}
		} else if strings.ToLower(parametros[0]) == "recovery" {
			if len(parametros) > 1 {
				bandera = "true"
				respuesta += AA.Recovery(parametros)
			} else {
				fmt.Println("RECOVERY ERROR: parametros no validos para el comando.")
				respuesta += "RECOVERY ERROR: parametros no validos para el comando."
			}
		} else if strings.ToLower(parametros[0]) == "exit" {
			fmt.Println("Salida exitosa")
			os.Exit(0)

		} else if strings.ToLower(parametros[0]) == "" {

		} else {
			fmt.Println("Comando no reconocible")
		}
	} else if bandera == "false" {
		tmp := strings.TrimRight(entrada, " ")
		parametros := strings.Split(tmp, " -")
		if strings.ToLower(parametros[0]) == "recovery" {
			if len(parametros) > 1 {
				bandera = "true"
				respuesta += AA.Recovery(parametros)
			} else {
				fmt.Println("RECOVERY ERROR: parametros no validos para el comando.")
				respuesta += "RECOVERY ERROR: parametros no validos para el comando."
			}
		}
	}
	return respuesta
}

func validarInicioSesion(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "application/json")

	if r.Method == http.MethodPost && (bandera == "" || bandera == "true") {
		var req LoginRequest

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Error al decodificar JSON", http.StatusBadRequest)
			json.NewEncoder(w).Encode(StatusResponse{
				Message: "Invalid JSON format",
				Type:    "error",
			})
			return
		}

		// Se construye el comando login para que funcione
		comandoLogin := fmt.Sprintf("login -user=%s -pass=%s -id=%s", req.NombreUsuario, req.Contrasenia, req.IdUsuario)

		// Se ejecuta el comando
		resultado := analizar(comandoLogin)

		// Se determina el resultado basado en la respuesta del comando
		var respuesta StatusResponse
		if strings.Contains(strings.ToLower(resultado), "exitoso") || strings.Contains(strings.ToLower(resultado), "correcto") {
			respuesta = StatusResponse{
				Message: "Login exitoso",
				Type:    "success",
			}
		} else {
			respuesta = StatusResponse{
				Message: strings.TrimSpace(resultado),
				Type:    "error",
			}
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(respuesta)
	} else if bandera == "false" {
		json.NewEncoder(w).Encode(StatusResponse{
			Message: "No se pudo iniciar sesión",
			Type:    "error",
		})
	} else {
		json.NewEncoder(w).Encode(StatusResponse{
			Message: "Método no permitido",
			Type:    "error",
		})
	}
}

func validarCerrarSesion(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "application/json")

	if r.Method == http.MethodPost && (bandera == "" || bandera == "true") {
		var req LogoutRequest

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Error al decodificar JSON", http.StatusBadRequest)
			json.NewEncoder(w).Encode(StatusResponse{
				Message: "Invalid JSON format",
				Type:    "error",
			})
			return
		}

		// Se construye el comando logout para que funcione
		comandoLogout := req.Comando

		// Se ejecuta el comando
		resultado := analizar(comandoLogout)

		// Se determina el resultado basado en la respuesta del comando
		var respuesta StatusResponse
		if strings.Contains(strings.ToLower(resultado), "exitoso") || strings.Contains(strings.ToLower(resultado), "correcto") {
			respuesta = StatusResponse{
				Message: "Logout exitoso",
				Type:    "success",
			}
		} else {
			respuesta = StatusResponse{
				Message: strings.TrimSpace(resultado),
				Type:    "error",
			}
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(respuesta)
	} else if bandera == "false" {
		json.NewEncoder(w).Encode(StatusResponse{
			Message: "No se pudo cerrar sesión",
			Type:    "error",
		})
	} else {
		json.NewEncoder(w).Encode(StatusResponse{
			Message: "Método no permitido",
			Type:    "error",
		})
	}
}

func obtenerDiscos(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// Se valida que sea un GET
	if r.Method != http.MethodGet && (bandera == "" || bandera == "true") {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	} else if bandera == "false" {
		http.Error(w, "No se pudieron obtener los discos", http.StatusMethodNotAllowed)
		return
	}

	carpeta := "/home/ubuntu/MIA_2S2025_PROYECTO2_202300722/Calificacion_MIA/Discos"
	var respuesta Respuesta

	// Se verifica si la carpeta existe
	if _, err := os.Stat(carpeta); os.IsNotExist(err) {
		fmt.Println("La carpeta no existe")
		respuesta.Error = "La carpeta no existe"
		json.NewEncoder(w).Encode(respuesta)
		return
	}

	// Se leen los archivos de la carpeta
	archivos, err := os.ReadDir(carpeta)
	if err != nil {
		fmt.Println("Error al leer la carpeta")
		respuesta.Error = "Error al leer la carpeta: " + err.Error()
		json.NewEncoder(w).Encode(respuesta)
		return
	}

	// Se procesa la información de cada archivo
	for _, archivo := range archivos {
		if !archivo.IsDir() {
			nombre := archivo.Name()

			// Se abre el archivo para obtener *os.File
			archivoAbierto, err := os.Open(filepath.Join(carpeta, nombre))
			if err != nil {
				fmt.Printf("Error abriendo archivo %s: %v\n", nombre, err)
				continue
			}
			defer archivoAbierto.Close()

			var mbr Structs.MBR

			err = mbr.Decodificar(archivoAbierto)

			if err != nil {
				fmt.Println("Error al deserializar el MBR.")
			}

			fitDisco := mbr.Fit[:]
			fitLimpio := string(bytes.TrimRight(fitDisco, "\x00"))

			var particionesFormateadas []Particion
			// Se validan que las particiones sean las que se formatearon
			for i := range mbr.Particiones {
				nombreParticion := mbr.Particiones[i].Nombre[:]
				nombreLimpio := string(bytes.TrimRight(nombreParticion, "\x00"))

				idParticion := mbr.Particiones[i].Id[:]
				idLimpio := string(bytes.TrimRight(idParticion, "\x00"))

				tipoParticion := mbr.Particiones[i].Tipo[:]
				tipoLimpio := string(bytes.TrimRight(tipoParticion, "\x00"))

				fitParticion := mbr.Particiones[i].Fit[:]
				fitLimpio := string(bytes.TrimRight(fitParticion, "\x00"))

				encontrado := Structs.ObtenerParticionesFormateadas(nombreLimpio)

				if encontrado {
					particionInfo := Particion{
						Nombre:    nombreLimpio,
						Id:        idLimpio,
						Tamaño:    int64(mbr.Particiones[i].Tamanio),
						Fit:       fitLimpio,
						Tipo:      tipoLimpio,
						Estado:    "Formateada",
						IsMounted: true,
					}

					particionesFormateadas = append(particionesFormateadas, particionInfo)
				}
			}

			/*fmt.Println(nombre)
			fmt.Println(int64(mbr.MbrSize))
			fmt.Println(fitLimpio)*/

			discoInfo := Discos{
				Nombre:      nombre,
				Tamaño:      int64(mbr.MbrSize),
				Fit:         fitLimpio,
				Particiones: particionesFormateadas,
			}

			respuesta.DiscosCreados = append(respuesta.DiscosCreados, discoInfo)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(respuesta)
}

func validarParticionEnInicioSesion(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "application/json")

	if r.Method == http.MethodPost {
		var req PartInicioSesionRequest

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Error al decodificar JSON", http.StatusBadRequest)
			json.NewEncoder(w).Encode(StatusResponse{
				Message: "Invalid JSON format",
				Type:    "error",
			})
			return
		}

		idParticion := req.IdParticion

		resultado := ""
		/*fmt.Printf("Partición seleccionada: '%s'\n", idParticion)
		fmt.Printf("Partición de globales: '%s'\n", globales.ParticionInicioSesion)*/

		if idParticion == globales.ParticionInicioSesion {
			resultado = "exitoso"
		} else {
			resultado = "No se inició sesión con esta partición"
		}

		// Se determina el resultado basado en si se inicio sesión con la particion indicada o no
		var respuesta StatusResponse
		if strings.Contains(strings.ToLower(resultado), "exitoso") || strings.Contains(strings.ToLower(resultado), "correcto") {
			respuesta = StatusResponse{
				Message: "Si se inició sesión con esta partición",
				Type:    "success",
			}
		} else {
			respuesta = StatusResponse{
				Message: strings.TrimSpace(resultado),
				Type:    "error",
			}
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(respuesta)
	} else {
		json.NewEncoder(w).Encode(StatusResponse{
			Message: "Método no permitido",
			Type:    "error",
		})
	}
}

func controladorSistemaArchivos(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "application/json")

	// Se valida que haya un inicio de sesión activo
	if !globales.HaIniciadoSesion() {
		json.NewEncoder(w).Encode(StatusResponse{
			Message: "No se ha iniciado sesión.",
			Type:    "error",
		})
		return
	}

	// Se crea el servicio de árbol de directorios
	dirService, err := ACAP.NewDirectoryTreeService()
	if err != nil {
		json.NewEncoder(w).Encode(StatusResponse{
			Message: "Error al acceder al sistema de archivos.",
			Type:    "error",
		})
		return
	}

	defer dirService.Close()

	// Obtener el árbol de directorios desde la raíz
	tree, err := dirService.GetDirectoryTree("/")
	if err != nil {
		json.NewEncoder(w).Encode(StatusResponse{
			Message: "Error al obtener el árbol del sistema de archivos.",
			Type:    "error",
		})
		return
	}

	json.NewEncoder(w).Encode(DirectoryTreeResponse{
		Success: true,
		Tree:    tree,
		Type:    "success",
	})
}
