# web/ — Alianzas por Colombia en React

App de [alianzasporcolombia.com](https://alianzasporcolombia.com): el mismo sitio ciudadano
del terremoto del 10 de agosto de 2026, migrado del `index.html` vanilla de la raíz a
**Vite + React 18 + TypeScript**.

Sin framework de UI, sin Tailwind, sin librería de animación, sin backend. CSS plano
sobre un sistema de tokens propio. Mapa con `react-leaflet` sobre OpenStreetMap (sin API
key).

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

## El sistema de diseño

Vive en tres archivos y se aplica en ese orden (ver `src/main.tsx`):

| Archivo | Qué contiene |
|---|---|
| `src/estilos/tokens.css` | Las variables: color, tipografía, espaciado, forma, movimiento. Modo claro y oscuro |
| `src/estilos/base.css` | Reset, tipografía base, foco, y las primitivas: `.btn` `.chip` `.campo` `.sello` `.tarjeta` `.esqueleto` |
| `src/estilos/{marco,lista,app}.css` | Lo específico del marco, del listado y de la página |

**Ningún componente escribe un color literal.** Todo sale de `tokens.css`, y por eso el modo
oscuro salió gratis. Si vas a agregar un valor y no está en la escala, casi siempre es
señal de que el problema está mal planteado.

### Las tres reglas que ordenan todo lo demás

**1. El rojo significa urgente. Nada más.** Antes el rojo era la marca, los bordes, los
títulos y los números; cuando un punto era urgente de verdad, no quedaba con qué decirlo.
Ahora la marca es tinta, lo interactivo es tinta, y el rojo es una señal. Los estados
(`--urgente` `--activo` `--confirmar` `--cubierto` `--cerrado`) están pensados como la
decisión que toma quien mira: ve ya · puedes ir · llama antes · no hace falta · no vayas.

**2. La herramienta va primero.** El orden es buscador → mapa → listado. Todo lo
explicativo va después. Antes había encabezado, franja de aviso, seis tarjetas de teléfonos
y un párrafo largo antes del mapa: tres pantallazos de scroll en un celular para llegar a
lo que la persona vino a buscar.

**3. El movimiento confirma, no decora.** Máximo 260 ms, y `prefers-reduced-motion` apaga
todo. En una lista de 878 elementos, animar la entrada de cada tarjeta es mareo, no diseño.

### Tipografía

**Geist** variable, autohospedada en `public/fuentes/geist.woff2`. Subseteada al español:
**25 KB por los nueve pesos**, contra 68 KB del archivo completo. Va autohospedada y no por
CDN porque un sitio de emergencia no debería depender de que un tercero esté arriba, y va
precargada en `index.html` porque define la primera impresión.

Para volver a subsetearla si algún día hace falta otro glifo, ver el histórico de git:
se hizo con `subset-font` en Node, sobre ASCII + acentos y eñes + comillas tipográficas.

### El color del mapa

Los 14 tipos de punto se agrupan en **4 decisiones + un neutro** (`GRUPOS_MAPA` en
`src/data/catalogo.ts`): dónde dormir · llevar o recoger · salud, agua y comida · rescate y
emergencia · otro.

Cuatro no es un número estético, es el techo medido. En un mapa cualquier par de marcadores
puede quedar pegado, así que hay que validar los 10 pares posibles, no solo los adyacentes.
Con esos 4 hues el peor par da ΔE 9,2 en visión deutan y 16,3 en visión normal — pasa todas
las compuertas, sobre las teselas claras y sobre las atenuadas del modo oscuro. **Un quinto
hue rompe el piso de visión normal.** Los 14 colores escogidos a ojo que había antes hacían
indistinguible la mitad de los pares, sobre todo para quien tiene daltonismo — 1 de cada 12
hombres.

El gris de "otro" queda fuera del set categórico a propósito: no compite como identidad,
marca la ausencia de una.

Dos detalles que se ven poco y se notan mucho: cada marcador lleva un anillo blanco de 2 px
(en el centro de Cali hay decenas encimados; sin el anillo el mapa es una mancha), y el modo
oscuro **atenúa** las teselas en vez de invertirlas — invertidas, el agua se ve como tierra
y el mapa deja de ser legible.

### Accesibilidad

Es un sitio de emergencia: se abre en celular, con afán, a veces caminando, a veces por
alguien con la vista cansada o con lector de pantalla.

- Área táctil mínima de 44 px en todo lo que se toca (`--toque`).
- Foco visible y consistente en todo el sitio (el del navegador desaparece sobre estas
  superficies claras).
- La identidad nunca depende solo del color: el mapa lleva leyenda con texto y el listado
  completo es la versión accesible de lo mismo.
- Modo oscuro real, no un filtro. No es estética: la gente abre esto de noche, con el
  celular en 8 % de batería.
