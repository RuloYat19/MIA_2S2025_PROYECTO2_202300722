import { useState, useCallback } from "react";
import { Partition } from "../types/partition";

export const usePartitions = () => {
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState<string | null>(null);
    const [respuesta, setRespuesta] = useState<string>("");

    const validarParticion = useCallback(async (idParticion: string) => {
        setLoading(true);
        setError(null);
        setRespuesta("");

        try {
            const apiUrl = import.meta.env.VITE_API_URL;
            console.log("Enviando solicitud con la partición:", idParticion);
            
            const response = await fetch(`${apiUrl}/validarParticionEnInicioSesion`, {
                method: "POST",
                headers: {
                    "Content-Type": "application/json",
                },
                body: JSON.stringify({ idParticion: idParticion }),
            });

            const data = await response.json();
            const resultado = data.type === "success" ? "success" : "error";
            
            setRespuesta(resultado);
            return resultado; // Retorna el resultado directamente
        } catch (err: any) {
            const errorMsg = err.message || "Error de conexión";
            setError(errorMsg);
            setRespuesta("error");
            return "error";
        } finally {
            setLoading(false);
        }
    }, []);

    return { 
        validarParticion, 
        respuesta, 
        loading, 
        error 
    };
};