package sources

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/JJGG310/alianzasporcolombia/scraper/internal/fetch"
	"github.com/JJGG310/alianzasporcolombia/scraper/internal/model"
)

// AquiHaceFalta lee el backend Convex de https://aqui-hace-falta.web.app/.
//
// El sitio habla con Convex por WebSocket, pero el mismo despliegue expone la
// query por HTTP (/api/query), que es estable y no necesita cliente de tiempo
// real: se pide la misma udf (needs:list) con los mismos argumentos.
type AquiHaceFalta struct{}

const (
	ahfSlug      = "aqui-hace-falta"
	ahfSitio     = "https://aqui-hace-falta.web.app/"
	ahfAPI       = "https://dusty-panda-452.convex.cloud/api/query"
	ahfUDF       = "needs:list"
	ahfMecanismo = "Convex HTTP query API (udf needs:list; descubierto interceptando el WebSocket)"
)

func (a *AquiHaceFalta) Slug() string      { return ahfSlug }
func (a *AquiHaceFalta) URL() string       { return ahfSitio }
func (a *AquiHaceFalta) Mecanismo() string { return ahfMecanismo }

// ahfRespuesta es el sobre de Convex: status "success" + value, o errorMessage.
type ahfRespuesta struct {
	Status       string            `json:"status"`
	Value        []json.RawMessage `json:"value"`
	ErrorMessage string            `json:"errorMessage"`
	ErrorData    json.RawMessage   `json:"errorData"`
}

type ahfRecurso struct {
	ID                string  `json:"id"`
	Description       string  `json:"description"`
	Type              string  `json:"type"`
	Unit              string  `json:"unit"`
	Status            string  `json:"status"`
	RequestedQuantity float64 `json:"requestedQuantity"`
	FulfilledQuantity float64 `json:"fulfilledQuantity"`
}

type ahfNecesidad struct {
	ID                 string       `json:"_id"`
	CreationTime       float64      `json:"_creationTime"` // epoch ms con fracción
	Address            string       `json:"address"`
	Categories         []string     `json:"categories"`
	CityID             string       `json:"cityId"`
	ContactEmail       string       `json:"contactEmail"`
	ContactName        string       `json:"contactName"`
	ContactPhone       string       `json:"contactPhone"`
	ContactWhatsapp    string       `json:"contactWhatsapp"`
	CreatedAt          string       `json:"createdAt"`
	Description        string       `json:"description"`
	EmergencyID        string       `json:"emergencyId"`
	IsDemoData         bool         `json:"isDemoData"`
	Latitude           float64      `json:"latitude"`
	Longitude          float64      `json:"longitude"`
	Neighborhood       string       `json:"neighborhood"`
	OperatingHours     string       `json:"operatingHours"`
	OrganizationName   string       `json:"organizationName"`
	PlaceType          string       `json:"placeType"`
	Priority           string       `json:"priority"`
	RequesterType      string       `json:"requesterType"`
	Resources          []ahfRecurso `json:"resources"`
	Source             string       `json:"source"`
	Status             string       `json:"status"`
	Title              string       `json:"title"`
	UpdatedAt          string       `json:"updatedAt"`
	VerificationStatus string       `json:"verificationStatus"`
	VerifiedBy         string       `json:"verifiedBy"`
}

// Vocabulario de la fuente -> canónico. Se cubren los valores observados en la
// muestra y los sinónimos plausibles que el formulario del sitio ofrece.
var ahfTipos = map[string]string{
	"EDIFICIO_AFECTADO":    model.TipoRescate,
	"DERRUMBE":             model.TipoRescate,
	"ZONA_DERRUMBE":        model.TipoRescate,
	"ESTRUCTURA_COLAPSADA": model.TipoRescate,
	"ALBERGUE":             model.TipoAlbergue,
	"REFUGIO":              model.TipoAlbergue,
	"ACOPIO":               model.TipoAcopio,
	"CENTRO_ACOPIO":        model.TipoAcopio,
	"CENTRO_DISTRIBUCION":  model.TipoAcopio,
	"PUNTO_LOGISTICO":      model.TipoAcopio,
	"SALUD":                model.TipoSalud,
	"HOSPITAL":             model.TipoSalud,
	"CENTRO_SALUD":         model.TipoSalud,
	"PUESTO_SALUD":         model.TipoSalud,
	"COCINA":               model.TipoOlla,
	"COCINA_COMUNITARIA":   model.TipoOlla,
	"OLLA":                 model.TipoOlla,
	"OLLA_COMUNITARIA":     model.TipoOlla,
	"COMUNIDAD_AFECTADA":   model.TipoNecesidad,
}

// Nombres legibles de placeType, para armar el nombre cuando no viene title.
var ahfLugarLegible = map[string]string{
	"EDIFICIO_AFECTADO":    "edificio afectado",
	"DERRUMBE":             "derrumbe",
	"ZONA_DERRUMBE":        "zona de derrumbe",
	"ESTRUCTURA_COLAPSADA": "estructura colapsada",
	"COMUNIDAD_AFECTADA":   "comunidad afectada",
	"PUNTO_LOGISTICO":      "punto logístico",
	"CENTRO_ACOPIO":        "centro de acopio",
	"CENTRO_DISTRIBUCION":  "centro de distribución",
	"ACOPIO":               "acopio",
	"REFUGIO":              "refugio",
	"ALBERGUE":             "albergue",
	"SALUD":                "punto de salud",
	"HOSPITAL":             "hospital",
	"CENTRO_SALUD":         "centro de salud",
	"PUESTO_SALUD":         "puesto de salud",
	"COCINA":               "cocina comunitaria",
	"COCINA_COMUNITARIA":   "cocina comunitaria",
	"OLLA":                 "olla comunitaria",
	"OLLA_COMUNITARIA":     "olla comunitaria",
}

var ahfPrioridades = map[string]string{
	"CRITICAL": model.PrioridadCritica,
	"HIGH":     model.PrioridadAlta,
	"MEDIUM":   model.PrioridadMedia,
	"LOW":      model.PrioridadBaja,
}

// La muestra solo trae status NEED_HELP_NOW; el resto son los estados que el
// esquema de la app permite y que conviene reconocer si aparecen.
var ahfEstados = map[string]string{
	"NEED_HELP_NOW": model.EstadoActivo,
	"OPEN":          model.EstadoActivo,
	"ACTIVE":        model.EstadoActivo,
	"IN_PROGRESS":   model.EstadoActivo,
	"HELP_ON_WAY":   model.EstadoActivo,
	"FULFILLED":     model.EstadoCubierto,
	"COVERED":       model.EstadoCubierto,
	"RESOLVED":      model.EstadoCubierto,
	"CLOSED":        model.EstadoCerrado,
	"CANCELLED":     model.EstadoCerrado,
	"CANCELED":      model.EstadoCerrado,
	"ARCHIVED":      model.EstadoCerrado,
}

// Categorías y tipos de recurso comparten vocabulario en esta fuente.
var ahfCategorias = map[string]string{
	"ESCOMBROS":                "Remoción de escombros",
	"AGUA":                     "Agua potable",
	"ALIMENTOS":                "Alimentos",
	"MEDICAMENTOS":             "Medicamentos",
	"LOGISTICA":                "Apoyo logístico",
	"VOLUNTARIADO_GENERAL":     "Voluntariado",
	"ROPA":                     "Ropa",
	"ALOJAMIENTO":              "Alojamiento",
	"HERRAMIENTAS":             "Herramientas",
	"MANO_OBRA":                "Mano de obra",
	"MANO_DE_OBRA":             "Mano de obra",
	"DINERO":                   "Donaciones en efectivo",
	"ANIMALES":                 "Atención a animales",
	"ATENCION_MEDICA":          "Atención médica",
	"APOYO_PSICOLOGICO":        "Apoyo psicológico",
	"OPERARIOS_MAQUINARIA":     "Operarios de maquinaria",
	"CLASIFICACION_DONACIONES": "Clasificación de donaciones",
	"TRANSPORTE":               "Transporte",
	"HIGIENE":                  "Aseo e higiene",
	"PANALES":                  "Pañales",
	"COMBUSTIBLE":              "Combustible",
	"OTRO":                     "Otras necesidades",
}

// Municipios que el cityId identifica sin ambigüedad. Si llega uno que no está
// aquí se deja vacío: es preferible sin ubicación que con una inventada.
var ahfCiudades = map[string][2]string{
	"cali":         {"Cali", "Valle del Cauca"},
	"jamundi":      {"Jamundí", "Valle del Cauca"},
	"yumbo":        {"Yumbo", "Valle del Cauca"},
	"palmira":      {"Palmira", "Valle del Cauca"},
	"buenaventura": {"Buenaventura", "Valle del Cauca"},
}

func (a *AquiHaceFalta) Obtener(ctx context.Context, c *fetch.Cliente) (*model.Resultado, error) {
	inicio := time.Now()
	res := &model.Resultado{
		Fuente:    a.Slug(),
		URL:       a.URL(),
		Mecanismo: a.Mecanismo(),
	}
	terminar := func() {
		res.Duracion = time.Since(inicio).String()
		res.Obtenido = time.Now().UTC().Format(time.RFC3339)
		res.Total = len(res.Puntos)
	}
	fallar := func(err error) (*model.Resultado, error) {
		res.OK = false
		res.Error = err.Error()
		terminar()
		return res, err
	}

	peticion := map[string]any{
		"path":   ahfUDF,
		"args":   map[string]any{"sortBy": "PRIORITY"},
		"format": "json",
	}
	cuerpo, err := c.PostJSON(ctx, ahfAPI, peticion)
	if err != nil {
		return fallar(fmt.Errorf("aqui-hace-falta consulta convex: %w", err))
	}

	var resp ahfRespuesta
	if err := json.Unmarshal(cuerpo, &resp); err != nil {
		return fallar(fmt.Errorf("aqui-hace-falta respuesta ilegible: %w", err))
	}
	if resp.Status != "success" {
		detalle := resp.ErrorMessage
		if detalle == "" {
			detalle = "sin errorMessage"
		}
		return fallar(fmt.Errorf("aqui-hace-falta convex devolvió status %q: %s", resp.Status, detalle))
	}

	corte := time.Now().UTC()
	puntos := make([]model.Punto, 0, len(resp.Value))
	var demo, ilegibles, sinID int
	porTipo := map[string]int{}
	porPrioridad := map[string]int{}
	porCategoria := map[string]int{}

	for _, crudo := range resp.Value {
		var n ahfNecesidad
		if err := json.Unmarshal(crudo, &n); err != nil {
			ilegibles++
			continue
		}
		if n.IsDemoData { // datos de prueba del sitio: no son ayuda real
			demo++
			continue
		}
		if n.ID == "" {
			sinID++
			continue
		}
		porTipo[n.PlaceType]++
		porPrioridad[n.Priority]++
		for _, cat := range n.Categories {
			porCategoria[cat]++
		}
		puntos = append(puntos, ahfAModelo(n, crudo, corte))
	}

	res.Puntos = puntos
	res.OK = true
	res.Extra = map[string]any{
		"por_place_type": porTipo,
		"por_prioridad":  porPrioridad,
		"por_categoria":  porCategoria,
		"descartados":    map[string]int{"demo": demo, "ilegibles": ilegibles, "sin_id": sinID},
	}

	var notas []string
	if demo > 0 {
		notas = append(notas, fmt.Sprintf("%d registros isDemoData descartados", demo))
	}
	if ilegibles > 0 {
		notas = append(notas, fmt.Sprintf("%d registros ilegibles descartados", ilegibles))
	}
	if sinID > 0 {
		notas = append(notas, fmt.Sprintf("%d registros sin _id descartados", sinID))
	}
	res.Nota = strings.Join(notas, "; ")

	terminar()
	return res, nil
}

func ahfAModelo(n ahfNecesidad, crudo json.RawMessage, corte time.Time) model.Punto {
	p := model.Punto{
		ID:           model.HacerID(ahfSlug, n.ID),
		Fuente:       ahfSlug,
		FuenteURL:    ahfSitio,
		Tipo:         ahfTipo(n.PlaceType),
		Nombre:       ahfNombre(n),
		Descripcion:  n.Description,
		Barrio:       n.Neighborhood,
		Direccion:    n.Address,
		Lat:          ahfCoord(n.Latitude),
		Lon:          ahfCoord(n.Longitude),
		Necesita:     ahfNecesita(n),
		Estado:       ahfEstado(n),
		Prioridad:    ahfPrioridades[strings.ToUpper(strings.TrimSpace(n.Priority))],
		Contacto:     ahfContacto(n),
		Organizacion: n.OrganizationName,
		Horario:      n.OperatingHours,
		Verificado:   strings.EqualFold(n.VerificationStatus, "VERIFIED"),
		Creado:       ahfCreado(n),
		Actualizado:  ahfFecha(n.UpdatedAt),
		Raw:          crudo,
	}
	if lugar, ok := ahfCiudades[strings.ToLower(strings.TrimSpace(n.CityID))]; ok {
		p.Municipio, p.Departamento = lugar[0], lugar[1]
	}
	model.Normalizar(&p, corte)
	return p
}

func ahfTipo(placeType string) string {
	if t, ok := ahfTipos[strings.ToUpper(strings.TrimSpace(placeType))]; ok {
		return t
	}
	// Sin tipo de lugar reconocible sigue siendo un pedido de ayuda ciudadano.
	return model.TipoNecesidad
}

func ahfEstado(n ahfNecesidad) string {
	if strings.EqualFold(n.Priority, "CRITICAL") {
		return model.EstadoUrgente
	}
	if e, ok := ahfEstados[strings.ToUpper(strings.TrimSpace(n.Status))]; ok {
		return e
	}
	return model.EstadoActivo
}

// ahfNombre nunca devuelve vacío: el nombre es lo único que el mapa alcanza a
// mostrar en la tarjeta, así que se arma con lo que haya.
func ahfNombre(n ahfNecesidad) string {
	if t := strings.TrimSpace(n.Title); t != "" {
		return t
	}
	lugar := ahfLugarLegible[strings.ToUpper(strings.TrimSpace(n.PlaceType))]
	if lugar == "" {
		lugar = ahfCapitalizar(n.PlaceType)
	}
	barrio := strings.TrimSpace(n.Neighborhood)
	switch {
	case barrio != "" && lugar != "":
		return barrio + " — " + ahfCapitalizar(lugar)
	case barrio != "":
		return "Necesidad en " + barrio
	case lugar != "":
		return ahfCapitalizar(lugar)
	case strings.TrimSpace(n.Address) != "":
		return "Necesidad en " + strings.TrimSpace(n.Address)
	}
	return "Necesidad reportada"
}

// ahfNecesita arma la lista de faltantes: primero los recursos concretos con su
// cantidad pendiente (solicitado - entregado), luego las categorías que no
// quedaron ya representadas por algún recurso.
func ahfNecesita(n ahfNecesidad) []string {
	var out []string
	for _, r := range n.Resources {
		if strings.EqualFold(r.Status, "FULFILLED") || strings.EqualFold(r.Status, "CANCELLED") {
			continue
		}
		desc := strings.TrimSpace(r.Description)
		if desc == "" {
			desc = ahfLegible(r.Type)
		}
		if desc == "" {
			continue
		}
		if pend := r.RequestedQuantity - r.FulfilledQuantity; pend > 0 {
			unidad := strings.TrimSpace(r.Unit)
			if unidad != "" {
				desc = fmt.Sprintf("%s (faltan %s %s)", desc, ahfNumero(pend), unidad)
			} else {
				desc = fmt.Sprintf("%s (faltan %s)", desc, ahfNumero(pend))
			}
		}
		out = append(out, desc)
	}
	for _, cat := range n.Categories {
		legible := ahfLegible(cat)
		if legible == "" || ahfYaMencionado(out, legible) {
			continue
		}
		out = append(out, legible)
	}
	return out
}

func ahfYaMencionado(lista []string, legible string) bool {
	buscado := strings.ToLower(legible)
	for _, s := range lista {
		if strings.Contains(strings.ToLower(s), buscado) {
			return true
		}
	}
	return false
}

func ahfLegible(clave string) string {
	clave = strings.ToUpper(strings.TrimSpace(clave))
	if clave == "" {
		return ""
	}
	if v, ok := ahfCategorias[clave]; ok {
		return v
	}
	return ahfCapitalizar(clave)
}

// ahfContacto junta solo lo que la fuente trae; no se completa ni se deduce.
func ahfContacto(n ahfNecesidad) string {
	var partes []string
	if v := strings.TrimSpace(n.ContactName); v != "" {
		partes = append(partes, v)
	}
	tel := strings.TrimSpace(n.ContactPhone)
	if tel != "" {
		partes = append(partes, tel)
	}
	if wa := strings.TrimSpace(n.ContactWhatsapp); wa != "" && wa != tel {
		partes = append(partes, "WhatsApp "+wa)
	}
	return strings.Join(partes, " — ")
}

func ahfCoord(v float64) *float64 {
	if v == 0 {
		return nil
	}
	return &v
}

func ahfCreado(n ahfNecesidad) *time.Time {
	if t := ahfFecha(n.CreatedAt); t != nil {
		return t
	}
	if n.CreationTime > 0 {
		t := time.UnixMilli(int64(n.CreationTime)).UTC()
		return &t
	}
	return nil
}

func ahfFecha(s string) *time.Time {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	for _, f := range []string{time.RFC3339Nano, time.RFC3339, "2006-01-02T15:04:05.000Z", "2006-01-02"} {
		if t, err := time.Parse(f, s); err == nil {
			u := t.UTC()
			return &u
		}
	}
	return nil
}

func ahfNumero(v float64) string {
	return strconv.FormatFloat(v, 'f', -1, 64)
}

func ahfCapitalizar(s string) string {
	s = strings.ToLower(strings.TrimSpace(strings.ReplaceAll(s, "_", " ")))
	if s == "" {
		return ""
	}
	r := []rune(s)
	return strings.ToUpper(string(r[0])) + string(r[1:])
}
