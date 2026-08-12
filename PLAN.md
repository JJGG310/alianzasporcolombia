# Alianzas por Colombia — Plan de producto

> **Actualización de alcance (12-ago, tarde):** marca definitiva **Alianzas por Colombia** (alianzasporcolombia.com). Foco del Valle en las primeras horas; expansión por regiones a todo el país (Pereira, Manizales, Armenia, Quibdó ya en barrido). Las menciones a "Cali S.O.S" abajo son el nombre de trabajo original.

## La tesis

En una emergencia, la utilidad de un sitio es **confianza × frescura × encontrabilidad**. Construir es lo fácil; lo que mata a estos sitios es el día 5, cuando los datos se vencen y el enlace que compartió todo el mundo apunta a un albergue que ya cerró. Todo el diseño gira alrededor de mantener los datos frescos con poco esfuerzo humano.

El barrido de 16 agentes (ver RECURSOS.md) encontró el nicho exacto: **no existe ningún sitio dedicado a Cali/Valle, y nadie —ni oficial ni ciudadano— publica datos estructurados**. Los albergues, acopios y cierres viales viven dispersos en artículos de prensa que se desactualizan entre sí. Ese es el hueco.

## Principios (qué NO hacemos)

1. **No manejamos plata.** Solo enlazamos campañas verificadas. Un sitio ciudadano que recibe donaciones pierde la confianza en el primer rumor.
2. **No duplicamos bases de desaparecidos.** Colombia Te Busca (5.100+ registros) ya es el estándar de facto y las autoridades la señalan. Federamos, no competimos. Datos de personas = responsabilidad legal (habeas data, Ley 1581) y humana (fotos de menores) que no necesitamos cargar.
3. **Toda cifra lleva fuente + hora.** Las cifras oficiales divergen (74 vs 95 muertos en Cali según el momento; 195 vs ~4.000 desaparecidos según quién cuenta). Mostrar la cifra sin su fuente es desinformar.
4. **Contactar antes de construir.** SOS Cali, ViveCali, Cuidar a Colombia y Encontrados ya existen. Cada función que ya hagan bien, se enlaza.

## Fase 0 — Ya está (publicar hoy)

- Directorio de 44 enlaces verificados en 7 secciones, líneas de emergencia, estática, liviana, sin build.
- Falta: comprar dominio, `vercel .` o arrastrar a Netlify.

## Fase 1 — El diferenciador (2-3 días)

**1. La tabla-mapa única de ayuda en Cali/Valle.** Albergues, puntos de acopio, bancos de sangre y ollas comunitarias en UNA tabla y UN mapa (Leaflet + OpenStreetMap, sin API key). Cada fila con: dirección, qué ofrece/recibe, **fuente, fecha de corte y estado** (`vigente / por confirmar / cerrado`). La materia prima ya está identificada: repositorio 193607 de la Alcaldía, los 2 artículos de El País, la guía de El Tiempo, El Valle Somos Todos, Diners (ollas).

**2. Datos abiertos: `datos/ayuda.json`.** La misma tabla, publicada como JSON documentado con licencia CC. Seríamos el único dato estructurado de la emergencia en todo el ecosistema (datos.gov.co tiene CERO datasets — verificado por API). La prensa y los sitios hermanos pueden consumirlo → cada consumidor es distribución y legitimidad.

**3. Módulo de cifras con fuente y hora.** Un panel: fallecidos / heridos / desaparecidos / albergados, cada uno con su fuente (UNGRD, PMU Cali, OCHA) y timestamp, más una línea explicando la brecha oficial-vs-ciudadano en desaparecidos. Nadie hace esto y es la pregunta #1 de todo el mundo.

**4. "Último boletín oficial".** La UNGRD no tiene página de situación consolidada y los boletines de la Alcaldía siguen el patrón scrapeable `cali.gov.co/boletines/publicaciones/{id}/`. Una sección que siempre muestre el boletín más reciente de Alcaldía + Gobernación + UNGRD, con hora.

**5. Réplicas en vivo.** El feed GeoJSON de USGS ya está verificado funcionando (CORS abierto, sin auth, con pronóstico de réplicas). Un widget de 30 líneas: última réplica, magnitud, hace cuánto. Antídoto directo contra las cadenas de pánico de "viene otro más fuerte".

**6. Compartir por WhatsApp.** Botón en cada tarjeta (albergue, acopio, cifra) que genera el texto listo para pegar. En Colombia la distribución real es WhatsApp; el sitio debe estar diseñado para ser citado por pedazos.

**7. Mantenimiento automatizado.** Rutina diaria programada (agente): re-verificar los liveblogs (rotan de URL a diario), detectar boletines nuevos, marcar enlaces caídos, y dejar el diff de `recursos.js` listo para revisión humana. Esto resuelve el problema del día 5.

### Diseño del monitoreo continuo (cadencia por fuente, no un solo reloj)

La cadencia correcta depende de qué tan rápido cambia cada fuente — un solo ciclo de 5 minutos para todo gasta 288 ejecuciones/día para ganar casi nada:

| Fuente | Mecanismo | Cadencia |
|---|---|---|
| Réplicas (USGS GeoJSON) | `fetch()` **desde el navegador del visitante** — sin backend, siempre fresco | En cada visita |
| Noticias (El País, Semana, Caracol…) | RSS de cada medio + Google News RSS (`news.google.com/rss/search?q=terremoto+cali`) — estructurado, gratis, sin scraping | Agente cada 30-60 min |
| Boletines oficiales (Alcaldía patrón `/boletines/publicaciones/{id}/`, Gobernación, UNGRD) | Scraping ligero del listado | Agente cada 60 min |
| Enlaces del directorio (liveblogs rotan a diario) | Re-verificación + diff de `recursos.js` para revisión humana | Agente 1-2 veces/día |
| GDACS / ReliefWeb | RSS-CAP / API pública | Agente 2 veces/día |
| **X / Twitter** | **No scrapeable**: bloquea bots (HTTP 402, verificado 2 veces en el barrido). API oficial de lectura es de pago. Nitter está muerto | Revisión humana de @numeral767 y @EMCALIoficial; lo demás llega a los liveblogs en minutos |

El agente de cada ciclo produce: (1) resumen de novedades, (2) boletines nuevos detectados, (3) enlaces caídos, (4) diff propuesto de datos — un humano aprueba antes de publicar.

## Fase 2 — Participación ciudadana (1 semana)

**8. Reportes ciudadanos con moderación.** "Este acopio ya no recibe", "este albergue está lleno", "aquí hay una olla nueva". Formulario mínimo → Supabase → cola de moderación (2-3 voluntarios) → publica. Nunca auto-publicar. Datos mínimos (sin nombres ni cédulas → sin carga de habeas data).

**9. Estado de servicios por comuna.** Emcali no tiene página de estado (solo X, bloqueado). Tabla agua/luz/gas por comuna alimentada por reportes ciudadanos moderados + lo que salga por prensa. Valor nuevo que no existe en ningún lado.

**10. Estado de vías en mapa.** Invías reporta por X (@numeral767) y #767; Diario Occidente lo transcribe. Pasarlo a mapa con los 13 puntos y estado.

**11. Salud mental y apoyo psicosocial.** Directorio: línea 106, apoyo Cruz Roja (01 8000 519 8534), puntos de atención psicosocial que anuncien Alcaldía/OPS. Nadie lo está centralizando y la demanda explota en la semana 2.

**12. Antidesinformación federada.** No escribir desmentidos propios: sección que agrega los de ViveCali y Cuidar a Colombia, con buscador. Hay desinformación activa documentada (videos de Filipinas/Guatemala, imágenes IA).

## Fase 3 — Federación y resiliencia (semanas)

**13. Desaparecidos: puerta única, no base nueva.** Una página "Buscar a alguien" que explique los 3 caminos (Colombia Te Busca para buscar/reportar, Encontrados para matching por foto, Cruz Roja RCF para gestión oficial) con deep-links. Contactos ya identificados: Encontrados es open source (GitHub), Cuidar a Colombia (sjimenezlon@gmail.com), SOS Cali (Grupo Monetae).

**14. Widgets embebibles para medios.** La tabla de albergues y el panel de cifras como iframe embebible. 90 Minutos y El País necesitan exactamente esto — la prensa se vuelve nuestra distribución.

**15. Versión ultraliviana + offline.** Una ruta `/lite` de solo texto (<50 KB) para redes caídas, y PWA con cache offline. Códigos QR impresos para pegar en albergues y acopios → el sitio llega a quien no tiene datos móviles.

**16. Después de la emergencia.** El sitio muta a guía de reconstrucción: censo de damnificados, subsidios (arriendo, servicios públicos 3 meses — ya anunciados), estado de escuelas, salud mental. Y la infraestructura queda lista para la próxima emergencia del Valle — que es zona sísmica.

## Arquitectura por fase

- **Fase 0-1:** sigue siendo estático + JSON + un fetch a USGS. Cero backend. La rutina de mantenimiento corre como agente programado, no como servidor.
- **Fase 2:** entra Supabase (formulario + moderación). Primer backend real, y solo para lo que lo necesita.
- **Fase 3:** nada nuevo — iframes, service worker y contactos humanos.

## Los tres primeros pasos concretos

1. Comprar dominio y publicar lo que ya hay (hoy).
2. Transcribir albergues/acopios/sangre/ollas a `datos/ayuda.json` + tabla + mapa (el diferenciador).
3. Escribir a los 4 sitios hermanos (Encontrados, Cuidar a Colombia, ViveCali, SOS Cali): "existimos, este es nuestro JSON abierto, ¿federamos?"
