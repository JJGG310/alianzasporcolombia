import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';

// Sitio de emergencia: build simple y liviano.
//
// El corte de chunks lo decide el `import()` diferido de MapaAyuda en App.tsx:
// leaflet + react-leaflet salen solos a su propio archivo y no bloquean el
// primer pintado. No forzamos manualChunks porque agrupar los vendors mete
// React dentro del chunk del mapa y lo vuelve a hacer bloqueante.
export default defineConfig({
  plugins: [react()],
  build: {
    target: 'es2019',
    sourcemap: false,
  },
  server: {
    port: 8942,
  },
});
