import { Routes, Route } from 'react-router-dom';
import Consola from './Consola'

import { Discos } from './Discos/Discos';
import { Particiones } from './Particiones/Particiones';
import FileSystemViewer from "./SistemaArchivos/SistemaArchivos";

function App() {
  return (
    <Routes>
      <Route path="/" element={<Consola />} />
      <Route path="/Discos" element={<Discos />} />
      <Route path="/Discos/Particiones" element={<Particiones />} />
      <Route path="/Discos/Particiones/SistemaArchivos" element={<FileSystemViewer />} />
    </Routes>
  );
}

export default App;