// TEMPORAL — banco de pruebas de los tres componentes de paridad. Se borra.
import { StrictMode } from 'react';
import { createRoot } from 'react-dom/client';

import Avisos from './componentes/Avisos';
import Necesidades from './componentes/Necesidades';
import BuscarAlguien from './componentes/BuscarAlguien';

import './estilos/tokens.css';
import './estilos/base.css';
import './estilos/marco.css';
import './estilos/lista.css';
import './estilos/app.css';

const raiz = document.getElementById('raiz')!;
createRoot(raiz).render(
  <StrictMode>
    <div className="app">
      <main className="app__cuerpo">
        <div className="contenedor">
          <Avisos />
          <Necesidades />
          <BuscarAlguien />
        </div>
      </main>
    </div>
  </StrictMode>,
);
