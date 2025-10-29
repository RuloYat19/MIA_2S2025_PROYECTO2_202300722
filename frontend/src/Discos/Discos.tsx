import React, { useState, useEffect } from 'react';
import { useNavigate, useLocation } from 'react-router-dom';
import { motion } from "framer-motion";
import { HardDrive, X } from "lucide-react";

interface Discos {
  nombre: string;
  tamaño: number;
  fit: string;
  particiones: Particion[];
}

interface Particion {
    nombre: string;
    id: string,
    tamaño: number;
    fit: string;
    tipo: string;
    estado: string;
    isMounted: boolean
}

interface RespuestaAPI {
  discoscreados: Discos[];
  error?: string;
}

export const Discos: React.FC = () => {
    const [archivos, setArchivos] = useState<Discos[]>([]);
    const [cargando, setCargando] = useState(true);
    const [error, setError] = useState<string>('');

    const API_BASE_URL = 'http://localhost:8080/';

    useEffect(() => {
        obtenerArchivos();
    }, []);

    const obtenerArchivos = async () => {
        try {
        setCargando(true);
        setError('');
        
        const respuesta = await fetch(`${API_BASE_URL}/obtenerDiscos`);
        const datos: RespuestaAPI = await respuesta.json();
        
        if (!respuesta.ok) {
            throw new Error(datos.error || 'Error al obtener archivos');
        }
        
        setArchivos(datos.discoscreados);
        } catch (err) {
        setError(err instanceof Error ? err.message : 'Error de conexión');
        } finally {
        setCargando(false);
        }
    };

    const navigate = useNavigate()

    const regresarConsola = () => {
        navigate('/')
    }

    const handleViewPartitions = (diskIndex: number) => {
        const selectedDisk = archivos[diskIndex];
        navigate(`/Discos/Particiones`, {
        state: {
            nombreDisco: selectedDisk.nombre,
            particionesDisco: selectedDisk.particiones,
        },
        });
    };

  if (cargando) {
    return (
      <div className="flex justify-center items-center p-8">
        <div className="text-lg">Cargando archivos...</div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="p-4">
        <div className="bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded">
          <strong>Error:</strong> {error}
        </div>
        <button
          onClick={obtenerArchivos}
          className="mt-4 bg-blue-500 text-white px-4 py-2 rounded hover:bg-blue-600"
        >
          Reintentar
        </button>
      </div>
    );
  }

    return (
        <div className="min-vh-100 p-2">
            <div className="flex justify-between items-center mb-6">
                <h2 className="h3 fw-bold text-dark mb-3">Discos en el Sistema</h2>

                {archivos.length > 0 ? (
                    <div className="space-y-4">
                    {archivos.map((disk, index) => (
                        <motion.div
                        key={index}
                        whileHover={{ scale: 1.02 }}
                        whileTap={{ scale: 0.98 }}
                        onClick={() => handleViewPartitions(index)}
                        className="relative group cursor-pointer border border-gray-300 rounded-xl p-4 bg-gray-50 "
                        >

                        <div className="flex items-center space-x-3">
                            <motion.div
                            initial={{ rotate: 0 }}
                            animate={{ rotate: [0, 5, 0, -5, 0] }}
                            transition={{ duration: 0.5, delay: 0.2 }}
                            >
                            <HardDrive size={28} className="text-indigo-500" />
                            </motion.div>

                            <div>
                            <h3 className="text-lg font-semibold text-gray-800 ">
                                Disco {index + 1}: {disk.nombre}
                            </h3>
                            <p className="text-sm text-gray-600 ">{"Tamaño: " + disk.tamaño}</p>
                            <p className="text-sm text-gray-600 ">{"Fit: " + disk.fit}</p>
                            <p className="text-sm text-gray-600 ">
                                {"Particiones Formateadas: " + (disk.particiones?.map(p => p.nombre).join(', ') || 'Ninguna')}
                            </p>
                            </div>
                        </div>
                        </motion.div>
                    ))}
                    </div>
                ) : (
                    <p className="text-gray-500 ">No se han agregado discos aún.</p>
                )}
                
                <button
                    onClick={regresarConsola}
                    className="btn btn-success btn-lg fw-bold text-dark"
                    >
                    Regresar a Consola
                </button>
            </div>
        </div>
    );
}