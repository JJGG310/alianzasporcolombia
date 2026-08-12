package sources

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/JJGG310/alianzasporcolombia/scraper/internal/fetch"
	"github.com/JJGG310/alianzasporcolombia/scraper/internal/model"
)

// SismoInfo — https://sismoinfo.co/
//
// Tiene API REST propia bajo /api. Se leen dos colecciones:
//   - buildings: edificios revisados, con su estado estructural
//   - anuncios:  tablón de avisos ciudadanos
//
// NO se lee /api/disappeared. Es el principio #2 del proyecto: no duplicamos
// bases de personas desaparecidas. Ese endpoint además sirve fotos
// (/disappeared/{id}/photo), muchas de menores. Se enlaza, no se copia.
type SismoInfo struct{}

const siBase = "https://sismoinfo.co/api"

func (s *SismoInfo) Slug() string { return "sismoinfo" }
func (s *SismoInfo) URL() string  { return "https://sismoinfo.co/" }
func (s *SismoInfo) Mecanismo() string {
	return "API REST propia (/api/buildings y /api/anuncios; /api/disappeared se omite a propósito)"
}

type siEdificio struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	Status       string   `json:"status"`
	Location     *string  `json:"location"`
	City         *string  `json:"city"`
	Lat          *float64 `json:"lat"`
	Lng          *float64 `json:"lng"`
	CreatedAt    string   `json:"created_at"`
	LastComment  *string  `json:"last_comment"`
	CommentCount int      `json:"comment_count"`
}

type siAnuncio struct {
	ID        string `json:"id"`
	Text      string `json:"text"`
	CreatedAt string `json:"created_at"`
}

func (s *SismoInfo) Obtener(ctx context.Context, c *fetch.Cliente) (*model.Resultado, error) {
	inicio := time.Now()
	res := &model.Resultado{
		Fuente: s.Slug(), URL: s.URL(), Mecanismo: s.Mecanismo(),
		Obtenido: time.Now().UTC().Format(time.RFC3339),
	}
	terminar := func() {
		res.Duracion = time.Since(inicio).String()
		res.Total = len(res.Puntos)
	}

	corte := time.Now().UTC()
	conteos := map[string]int{}
	var notas []string

	var edificios []siEdificio
	if err := c.GetJSON(ctx, siBase+"/buildings", &edificios); err != nil {
		res.OK = false
		res.Error = fmt.Sprintf("buildings: %v", err)
		terminar()
		return res, err
	}
	for _, e := range edificios {
		if p, ok := siPuntoDeEdificio(e, corte); ok {
			res.Puntos = append(res.Puntos, p)
			conteos["buildings"]++
		}
	}

	// Que falle el tablón no debe tumbar los edificios.
	var anuncios []siAnuncio
	if err := c.GetJSON(ctx, siBase+"/anuncios", &anuncios); err != nil {
		notas = append(notas, "anuncios no disponibles: "+err.Error())
	} else {
		for _, a := range anuncios {
			if p, ok := siPuntoDeAnuncio(a, corte); ok {
				res.Puntos = append(res.Puntos, p)
				conteos["anuncios"]++
			}
		}
	}

	res.OK = true
	notas = append(notas, "no se leyó /api/disappeared por política de datos personales")
	res.Nota = strings.Join(notas, "; ")
	res.Extra = map[string]any{
		"colecciones_leidas":   conteos,
		"coleccion_omitida":    "disappeared",
		"motivo_de_la_omision": "datos personales y fotos de personas: no duplicamos bases de desaparecidos (principio #2). Se enlaza a Colombia Te Busca.",
	}
	terminar()
	return res, nil
}

// El estado estructural que reporta el sitio. "seguro" es tan valioso como
// "colapsado": un edificio revisado y seguro es a donde la gente puede volver.
var siEstadoEdificio = map[string]struct {
	estado, prioridad, nota string
}{
	"colapsado": {model.EstadoUrgente, model.PrioridadCritica, "Edificio reportado como colapsado"},
	"danado":    {model.EstadoActivo, model.PrioridadAlta, "Edificio reportado como dañado"},
	"dañado":    {model.EstadoActivo, model.PrioridadAlta, "Edificio reportado como dañado"},
	"seguro":    {model.EstadoCubierto, "", "Edificio revisado y reportado como seguro"},
}

func siPuntoDeEdificio(e siEdificio, corte time.Time) (model.Punto, bool) {
	if e.ID == "" || strings.TrimSpace(e.Name) == "" {
		return model.Punto{}, false
	}

	p := model.Punto{
		ID:        model.HacerID("sismoinfo", e.ID),
		Fuente:    "sismoinfo",
		FuenteURL: "https://sismoinfo.co/",
		// Un edificio colapsado o dañado es un punto de rescate; uno revisado y
		// seguro es información para quien no sabe si puede volver a su casa.
		Tipo:   model.TipoRescate,
		Nombre: e.Name,
		Lat:    e.Lat,
		Lon:    e.Lng,
		Creado: siFecha(e.CreatedAt),
	}
	if e.Location != nil {
		p.Direccion = *e.Location
	}
	if e.City != nil {
		p.Municipio = *e.City
		if d := siDepartamento(*e.City); d != "" {
			p.Departamento = d
		}
	}

	est, ok := siEstadoEdificio[strings.ToLower(strings.TrimSpace(e.Status))]
	if ok {
		p.Estado = est.estado
		p.Prioridad = est.prioridad
		p.Descripcion = est.nota
		if est.estado == model.EstadoCubierto {
			p.Tipo = model.TipoInfo
		}
	} else {
		p.Estado = model.EstadoPorConfirmar
	}

	// El último comentario suele traer el dato operativo real ("sin comer desde
	// ayer", "no han podido entrar"), que vale más que la etiqueta de estado.
	if e.LastComment != nil && strings.TrimSpace(*e.LastComment) != "" {
		if p.Descripcion != "" {
			p.Descripcion += ". "
		}
		p.Descripcion += strings.TrimSpace(*e.LastComment)
	}
	p.Confirmaciones = e.CommentCount

	p.Raw, _ = json.Marshal(e)
	model.Normalizar(&p, corte)
	return p, true
}

func siPuntoDeAnuncio(a siAnuncio, corte time.Time) (model.Punto, bool) {
	texto := strings.TrimSpace(a.Text)
	if a.ID == "" || texto == "" {
		return model.Punto{}, false
	}

	p := model.Punto{
		ID:        model.HacerID("sismoinfo", "anuncio", a.ID),
		Fuente:    "sismoinfo",
		FuenteURL: "https://sismoinfo.co/",
		// Son avisos de gente que ofrece algo ("sirvo gratis", "llevo agua").
		Tipo:        model.TipoOferta,
		Nombre:      siTitulo(texto),
		Descripcion: texto,
		Estado:      model.EstadoActivo,
		Creado:      siFecha(a.CreatedAt),
	}
	p.Raw, _ = json.Marshal(a)
	model.Normalizar(&p, corte)
	return p, true
}

// siTitulo saca un nombre corto de un texto libre: la primera frase, recortada.
// Sin esto la tarjeta muestra un párrafo entero como título.
func siTitulo(texto string) string {
	limpio := strings.Join(strings.Fields(texto), " ")
	for _, corte := range []string{". ", "\n", " - "} {
		if i := strings.Index(limpio, corte); i > 12 && i < 90 {
			return limpio[:i]
		}
	}
	if len(limpio) > 90 {
		// Se corta en espacio para no partir una palabra por la mitad.
		if i := strings.LastIndex(limpio[:90], " "); i > 30 {
			return limpio[:i] + "…"
		}
		return limpio[:90] + "…"
	}
	return limpio
}

func siFecha(iso string) *time.Time {
	t, err := time.Parse(time.RFC3339, iso)
	if err != nil {
		return nil
	}
	t = t.UTC()
	return &t
}

// Solo municipios que aparecen de verdad en la fuente. No se adivina.
var siDeptos = map[string]string{
	"cali":      "Valle del Cauca",
	"pereira":   "Risaralda",
	"palmira":   "Valle del Cauca",
	"jamundí":   "Valle del Cauca",
	"jamundi":   "Valle del Cauca",
	"yumbo":     "Valle del Cauca",
	"manizales": "Caldas",
	"armenia":   "Quindío",
	"quibdó":    "Chocó",
	"quibdo":    "Chocó",
}

func siDepartamento(ciudad string) string {
	return siDeptos[strings.ToLower(strings.TrimSpace(ciudad))]
}
