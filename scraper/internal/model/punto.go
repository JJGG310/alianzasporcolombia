// Package model define el esquema normalizado al que toda fuente se traduce.
//
// Regla del proyecto: cada punto lleva SIEMPRE fuente + fecha de corte, y jamás
// datos de pago (ver sanitize.go).
package model

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"sort"
	"strings"
	"time"
)

// Tipos canónicos. Los adaptadores mapean su vocabulario a estos.
const (
	TipoAlbergue    = "albergue"
	TipoAcopio      = "acopio"
	TipoSangre      = "sangre"
	TipoOlla        = "olla"
	TipoDonacionOrg = "donacion-org"
	TipoRescate     = "rescate"
	TipoSalud       = "salud"
	TipoAgua        = "agua"
	TipoVeterinario = "veterinario"
	TipoPsicologico = "psicologico"
	TipoNecesidad   = "necesidad" // pedido ciudadano puntual
	TipoOferta      = "oferta"    // alguien ofrece ayuda
	TipoInfo        = "info"
	TipoOtro        = "otro"
)

// Estados canónicos.
const (
	EstadoUrgente      = "urgente"
	EstadoActivo       = "activo"
	EstadoCubierto     = "cubierto"
	EstadoCerrado      = "cerrado"
	EstadoPorConfirmar = "por-confirmar"
	EstadoDescartado   = "descartado"
)

// Prioridades canónicas.
const (
	PrioridadCritica = "critica"
	PrioridadAlta    = "alta"
	PrioridadMedia   = "media"
	PrioridadBaja    = "baja"
)

// Punto es un ítem de ayuda normalizado, venga de donde venga.
type Punto struct {
	ID        string `json:"id"`         // estable: <fuente>-<hash>
	Fuente    string `json:"fuente"`     // slug del sitio de origen
	FuenteURL string `json:"fuente_url"` // URL citable del ítem o del sitio

	Tipo        string `json:"tipo"`
	Nombre      string `json:"nombre"`
	Descripcion string `json:"descripcion,omitempty"`

	Departamento string   `json:"departamento,omitempty"`
	Municipio    string   `json:"municipio,omitempty"`
	Comuna       string   `json:"comuna,omitempty"`
	Barrio       string   `json:"barrio,omitempty"`
	Direccion    string   `json:"direccion,omitempty"`
	Lat          *float64 `json:"lat,omitempty"`
	Lon          *float64 `json:"lon,omitempty"`

	Necesita []string `json:"necesita,omitempty"` // qué le hace falta a este punto
	Ofrece   []string `json:"ofrece,omitempty"`   // qué entrega o recibe

	Estado    string `json:"estado,omitempty"`
	Prioridad string `json:"prioridad,omitempty"`

	VoluntariosHay    int `json:"voluntarios_hay,omitempty"`
	VoluntariosFaltan int `json:"voluntarios_faltan,omitempty"`

	Contacto     string `json:"contacto,omitempty"`
	Organizacion string `json:"organizacion,omitempty"`
	Horario      string `json:"horario,omitempty"`

	Verificado     bool `json:"verificado"`
	Confirmaciones int  `json:"confirmaciones,omitempty"`
	Desmentidos    int  `json:"desmentidos,omitempty"`

	Creado      *time.Time `json:"creado,omitempty"`
	Actualizado *time.Time `json:"actualizado,omitempty"`
	FechaCorte  string     `json:"fecha_corte"` // ISO8601 del momento del scrape

	// PoliticaAlerta es BLOQUEANTE: el punto no se publica sin que un humano lo
	// revise. Solo se llena por violaciones de política (posible dato de pago) o
	// por señales de abuso del sitio de origen.
	PoliticaAlerta []string `json:"politica_alerta,omitempty"`

	// Avisos es informativo y NO bloquea la publicación: cosas que conviene
	// saber del dato (coordenadas descartadas, marcado como viejo en su sitio).
	// Separado de PoliticaAlerta a propósito: si todo cae en el mismo cajón, un
	// punto perfectamente publicable se esconde por tener mal el lat/lon.
	Avisos []string `json:"avisos,omitempty"`

	// TambienEn lista las otras fuentes que reportan este mismo punto. Que dos
	// sitios ciudadanos independientes coincidan es la mejor señal de confianza
	// que tenemos sin llamar por teléfono.
	TambienEn []Referencia `json:"tambien_en,omitempty"`

	// Raw conserva el registro original para poder reprocesar sin re-scrapear.
	Raw json.RawMessage `json:"raw,omitempty"`
}

// Resultado es lo que devuelve cada adaptador.
type Resultado struct {
	Fuente    string  `json:"fuente"`
	URL       string  `json:"url"`
	Mecanismo string  `json:"mecanismo"` // cómo se obtuvo (api rest, rsc, csv, convex, html...)
	OK        bool    `json:"ok"`
	Error     string  `json:"error,omitempty"`
	Nota      string  `json:"nota,omitempty"`
	Duracion  string  `json:"duracion"`
	Obtenido  string  `json:"obtenido"` // ISO8601
	Total     int     `json:"total"`
	Puntos    []Punto `json:"puntos"`
	Extra     any     `json:"extra,omitempty"` // novedades, cifras, avisos sueltos
}

// HacerID genera un id estable y determinista para un punto.
// Determinista importa: si el ítem no cambió, el id no cambia entre corridas,
// y el diff contra la corrida anterior es legible.
func HacerID(fuente string, partes ...string) string {
	h := sha1.New()
	for _, p := range partes {
		h.Write([]byte(strings.ToLower(strings.TrimSpace(p))))
		h.Write([]byte("\x00"))
	}
	return fuente + "-" + hex.EncodeToString(h.Sum(nil))[:10]
}

// Normalizar limpia espacios, ordena listas y aplica el saneamiento de política.
// Todo adaptador debe llamarlo antes de devolver.
func Normalizar(p *Punto, corte time.Time) {
	p.Nombre = limpiar(p.Nombre)
	p.Descripcion = limpiar(p.Descripcion)
	p.Direccion = limpiar(p.Direccion)
	p.Barrio = limpiar(p.Barrio)
	p.Comuna = limpiar(p.Comuna)
	p.Municipio = limpiar(p.Municipio)
	p.Organizacion = limpiar(p.Organizacion)
	p.Contacto = limpiar(p.Contacto)

	p.Necesita = limpiarLista(p.Necesita)
	p.Ofrece = limpiarLista(p.Ofrece)

	if p.Tipo == "" {
		p.Tipo = TipoOtro
	}
	if p.Estado == "" {
		p.Estado = EstadoPorConfirmar
	}
	if p.FechaCorte == "" {
		p.FechaCorte = corte.Format(time.RFC3339)
	}
	// Coordenadas fuera de Colombia continental = dato malo, mejor sin mapa.
	if p.Lat != nil && p.Lon != nil {
		if *p.Lat < -5 || *p.Lat > 14 || *p.Lon < -82 || *p.Lon > -66 {
			p.Lat, p.Lon = nil, nil
			p.Avisos = append(p.Avisos, "coordenadas fuera de Colombia: descartadas")
		}
	}
	Sanear(p)
}

func limpiar(s string) string {
	return strings.TrimSpace(strings.Join(strings.Fields(s), " "))
}

func limpiarLista(in []string) []string {
	visto := map[string]bool{}
	out := make([]string, 0, len(in))
	for _, s := range in {
		s = limpiar(s)
		if s == "" || visto[strings.ToLower(s)] {
			continue
		}
		visto[strings.ToLower(s)] = true
		out = append(out, s)
	}
	sort.Strings(out)
	return out
}
