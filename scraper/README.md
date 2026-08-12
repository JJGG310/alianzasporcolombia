# scraper — agregador de sitios ciudadanos

Recorre los sitios ciudadanos que surgieron por el terremoto del 10 de agosto de 2026, normaliza todo a un mismo esquema, fusiona duplicados entre sitios y escribe [`../extracted-data/`](../extracted-data/).

**Un binario, cero dependencias externas.** Se compila una vez y se corre en cualquier computador del equipo: quien lo corra actualiza los datos y sube el diff.

```bash
cd scraper
make correr          # corre todas las fuentes y escribe ../extracted-data/
make construir       # binario en bin/scrape
make todas           # binarios para mac (arm/intel), linux y windows
```

Sin `make`:

```bash
go run ./cmd/scrape -salida ../extracted-data
go build -o bin/scrape ./cmd/scrape && ./bin/scrape
```

Opciones útiles:

```bash
./bin/scrape -listar                        # qué fuentes hay y cómo se leen
./bin/scrape -fuentes mapa-emergencia       # solo una (o varias, con comas)
./bin/scrape -seco                          # corre sin escribir nada
./bin/scrape -paralelas 2                   # baja la concurrencia
./bin/scrape -curados ""                    # sin fusionar datos/ayuda.json
```

## Las fuentes

| Fuente | Cómo se leen sus datos |
|---|---|
| `mapa-emergencia` | API REST propia: `/api/snapshot`, `/api/novedades`, `/api/config` |
| `aqui-hace-falta` | Convex. La app habla por WebSocket; los mismos datos salen por `POST /api/query` con `udfPath: needs:list` |
| `haciendo-comunidad` | Firestore con lectura anónima, vía API REST. Config pública en su `js/config.js` |
| `calisolidario` | Next.js App Router: los datos vienen dentro del stream RSC embebido en el HTML de `/necesidades` y `/ofertas` |
| `quesenecesita` | Google Sheets publicado como CSV (la URL está en el HTML del sitio) |
| `puntos-criticos` | Arreglos JS escritos a mano dentro del HTML (React por Babel standalone). Se parsean los literales |
| `aidtrace` | Su `/api/timeline` responde 500 y su Supabase exige autenticación. Sin datos públicos; se reintenta en cada corrida |
| `window-of-hope` | Página informativa sin dataset. Se conserva como fuente editorial |

Explícitamente **no** se leen datos de personas desaparecidas, ni de este ni de ningún otro sitio: es el principio #2 del proyecto (ver [PLAN.md](../PLAN.md)). `haciendo-comunidad` tiene una colección `desaparecidos` y el adaptador la salta a propósito. A ese esfuerzo se enlaza; no se copia.

## Qué escribe

En `../extracted-data/`:

| Archivo | Qué es |
|---|---|
| `puntos.json` | Lo que consume el sitio: todo normalizado y fusionado, con el informe de la corrida |
| `puntos.geojson` | Solo lo que tiene coordenadas, para que prensa y aliados lo usen sin escribir código |
| `fuentes.json` | Ficha de auditoría: qué fuente respondió, cuántos puntos, cuánto tardó, qué falló |
| `raw/<fuente>.json` | La respuesta original de cada fuente, tal cual llegó |

`raw/` existe para poder reprocesar sin volver a golpear los sitios. Si mañana cambiamos cómo mapeamos un campo, se recalcula desde ahí.

La salida es **determinista**: mismo dato de entrada, mismos bytes de salida. Sin eso el diff de git no serviría para revisar qué cambió de verdad, y revisar el diff es el control humano que exige el proyecto.

## Dos reglas que están en el código, no solo en el README

**Nunca publicamos datos de pago.** [`internal/model/sanitize.go`](internal/model/sanitize.go) detecta cuentas bancarias, Nequi/Daviplata, enlaces de recaudo y direcciones cripto en los campos de texto libre, los redacta y marca el punto con `politica_alerta`. Agregamos texto que escribe cualquiera; no podemos confiar en que venga limpio. Un punto con `politica_alerta` **no se publica sin que alguien lo mire**.

**Todo punto lleva fuente y fecha de corte.** `fuente_url` y `fecha_corte` se llenan siempre, en todos los adaptadores.

## Fusión de duplicados

Los sitios ciudadanos se copian entre sí: el mismo albergue aparece en tres tableros con el nombre escrito distinto. [`internal/merge`](internal/merge/merge.go) los colapsa en uno y deja constancia en `tambien_en`.

El criterio es deliberadamente conservador — preferimos dejar dos puntos separados (ruido visible) a fusionar dos puntos distintos (dato perdido):

- Solo se fusionan puntos del **mismo tipo** y de **fuentes distintas**.
- Por coincidencia exacta de nombre + dirección normalizados, o por estar a **menos de 120 m** con nombres parecidos.
- Los avisos ciudadanos (`necesidad`, `oferta`) **nunca** se fusionan: son mensajes de personas distintas, no lugares. Dos vecinos pidiendo agua son dos pedidos.

Que dos sitios independientes reporten el mismo punto es la mejor señal de confianza que hay sin llamar por teléfono; por eso `tambien_en` se publica.

## Cuando una fuente se rompe (o aparece un sitio nuevo)

Se va a romper: son sitios hechos a las carreras que cambian todos los días. Hay dos herramientas para volver a encontrar de dónde salen los datos.

### `husmear` — sin navegador, sin dependencias

```bash
go run ./cmd/husmear https://un-sitio-nuevo.com
```

Descarga el HTML y **sigue los imports de los bundles JS en cascada**, buscando las huellas de los backends que usa la gente para montar estas apps: Convex, Firestore, Supabase, Sheets publicados, rutas `/api/…`, payloads RSC de Next.js, arreglos JS escritos a mano. Por cada hallazgo dice qué hacer con él.

Acierta en casi todos estos sitios. Ejemplo real:

```
● Firebase / Firestore
   config.js → projectId: "haciendo-comunidad-db"
   ⇒ Si la base permite lectura anónima, se lee por REST: …
```

Ojo con el crawl en cascada: en `haciendo-comunidad` el `config.js` con la clave está a **tres saltos** del script que enlaza el HTML (`app.js → mapa.js → config.js`). Mirando solo el primer nivel no aparece nada.

### `descubrir/` — navegador real (Playwright)

Lo que `husmear` no puede ver: una app que habla **solo por WebSocket** sin dejar la URL en el bundle.

```bash
cd descubrir
go run . -instalar     # una sola vez: baja Chromium
go run .               # audita todos los sitios conocidos
go run . -url https://un-sitio-nuevo.com
```

Abre cada sitio, registra **todas** sus peticiones de datos y los mensajes de WebSocket, y traduce lo observado a una recomendación de cómo scrapear. Va en su propio módulo Go a propósito: Playwright arrastra dependencias y la descarga de un navegador; el scraper de producción no depende de nada de eso.

> ⚠️ **Hoy la instalación del driver falla.** `go run . -instalar` devuelve 404 en todas las versiones: Microsoft retiró los hosts `*.azureedge.net` desde donde el port de Go descarga su driver, y no hay reemplazo publicado. El código está listo y compila; cuando el CDN se arregle, funciona. Mientras tanto, para el caso del WebSocket se hace a mano en el navegador — así se descubrió Convex en `aqui-hace-falta`:
>
> ```js
> // consola del navegador, en el sitio ya cargado
> const orig = WebSocket.prototype.send;
> window.__sent = [];
> WebSocket.prototype.send = function (d) { window.__sent.push(String(d)); return orig.apply(this, arguments); };
> // toca un filtro para que la app pida datos, y luego:
> window.__sent.slice(-3);
> // → {"type":"ModifyQuerySet",…,"udfPath":"needs:list","args":[{…}]}
> ```
>
> Ese `udfPath` es el nombre de la función que va en `POST /api/query`.

## Agregar una fuente

1. Crea `internal/sources/<slug>.go` con un tipo que implemente `Fuente` (mira `mapaemergencia.go`, es el más simple).
2. Regístralo en `Todas()` de `internal/sources/registry.go`.
3. Ponle prioridad en `prioridadFuente` de `internal/merge/merge.go` (define quién gana ante un conflicto de datos).
4. Agrégalo a la tabla de arriba y a la lista de `descubrir/main.go`.

Reglas para el adaptador: solo stdlib, llamar a `model.Normalizar` en cada punto, y ante fallo devolver `*model.Resultado` con `OK=false` **además** del error — así el runner registra el fallo de esa fuente sin abortar las demás.

## Cortesía con los sitios de origen

Son proyectos ciudadanos hechos por gente que está respondiendo a una emergencia, muchos en hosting gratis. El scraper se identifica con un User-Agent que dice quiénes somos y cómo contactarnos, limita la concurrencia (4 fuentes a la vez por defecto) y no reintenta los 4xx. Córrelo cada 30-60 minutos, no cada minuto.
