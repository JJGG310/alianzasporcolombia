import { StrictMode } from 'react';
import { createRoot } from 'react-dom/client';

import App from './App';

// El orden importa: primero los tokens (las variables), luego la base (reset y
// primitivas), y de últimas lo de cada componente.
import './estilos/tokens.css';
import './estilos/base.css';
import './estilos/marco.css';
import './estilos/lista.css';
import './estilos/app.css';

const raiz = document.getElementById('raiz');
if (!raiz) throw new Error('Falta el nodo #raiz en index.html');

createRoot(raiz).render(
  <StrictMode>
    <App />
  </StrictMode>,
);
