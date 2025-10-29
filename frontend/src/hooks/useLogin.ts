import { useState } from "react";

export const useLogin = () => {
  const [loading, setLoading] = useState(false);
  const [backendMessage, setBackendMessage] = useState<string | null>(null); 
  const [messageType, setMessageType] = useState<"success" | "error" | "info" | "">("");

  const login = async (nombreUsuario: string, contrasenia: string, idUsuario: string): Promise<string> => {
    setLoading(true);
    setBackendMessage(null); 
    setMessageType("");

    try {

      const apiUrl = import.meta.env.VITE_API_URL ?? 'http://localhost:8080';
      console.log('Login -> calling API at', apiUrl);
      const response = await fetch(`${apiUrl}/validarInicioSesion`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        //body: JSON.stringify({ username: nombreUsuario, password: contrasenia, userId: idUsuario }),
        body: JSON.stringify({ nombreUsuario: nombreUsuario, contrasenia: contrasenia, idUsuario: idUsuario }),
      });

      if (response.status === 404) {
        setBackendMessage("Endpoint no encontrado (404)");
        setMessageType("error");
        return "not_found";
      }

      const data = await response.json(); 

      console.log(data.type)
      if (data.type === "error") {
        setBackendMessage(data.message);
        setMessageType("error");
        return "error"; 
      } else if (data.type === "success") {
        setBackendMessage(data.message);
        setMessageType("success");
        return "success"; 
      }

      return "error"; 
    } catch (error) {
      if (error instanceof Error) {
        setBackendMessage(`Error: ${error.message}`);
        setMessageType("error");
      } else {
        setBackendMessage("Ocurrió un error desconocido.");
        setMessageType("error");
      }
      return "error"; 
    } finally {
      setLoading(false);
    }
  };

  return {
    loading,
    backendMessage,
    messageType,
    login,
  };
};