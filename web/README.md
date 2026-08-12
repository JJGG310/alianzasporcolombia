# web/ — Alianzas por Colombia en React

App de [alianzasporcolombia.com](https://alianzasporcolombia.com): el mismo sitio ciudadano
del terremoto del 10 de agosto de 2026, migrado del `index.html` vanilla de la raíz a
**Vite + React 18 + TypeScript**.

Sin framework de UI, sin Tailwind, sin backend. CSS plano que porta la identidad visual
del sitio original. Mapa con `react-leaflet` sobre OpenStreetMap (sin API key).

> El `index.html`, `recursos.js` y `datos/` de la raíz siguen ahí a propósito: son el sitio
> en producción hasta que esta migración se valide. No los borres.

## Correr en desarrollo

```bash
cd web
npm install
npm run dev      # http://localhost:8942
```

`npm run dev` corre antes `npm run sync-datos`, que copia los datos de la raíz del repo a
`public/` (ver abajo). Sin ese paso el mapa sale vacío.

## Construir y previsualizar

```bash
npm run build     # sync-datos + tsc --noEmit + vite build  →  dist/
npm run preview   # sirve dist/ en http://localhost:4173
npm run typecheck # solo tsc --noEmit
```

El build **falla** si TypeScript encuentra un error: `tsc --noEmit` corre antes de Vite a
propósito.

## De dónde salen los datos

La app consume **dos** archivos JSON estáticos, en paralelo, desde el navegador:

| Ruta pública | Qué es | Quién lo genera |
|---|---|---|
| `/datos/ayuda.json` | Los 91 puntos curados y verificados a mano | Humanos, editando `datos/ayuda.json` en la raíz |
| `/extracted-data/puntos.json` | Los puntos agregados automáticamente de otros sitios ciudadanos | El scraper en Go (`scraper/`), que escribe `extracted-data/puntos.json` en la raíz |

**La fuente de verdad NO vive en `web/public/`.** Vive en la raíz del repo. El script
`scripts/sync-datos.mjs` (que corre solo en `dev` y `build`) las copia:

```
../datos/ayuda.json            →  web/public/datos/ayuda.json          (ignorado por git)
../extracted-data/puntos.json  →  web/public/extracted-data/puntos.json (si existe)
```

### El archivo de muestra

`web/public/extracted-data/puntos.json` está versionado con **5 puntos ficticios marcados
`[MUESTRA]`** y `"muestra": true`, solo para que `npm run dev` funcione antes de que el
scraper corra por primera vez. Cuando el archivo real existe en la raíz, `sync-datos` lo
sobrescribe. **En producción ese archivo lo genera el scraper, nunca se escribe a mano.**
Para volver a la muestra: `git checkout web/public/extracted-data/puntos.json`.

Mientras la app detecta la muestra (o no encuentra el archivo del todo), muestra un aviso
discreto y sigue funcionando solo con los 91 curados. **Un 404 en
`/extracted-data/puntos.json` nunca rompe el sitio.**

### Esquema

`src/types.ts` refleja **exactamente** `model.Punto` de
`scraper/internal/model/punto.go`, con los mismos nombres de campo JSON. Si ese archivo
cambia en Go, hay que cambiar `src/types.ts`.

Los 91 curados se normalizan al mismo tipo en `src/data/carga.ts`:
`que` → `descripcion`, `fuente` → `fuente_url`, `fuente = "curado"`, `verificado = true`.
Si el scraper trae un punto con el mismo nombre y municipio que uno curado, **gana el
curado** y el agregado se descarta.

## Reglas del proyecto que están escritas en el código

1. **Nunca cuentas bancarias ni datos de pago.** Aviso fijo en la sección de donaciones
   (`Directorio.tsx`), en la descripción del mapa y en el pie.
2. **Todo punto lleva fuente + fecha de corte + estado**, en la tarjeta, en la fila de la
   tabla y en el popup del mapa. No hay vista donde falte.
3. **Verificado a mano ≠ agregado automáticamente.** Sello verde `✓ verificado a mano` y
   borde izquierdo sólido para los curados; sello gris `⟳ agregado automáticamente`, borde
   punteado y relleno tenue en el mapa para los del scraper. Hay filtro
   *Solo verificados a mano*.
4. **`politica_alerta` no se publica.** Un punto con alertas del saneamiento (posible dato
   bancario) sale del listado público y solo aparece con el filtro *Ver los que requieren
   revisión*, apagado por defecto, con su motivo a la vista y la advertencia de no
   compartirlo.
5. **Los puntos sin coordenadas no se esconden.** Hoy son mayoría (73 de 96). Salen en la
   tabla/tarjetas igual, y el mapa dice cuántos quedaron fuera.
6. **Compartir por WhatsApp** en cada punto (`https://wa.me/?text=…`), con nombre,
   dirección, qué necesita/ofrece, contacto, estado, fecha de corte, si está verificado o
   no, y el enlace a la fuente.

## Accesibilidad y peso

- HTML semántico (`header`, `main`, `section` con `aria-labelledby`, `table` con
  `caption` y `th scope`), enlace *saltar al mapa*, `aria-label` en las líneas de
  emergencia y en cada botón de compartir, `aria-live` en el contador de resultados.
- Foco visible en todo elemento interactivo; objetivos táctiles de 40 px mínimo.
- Mobile-first: en pantallas < 900 px se renderizan tarjetas; en escritorio, la tabla
  densa. Solo se pinta una de las dos.
- Se muestran 60 puntos a la vez con botón «mostrar más», para no bloquear celulares
  gama baja cuando el scraper traiga 700+.
- Leaflet (157 kB) se carga con `import()` diferido: no bloquea el primer pintado.
- `Replicas.tsx` consulta el feed público de USGS **desde el navegador del visitante**
  (CORS abierto, sin auth), filtrado a la caja de Colombia (lat −5..14, lon −82..−66). Si
  el feed falla, el widget desaparece en vez de mostrar datos viejos.

## Estructura

```
index.html                 esqueleto + meta/OG
vite.config.ts
scripts/sync-datos.mjs     copia los datos de la raíz a public/
public/
  favicon.svg
  datos/ayuda.json               ← generado por sync-datos (gitignored)
  extracted-data/puntos.json     ← MUESTRA versionada; en prod la escribe el scraper
src/
  main.tsx  App.tsx  types.ts
  data/
    carga.ts       fetch en paralelo, fusión, dedup, orden, cola de revisión
    catalogo.ts    etiquetas, colores, fechas, texto de WhatsApp, búsqueda
    recursos.ts    el directorio de enlaces (puerto TS de recursos.js)
  componentes/
    Encabezado.tsx  LineasEmergencia.tsx  Replicas.tsx
    Filtros.tsx     MapaAyuda.tsx (diferido)  TablaPuntos.tsx  TarjetaPunto.tsx
    Directorio.tsx  PieDePagina.tsx
  estilos/
    global.css     base, encabezado, líneas, directorio, pie
    ayuda.css      filtros, mapa, tarjetas, tabla, réplicas
```

Para actualizar el directorio de enlaces se edita `src/data/recursos.ts`, igual que antes
se editaba `recursos.js`.

## Desplegar

Estático puro, sin servidor. La salida es `dist/`.

**Vercel**

```bash
cd web && vercel --prod
```
o conectando el repo con: *Root Directory* `web`, *Build Command* `npm run build`,
*Output Directory* `dist`.

**Netlify**

*Base directory* `web`, *Build command* `npm run build`, *Publish directory* `dist`.

En ambos casos el build necesita ver `../datos/` y `../extracted-data/` (`sync-datos` los
copia), así que hay que desplegar desde la raíz del repo, no desde `web/` aislado.

> Nota de despliegue: el scraper debe escribir `extracted-data/puntos.json` **antes** del
> build, o publicar ese archivo aparte en la misma ruta del CDN. Si no está, el sitio sale
> con los 91 curados y el aviso correspondiente.

## Licencias

Código MIT. Datos CC BY 4.0 — libres de usar citando la fuente de cada ítem.
