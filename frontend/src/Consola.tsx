import { useState, useRef, useEffect } from 'react';
import { useNavigate, useLocation } from 'react-router-dom';
import { useLogin } from "./hooks/useLogin";
import { useLogout } from "./hooks/useLogout";
import { useAuth } from "./hooks/useAuth"
import { Discos } from "./Discos/Discos"
import { API_BASE_URL, API_ENDPOINTS } from './config/api';

function Consola() {
  const [text, setText] = useState('');
  const [textExit, setTextExit] = useState('');
  const [fileInfo, setFileInfo] = useState('');
  const fileInputRef = useRef<HTMLInputElement>(null);
  const textareaRef = useRef<HTMLTextAreaElement>(null);
  const { login, loading, backendMessage, messageType } = useLogin();
  const [nombreUsuario, setNombreUsuario] = useState("");
  const [contrasenia, setContrasenia] = useState("");
  const [idUsuario, setIDUsuario] = useState("");
  
  // Estados separados para alerts
  const [error, setError] = useState<string | null>(null);
  const [showModal, setShowModal] = useState(false);
  
  // Alert principal (para toda la aplicación)
  const [showMainAlert, setShowMainAlert] = useState(false);
  const [mainAlertMessage, setMainAlertMessage] = useState("");
  const [mainAlertType, setMainAlertType] = useState<"primary" | "success" | "danger" | "warning" | "info">("primary");
  
  // Alert del modal (solo para el modal de login)
  const [showModalAlert, setShowModalAlert] = useState(false);
  const [modalAlertMessage, setModalAlertMessage] = useState("");
  const [modalAlertType, setModalAlertType] = useState<"warning" | "danger" | "success">("warning");

  const { logout, loading: logoutLoading, messageType: logoutMessageType } = useLogout()
  const { isLogged } = useAuth()
  const navigate = useNavigate()
  const location = useLocation()

  // Efecto para mostrar alerts principales automáticamente desde el backend
  {/*useEffect(() => {
    if (backendMessage) {
      setMainAlertMessage(backendMessage);
      setMainAlertType(messageType === "success" ? "success" : "danger");
      setShowMainAlert(true);
      
      // Auto-ocultar después de 7 segundos
      const timer = setTimeout(() => {
        setShowMainAlert(false);
      }, 7000);
      
      return () => clearTimeout(timer);
    }
  }, [backendMessage, messageType]);*/}

  const timerAlertMainAlert = () => {
    const timer = setTimeout(() => {
        setShowMainAlert(false);
      }, 5000);
    return () => clearTimeout(timer);
  }

  const handleCloseMainAlert = () => {
    setShowMainAlert(false);
  };

  const handleCloseModalAlert = () => {
    setShowModalAlert(false);
  };

  const handleValidarIniciarSesion = (e: React.FormEvent) => {
    e.preventDefault();
    if (!nombreUsuario || !contrasenia || !idUsuario) {
      setModalAlertMessage("Todos los campos son obligatorios: ID de usuario, nombre de usuario y contraseña.");
      setModalAlertType("warning");
      setShowModalAlert(true);
      return;
    }
    setError(null);
    setShowModalAlert(false); // Ocultar alert del modal al iniciar validación
    
    login(nombreUsuario, contrasenia, idUsuario).then((result) => {
      if (result === "success") {
        setMainAlertMessage("Se ha iniciado sesión con éxito");
        setMainAlertType("success");
        setShowMainAlert(true);
        timerAlertMainAlert()

        setShowModal(false);
        
        setNombreUsuario("");
        setContrasenia("");
        setIDUsuario("");
      } else {
        // Error: usar alert del modal para mostrar error específico
        setModalAlertMessage("Hubo problemas al iniciar sesión. Checa eso XD");
        setModalAlertType("danger");
        setShowModalAlert(true);
      }
    }).catch((error) => {
      // Error de conexión: usar alert del modal
      setModalAlertMessage("Error de conexión. Intente nuevamente.");
      setModalAlertType("danger");
      setShowModalAlert(true);
    });
  };

  const handleCerrarSesion = () => {
    setError(null);
    logout().then((result) => {
      if (result === "success") {
        setMainAlertMessage("¡Sesión cerrada exitosamente!");
        setMainAlertType("success");
        setShowMainAlert(true);
      } else {
        setMainAlertMessage("No se pudo cerrar la sesión.");
        setMainAlertType("danger");
        setShowMainAlert(true);
      }
    }).catch((error) => {
      setMainAlertMessage("Error al cerrar sesión: " + error.message);
      setMainAlertType("danger");
      setShowMainAlert(true);
    });
  };

  const handleModalIniciarSesion = () => {
    setShowModal(true);
    setShowModalAlert(false);
  };

  const handleCloseModal = () => {
    setShowModal(false);
    setError(null);
    setShowModalAlert(false); // Limpiar alert del modal al cerrar
  };

  const handleTextChange = (e: React.ChangeEvent<HTMLTextAreaElement>) => {
    setText(e.target.value);
  };

  const clearTextArea = () => {
    setText('');
    setTextExit('');
    setFileInfo(''); 
    if (textareaRef.current) {
      textareaRef.current.focus();
    }
  };

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;

    if (!file.name.endsWith('.smia')) {
      setMainAlertMessage('Por favor, selecciona un archivo .smia');
      setMainAlertType("warning");
      setShowMainAlert(true);
      return;
    }

    setFileInfo(file.name);
    setMainAlertMessage(`Archivo "${file.name}" cargado correctamente`);
    setMainAlertType("success");
    setShowMainAlert(true);

    const reader = new FileReader();
    reader.onload = (event) => {
      const content = event.target?.result as string;
      setText(content);
      if (textareaRef.current) {
        textareaRef.current.focus();
      }
    };
    reader.readAsText(file);
  };

  const handleEnviarComando = async () => {
    if (!text.trim()) {
      setMainAlertMessage('Por favor ingresa algún texto');
      setMainAlertType("warning");
      setShowMainAlert(true);
      textareaRef.current?.focus();
      return;
    }

    try {
      console.log(API_ENDPOINTS.analizar)
      console.log(API_BASE_URL)  
      const response = await fetch( API_ENDPOINTS.analizar, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ text }),
      });

      if (!response.ok) {
        throw new Error(`Error HTTP: ${response.status}`);
      }

      const data = await response.json();
      console.log('Respuesta del servidor:', data);
      setText('')
      setTextExit(data.textsalida);
      
      // Mostrar alerta principal de éxito
      setMainAlertMessage("Comando ejecutado correctamente");
      setMainAlertType("success");
      setShowMainAlert(true);
      timerAlertMainAlert();
    } catch (error) {
      console.error('Error al enviar datos:', error);
      setMainAlertMessage('Error al conectar con el servidor');
      setMainAlertType("danger");
      setShowMainAlert(true);
      timerAlertMainAlert()
    }
  };

  const handleFileSelectClick = () => {
    fileInputRef.current?.click();
  };

  const handleDiscos = () => {
    navigate('/Discos')
  }
  return (
    <div className="min-vh-100 p-2">
      {/* Alert principal - se muestra en toda la aplicación */}
      {showMainAlert && (
        <div className={`alert alert-${mainAlertType} alert-dismissible fade show`} role="alert">
          {mainAlertMessage}
          <button 
            type="button" 
            className="btn-close" 
            onClick={handleCloseMainAlert}
            aria-label="Close"
          ></button>
        </div>
      )}

      {/* Contenedor de botones */}
      <div className="d-flex flex-wrap align-items-center gap-3 mb-3 p-1 bg-white rounded justify-content-between">
        
        <input
          type="file"
          ref={fileInputRef}
          onChange={handleFileChange}
          accept=".smia"
          className="d-none"
        />

        {/* Botón Seleccionar Archivo */}
        <button
          onClick={handleFileSelectClick}
          className="btn btn-success btn-lg fw-bold text-dark"
        >
          Seleccionar archivo
        </button>

        {/* Botón Enviar Comando */}
        <button
          onClick={handleEnviarComando}
          className="btn btn-success btn-lg fw-bold text-dark"
        >
          Enviar Comando
        </button>

        {/* Botón Limpiar */}
        <button
          onClick={clearTextArea}
          className="btn btn-success btn-lg fw-bold text-dark"
        >
          Limpiar
        </button>

        {/* Botón Iniciar Sesión */}
        <button 
          type="button" 
          onClick={handleModalIniciarSesion}
          className="btn btn-success btn-lg fw-bold text-dark" 
        >
          Iniciar Sesión
        </button>

        {/* Botón Cerrar Sesión */}
        <button
          onClick={handleCerrarSesion}
          className="btn btn-success btn-lg fw-bold text-dark"
          disabled={logoutLoading}
        >
          {logoutLoading ? "Cerrando..." : "Cerrar Sesión"}
        </button>

        {/* Botón Ver Sistema de Archivos */}
        <button
          onClick={handleDiscos}
          className="btn btn-success btn-lg fw-bold text-dark"
        >
          Ver Sistema de Archivos
        </button>
      </div>

      {/* Información del archivo seleccionado */}
      {fileInfo && (
        <div className="mb-3 text-center">
          <span className="badge bg-primary text-wrap">
            📁 Archivo seleccionado: {fileInfo}
          </span>
        </div>
      )}

      <div className="w-100 bg-white rounded shadow-sm p-2">
        <h2 className="h3 fw-bold text-dark mb-3">Entrada de Comandos</h2>
        
        <div className="mb-4">
          <textarea
            ref={textareaRef}
            value={text}
            onChange={(e) => setText(e.target.value)}
            className="form-control"
            style={{ minHeight: '205px', minWidth: '1400px', fontSize: '24px'}}
            placeholder="Escribe tus comandos aquí o carga un archivo .smia"
          />
        </div>

        <h2 className="h3 fw-bold text-dark mb-3">Salida de Comandos</h2>

        <div>
          <textarea
            value={textExit}
            readOnly
            className="form-control bg-light"
            style={{ minHeight: '205px', minWidth: '1400px', fontSize: '24px' }}
            placeholder="La salida de los comandos aparecerá aquí..."
          />
        </div>
      </div>

      {showModal && (
        <div className="modal show d-block" tabIndex={-1}>
          <div className="modal-dialog">
            <div className="modal-content">
              <div className="modal-header">
                <h1 className="modal-title fs-5">Inicio de Sesión</h1>
                <button type="button" className="btn-close" onClick={handleCloseModal}></button>
              </div>
              <div className="modal-body">
                <form>
                  <div className="mb-3">
                    <label htmlFor="nombreUsuario" className="col-form-label">Nombre de Usuario:</label>
                    <input type="text" className="form-control" id="nombreUsuario" name="nombreUsuario" value={nombreUsuario} onChange={(e) => setNombreUsuario(e.target.value)}/>
                  </div>
                  <div className="mb-3">
                    <label htmlFor="contrasenia" className="col-form-label">Contraseña:</label>
                    <input type="password" className="form-control" id="contrasenia" name="contrasenia" value={contrasenia} onChange={(e) => setContrasenia(e.target.value)}/>
                  </div>
                  <div className="mb-3">
                    <label htmlFor="idUsuario" className="col-form-label">ID de Partición:</label>
                    <input type="text" className="form-control" id="idUsuario" name="idUsuario" value={idUsuario} onChange={(e) => setIDUsuario(e.target.value)}/>
                  </div>
                </form>
                
                {/* Alert específico del modal */}
                {showModalAlert && (
                  <div className={`alert alert-${modalAlertType} alert-dismissible fade show mt-3`} role="alert">
                    {modalAlertMessage}
                    <button 
                      type="button" 
                      className="btn-close" 
                      onClick={handleCloseModalAlert}
                      aria-label="Close"
                    ></button>
                  </div>
                )}

              </div>
              <div className="modal-footer">
                <button type="button" className="btn btn-secondary" onClick={handleCloseModal}>Close</button>
                <button type="button" className="btn btn-success" onClick={handleValidarIniciarSesion}>
                  {loading ? "Validando..." : "Validar Datos"}
                </button>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}

export default Consola;