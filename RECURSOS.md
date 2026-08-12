# Inventario de fuentes — Terremoto de Colombia, 10 de agosto de 2026

Censo de páginas que ya recolectan o publican información sobre la emergencia, para que Cali S.O.S centralice sin duplicar. Barrido de 14 agentes con verificación enlace por enlace, **corte: 12 de agosto de 2026, ~1:45 p.m.**

Lo que está en `recursos.js` (la página) es la selección curada de esto. Aquí está todo, incluidas fuentes caídas y el análisis de vacíos.

## El hallazgo central

La recolección ciudadana **no** se consolidó en hashtags ni cadenas de WhatsApp, sino en dos plataformas web:

1. **[Colombia Te Busca](https://www.colombiatebusca.com)** — el registro de desaparecidos de facto: 5.114 personas registradas (4.155 por localizar, 957 localizadas al corte). Asocapitales y Encontrados.co ya la consumen. Las autoridades la señalan como canal y desaconsejan las cadenas de WhatsApp.
2. **[Encontrados.co](https://encontrados.co)** — matching de fotos con IA (rescatistas suben fotos de personas recuperadas, el sistema cruza contra reportes). **Código abierto con repo público en GitHub** — el primer socio técnico a contactar para federar (creadores: Ni500 y Torrenegra, contactables por X).

**El nicho de Cali S.O.S está libre**: no existe ningún sitio ciudadano enfocado exclusivamente en Cali/Valle. Los albergues y acopios de Cali hoy solo viven en artículos de prensa y boletines de la Alcaldía, sin mapa consolidado ni datos estructurados. Nadie —ni oficial ni ciudadano— publica JSON/CSV abierto de albergues y acopios: publicarlo sería la diferenciación inmediata.

## Fuentes verificadas vivas

### Oficiales
| Fuente | URL | Qué ofrece |
|---|---|---|
| Repositorio oficial Alcaldía de Cali | https://www.cali.gov.co/gobierno/publicaciones/193607/terremoto-de-cali-repositorio-oficial-de-informacion/ | LA fuente primaria de Cali: cifras, líneas de desaparecidos 24/7 (Personería 318 335 5722), albergues, acopios, boletines. Boletines scrapeables con patrón `cali.gov.co/boletines/publicaciones/{id}/` |
| Boletín calamidad pública + toque de queda | https://www.cali.gov.co/boletines/publicaciones/193596/... | Decreto oficial en PDF descargable |
| Boletín medidas de emergencia | https://www.cali.gov.co/boletines/publicaciones/193598/... | Toque de queda, alerta roja hospitalaria, pico y placa |
| Secretaría Gestión del Riesgo Cali | https://www.cali.gov.co/gestiondelriesgo/ | Coordinación distrital; tel. (602) 653 3801 |
| Gobernación del Valle | https://www.valledelcauca.gov.co/ | Calamidad departamental, albergues regionales, líneas 01-8000-972033 / (602) 620-0000 |
| Boletín coordinación Nación-Gobernación-Alcaldía | https://valledelcauca.gov.co/publicaciones/90168/... | Cifras interinstitucionales |
| Campaña El Valle Somos Todos | https://www.valledelcauca.gov.co/publicaciones/90172/... | Qué donar y dónde (Palacio Gobernación, Antigua Licorera) |
| Asocapitales — Terremoto Colombia | https://www.asocapitales.co/terremoto-colombia.html | Hub nacional: dashboard, buscador de desaparecidos (integra Colombia Te Busca), reporte al PMU línea 300 761 6647, acopios por capital |

### Ciudadanas
| Fuente | URL | Qué ofrece |
|---|---|---|
| Colombia Te Busca | https://www.colombiatebusca.com | Registro ciudadano de desaparecidos (ver arriba). Sin API pública |
| Encontrados.co | https://encontrados.co | Matching facial IA, open source (ver arriba) |
| Cuidar a Colombia | https://cuidarcolombia.vercel.app/ | **El sitio hermano en Vercel más parecido a Cali S.O.S**, alcance nacional: mapa de acopios/bancos de sangre, cifras verificadas, antidesinformación. Creador: Santiago Jiménez Londoño (sjimenezlon@gmail.com) |
| ViveCali — especial terremoto | https://www.vivecali.com/terremoto-cali/ | Hub local Cali: líneas, pico y placa, aeropuerto, **sección de desmentidos**. Contactar antes que competir |
| SOS Cali | https://soscali.co | Tablero ciudadano de Grupo Monetae: desaparecidos, info verificada, pedidos de ayuda. **Ojo: nombre casi idéntico a "Cali S.O.S"** — coordinar o diferenciarse |

### Albergues, acopio y ayuda (prensa)
| Fuente | URL | Qué ofrece |
|---|---|---|
| El País — albergues habilitados | https://www.elpais.com.co/cali/terremoto-en-cali-estos-son-los-lugares-habilitados-para-recibir-a-los-damnificados-1038.html | Cancha Hockey Miguel Calero, Diamante de Béisbol, UD Jaime Aparicio |
| El País — nuevos acopios | https://www.elpais.com.co/cali/alcaldia-habilita-nuevos-espacios-para-recibir-ayudas-y-atender-a-afectados-en-cali-1154.html | Plazoleta Jairo Varela, Escuela Nacional del Deporte + lista de insumos |
| El Tiempo — guía de ayudas | https://www.eltiempo.com/colombia/otras-ciudades/ayudas-tras-terremoto-en-colombia-...-3577631 | Acopios + bancos de sangre por ciudad (Cali: Plaza de Cayzedo, CAM, 5 bancos permanentes) |
| El Tiempo — mapa nacional de acopios | https://www.eltiempo.com/datos/...-3577654 | 40+ centros en 24 municipios |
| Infobae — acopios por ciudad | https://www.infobae.com/colombia/2026/08/10/centros-de-acopio-habilitados-en-colombia-... | Guía multiciudad |
| Revista Diners — ollas comunitarias | https://revistadiners.com.co/gastronomia/donde-comer/ollas-comunitarias-terremoto-cali | Comida comunitaria por barrio |

### Desaparecidos (canales)
- Cruz Roja RCF: línea **132**, WhatsApp **+57 321 213 9525**, rcf@cruzrojacolombiana.org — detalle en [El Tiempo](https://www.eltiempo.com/colombia/otras-ciudades/cruz-roja-habilito-canales-oficiales-de-ayuda-para-contactar-a-familiares-desaparecidos-tras-terremoto-de-7-4-en-colombia-esto-debe-saber-3577533)
- Guía de reporte: [El País](https://www.elpais.com.co/colombia/asi-puede-reportar-a-una-persona-desaparecida-y-ayudar-a-encontrarla-tras-el-terremoto-en-colombia-1107.html)
- Personería de Cali 24/7: 318 335 5722

### Donaciones verificadas
- El Valle Somos Todos (Gobernación) — en especie
- Propacífico "Solidaridad con el Pacífico" — insumos médicos; canales en su web (política del proyecto: no copiamos cuentas) — [nota El País](https://www.elpais.com.co/cali/propacifico-activa-campana-de-solidaridad-para-atender-a-damnificados-por-el-terremoto-1129.html)
- UNICEF Colombia — https://unicef.org.co/terremoto-colombia (pesos, Nequi)
- Vaki "Yo Tengo Fe por el Pacífico" — https://vaki.co/vaki/yo-tengo-fe-por-el-pacifico (Buenaventura/Chocó)
- GoFundMe hub verificado — https://www.gofundme.com/c/act/colombia-earthquake-relief (internacional)
- Direct Relief — https://www.directrelief.org/emergency/colombia-earthquake-2026/ (médico internacional)

### Cobertura en vivo
| Medio | URL | Nota |
|---|---|---|
| El País Cali — día 3 | https://www.elpais.com.co/cali/en-vivo-tercer-dia-tras-el-terremoto-en-cali-y-el-valle-...-1224.html | El liveblog local más completo; El País además tiene sección permanente "Terremoto en Colombia" |
| El País Valle | https://www.elpais.com.co/valle/en-vivo-terremoto-en-el-valle-...-1056.html | Desglose por municipio |
| Semana | https://www.semana.com/nacion/articulo/terremoto-en-colombia-en-vivo-...-/202646/ | Nacional, incluye donaciones y acopios |
| Noticias Caracol | https://www.noticiascaracol.com/colombia/en-vivo-colombia-hoy-tras-terremoto-de-magnitud-7-4-...-so35 | Cifras UNGRD/SGC |
| El Tiempo | https://www.eltiempo.com/colombia/otras-ciudades/fuerte-terremoto-de-7-4-...-3577255 | Balance por ciudad |
| El Espectador | https://www.elespectador.com/ciencia/en-vivo-asi-amanece-colombia-tras-el-terremoto-... | Único con medidas de alivio (subsidios de arriendo, servicios públicos 3 meses) |
| Blu Radio | https://www.bluradio.com/nacion/terremoto-hoy-en-colombia-en-vivo-...-so35 | Énfasis Cali/Valle |
| El Heraldo | https://www.elheraldo.co/colombia/2026/08/10/en-vivo-minuto-a-minuto-... | Estado hospitalario/alerta roja |
| Diario Occidente | https://occidente.co/regionales/valle-del-cauca/sismo-en-cali-... y .../vias-afectadas-por-el-sismo-segun-invias-... | Galería de daños + estado vial Invías (13 puntos) |
| Univision / Telemundo | (ver liveblogs por día) | Internacional en español, para diáspora |

### Mapas, datos y APIs
| Fuente | URL | Formato |
|---|---|---|
| USGS evento us6000tjl2 | https://earthquake.usgs.gov/earthquakes/eventpage/us6000tjl2 | ShakeMap, PAGER, GeoJSON API (réplicas en vivo) |
| GDACS alerta naranja | https://gdacs.org/report.aspx?eventid=1557236&episodeid=1724218&eventtype=EQ | API GeoJSON, RSS/CAP, mapas UNITAR-UNOSAT |
| Copernicus EMS EMSR916 | https://mapping.emergency.copernicus.eu/news/earthquake-in-colombia-emsr916/ | Productos GIS de daños descargables (Shapefile/GeoPackage) + [StoryMap ArcGIS](https://storymaps.arcgis.com/stories/faac12172e564e31b558b1dff08c91d6) |
| ReliefWeb | https://reliefweb.int/disaster/eq-2026-000146-col | Sitreps; **API pública api.reliefweb.int** (la web bloquea bots, la API no) |
| Wikipedia ES/EN | es.wikipedia.org/wiki/Terremoto_de_Colombia_de_2026 | Infobox estructurado vía API MediaWiki, referencias consolidadas |
| EarthquakeTrack | https://earthquaketrack.com/quakes/2026-08-10-12-34-28-utc-7-4-110 | Ficha técnica + réplicas (espejo de USGS) |
| UN News | https://news.un.org/en/story/2026/08/1168112 | Cifras ONU consolidadas |
| NRC / IRC | nrc.no/news/2026/colombia-earthquake · rescue.org/press-release/earthquake-strikes-colombias-poorest-region | Análisis humanitario por sector |

### Oficiales nacionales (re-barrido, completado 2:05 p.m.)
| Fuente | URL | Qué ofrece |
|---|---|---|
| SGC — noticia oficial del sismo | https://www2.sgc.gov.co/Noticias/Paginas/SGC-actualiza-la-informacion-sobre-el-sismo-ocurrido-en-San-Jose-del-Palmar-Choco.aspx | Parámetros verificados (M 7,4, prof. 103 km, 18+ réplicas, intensidad máx. 7) |
| SGC — Sismo Sentido | https://sismosentido.sgc.gov.co | Formulario ciudadano oficial de intensidad; alimenta los mapas del SGC |
| UNGRD — comunicado de respuesta | https://portal.gestiondelriesgo.gov.co/Paginas/Noticias/2026/Gobierno-nacional-despliega-respuesta-ante-sismo-en-Chocó-... | Sala de Crisis Nacional, 4 grupos USAR. **Sin página consolidada de situación** |
| Cruz Roja — respuesta al sismo | https://www.cruzrojacolombiana.org/cruz-roja-colombiana-despliega-capacidades-para-dar-respuesta-tras-las-afectaciones-por-sismo-en-colombia/ | RCF (132 / WhatsApp 321 213 9525 / 01 8000 519 8534 / rcf@...), sangre por ciudad, donaciones vía su web (política del proyecto: no copiamos cuentas) |
| Cruz Roja — donación en línea | https://ayuda.cruzrojacolombiana.org/emergencia-colombia-terremoto | Landing de recaudo dedicado (verificar pagos in situ antes de promover) |
| OPS/OMS — Informe de Situación 1 | https://www.paho.org/es/documentos/informe-situacion-1-colombia-terremoto-agosto-2026-10-agosto-2026 | Sit-rep sector salud, serie numerada, PDF. Monitorear informes 2+ |
| ONU Noticias — balance OCHA | https://news.un.org/es/story/2026/08/1541802 | Cifras consolidadas al 12-ago: 224 fallecidos, 195 desaparecidos, 2.595 heridos, 357 municipios en 13 departamentos |
| Defensa Civil | https://www.defensacivil.gov.co/ | Solo línea 144; **sin sección dedicada al sismo** — no usar como fuente primaria |

### Datos y APIs (re-barrido)
| Fuente | Endpoint | Formato |
|---|---|---|
| USGS — detalle del evento | https://earthquake.usgs.gov/earthquakes/feed/v1.0/detail/us6000tjl2.geojson | **GeoJSON verificado funcionando**, CORS abierto, sin auth. Incluye ShakeMap, PAGER (alerta roja), pronóstico de réplicas (OAF), 1.107 reportes DYFI |
| USGS — réplicas | https://earthquake.usgs.gov/fdsnws/event/1.0/query | API FDSN: filtrar por radio alrededor de 4.844 N, -76.242 W |
| SGC — visor de sismos | https://www.sgc.gov.co/sismos | App JS sin API documentada; el SGC registra 70+ réplicas (más que USGS). Descubrir el endpoint interno con DevTools |
| SGC — catálogo | https://sismo.sgc.gov.co/ | Consulta tabular del catálogo revisado; sin API JSON |
| SGC — datos abiertos | https://datos.sgc.gov.co/ (servicios en srvags.sgc.gov.co/arcgis/rest/services) | ArcGIS REST, JSON; capas de amenaza sísmica |
| datos.gov.co | API Socrata | **CERO datasets de la emergencia al 12-ago** (verificado). Dejar un poller: cuando publiquen censos, salen por API SODA |
| INVÍAS | https://www.invias.gov.co/ + cuenta @numeral767 en X + línea #767 | Sin API de incidentes; la home no muestra alertas del sismo — los cierres salen por X |
| Wikipedia | API MediaWiki / Wikidata | Infobox estructurado, cifras agregadas con referencias |

## Fuentes caídas o bloqueadas (no usar sin re-verificar)

| Fuente | Problema |
|---|---|
| bomberoscali.org | Sin contenido del evento (última noticia: 2023) |
| @EMCALIoficial en X | X bloquea acceso (HTTP 402); el estado de servicios de Emcali solo sale por X y prensa |
| 90minutos.co | HTTP 403 **solo para bots** — funciona en navegador; incluido en la página |
| CNN en Español / CNN | HTTP 451 (bloqueo legal/regional del fetcher; puede funcionar desde Colombia) |
| ReliefWeb (web) | 403 para bots; usar la API |
| IFRC, OCHA, Save the Children, Americares | 403 para bots; re-verificar a mano en 24-48 h |
| infobea.com (con typo) | ¡Dominio falso! Redirige a tracker t.tr4k.live. El real es infobae.com |
| YouTube (transmisiones Blu/Caracol) | Bloqueado para fetchers con CAPTCHA; embeber directo desde la página debería funcionar |

## Vacíos detectados = oportunidades para Cali S.O.S

1. **Nadie publica datos estructurados.** Ni oficiales ni ciudadanos exponen API/CSV/JSON de albergues, acopios o cifras. Todo es HTML narrativo. → Publicar `datos.json` abierto de albergues/acopios de Cali sería único en el ecosistema.
2. **Los albergues/acopios de Cali están dispersos y desactualizados entre artículos** (El País tiene 2 listas distintas). → Una tabla única con fecha de corte por ítem.
3. **Las cifras divergen fuerte según fuente y hora** (74 vs 95 muertos en Cali; 195 vs ~4.000-5.000 desaparecidos según se cuente oficial o ciudadano). → Mostrar cada cifra con fuente + timestamp, y explicar la brecha oficial/ciudadano para no alarmar.
4. **Estado de servicios públicos (Emcali) no tiene página** — solo X y prensa. → Una tabla de servicios por comuna sería valor nuevo.
5. **Desinformación activa** (videos de Filipinas/Guatemala, imágenes IA). ViveCali y Cuidar a Colombia ya tienen secciones de desmentidos. → Enlazarlas, no duplicar.
6. **Hashtags**: no hay confirmación de prensa de #CaliSOS ni #FuerzaCali; no publicar hashtags sin verificar.
7. **Los liveblogs rotan de URL a diario** (Caracol, CNN, Univision). → Re-verificar enlaces cada mañana.
8. **La UNGRD no publica situación consolidada** — todo son noticias sueltas. → Centralizar "el último boletín oficial" con fecha/hora tiene mucho valor.
9. **datos.gov.co está vacío de la emergencia** (0 datasets, verificado vía API Socrata). → Transcribir las listas de acopios de El Tiempo/Infobae a un JSON propio es la materia prima inmediata; dejar un poller por si publican censos.
10. **El SGC no expone API pública de réplicas** — su visor consume un endpoint interno. → Descubrirlo con DevTools o usar USGS FDSN mientras tanto.

## Pendiente

- Verificar a mano (bloqueados para bots, pueden funcionar en navegador): HOT Tasking Manager (tasks.hotosm.org — buscar "Colombia earthquake"), uMap HOT (umap.hotosm.org/en/map/colombia-m-74-earthquake-10-ago-2026_3482), IFRC GO, Canal Trece (cifras UNGRD/Medicina Legal), @numeral767 e @EMCALIoficial en X.
- Verificar a mano: censo de damnificados de la Alcaldía (anunciado en el repositorio 193607, sin URL directa), informes OPS 2+ en paho.org.
- Contactos para federar: Encontrados.co (GitHub/X), Cuidar a Colombia (sjimenezlon@gmail.com), ViveCali, El Tiempo (uso del mapa de acopios).
