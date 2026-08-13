// Directorio de recursos de Alianzas por Colombia — editar este archivo es actualizar la página.
// Todos los enlaces fueron verificados en vivo el 12 de agosto de 2026 (barrido de 14 agentes).
// Inventario completo con fuentes descartadas y análisis de vacíos: ver RECURSOS.md.
const RECURSOS = {
  actualizado: "12 de agosto de 2026, 2:05 p.m.",
  secciones: [
    {
      id: "oficiales",
      titulo: "Fuentes oficiales",
      desc: "Cifras, boletines y órdenes de las autoridades. Empieza siempre aquí. Las cifras varían según fuente y hora — fíjate en la fecha de cada reporte.",
      items: [
        { nombre: "Repositorio oficial del terremoto — Alcaldía de Cali", url: "https://www.cali.gov.co/gobierno/publicaciones/193607/terremoto-de-cali-repositorio-oficial-de-informacion/", desc: "La fuente primaria de Cali: cifras verificadas, líneas de desaparecidos 24/7, albergues, puntos de acopio y boletines cronológicos.", badge: "oficial" },
        { nombre: "Secretaría de Gestión del Riesgo de Cali", url: "https://www.cali.gov.co/gestiondelriesgo/", desc: "Coordinación distrital de la respuesta. Publica actualizaciones periódicas.", badge: "oficial" },
        { nombre: "Gobernación del Valle del Cauca", url: "https://www.valledelcauca.gov.co/", desc: "Situación departamental: calamidad pública, albergues regionales y ayuda a municipios (El Águila, El Cairo, Roldanillo, Zarzal). Líneas 01-8000-972033 y (602) 620-0000.", badge: "oficial" },
        { nombre: "Asocapitales — Terremoto Colombia", url: "https://www.asocapitales.co/terremoto-colombia.html", desc: "Hub nacional por ciudades: dashboard de cifras, buscador de desaparecidos conectado a Colombia Te Busca, reporte al PMU (línea 300 761 6647) y acopios por capital.", badge: "oficial" },
        { nombre: "SGC — Información oficial del sismo", url: "https://www2.sgc.gov.co/Noticias/Paginas/SGC-actualiza-la-informacion-sobre-el-sismo-ocurrido-en-San-Jose-del-Palmar-Choco.aspx", desc: "Parámetros verificados del evento (M 7,4, profundidad 103 km) y réplicas. ¿Lo sentiste? Repórtalo en sismosentido.sgc.gov.co.", badge: "oficial" },
        { nombre: "UNGRD — Gestión del Riesgo", url: "https://portal.gestiondelriesgo.gov.co", desc: "Comunicados de la Sala de Crisis Nacional. No tiene página consolidada de situación: revisar sus noticias.", badge: "oficial" },
        { nombre: "Cruz Roja Colombiana — Respuesta al sismo", url: "https://www.cruzrojacolombiana.org/cruz-roja-colombiana-despliega-capacidades-para-dar-respuesta-tras-las-afectaciones-por-sismo-en-colombia/", desc: "Canales activos: contactos familiares, donación de sangre por ciudad (incluida Cali) y donaciones monetarias.", badge: "oficial" },
        { nombre: "OPS/OMS — Informes de situación del sector salud", url: "https://www.paho.org/es/documentos/informe-situacion-1-colombia-terremoto-agosto-2026-10-agosto-2026", desc: "Serie numerada de reportes sobre la red hospitalaria. Buscar los informes siguientes en paho.org.", badge: "oficial" },
      ],
    },
    {
      id: "albergues",
      titulo: "Albergues y puntos de acopio",
      desc: "Dónde recibir ayuda y dónde llevar donaciones físicas en Cali y el Valle. Lleva solo lo solicitado en las listas oficiales.",
      items: [
        { nombre: "Albergues habilitados en Cali (El País)", url: "https://www.elpais.com.co/cali/terremoto-en-cali-estos-son-los-lugares-habilitados-para-recibir-a-los-damnificados-1038.html", desc: "Cancha de Hockey Miguel Calero (Calle 9 con Cra. 37A), Diamante de Béisbol (Cra. 39 con Calle 9) y Unidad Deportiva Jaime Aparicio." },
        { nombre: "Nuevos puntos de acopio en Cali (El País)", url: "https://www.elpais.com.co/cali/alcaldia-habilita-nuevos-espacios-para-recibir-ayudas-y-atender-a-afectados-en-cali-1154.html", desc: "Plazoleta Jairo Varela y Escuela Nacional del Deporte, con la lista de insumos que se necesitan (agua, víveres, colchonetas, cobijas)." },
        { nombre: "Campaña 'El Valle Somos Todos' — Gobernación", url: "https://www.valledelcauca.gov.co/publicaciones/90172/campana-el-valle-somos-todos-estos-son-los-elementos-que-se-pueden-donar-para-los-damnificados-por-el-sismo/", desc: "Qué donar y dónde: Palacio de la Gobernación, Antigua Licorera del Valle y más puntos departamentales.", badge: "oficial" },
        { nombre: "Guía de ayudas y bancos de sangre (El Tiempo)", url: "https://www.eltiempo.com/colombia/otras-ciudades/ayudas-tras-terremoto-en-colombia-centros-de-acopio-bancos-de-sangre-bancos-de-alimentos-canales-oficiales-y-puntos-de-donaciones-en-el-pais-3577631", desc: "Consolidado por ciudad. En Cali: donación de sangre en Plaza de Cayzedo y CAM (9 a.m.–5 p.m.), 5 bancos de sangre permanentes y acopios." },
        { nombre: "Mapa nacional de centros de acopio (El Tiempo)", url: "https://www.eltiempo.com/datos/este-es-el-mapa-completo-de-los-centros-de-acopio-habilitados-en-colombia-para-ayudar-a-los-damnificados-del-terremoto-de-magnitud-7-3577654", desc: "Más de 40 centros de acopio en 24 municipios de 22 departamentos." },
        { nombre: "Ollas comunitarias en Cali (Revista Diners)", url: "https://revistadiners.com.co/gastronomia/donde-comer/ollas-comunitarias-terremoto-cali", desc: "Puntos de comida comunitaria: San Antonio, Puerto Resistencia, Salomía, Llano Verde, Capri y más.", badge: "ciudadano" },
      ],
    },
    // Personas desaparecidas: tiene su propia sección destacada arriba, en index.html.
    {
      id: "donaciones",
      titulo: "Donaciones verificadas",
      desc: "Campañas y entidades confirmadas. Desconfía de cuentas personales sin respaldo y de cadenas de WhatsApp.",
      items: [
        { nombre: "Cruz Roja Colombiana — Donación en línea", url: "https://ayuda.cruzrojacolombiana.org/emergencia-colombia-terremoto", desc: "Landing oficial de recaudo para la emergencia (tarjeta, Daviplata, transferencia).", badge: "oficial" },
        { nombre: "Fondo Milagro — Gobierno Nacional", url: "https://www.eltiempo.com/colombia/otras-ciudades/presidente-abelardo-de-la-espriella-anuncio-la-creacion-del-fondo-milagro-para-atender-emergencia-tras-terremoto-de-7-4-en-colombia-asi-funcionara-3577993", desc: "Recién creado (12-ago) para canalizar aportes a la reconstrucción. Aún sin canales de donación anunciados: infórmate aquí y espera los oficiales.", badge: "oficial" },
        { nombre: "Dona Hoy — corredor humanitario ABACO", url: "https://donahoy.abaco.org.co/colombia2026", desc: "El corredor del Consejo Gremial y la red de bancos de alimentos: qué se necesita y dónde donar en especie, con puntos en 10 ciudades.", badge: "ciudadano" },
        { nombre: "El Valle Somos Todos — Gobernación del Valle", url: "https://www.valledelcauca.gov.co/publicaciones/90172/campana-el-valle-somos-todos-estos-son-los-elementos-que-se-pueden-donar-para-los-damnificados-por-el-sismo/", desc: "Campaña oficial departamental: donaciones en especie en puntos verificados.", badge: "oficial" },
        { nombre: "Propacífico — Solidaridad con el Pacífico", url: "https://www.elpais.com.co/cali/propacifico-activa-campana-de-solidaridad-para-atender-a-damnificados-por-el-terremoto-1129.html", desc: "Enfocada en insumos médicos. Punto físico en San Nicolás (Cali), transferencia, PSE y tarjeta." },
        { nombre: "UNICEF Colombia", url: "https://unicef.org.co/terremoto-colombia", desc: "Donaciones en pesos (Nequi, tarjeta, cuenta bancaria) para agua, saneamiento y protección de niñez en Chocó y Buenaventura." },
        { nombre: "Vaki — Yo Tengo Fe por el Pacífico", url: "https://vaki.co/vaki/yo-tengo-fe-por-el-pacifico", desc: "Recaudo para damnificados de Buenaventura y Chocó.", badge: "ciudadano" },
        { nombre: "GoFundMe — Colombia Earthquake Relief", url: "https://www.gofundme.com/c/act/colombia-earthquake-relief", desc: "Hub de campañas internacionales verificadas por GoFundMe (Airlink, Americares, Direct Relief y más). Para donar desde el exterior." },
        { nombre: "Direct Relief — Colombia Earthquake Response", url: "https://www.directrelief.org/emergency/colombia-earthquake-2026/", desc: "Fondo internacional 100% destinado a la respuesta médica." },
      ],
    },
    {
      id: "alivios",
      titulo: "Alivios para damnificados",
      desc: "Beneficios ya activos que quizás no sabías que tienes.",
      items: [
        { nombre: "Claro — paquete gratis automático", url: "https://www.claro.com.co", desc: "Si eres prepago en Chocó, Valle, Risaralda, Quindío o Caldas: 7 días de servicio, 2 GB y minutos ilimitados gratis, sin hacer ningún trámite.", badge: "oficial" },
        { nombre: "Avianca — cambios de tiquete sin penalidad", url: "https://www.avianca.com", desc: "Hasta el 16 de agosto: cambio de fecha (hasta 15 días), cambio de ruta o reembolso de trayectos no volados. Además ~2.500 sillas extra hacia Cali, Armenia y Quibdó.", badge: "oficial" },
        { nombre: "Subsidios del Gobierno Nacional", url: "https://www.elespectador.com/ciencia/en-vivo-asi-amanece-colombia-tras-el-terremoto-balance-de-victimas-y-afectaciones/", desc: "Anunciados: subsidio de arriendo para damnificados y pago de servicios públicos por 3 meses. Detalles por confirmar en los boletines oficiales." },
        { nombre: "Universidad de Caldas — servicios gratuitos (Manizales)", url: "https://www.lapatria.com/manizales/universidad-de-caldas-habilita-centro-de-acopio-en-manizales-para-damnificados-del-sismo", desc: "En su coliseo (zona Velódromo): atención veterinaria, consultas de salud, primeros auxilios psicológicos y asesoría jurídica sobre pólizas. Tel. (606) 893 2880, WhatsApp 320 727 3645." },
      ],
    },
    {
      id: "cobertura",
      titulo: "Cobertura en vivo",
      desc: "Liveblogs con seguimiento minuto a minuto. Ojo: algunos medios abren URL nueva cada día.",
      items: [
        { nombre: "El País de Cali — Día 3 en vivo", url: "https://www.elpais.com.co/cali/en-vivo-tercer-dia-tras-el-terremoto-en-cali-y-el-valle-desaparecidos-rescates-ayudas-y-ultimas-noticias-1224.html", desc: "El liveblog local más completo: rescates, desaparecidos, ayudas y medidas en Cali, actualizado en tiempo real.", badge: "en vivo" },
        { nombre: "El País — El Valle en vivo", url: "https://www.elpais.com.co/valle/en-vivo-terremoto-en-el-valle-deja-hasta-el-momento-27-muertos-tres-de-las-victimas-eran-ninos-1056.html", desc: "Cobertura regional: desglose por municipio (El Águila, El Cairo, Roldanillo, Zarzal, Buenaventura).", badge: "en vivo" },
        { nombre: "Semana — Minuto a minuto", url: "https://www.semana.com/nacion/articulo/terremoto-en-colombia-en-vivo-ultimas-noticias-de-cali-pereira-manizales-y-choco/202646/", desc: "Liveblog nacional con cifras actualizadas, donaciones y puntos de acopio.", badge: "en vivo" },
        { nombre: "Noticias Caracol — En vivo", url: "https://www.noticiascaracol.com/colombia/en-vivo-colombia-hoy-tras-terremoto-de-magnitud-7-4-continuan-las-labores-de-rescate-so35", desc: "Labores de rescate con cifras de UNGRD y SGC (96+ réplicas registradas).", badge: "en vivo" },
        { nombre: "El Tiempo — En vivo", url: "https://www.eltiempo.com/colombia/otras-ciudades/fuerte-terremoto-de-7-4-en-colombia-en-la-manana-de-este-lunes-10-de-agosto-videos-en-redes-sociales-captaron-el-sismo-que-tuvo-epicentro-en-choco-3577255", desc: "Balance por ciudad y cobertura en video.", badge: "en vivo" },
        { nombre: "El Espectador — En vivo", url: "https://www.elespectador.com/ciencia/en-vivo-asi-amanece-colombia-tras-el-terremoto-balance-de-victimas-y-afectaciones/", desc: "El único con detalle de las medidas de alivio: subsidios de arriendo y pago de servicios públicos por 3 meses.", badge: "en vivo" },
        { nombre: "Blu Radio — En vivo", url: "https://www.bluradio.com/nacion/terremoto-hoy-en-colombia-en-vivo-victimas-reportes-y-ultimas-noticias-de-la-tragedia-so35", desc: "Liveblog nacional con énfasis en Cali y el Valle.", badge: "en vivo" },
        { nombre: "90 Minutos", url: "https://90minutos.co/", desc: "El noticiero regional del suroccidente: transmisión en vivo digital (8 a.m.) y emisión central (1 p.m.) por Telepacífico, Facebook y YouTube.", badge: "en vivo" },
        { nombre: "La Patria — Manizales y Eje Cafetero", url: "https://www.lapatria.com", desc: "El diario del Eje Cafetero: detalle local de acopios, necesidades y daños en Caldas que los medios nacionales no cubren.", badge: "en vivo" },
        { nombre: "La República — especial Catástrofe Nacional", url: "https://www.larepublica.co/especiales/catastrofe-nacional", desc: "Los anuncios de ayuda del sector empresarial, consolidados." },
      ],
    },
    {
      id: "hermanos",
      titulo: "Sitios aliados",
      desc: "Otras páginas ciudadanas centralizando información. Antes de duplicar una función, contactarlos y federar.",
      items: [
        { nombre: "Cuidar a Colombia", url: "https://cuidarcolombia.vercel.app/", desc: "Sitio ciudadano nacional (en Vercel): mapa interactivo de acopios y bancos de sangre, cifras verificadas y sección antidesinformación.", badge: "ciudadano" },
        { nombre: "ViveCali — Especial terremoto", url: "https://www.vivecali.com/terremoto-cali/", desc: "Hub local de Cali: líneas de emergencia, pico y placa, estado del aeropuerto y desmentidos de noticias falsas.", badge: "ciudadano" },
        { nombre: "SOS Cali", url: "https://soscali.co", desc: "Tablero ciudadano (Grupo Monetae): reportes de desaparecidos, información verificada y pedidos de ayuda.", badge: "ciudadano" },
      ],
    },
    {
      id: "mapas-datos",
      titulo: "Mapas y datos",
      desc: "Mapas de daños, réplicas y datos abiertos para quien construye herramientas.",
      items: [
        { nombre: "USGS — Evento M 7,4 (us6000tjl2)", url: "https://earthquake.usgs.gov/earthquakes/eventpage/us6000tjl2", desc: "ShakeMap, PAGER y pronóstico de réplicas. Feed GeoJSON directo del evento y API FDSN para réplicas, sin autenticación ni límite.", badge: "datos" },
        { nombre: "SGC — Visor de sismos", url: "https://www.sgc.gov.co/sismos", desc: "Visor oficial de sismicidad colombiana; la red del SGC registra más réplicas que USGS (70+). Catálogo consultable en sismo.sgc.gov.co.", badge: "datos" },
        { nombre: "SGC — Datos abiertos", url: "https://datos.sgc.gov.co/", desc: "Servicios ArcGIS REST con capas de amenaza sísmica, consumibles desde Leaflet o Mapbox.", badge: "datos" },
        { nombre: "GDACS — Alerta naranja Colombia", url: "https://gdacs.org/report.aspx?eventid=1557236&episodeid=1724218&eventtype=EQ", desc: "Población expuesta (~5,2 millones en MMI ≥ VII), mapas satelitales y 48+ productos. API GeoJSON y feeds RSS/CAP.", badge: "datos" },
        { nombre: "Copernicus EMS — Activación EMSR916", url: "https://mapping.emergency.copernicus.eu/news/earthquake-in-colombia-emsr916/", desc: "Mapeo satelital de daños a edificios, con productos GIS descargables (Shapefile, GeoPackage) y StoryMap en vivo.", badge: "datos" },
        { nombre: "Vías afectadas según Invías (Diario Occidente)", url: "https://occidente.co/regionales/valle-del-cauca/vias-afectadas-por-el-sismo-segun-invias-reporte-de-emergencias/", desc: "13 puntos afectados en la red nacional, con cierres totales en Dagua (vía a Loboguerrero) y Cartago–Alcalá." },
        { nombre: "ReliefWeb — Reportes de situación", url: "https://reliefweb.int/disaster/eq-2026-000146-col", desc: "Sitreps humanitarios de la ONU. Tiene API pública (api.reliefweb.int) filtrable por desastre.", badge: "datos" },
        { nombre: "Wikipedia — Terremoto de Colombia de 2026", url: "https://es.wikipedia.org/wiki/Terremoto_de_Colombia_de_2026", desc: "Consolidado colaborativo con referencias verificables, actualizado por la comunidad. Datos estructurados vía API.", badge: "datos" },
      ],
    },
  ],
};
