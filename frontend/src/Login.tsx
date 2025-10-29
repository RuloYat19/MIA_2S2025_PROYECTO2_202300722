import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useAuth } from './Autentificacion';

const Login: React.FC = () => {
  const { login } = useAuth();
  const [username, setUsername] = useState('grupo4_seccion_proy1');
  const [password, setPassword] = useState('HalaMadrid');
  const [error, setError] = useState<string | null>(null);
  const navigate = useNavigate();

  const submit = async (e: React.FormEvent) => {
    e.preventDefault();
    const ok = await login(username.trim(), password);
    if (ok) navigate('/');
    else setError('Credenciales inválidas');
  };

  const handleRegresarClick = () => {
    navigate('/');
  };

  return (
    <div className="min-vh-100 p-2 align-text-top">
      <div className="fixed-top rounded p-5">
        <h1 className="h3 text-center mb-4">Inicio de sesión</h1>
        
        {error && (
          <div className="alert alert-danger text-center">
            {error}
          </div>
        )}

        <form onSubmit={submit}>
          <div className="mb-3">
            <label className="form-label">Usuario</label>
            <input 
              value={username} 
              onChange={e => setUsername(e.target.value)} 
              className="form-control"
            />
          </div>

          <div className="mb-3">
            <label className="form-label">Contraseña</label>
            <input 
              type="password" 
              value={password} 
              onChange={e => setPassword(e.target.value)} 
              className="form-control"
            />
          </div>

          <button 
            type="submit"
            className="btn btn-primary w-100 mb-2"
          >
            Entrar
          </button>

          <button
            type="button"
            onClick={handleRegresarClick}
            className="btn btn-secondary w-100"
          >
            Regresar
          </button>
        </form>
      </div>
    </div>
  );
};

export default Login;