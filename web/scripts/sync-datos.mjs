// Copia los datos de la raíz del repo a web/public/ para que Vite los sirva.
//
// La fuente de verdad NO vive en web/public/: vive en
//   ../public/datos/           (lo que mantiene el equipo a mano)
//   ../public/extracted-data/  (lo que genera el scraper en Go)
//
// Es la MISMA carpeta que despliega Cloudflare Pages para el sitio en vivo, así
// que los dos frontends leen exactamente los mismos archivos. Si algún día se
// decide que React es el sitio, esto deja de hacer falta.
//
// web/public/ es copia de trabajo y no se versiona. Este script corre solo en
// `npm run dev` y `npm run build`.
//
// Los curados son obligatorios; lo del scraper no: si nadie ha corrido
// `cd scraper && make correr` en este clon, la app arranca igual y muestra el
// aviso de "solo datos curados".
import { copyFileSync, existsSync, mkdirSync } from 'node:fs';
import { dirname, join } from 'node:path';
import { fileURLToPath } from 'node:url';

const aqui = dirname(fileURLToPath(import.meta.url));
const web = join(aqui, '..');
const raiz = join(web, '..');

const copias = [
  {
    de: join(raiz, 'public', 'datos', 'ayuda.json'),
    a: join(web, 'public', 'datos', 'ayuda.json'),
    obligatorio: true,
  },
  {
    de: join(raiz, 'public', 'extracted-data', 'puntos.json'),
    a: join(web, 'public', 'extracted-data', 'puntos.json'),
    obligatorio: false,
  },
  {
    // El pie de página lo enlaza como dato abierto: sin esta copia da 404.
    de: join(raiz, 'public', 'extracted-data', 'puntos.geojson'),
    a: join(web, 'public', 'extracted-data', 'puntos.geojson'),
    obligatorio: false,
  },
];

let faltoAlguno = false;

for (const { de, a, obligatorio } of copias) {
  if (!existsSync(de)) {
    if (obligatorio) {
      console.error(`[sync-datos] FALTA archivo obligatorio: ${de}`);
      faltoAlguno = true;
    } else {
      console.warn(`[sync-datos] sin ${de} — se conserva la muestra de ${a}`);
    }
    continue;
  }
  mkdirSync(dirname(a), { recursive: true });
  copyFileSync(de, a);
  console.log(`[sync-datos] ${de} -> ${a}`);
}

if (faltoAlguno) process.exit(1);
