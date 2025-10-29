import { useNavigate, useLocation } from "react-router-dom";
import { motion } from "framer-motion";
import { usePartitions } from "../hooks/usePartition";
import { ArrowLeft, HardDrive } from "lucide-react";
import FileSystemViewer from "../SistemaArchivos/SistemaArchivos";
import { useState } from "react";

interface Particion {
    nombre: string;
    id: string,
    tamaño: number;
    fit: string;
    tipo: string;
    estado: string;
    isMounted: boolean
}

export const Particiones = () => {
    const navigate = useNavigate();
    const location = useLocation();
    const booleano = true;

    const nombreDisco = location.state?.nombreDisco;
    const particionesDisco = location.state?.particionesDisco

    const [selectedPartition, setSelectedPartition] = useState<string | null>(null);
    const [showFileSystem, setShowFileSystem] = useState(false);

    const { validarParticion, respuesta, loading, error } = usePartitions();
    
    const regresarConsola = () => {
        navigate('/')
    }

    const regresarDiscos = () => {
        navigate('/Discos')
    }

    const validarParticionEnInicioSesion = async (indiceParticion: number) => {
        const particionSeleccionada = particionesDisco[indiceParticion];
        const idParticion = particionSeleccionada.id;
        const resultado = await validarParticion(idParticion);
        if (resultado === "success") {
          navigate('/Discos/Particiones/SistemaArchivos', {
            state: {
              partitionId: idParticion, 
              nombreDisco: nombreDisco,
              nombreParticion: particionSeleccionada.nombre
            }
          });
        }
    };

    return (<div className="min-vh-100 p-2">
                <div className="flex justify-between items-center mb-6">
                    <h2 className="h3 fw-bold text-dark mb-3">Particiones formateadas del disco</h2>

                    <motion.div
                        initial={{ opacity: 0, y: 20 }}
                        animate={{ opacity: 1, y: 0 }}
                        transition={{ duration: 0.4 }}
                        className="w-full max-w-3xl bg-light-50 p-8 rounded-2xl shadow-md border border-gray-200"
                    >

                    {particionesDisco.length > 0 ? (
                    <div className="space-y-4">
                        {particionesDisco.map((partition: Particion, index: number) => (
                        <motion.div
                            key={index}
                            whileHover={partition.isMounted ? { scale: 1.02 } : undefined}
                            whileTap={partition.isMounted ? { scale: 0.98 } : undefined}
                            onClick={() =>
                            validarParticionEnInicioSesion(index)
                            }
                            title={
                            partition.isMounted
                                ? "Ver árbol de partición"
                                : "La partición no está montada"
                            }
                            className={`p-4 border rounded-lg transition-all ${
                            partition.isMounted
                                ? "cursor-pointer bg-gray-50 hover:bg-gray-100 border-gray-300"
                                : "cursor-not-allowed bg-gray-100 border-gray-200 opacity-60"
                            }`}
                        >
                            <div className="flex items-center gap-3">
                            <HardDrive className="text-indigo-500" size={28} />
                            <div>
                                <h3 className="text-lg font-bold text-gray-800">
                                Partición {index + 1}: {partition.nombre}
                                </h3>
                                <p className="text-sm text-gray-600">Tamaño: {partition.tamaño}</p>
                                <p className="text-sm text-gray-600">Fit: {partition.fit}</p>
                                <p className="text-sm text-gray-600">Tipo: {partition.tipo}</p>
                                <p className="text-sm text-gray-600">Estado: {partition.estado}</p>
                            </div>
                            </div>
                        </motion.div>
                        ))}
                    </div>
                    ) : (
                    !booleano && (
                        <p className="text-gray-500">No se encontraron particiones.</p>
                    )
                    )}
                </motion.div>
    
                    <button
                        onClick={regresarConsola}
                        className="btn btn-success btn-lg fw-bold text-dark"
                        >
                        Regresar a Consola
                    </button>
                    <button
                        onClick={regresarDiscos}
                        className="btn btn-success btn-lg fw-bold text-dark"
                        >
                        Regresar a Discos
                    </button>
                </div>
    </div>);
}