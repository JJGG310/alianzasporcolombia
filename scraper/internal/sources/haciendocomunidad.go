package sources

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/JJGG310/alianzasporcolombia/scraper/internal/fetch"
	"github.com/JJGG310/alianzasporcolombia/scraper/internal/model"
)

// HaciendoComunidad — https://personnofound.github.io/HaciendoComunidad/
//
// PWA estática en GitHub Pages con Firestore detrás. La configuración de
// Firebase viene en js/config.js del propio sitio (es pública por diseño: toda
// app web de Firebase la expone), y la base permite lectura anónima, así que
// usamos la API REST de Firestore en vez de abrir un navegador.
//
// DELIBERADAMENTE NO leemos la colección "desaparecidos". El principio #2 del
// proyecto es no duplicar bases de personas desaparecidas: son datos personales
// con carga legal (Ley 1581) y humana que no necesitamos asumir, y ya existe un
// estándar de facto (Colombia Te Busca). A ese esfuerzo se enlaza, no se copia.
type HaciendoComunidad struct{}

const (
	hcProyecto = "haciendo-comunidad-db"
	// Clave web pública del sitio, tomada de su js/config.js. No es un secreto:
	// identifica el proyecto, no autoriza nada por sí sola.
	hcClave = "AIzaSyDx9QqALGXfGLL74K5qqbfupBqrGOVlr0I"
	hcBase  = "https://firestore.googleapis.com/v1/projects/" + hcProyecto + "/databases/(default)/documents"
)

func (h *HaciendoComunidad) Slug() string { return "haciendo-comunidad" }
func (h *HaciendoComunidad) URL() string {
	return "https://personnofound.github.io/HaciendoComunidad/"
}
func (h *HaciendoComunidad) Mecanismo() string {
	return "API REST de Firestore con lectura anónima (colecciones reportes_ayuda y comunidad_servicios; la colección desaparecidos se omite a propósito)"
}

func (h *HaciendoComunidad) Obtener(ctx context.Context, c *fetch.Cliente) (*model.Resultado, error) {
	inicio := time.Now()
	res := &model.Resultado{
		Fuente: h.Slug(), URL: h.URL(), Mecanismo: h.Mecanismo(),
		Obtenido: time.Now().UTC().Format(time.RFC3339),
	}
	terminar := func() {
		res.Duracion = time.Since(inicio).String()
		res.Total = len(res.Puntos)
	}

	corte := time.Now().UTC()
	conteos := map[string]int{}

	reportes, err := hcLeerColeccion(ctx, c, "reportes_ayuda")
	if err != nil {
		res.OK = false
		res.Error = fmt.Sprintf("leyendo reportes_ayuda: %v", err)
		terminar()
		return res, err
	}
	for _, d := range reportes {
		if p, ok := h.puntoDeReporte(d, corte); ok {
			res.Puntos = append(res.Puntos, p)
			conteos["reportes_ayuda"]++
		}
	}

	// Que falle la segunda colección no debe tumbar la primera.
	servicios, err := hcLeerColeccion(ctx, c, "comunidad_servicios")
	if err != nil {
		res.Nota = strings.TrimSpace(res.Nota + " comunidad_servicios falló: " + err.Error())
	} else {
		for _, d := range servicios {
			if p, ok := h.puntoDeServicio(d, corte); ok {
				res.Puntos = append(res.Puntos, p)
				conteos["comunidad_servicios"]++
			}
		}
	}

	res.OK = true
	res.Extra = map[string]any{
		"colecciones_leidas":   conteos,
		"coleccion_omitida":    "desaparecidos",
		"motivo_de_la_omision": "datos personales: no duplicamos bases de desaparecidos (Ley 1581 y principio #2 del proyecto). Se enlaza a Colombia Te Busca.",
	}
	res.Nota = strings.TrimSpace(res.Nota + " No se leyó la colección 'desaparecidos' por política de datos personales.")
	terminar()
	return res, nil
}

// --- mapeo ---

// Tipos de reporte del sitio -> vocabulario canónico.
var hcTipos = map[string]string{
	"medica":    model.TipoSalud,
	"rescate":   model.TipoRescate,
	"agua":      model.TipoAgua,
	"alimentos": model.TipoAcopio,
	"refugio":   model.TipoAlbergue,
	"colapso":   model.TipoRescate,
	"via":       model.TipoInfo, // vía cerrada: es información, no un punto de ayuda
	"otro":      model.TipoOtro,
}

var hcEtiquetas = map[string]string{
	"medica": "Atención médica", "rescate": "Rescate de personas atrapadas",
	"agua": "Agua", "alimentos": "Alimentos", "refugio": "Refugio o techo",
	"colapso": "Colapso de infraestructura", "via": "Vía cerrada", "otro": "Otro",
	"transporte": "Transporte", "herramientas": "Herramientas o maquinaria",
	"voluntariado": "Voluntariado", "ropa": "Ropa o abrigo",
}

// El propio sitio considera un reporte desactualizado con 15 marcas de la
// comunidad (UMBRAL_DESACTUALIZADO en su config.js). Respetamos su umbral.
const hcUmbralDesactualizado = 15

func (h *HaciendoComunidad) puntoDeReporte(d hcDoc, corte time.Time) (model.Punto, bool) {
	id := d.id()
	if id == "" {
		return model.Punto{}, false
	}
	tipo := d.texto("tipo")
	canon, ok := hcTipos[tipo]
	if !ok {
		canon = model.TipoOtro
	}

	nombre := d.texto("localidad")
	if nombre == "" {
		nombre = hcEtiquetas[tipo]
	}
	if nombre == "" {
		nombre = "Reporte ciudadano"
	}

	p := model.Punto{
		ID:          model.HacerID(h.Slug(), id),
		Fuente:      h.Slug(),
		FuenteURL:   h.URL(),
		Tipo:        canon,
		Nombre:      nombre,
		Descripcion: d.texto("descripcion"),
		Barrio:      d.texto("localidad"),
		Contacto:    d.texto("contacto"),
		Lat:         d.numero("lat"),
		Lon:         d.numero("lng"),
		Creado:      d.fecha("timestamp"),
		Raw:         d.crudo(),
	}
	if e := hcEtiquetas[tipo]; e != "" {
		p.Necesita = append(p.Necesita, e)
	}

	// El sitio es de Cali (su mapa está centrado ahí), pero solo lo afirmamos
	// cuando las coordenadas lo respaldan: Colombia entera está reportando.
	if p.Lat != nil && p.Lon != nil && enValleDelCauca(*p.Lat, *p.Lon) {
		p.Departamento = "Valle del Cauca"
	}

	switch d.texto("estado") {
	case "activo":
		p.Estado = model.EstadoActivo
	case "resuelto", "atendido":
		p.Estado = model.EstadoCubierto
	case "cerrado":
		p.Estado = model.EstadoCerrado
	default:
		p.Estado = model.EstadoPorConfirmar
	}

	// Señales de la comunidad del sitio origen: las respetamos en vez de
	// republicar como vigente algo que allá ya está en duda.
	if n := d.entero("marcasDesactualizado"); n >= hcUmbralDesactualizado {
		p.Estado = model.EstadoPorConfirmar
		p.Avisos = append(p.Avisos,
			fmt.Sprintf("marcado como desactualizado %d veces en el sitio origen", n))
	}
	if n := d.entero("reportesAbuso"); n > 0 {
		p.PoliticaAlerta = append(p.PoliticaAlerta,
			fmt.Sprintf("reportado como abuso %d veces en el sitio origen", n))
	}

	model.Normalizar(&p, corte)
	return p, true
}

func (h *HaciendoComunidad) puntoDeServicio(d hcDoc, corte time.Time) (model.Punto, bool) {
	id := d.id()
	if id == "" {
		return model.Punto{}, false
	}

	nombre := d.texto("servicio")
	if nombre == "" {
		nombre = hcEtiquetas[d.texto("categoria")]
	}
	if nombre == "" {
		nombre = "Servicio comunitario"
	}

	p := model.Punto{
		ID:          model.HacerID(h.Slug(), id),
		Fuente:      h.Slug(),
		FuenteURL:   h.URL(),
		Nombre:      nombre,
		Descripcion: d.texto("descripcion"),
		Barrio:      d.texto("zona"),
		Contacto:    d.texto("contacto"),
		Lat:         d.numero("lat"),
		Lon:         d.numero("lng"),
		Creado:      d.fecha("timestamp"),
		Raw:         d.crudo(),
	}

	categoria := d.texto("categoria")
	etiqueta := hcEtiquetas[categoria]

	// "oferta" = alguien pone un servicio a disposición; "necesidad" = alguien
	// lo pide. Son cosas distintas y el sitio las debe mostrar distinto.
	if d.texto("tipoPublicacion") == "necesidad" {
		p.Tipo = model.TipoNecesidad
		if etiqueta != "" {
			p.Necesita = append(p.Necesita, etiqueta)
		}
	} else {
		p.Tipo = model.TipoOferta
		if etiqueta != "" {
			p.Ofrece = append(p.Ofrece, etiqueta)
		}
		// Una olla comunitaria ofrecida es un punto de ayuda de verdad, no un
		// aviso: merece salir en el mapa como tal.
		if categoria == "alimentos" && strings.Contains(strings.ToLower(nombre), "olla") {
			p.Tipo = model.TipoOlla
		}
	}

	switch d.texto("estado") {
	case "abierto":
		p.Estado = model.EstadoActivo
	case "cerrado":
		p.Estado = model.EstadoCerrado
	default:
		p.Estado = model.EstadoPorConfirmar
	}
	if n := d.entero("marcasDesactualizado"); n >= hcUmbralDesactualizado {
		p.Estado = model.EstadoPorConfirmar
		p.Avisos = append(p.Avisos,
			fmt.Sprintf("marcado como desactualizado %d veces en el sitio origen", n))
	}
	if n := d.entero("reportesAbuso"); n > 0 {
		p.PoliticaAlerta = append(p.PoliticaAlerta,
			fmt.Sprintf("reportado como abuso %d veces en el sitio origen", n))
	}

	model.Normalizar(&p, corte)
	return p, true
}

// enValleDelCauca es una caja aproximada del departamento. Sirve para no
// atribuir un municipio a ciegas, no para geocodificar.
func enValleDelCauca(lat, lon float64) bool {
	return lat > 3.0 && lat < 5.1 && lon > -77.6 && lon < -75.7
}

// --- cliente REST de Firestore ---

// hcDoc envuelve un documento de Firestore, que llega con los valores tipados
// ({"stringValue": "..."} en vez de "..."). Los accesores esconden ese ruido.
type hcDoc struct {
	Name       string                     `json:"name"`
	Fields     map[string]json.RawMessage `json:"fields"`
	CreateTime string                     `json:"createTime"`
	UpdateTime string                     `json:"updateTime"`
}

func (d hcDoc) id() string {
	i := strings.LastIndex(d.Name, "/")
	if i < 0 {
		return d.Name
	}
	return d.Name[i+1:]
}

func (d hcDoc) crudo() json.RawMessage {
	b, err := json.Marshal(d)
	if err != nil {
		return nil
	}
	return b
}

// valor de Firestore: solo llega uno de estos campos poblado.
type hcValor struct {
	String    *string  `json:"stringValue"`
	Double    *float64 `json:"doubleValue"`
	Integer   *string  `json:"integerValue"` // Firestore manda los enteros como cadena
	Boolean   *bool    `json:"booleanValue"`
	Timestamp *string  `json:"timestampValue"`
}

func (d hcDoc) valor(campo string) (hcValor, bool) {
	crudo, ok := d.Fields[campo]
	if !ok {
		return hcValor{}, false
	}
	var v hcValor
	if err := json.Unmarshal(crudo, &v); err != nil {
		return hcValor{}, false
	}
	return v, true
}

func (d hcDoc) texto(campo string) string {
	v, ok := d.valor(campo)
	if !ok || v.String == nil {
		return ""
	}
	return strings.TrimSpace(*v.String)
}

func (d hcDoc) numero(campo string) *float64 {
	v, ok := d.valor(campo)
	if !ok {
		return nil
	}
	if v.Double != nil {
		return v.Double
	}
	if v.Integer != nil {
		if n, err := strconv.ParseFloat(*v.Integer, 64); err == nil {
			return &n
		}
	}
	return nil
}

func (d hcDoc) entero(campo string) int {
	v, ok := d.valor(campo)
	if !ok {
		return 0
	}
	if v.Integer != nil {
		if n, err := strconv.Atoi(*v.Integer); err == nil {
			return n
		}
	}
	if v.Double != nil {
		return int(*v.Double)
	}
	return 0
}

func (d hcDoc) fecha(campo string) *time.Time {
	v, ok := d.valor(campo)
	if !ok || v.Timestamp == nil {
		return nil
	}
	t, err := time.Parse(time.RFC3339, *v.Timestamp)
	if err != nil {
		return nil
	}
	t = t.UTC()
	return &t
}

type hcRespuesta struct {
	Documents     []hcDoc `json:"documents"`
	NextPageToken string  `json:"nextPageToken"`
}

// hcLeerColeccion pagina una colección completa. El tope de páginas evita que
// un token que no avanza nos deje girando para siempre.
func hcLeerColeccion(ctx context.Context, c *fetch.Cliente, coleccion string) ([]hcDoc, error) {
	var todos []hcDoc
	token := ""
	for pagina := 0; pagina < 20; pagina++ {
		q := url.Values{}
		q.Set("key", hcClave)
		q.Set("pageSize", "300")
		if token != "" {
			q.Set("pageToken", token)
		}
		u := hcBase + "/" + coleccion + "?" + q.Encode()

		var r hcRespuesta
		if err := c.GetJSON(ctx, u, &r); err != nil {
			return todos, err
		}
		todos = append(todos, r.Documents...)
		if r.NextPageToken == "" || r.NextPageToken == token {
			break
		}
		token = r.NextPageToken
	}
	return todos, nil
}
