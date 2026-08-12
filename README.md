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
index.html        La página completa (HTML + CSS + JS vanilla, sin build)
recursos.js       Directorio de enlaces por sección — editarlo ES actualizar la página
datos/ayuda.json  Datos abiertos (CC BY 4.0): albergues, acopios, sangre, ollas, orgs de donación
RECURSOS.md       Inventario completo del barrido de fuentes (incluye caídas y vacíos)
PLAN.md           Plan de producto por fases
TRABAJO.md        Ruta de trabajo del equipo (quién hace qué, flujo git)
docs/aliados.md   Borradores de contacto para los sitios hermanos
```

## Correr local

```bash
python3 -m http.server 8942
```

y abrir http://localhost:8942 (hace falta servidor por el `fetch` de `datos/ayuda.json`; abrir el archivo directo no carga el mapa).

## Editar datos

- **Enlaces del directorio** → `recursos.js` (cada item: nombre, url, desc, badge opcional).
- **Puntos de ayuda / mapa** → `datos/ayuda.json`. Reglas: `tipo` ∈ albergue|acopio|sangre|olla|donacion-org; `fuente` y `fecha_corte` obligatorios; `lat/lon` solo si son confiables (si no, `null` y sale solo en la tabla); **jamás cuentas bancarias en ningún campo**.

## Deploy

Automático: cada push a `main` dispara [.github/workflows/deploy.yml](.github/workflows/deploy.yml), que publica el sitio a **Cloudflare Pages** (proyecto `alianzasporcolombia`) con `wrangler`. En ~25 segundos el cambio queda en vivo en alianzasporcolombia.com — no hay que hacer nada más que el push.

Para desplegar a mano desde tu máquina (si alguna vez hace falta):
```bash
npx wrangler pages deploy . --project-name=alianzasporcolombia --branch=main
```
(pide login de Cloudflare la primera vez).

## Licencias

Código: MIT. Datos (`datos/`): CC BY 4.0 — libres de usar citando la fuente de cada ítem.
