import { useState } from "react";
import { useAuth } from "./useAuth"; 
import { useDisksStore } from "./useDiskStore"; 

export const useLogout = () => {
  const [loading, setLoading] = useState(false); 
  const [backendMessage, setBackendMessage] = useState<string | null>(null); 
  const [messageType, setMessageType] = useState<"success" | "error" | "info" | "">(""); 
  const { Logout } = useAuth(); 
  const { clearDisks } = useDisksStore(); 

  const logout = async () => {
    setLoading(true);
    setBackendMessage(null); 
    setMessageType(""); 
    try {
      // Realizar la solicitud al backend para enviar el comando logout
      const apiUrl = import.meta.env.VITE_API_URL;
      console.log('Logout -> calling API at', apiUrl);
      const response = await fetch(`${apiUrl}/validarCerrarSesion`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({ Comando: "logout" }), 
      });

      const data = await response.json();
      console.log(data.type)
      if (data.type === "success") {
        setBackendMessage(data.message);
        setMessageType("success");
        //Logout(); 
        //clearDisks();
        return "success" 
      } else {
        setBackendMessage("Error: No se pudo cerrar la sesión.");
        setMessageType("error");
        return "fail"
      }
    } catch (error) {
      if (error instanceof Error) {
        setBackendMessage(`Error: ${error.message}`);
        setMessageType("error");
      } else {
        setBackendMessage("Ocurrió un error desconocido.");
        setMessageType("error");
      }
    } finally {
      setLoading(false);
    }
  };

  return {
    loading,        
    backendMessage, 
    messageType,    
    logout,         
  };
};