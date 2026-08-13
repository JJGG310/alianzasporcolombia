# Ruta de trabajo — equipo de 3

Diseñada para trabajar **en paralelo sin pisarse**: cada persona es dueña de archivos distintos.

## Frentes (dueño por archivo = cero conflictos de merge)

### Persona A — Datos y verificación 📊
**Dueña de:** `public/datos/*.json` (ayuda, necesidades, avisos, cifras), `public/recursos.js`, `RECURSOS.md`
- **Ojo:** `public/lite.html` lleva los datos EMBEBIDOS (es la versión para 2G) — cuando cambien mucho los datos, pedir regenerarla.
- Mantener cada punto con fuente + fecha de corte; marcar vencidos.
- Correr el scraper (`cd scraper && make correr`) y **revisar el diff** de `extracted-data/` antes de subirlo. La salida es determinista: lo que cambie en el diff es lo que cambió en la realidad.
- Revisar los puntos con `politica_alerta` — no se publican sin que alguien los mire.
- Re-verificar liveblogs a diario (rotan de URL) y boletines oficiales.
- Llamar/escribir a albergues y acopios para confirmar que siguen activos (el dato más valioso del sitio).

### Persona B — Producto y frontend 🛠
**Dueña de:** `web/src/` (el sitio, en React — es lo que se despliega)
- Mapa y tabla (hecho), compartir por WhatsApp (hecho), réplicas en vivo del USGS (hecho).
- Panel de cifras con fuente + hora.
- Versión `/lite` de solo texto (<50 KB) y luego PWA offline.
- El scraper (`scraper/`) es de quien lo toque: si una fuente se rompe, `go run ./cmd/husmear <url>` dice de dónde saca sus datos ahora.

### Persona C — Alianzas, difusión y moderación 🤝
**Dueña de:** `docs/`
- Enviar los mensajes de `docs/aliados.md` (Encontrados, Cuidar a Colombia, ViveCali, SOS Cali) y coordinar federación.
- Ofrecer los widgets embebibles a 90 Minutos y El País.
- Difusión por WhatsApp/redes; recoger correcciones de la gente y pasarlas a Persona A.
- Cuando entre el formulario ciudadano (Fase 2): moderar reportes.

## Trabajar desde varias máquinas / sesiones a la vez

El repo es **público** (github.com/JJGG310/alianzasporcolombia), así que cualquiera lo clona sin permisos. Para *empujar* cambios hace falta estar autenticado en esa máquina:

```bash
gh auth login          # una sola vez por máquina
git clone https://github.com/JJGG310/alianzasporcolombia
```

**La regla que evita el 90% de los problemas:** `git pull` antes de empezar a editar, y `git push` apenas termines un cambio. Nada de acumular trabajo local por horas.

Si dos sesiones tocan el mismo archivo a la vez, git rechaza el segundo push. No es grave — se arregla así:

```bash
git pull --rebase    # trae lo del otro y pone tus cambios encima
git push
```

Si hay conflicto real (ambos editaron la misma línea), git marca el archivo; se edita a mano dejando la versión correcta y luego `git add <archivo> && git rebase --continue`.

Por eso cada persona es dueña de archivos distintos (ver arriba): así el conflicto casi nunca ocurre.

## Monitoreo automático

Cada hora corre un agente en la nube que revisa boletines oficiales y liveblogs, verifica que los enlaces sigan vivos, y escribe lo que encuentre en `docs/monitoreo.md` (solo si hay novedades). **No modifica datos publicados** — solo reporta; Persona A revisa y aplica.

Ver o pausar la rutina: https://claude.ai/code/routines/trig_014Qg4TSWyn1W3vxhGqjTDRo

Lo que cambia minuto a minuto (réplicas) no lo maneja el agente: la página consulta al USGS directamente en cada visita, así que siempre está al día sin depender de nadie.

## Flujo git (simple, rápido)

- `main` = lo que está publicado. Repo conectado a Cloudflare Pages con deploy automático en cada push (ver `.github/workflows/deploy.yml`); solo se despliega `public/`.
- Cambios chicos y frecuentes directo a `main` (somos 3 y cada quien toca sus archivos). Rama + PR solo para cambios grandes de `index.html`.
- Mensajes de commit en español, cortos: `datos: 4 acopios nuevos Palmira (fuente El País)`.
- Regla de oro: **si tocas un archivo del que no eres dueño, avisa por WhatsApp antes.**

## Cadencia

- **Sync de 10 min, 2 veces al día** (mañana y noche) por WhatsApp/llamada: qué cambió, qué se venció, qué bloquea.
- El barrido automático de agentes corre cada 30-60 min y deja novedades + enlaces caídos + diff propuesto; Persona A lo revisa en cada sync.
- GitHub Issues con etiquetas `datos` / `frontend` / `alianzas` / `urgente` para lo que no se resuelva en el sync.

## Definición de "hecho" para un dato

Publicable solo si tiene: fuente (URL) + fecha de corte + municipio + tipo. Si es organización de donación: además contacto verificado y **cero datos de pago**.

## Deuda técnica conocida

- **El scraper reescribe `fecha_corte` de TODOS los puntos en cada corrida** (no
  solo los que cambiaron). Efecto medido en el commit `2b81793`: 1.844 cambios de
  `fecha_corte` contra ~40 cambios reales de contenido → cada corrida (cada ~90
  min por throttling de GitHub) commitea ~18k líneas. Rompe el principio "el diff
  es el control de calidad" (los cambios reales quedan enterrados) e infla el
  repo público. **Causa:** `scraper/internal/model/punto.go:159` pone
  `time.Now()` cuando la fuente no trae timestamp, así que los puntos agregados
  sin fecha de origen se re-sellan cada vez. **Arreglo propuesto** (necesita Go
  local para probar, por eso queda para quien tenga el toolchain): que el scraper
  lea su salida anterior y conserve `fecha_corte` para los puntos cuyo contenido
  (todo menos el timestamp) no cambió; solo bumpea cuando el dato de verdad
  cambia. Alternativa barata si urge: redondear el fallback a día
  (`Format("2006-01-02")`) — baja el ruido ~16x pero sigue habiendo un diff
  grande al pasar la medianoche UTC. Validable rápido con `gh workflow run
  scraper.yml` y viendo si el diff encoge. No lo toqué porque no puedo compilar
  Go en esta sesión y es el pipeline autónomo de un sitio en vivo.

## Esta semana (en orden)

1. ~~Estructura, directorio verificado, mapa+tabla, datos abiertos~~ ✅
2. ~~Conectar repo a Cloudflare Pages → publicar en alianzasporcolombia.com~~ ✅ (dominio comprado, DNS y HTTPS activos, deploy automático en cada push)
3. ~~Scraper que federa 9 sitios ciudadanos → 957 puntos, publicados en `public/extracted-data/`~~ ✅
4. ~~Decidir cuál frontend queda~~ ✅ **React.** El deploy publica `web/dist`; el vanilla (`public/index.html`) se retiró. Quien edite HTML ahora trabaja en `web/src/`, no en `public/`.
5. ~~Pasarle al sitio en vivo los puntos agregados~~ ✅ (729 reportes en mapa y tabla, marcados «sin verificación propia», con tipos nuevos: 🆘 pide ayuda, ofrece ayuda, salud; excluidos los de `politica_alerta`)
6. Revisar los **10** puntos con `politica_alerta` (Persona A, hoy). Ojo: el mismo Nequi repetido en varios puntos de mapa-emergencia — puede ser recaudo legítimo o puede no serlo. (El render los excluye hasta que alguien los mire, así que no hay urgencia de que se publiquen mal, pero sí de decidir si entran.)
7. Enviar mensajes a los sitios aliados (Persona C, hoy). Ahora hay algo concreto que ofrecer: **ya los estamos federando** y los datos están en CC BY 4.0 para que ellos nos consuman de vuelta. El ing. de Univalle de Haciendo Comunidad ya escribió pidiendo justamente esto.
8. ~~Panel de cifras con fuente + hora~~ ✅ (balance UNGRD 12-ago en pestaña Ahora, con la brecha oficial/ciudadano explicada)
9. Rutina automática de monitoreo: `cd scraper && make correr` cada 30-60 min, diff a revisión humana
10. Confirmación telefónica de albergues/acopios listados (Persona A, continuo)
11. ~~Limpiar los `municipio` que no son municipios~~ ✅ (13 normalizados a «Nacional»; guardia también en la UI)
