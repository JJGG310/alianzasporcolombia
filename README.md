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
web/              El sitio, en React + TypeScript (Vite)
scraper/          Agregador en Go: lee los otros sitios ciudadanos y normaliza sus datos
extracted-data/   Salida del scraper (generada, no se edita a mano)
datos/ayuda.json  Los 91 puntos verificados a mano por el equipo (CC BY 4.0)
recursos.js       Directorio de enlaces por sección
index.html        El sitio original en HTML plano — se conserva hasta validar la migración
RECURSOS.md       Inventario completo del barrido de fuentes (incluye caídas y vacíos)
PLAN.md           Plan de producto por fases
TRABAJO.md        Ruta de trabajo del equipo (quién hace qué, flujo git)
docs/aliados.md   Borradores de contacto para los sitios hermanos
```

Hay **dos orígenes de datos y son distintos a propósito**: `datos/ayuda.json` son los puntos que alguien del equipo confirmó (`verificado: true`), y `extracted-data/` es lo agregado automáticamente de otros sitios ciudadanos (`verificado: false`, siempre con enlace a su fuente). El sitio los muestra diferenciados: la confianza es el producto, y el usuario tiene que poder distinguirlos de un vistazo.

## Correr local

El sitio (React):

```bash
cd web && npm install && npm run dev
```

Actualizar los datos agregados de los otros sitios:

```bash
cd scraper && make correr     # escribe extracted-data/, revisa el diff antes de subirlo
```

El scraper compila a **un binario sin dependencias**: `make construir` (o `make todas` para mac/linux/windows). Cualquiera del equipo lo corre desde su computador y sube el diff. Ver [scraper/README.md](scraper/README.md).

## Editar datos

- **Enlaces del directorio** → `recursos.js` (cada item: nombre, url, desc, badge opcional).
- **Puntos de ayuda / mapa** → `datos/ayuda.json`. Reglas: `tipo` ∈ albergue|acopio|sangre|olla|donacion-org; `fuente` y `fecha_corte` obligatorios; `lat/lon` solo si son confiables (si no, `null` y sale solo en la tabla); **jamás cuentas bancarias en ningún campo**.
- **`extracted-data/` no se edita.** Lo genera el scraper. Si un punto agregado está mal, se corrige en el adaptador de esa fuente o se le escribe al sitio de origen.

Un punto agregado que traiga `politica_alerta` **no se publica sin que alguien lo mire**: es la marca que deja el saneamiento cuando detecta algo que parece dato de pago, o cuando el sitio de origen lo tiene reportado.

## Deploy

Estático puro: `cd web && npm run build` y publicar `web/dist/` en Netlify o Vercel. Cada merge a `main` puede desplegar automático conectando el repo.

`extracted-data/` y `datos/` se sirven como archivos estáticos junto al sitio, así que también son la API pública del proyecto: cualquiera puede consumirlos.

## Licencias

Código: MIT. Datos (`datos/` y `extracted-data/`): CC BY 4.0 — libres de usar citando la fuente de cada ítem.

Los datos agregados vienen de proyectos ciudadanos de terceros y **no** están verificados uno por uno por nosotros: por eso cada punto trae su `fuente_url`, para que cualquiera pueda confirmarlo. No agregamos datos de personas desaparecidas de ninguna fuente — ver el principio #2 en [PLAN.md](PLAN.md).
