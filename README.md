# Alianzas por Colombia — alianzasporcolombia.com

Sitio ciudadano que centraliza información verificada sobre el terremoto del 10 de agosto de 2026: líneas de emergencia, albergues, puntos de acopio, bancos de sangre, ollas comunitarias, dónde donar, desaparecidos, cobertura en vivo y datos abiertos.

**Alcance:** empezamos por Cali y el Valle del Cauca (las primeras horas), y crecemos por regiones a todo el país: Pereira/Risaralda, Manizales/Caldas, Armenia/Quindío, Quibdó/Chocó.

**Principios innegociables:**
1. **Nunca recibimos ni publicamos cuentas bancarias.** Solo el contacto de la organización, con la advertencia de validar con ella antes de entregar dinero.
2. **Toda cifra y todo punto lleva fuente + fecha de corte.**
3. **No duplicamos lo que ya funciona** (Colombia Te Busca, Encontrados.co): enlazamos y federamos.
4. Nada se auto-publica: un humano revisa antes de cada actualización.

## Estructura

```
web/                      El sitio, en React + TypeScript (Vite). Es lo que se despliega
public/                   Los datos que consume el sitio (se copian al build)
  datos/ayuda.json        Los puntos verificados a mano por el equipo (CC BY 4.0)
  datos/avisos.json       Avisos operativos vigentes (toque de queda, vías, aeropuertos)
  datos/necesidades.json  Qué se necesita ahora
  extracted-data/         Lo que produce el scraper, publicado como datos abiertos
  _headers                Reglas de caché y CORS (se copian al build)
scraper/                  Agregador en Go: lee otros sitios ciudadanos y normaliza sus datos
extracted-data/           Salida completa del scraper, incluido el crudo por fuente
RECURSOS.md               Inventario del barrido de fuentes (incluye caídas y vacíos)
PLAN.md                   Plan de producto por fases
TRABAJO.md                Ruta de trabajo del equipo (quién hace qué, flujo git)
docs/aliados.md           Borradores de contacto para los sitios hermanos
docs/monitoreo.md         Bitácora del agente de monitoreo automático
```

El sitio se construye desde `web/` y lee sus datos de `public/datos/` y
`public/extracted-data/`: por eso `public/` se queda, aunque ya no tenga HTML.

Hay **dos orígenes de datos y son distintos a propósito**: `public/datos/ayuda.json` son
los puntos que alguien del equipo confirmó (`verificado: true`), y `extracted-data/` es lo
agregado automáticamente de otros sitios ciudadanos (`verificado: false`, siempre con
enlace a su fuente). Se muestran diferenciados: la confianza es el producto, y el usuario
tiene que poder distinguirlos de un vistazo.

## Correr local

El sitio:

```bash
cd web && npm install && npm run dev
```

Actualizar los datos agregados de los otros sitios:

```bash
cd scraper && make correr     # escribe extracted-data/, revisa el diff antes de subirlo
```

El scraper compila a **un binario sin dependencias**: `make construir` (o `make todas` para mac/linux/windows). Cualquiera del equipo lo corre desde su computador y sube el diff. Ver [scraper/README.md](scraper/README.md).

## Federación con los otros sitios ciudadanos

Después del terremoto aparecieron una docena de sitios ciudadanos, cada uno con su pedazo de la información y ninguno hablándose con los demás. En vez de pedirles que se pasen a nuestro formato, los leemos: el scraper agrega **8 sitios** y los normaliza al mismo esquema.

| Sitio | De dónde salen sus datos | Puntos |
|---|---|---:|
| [mapa-emergencia](https://mapa-emergencia.artefactofilms.workers.dev/) | API REST propia | 502 |
| [aidtrace](https://aidtrace-rastroayuda.vercel.app) | `/api/timeline` (trazabilidad de lotes, intermitente) | 100 |
| [quesenecesita.org](https://quesenecesita.org/) | Google Sheets publicado como CSV | 49 |
| [calisolidario](https://calisolidario.triadaaliados.com/) | payload RSC de Next.js embebido en el HTML | 45 |
| [aqui-hace-falta](https://aqui-hace-falta.web.app/) | Convex — WebSocket, y también `POST /api/query` | 42 |
| [Haciendo Comunidad](https://personnofound.github.io/HaciendoComunidad/) | Firestore con lectura anónima | 38 |
| [puntos-criticos](https://terremoto-cali-puntos-criticos.netlify.app/) | arreglos JS escritos a mano en el HTML | 37 |
| [Ventana de Vida](https://window-of-hope-countdown.lovable.app) | página editorial, sin dataset | 1 |

**886 puntos** tras fusionar: 500 con coordenadas, 20 duplicados colapsados y **19 corroborados por dos sitios independientes** — que dos equipos que no se conocen reporten el mismo albergue es la mejor señal de confianza que hay sin llamar por teléfono.

Esto es federación, no apropiación: cada punto conserva el enlace a su sitio de origen, y `extracted-data/` está en CC BY 4.0 para que ellos también puedan consumirnos. Si mantienes uno de estos sitios y prefieres que no te leamos, escríbenos y listo.

**No agregamos datos de personas desaparecidas de ninguna fuente** — es el principio #2. Dos de estos sitios tienen su propia base; los adaptadores la saltan a propósito.

## Editar datos

- **Enlaces del directorio** → `web/src/data/recursos.ts`.
- **Puntos de ayuda / mapa** → `public/datos/ayuda.json`. Reglas: `tipo` ∈ albergue|acopio|sangre|olla|donacion-org; `fuente` y `fecha_corte` obligatorios; `lat/lon` solo si son confiables (si no, `null` y sale solo en la tabla); **jamás cuentas bancarias en ningún campo**.
- **`extracted-data/` no se edita.** Lo genera el scraper. Si un punto agregado está mal, se corrige en el adaptador de esa fuente o se le escribe al sitio de origen.

Un punto agregado que traiga `politica_alerta` **no se publica sin que alguien lo mire**: es la marca que deja el saneamiento cuando detecta algo que parece dato de pago, o cuando el sitio de origen lo tiene reportado.

## Deploy

El sitio en vivo es **`web/`** (React). El sitio vanilla que hubo al principio
se retiró cuando React quedó a la par; sigue en el historial de git
(`git show HEAD~1:public/index.html`) por si alguna vez hace falta consultarlo.

Automático: cada push a `main` dispara [.github/workflows/deploy.yml](.github/workflows/deploy.yml),
que construye `web/`, copia `public/_headers` al build y publica `web/dist` en
**Cloudflare Pages** (proyecto `alianzasporcolombia`). El build corre desde la
raíz del repo, no desde `web/`: `sync-datos.mjs` lee `public/datos/` y
`public/extracted-data/`, que están afuera de `web/`. Si algún archivo de datos
no queda en el build, el deploy **falla** a propósito — mejor la versión
anterior que un sitio mudo en plena emergencia.

`/extracted-data/` y `/datos/` se sirven como archivos estáticos con CORS
abierto: son también la API pública del proyecto. La prensa y los sitios
aliados pueden consumirlos sin pedir permiso.

## Licencias

Código: MIT. Datos (`datos/` y `extracted-data/`): CC BY 4.0 — libres de usar citando la fuente de cada ítem.

Los datos agregados vienen de proyectos ciudadanos de terceros y **no** están verificados uno por uno por nosotros: por eso cada punto trae su `fuente_url`, para que cualquiera pueda confirmarlo. No agregamos datos de personas desaparecidas de ninguna fuente — ver el principio #2 en [PLAN.md](PLAN.md).
