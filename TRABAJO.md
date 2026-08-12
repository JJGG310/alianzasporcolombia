# Ruta de trabajo — equipo de 3

Diseñada para trabajar **en paralelo sin pisarse**: cada persona es dueña de archivos distintos.

## Frentes (dueño por archivo = cero conflictos de merge)

### Persona A — Datos y verificación 📊
**Dueña de:** `datos/ayuda.json`, `recursos.js`, `RECURSOS.md`
- Mantener cada punto con fuente + fecha de corte; marcar vencidos.
- Procesar los diffs que producen los barridos de agentes (revisar → aprobar → merge).
- Re-verificar liveblogs a diario (rotan de URL) y boletines oficiales.
- Llamar/escribir a albergues y acopios para confirmar que siguen activos (el dato más valioso del sitio).

### Persona B — Producto y frontend 🛠
**Dueña de:** `index.html`
- Mapa y tabla (hecho — pulir), botón compartir por WhatsApp por tarjeta.
- Panel de cifras con fuente + hora.
- Widget de réplicas en vivo (feed USGS ya verificado, `fetch` directo).
- Versión `/lite` de solo texto (<50 KB) y luego PWA offline.

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
2. Comprar dominio → conectar repo a Vercel/Netlify → publicar (Persona B, 1 h)
3. Aprobar el JSON del barrido Valle-wide de fundaciones y puntos (Persona A, hoy)
4. Enviar mensajes a los 4 sitios aliados (Persona C, hoy)
5. Compartir por WhatsApp + panel de cifras + réplicas en vivo (Persona B, 1-2 días)
6. Rutina automática de monitoreo aprobada y corriendo (ya diseñada en PLAN.md)
7. Confirmación telefónica de albergues/acopios listados (Persona A, continuo)
