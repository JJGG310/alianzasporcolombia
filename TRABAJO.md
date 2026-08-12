# Ruta de trabajo — equipo de 3

Diseñada para trabajar **en paralelo sin pisarse**: cada persona es dueña de archivos distintos.

## Frentes (dueño por archivo = cero conflictos de merge)

### Persona A — Datos y verificación 📊
**Dueña de:** `datos/ayuda.json`, `web/src/data/recursos.ts`, `RECURSOS.md`
- Mantener cada punto con fuente + fecha de corte; marcar vencidos.
- Correr el scraper (`cd scraper && make correr`) y **revisar el diff** de `extracted-data/` antes de subirlo. La salida es determinista: lo que cambie en el diff es lo que cambió en la realidad.
- Revisar los puntos con `politica_alerta` — no se publican sin que alguien los mire.
- Re-verificar liveblogs a diario (rotan de URL) y boletines oficiales.
- Llamar/escribir a albergues y acopios para confirmar que siguen activos (el dato más valioso del sitio).

### Persona B — Producto y frontend 🛠
**Dueña de:** `web/src/`
- Mapa y tabla unificados (hecho), compartir por WhatsApp (hecho), réplicas en vivo del USGS (hecho).
- Panel de cifras con fuente + hora.
- Versión `/lite` de solo texto (<50 KB) y luego PWA offline.
- El scraper (`scraper/`) es de quien lo toque: si una fuente se rompe, `go run ./cmd/husmear <url>` dice de dónde saca sus datos ahora.

### Persona C — Alianzas, difusión y moderación 🤝
**Dueña de:** `docs/`
- Enviar los mensajes de `docs/aliados.md` (Encontrados, Cuidar a Colombia, ViveCali, SOS Cali) y coordinar federación.
- Ofrecer los widgets embebibles a 90 Minutos y El País.
- Difusión por WhatsApp/redes; recoger correcciones de la gente y pasarlas a Persona A.
- Cuando entre el formulario ciudadano (Fase 2): moderar reportes.

## Flujo git (simple, rápido)

- `main` = lo que está publicado. Conectar el repo a Vercel/Netlify para deploy automático en cada push.
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
2. ~~Sitio migrado a React; compartir por WhatsApp y réplicas en vivo~~ ✅
3. ~~Scraper que federa 8 sitios ciudadanos → 886 puntos en `extracted-data/`~~ ✅
4. **Comprar dominio → conectar repo a Vercel/Netlify → publicar** (Persona B, 1 h). `netlify.toml` y `vercel.json` ya están listos; el build corre desde la raíz, no desde `web/`.
5. Revisar los 8 puntos con `politica_alerta` (Persona A, hoy). Ojo: el mismo Nequi repetido en 4 puntos de mapa-emergencia — puede ser recaudo legítimo o puede no serlo.
6. Enviar mensajes a los sitios aliados (Persona C, hoy). Ahora hay algo concreto que ofrecer: **ya los estamos federando** y `extracted-data/` está en CC BY 4.0 para que ellos nos consuman de vuelta. El ing. de Univalle de Haciendo Comunidad ya escribió pidiendo justamente esto.
7. Panel de cifras con fuente + hora (Persona B, 1-2 días)
8. Rutina automática de monitoreo: `make correr` cada 30-60 min, diff a revisión humana (ya diseñada en PLAN.md)
9. Confirmación telefónica de albergues/acopios listados (Persona A, continuo)
10. Limpiar los `municipio` de `datos/ayuda.json` que no son municipios (`"Nacional (canaliza a zonas afectadas)"`, `"Valle del Cauca"`): ensucian el desplegable del sitio (Persona A)
