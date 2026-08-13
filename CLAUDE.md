# Orientación para agentes

Léelo antes de tocar código. Es el mapa mínimo para no romper el sitio, que está
en vivo en alianzasporcolombia.com durante una emergencia real.

## El frontend: React, en `web/`

**El sitio es `web/` (Vite + React 18 + TypeScript). Es lo único que se
despliega.** Hubo un sitio vanilla en `public/index.html`; se retiró el 12-ago
cuando React quedó a la par. Si vas a cambiar la página, es en `web/src/`.

- No queda `public/index.html`, `public/recursos.js` ni `public/lite.html`.
  Están en el historial de git (`git show <commit>~1:public/index.html`).
- El directorio de enlaces vive en `web/src/data/recursos.ts`, no en un `.js`.

## `public/` sigue existiendo, pero es SOLO datos

Aunque ya no tiene página, `public/` se queda porque el build de React copia sus
datos. **No borres nada de acá o el sitio queda mudo:**

- `public/datos/` — lo que el equipo mantiene a mano (ayuda, avisos, necesidades)
- `public/extracted-data/` — lo que genera el scraper (no se edita a mano)
- `public/_headers` — reglas de caché y CORS, se copian al build

`web/scripts/sync-datos.mjs` los copia a `web/dist/` en cada `dev` y `build`.
Por eso el build corre **desde la raíz del repo, no desde `web/`**.

## Deploy

Cada push a `main` dispara `.github/workflows/deploy.yml`: construye `web/`,
copia `public/_headers`, y publica `web/dist` en Cloudflare Pages. **Falla a
propósito si algún archivo de datos no quedó en el build** — mejor la versión
anterior que un sitio sin datos.

## Dos orígenes de datos, distintos a propósito

- **Curado** (`public/datos/ayuda.json`, `fuente:"curado"`, `verificado:true`):
  lo confirmó una persona del equipo.
- **Agregado** (`public/extracted-data/`, `verificado:false`): lo saca el
  scraper de otros sitios ciudadanos, siempre con enlace a su fuente.

El sitio los muestra diferenciados. La confianza es el producto: el usuario tiene
que poder distinguir a un vistazo lo verificado de lo agregado. No los mezcles.

## Reglas que están en el código, no solo en el README

- **El rojo (`--urgente`) significa URGENTE y nada más.** No es la marca, ni los
  bordes, ni los títulos, ni lo interactivo (eso es tinta, `--accion`). Todo el
  color sale de `web/src/estilos/tokens.css`; no escribas colores literales.
- **Cero marcas de front generado con IA**: nada de barras de color al costado
  de las tarjetas, degradados, puntos de color decorativos ni animaciones en
  bucle infinito. Se quitaron a propósito; no las reintroduzcas.
- **Nunca se publican datos de pago** (cuentas, Nequi/Daviplata, recaudo,
  cripto). El scraper los detecta y redacta; un punto con `politica_alerta` no se
  publica sin revisión humana.
- **No se agregan datos de personas desaparecidas de ninguna fuente**
  (principio #2, Ley 1581). Los adaptardores que tienen esa colección la saltan.
  Se enlaza a Colombia Te Busca / Encontrados, no se copia.

## El scraper (`scraper/`, Go)

Un binario sin dependencias. `cd scraper && make correr` lee los curados, agrega
9 sitios ciudadanos, fusiona duplicados y escribe `extracted-data/` +
`public/extracted-data/`. La salida es determinista: el diff de git es el control
de calidad. Si una fuente se rompe, `go run ./cmd/husmear <url>` dice de dónde
saca sus datos ahora.

## Flujo git

`main` se despliega solo. Cambios chicos directo a `main`; `git pull --rebase`
antes de empezar porque el equipo trabaja en paralelo. Commits en español,
cortos. Ver `TRABAJO.md` para el reparto por persona y `README.md` para el
detalle de cada parte.
