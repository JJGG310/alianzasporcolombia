# Ruta de trabajo — equipo de 3

Diseñada para trabajar **en paralelo sin pisarse**: cada persona es dueña de archivos distintos.

## Frentes (dueño por archivo = cero conflictos de merge)

### Persona A — Datos y verificación 📊
**Dueña de:** `public/datos/*.json`, `public/recursos.js`, `RECURSOS.md`
- Mantener cada punto con fuente + fecha de corte; marcar vencidos.
- Correr el scraper (`cd scraper && make correr`) y **revisar el diff** de `extracted-data/` antes de subirlo. La salida es determinista: lo que cambie en el diff es lo que cambió en la realidad.
- Revisar los puntos con `politica_alerta` — no se publican sin que alguien los mire.
- Re-verificar liveblogs a diario (rotan de URL) y boletines oficiales.
- Llamar/escribir a albergues y acopios para confirmar que siguen activos (el dato más valioso del sitio).

### Persona B — Producto y frontend 🛠
**Dueña de:** `public/index.html` (el sitio en vivo) y `web/src/` (la versión en React)
- Mapa y tabla (hecho), compartir por WhatsApp (hecho), réplicas en vivo del USGS (hecho).
- **Decidir cuál frontend queda** y pasarle al ganador lo que tenga el otro. Ver README.
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

## Esta semana (en orden)

1. ~~Estructura, directorio verificado, mapa+tabla, datos abiertos~~ ✅
2. ~~Conectar repo a Cloudflare Pages → publicar en alianzasporcolombia.com~~ ✅ (dominio comprado, DNS y HTTPS activos, deploy automático en cada push)
3. ~~Scraper que federa 9 sitios ciudadanos → 957 puntos, publicados en `public/extracted-data/`~~ ✅
4. **Decidir cuál frontend queda**: el vanilla que está en vivo o el de React (Persona B + equipo). Ver la tabla comparativa en el README. Mientras no se decida, el deploy NO se cambia.
5. Pasarle al sitio en vivo los 957 puntos agregados: ya están publicados en `public/extracted-data/puntos.json`, falta que `index.html` los lea (Persona B)
6. Revisar los 8 puntos con `politica_alerta` (Persona A, hoy). Ojo: el mismo Nequi repetido en 4 puntos de mapa-emergencia — puede ser recaudo legítimo o puede no serlo.
7. Enviar mensajes a los sitios aliados (Persona C, hoy). Ahora hay algo concreto que ofrecer: **ya los estamos federando** y los datos están en CC BY 4.0 para que ellos nos consuman de vuelta. El ing. de Univalle de Haciendo Comunidad ya escribió pidiendo justamente esto.
8. Panel de cifras con fuente + hora (Persona B, 1-2 días)
9. Rutina automática de monitoreo: `cd scraper && make correr` cada 30-60 min, diff a revisión humana
10. Confirmación telefónica de albergues/acopios listados (Persona A, continuo)
11. Limpiar los `municipio` de `public/datos/ayuda.json` que no son municipios (`"Nacional (canaliza a zonas afectadas)"`, `"Valle del Cauca"`): ensucian el desplegable (Persona A)
