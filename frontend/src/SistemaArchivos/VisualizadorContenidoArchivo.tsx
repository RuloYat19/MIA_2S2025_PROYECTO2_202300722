import { useState, useEffect } from 'react';
import { API_ENDPOINTS } from '../config/api';

interface FileContentViewerProps {
  partitionId: string;
  filePath: string;
  fileName: string;
  onClose: () => void;
  onRefresh?: () => void;
}

const FileContentViewer = ({ partitionId, filePath, fileName, onClose, onRefresh }: FileContentViewerProps) => {
  const [content, setContent] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const [isEditing, setIsEditing] = useState(false);
  const [editContent, setEditContent] = useState('');
  const [isSaving, setIsSaving] = useState(false);

  const loadContent = async () => {
    setIsLoading(true);
    try {
      // ✅ SOLUCIÓN: Enviar como JSON en lugar de texto plano
      const command = `cat -file=${filePath} -id=${partitionId}`;
      
      console.log('🔍 Comando enviado:', command);
      console.log('🔍 FilePath:', filePath);
      console.log('🔍 PartitionId:', partitionId);
      
      const response = await fetch(API_ENDPOINTS.analizar, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json', // ✅ Cambiar a JSON
        },
        body: JSON.stringify({ // ✅ Envolver en objeto JSON
          text: command
        }),
      });

      console.log('🔍 Status de respuesta:', response.status);

      // ✅ Leer la respuesta UNA sola vez
      const responseText = await response.text();
      console.log('🔍 Respuesta cruda:', responseText);

      // ✅ Verificar si hay error HTTP
      if (!response.ok) {
        let errorDetails = 'Error del servidor';
        try {
          const errorData = JSON.parse(responseText);
          errorDetails = errorData.message || errorData.error || JSON.stringify(errorData);
        } catch {
          errorDetails = responseText || `Error ${response.status}`;
        }
        throw new Error(errorDetails);
      }

      // ✅ Procesar respuesta exitosa
      let data;
      try {
        data = JSON.parse(responseText);
        console.log('🔍 Respuesta parseada:', data);
      } catch (parseError) {
        console.error('❌ Error parseando JSON:', parseError);
        throw new Error(`Respuesta inválida del servidor: ${responseText.substring(0, 100)}...`);
      }
      
      // ✅ Verificar estructura de respuesta esperada
      // Tu backend devuelve { text: "resultado" } o { textsalida: "resultado" }
      if (data.text || data.textsalida) {
        const fileContent = data.text || data.textsalida || 'Archivo vacío';
        setContent(fileContent);
        setEditContent(fileContent);
      } else {
        throw new Error('Formato de respuesta inesperado del servidor');
      }
    } catch (error) {
      console.error('❌ Error loading file content:', error);
      setContent(`Error al cargar el archivo: ${error instanceof Error ? error.message : 'Error desconocido'}`);
    } finally {
      setIsLoading(false);
    }
  };

  const saveContent = async () => {
    setIsSaving(true);
    try {
      // ✅ SOLUCIÓN: Enviar como JSON
      const command = `edit -path=${filePath} -cont=${editContent} -id=${partitionId}`;
      
      console.log('💾 Comando guardar:', command);
      
      const response = await fetch(API_ENDPOINTS.analizar, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json', // ✅ Cambiar a JSON
        },
        body: JSON.stringify({ // ✅ Envolver en objeto JSON
          text: command
        }),
      });

      console.log('💾 Status guardar:', response.status);

      // ✅ Leer la respuesta UNA sola vez
      const responseText = await response.text();

      if (!response.ok) {
        let errorDetails = 'Error del servidor';
        try {
          const errorData = JSON.parse(responseText);
          errorDetails = errorData.message || errorData.error || JSON.stringify(errorData);
        } catch {
          errorDetails = responseText || `Error ${response.status}`;
        }
        throw new Error(errorDetails);
      }

      // ✅ Procesar respuesta exitosa
      const data = JSON.parse(responseText);
      
      // Verificar éxito basado en la respuesta del comando
      if (data.text && (data.text.includes('exitoso') || data.text.includes('correcto'))) {
        setContent(editContent);
        setIsEditing(false);
        if (onRefresh) onRefresh();
        alert('Archivo guardado exitosamente');
      } else {
        const errorMessage = data.text || data.message || 'Error desconocido al guardar';
        throw new Error(errorMessage);
      }
    } catch (error) {
      console.error('❌ Error saving file content:', error);
      alert(`Error al guardar el archivo: ${error instanceof Error ? error.message : 'Error desconocido'}`);
    } finally {
      setIsSaving(false);
    }
  };

  useEffect(() => {
    loadContent();
  }, [filePath, partitionId]);

  return (
    <div className="modal-overlay">
      <div className="modal-content file-content-viewer">
        <div className="modal-header">
          <h3>📄 {fileName}</h3>
          <div className="file-actions">
            {!isEditing ? (
              <>
                <button
                  className="btn btn-sm btn-primary"
                  onClick={() => setIsEditing(true)}
                  disabled={isLoading}
                >
                  ✏️ Editar
                </button>
                <button
                  className="btn btn-sm btn-secondary"
                  onClick={loadContent}
                  disabled={isLoading}
                >
                  🔄 Recargar
                </button>
              </>
            ) : (
              <>
                <button
                  className="btn btn-sm btn-success"
                  onClick={saveContent}
                  disabled={isSaving}
                >
                  {isSaving ? 'Guardando...' : '💾 Guardar'}
                </button>
                <button
                  className="btn btn-sm btn-secondary"
                  onClick={() => {
                    setIsEditing(false);
                    setEditContent(content);
                  }}
                  disabled={isSaving}
                >
                  ❌ Cancelar
                </button>
              </>
            )}
            <button className="close-btn" onClick={onClose}>
              ×
            </button>
          </div>
        </div>

        <div className="file-content-body">
          <div className="file-info">
            <span className="file-path">📍 Ruta: {filePath}</span>
            <span className="file-partition">🆔 Partición: {partitionId}</span>
          </div>

          <div className="content-container">
            {isLoading ? (
              <div className="loading-content">
                <div className="loading-spinner">⏳</div>
                <p>Cargando contenido...</p>
              </div>
            ) : isEditing ? (
              <textarea
                className="content-editor"
                value={editContent}
                onChange={(e) => setEditContent(e.target.value)}
                placeholder="Escribe el contenido del archivo..."
                rows={20}
                disabled={isSaving}
              />
            ) : (
              <pre className="content-display">
                {content || 'Archivo vacío'}
              </pre>
            )}
          </div>
        </div>

        <div className="modal-footer">
          <div className="file-stats">
            <span>Caracteres: {(isEditing ? editContent : content).length}</span>
            <span>Líneas: {(isEditing ? editContent : content).split('\n').length}</span>
          </div>
        </div>
      </div>
    </div>
  );
};

export default FileContentViewer;