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

// MapaEmergencia lee el mapa colaborativo del sismo (Valle, Chocó y Cauca).
// Publica su estado completo en un endpoint JSON, así que no hay que raspar HTML.
type MapaEmergencia struct{}

const (
	meBase           = "https://mapa-emergencia.artefactofilms.workers.dev"
	meURLSnapshot    = meBase + "/api/snapshot"
	meURLNovedades   = meBase + "/api/novedades"
	meURLConfig      = meBase + "/api/config"
	meMecanismoTexto = "API REST pública (/api/snapshot, /api/novedades, /api/config)"
)

func (m *MapaEmergencia) Slug() string      { return "mapa-emergencia" }
func (m *MapaEmergencia) URL() string       { return meBase }
func (m *MapaEmergencia) Mecanismo() string { return meMecanismoTexto }

// meSnapshot: los puntos se leen dos veces, tipados y crudos, para poder guardar
// el registro original en Punto.Raw sin re-serializar (y sin perder campos nuevos).
type meSnapshot struct {
	Rev           int               `json:"rev"`
	Generado      int64             `json:"generado"`
	VigenciaHoras int               `json:"vigencia_horas"`
	Resumen       json.RawMessage   `json:"resumen"`
	Puntos        []json.RawMessage `json:"puntos"`
}

type mePunto struct {
	ID                string   `json:"id"`
	Nombre            string   `json:"nombre"`
	Tipo              string   `json:"tipo"`
	Lat               *float64 `json:"lat"`
	Lng               *float64 `json:"lng"`
	Direccion         string   `json:"direccion"`
	Barrio            string   `json:"barrio"`
	Estado            string   `json:"estado"`
	Necesidades       []string `json:"necesidades"`
	VoluntariosHay    int      `json:"voluntarios_hay"`
	VoluntariosFaltan int      `json:"voluntarios_faltan"`
	Contacto          string   `json:"contacto"`
	Notas             string   `json:"notas"`
	Fuente            string   `json:"fuente"`
	Verificado        int      `json:"verificado"`
	Autor             string   `json:"autor"`
	Creado            int64    `json:"creado"`
	Actualizado       int64    `json:"actualizado"`
	Confirmado        int64    `json:"confirmado"`
	Ubicado           bool     `json:"ubicado"`
	Fresco            bool     `json:"fresco"`
	Saturacion        string   `json:"saturacion"`
	Confirma          int      `json:"confirma"`
	Desmiente         int      `json:"desmiente"`
	MediaTotal        int      `json:"media_total"`
}

// Vocabulario propio de la fuente -> vocabulario canónico del proyecto.
var meTipos = map[string]string{
	"rescate":     model.TipoRescate,
	"albergue":    model.TipoAlbergue,
	"acopio":      model.TipoAcopio,
	"agua":        model.TipoAgua,
	"salud":       model.TipoSalud,
	"cocina":      model.TipoOlla,
	"psicologico": model.TipoPsicologico,
	"veterinario": model.TipoVeterinario,
	"info":        model.TipoInfo,
	"otro":        model.TipoOtro,
}

var meEstados = map[string]string{
	"urgente":    model.EstadoUrgente,
	"necesita":   model.EstadoActivo,
	"cubierto":   model.EstadoCubierto,
	"cerrado":    model.EstadoCerrado,
	"descartado": model.EstadoDescartado,
}

func (m *MapaEmergencia) Obtener(ctx context.Context, c *fetch.Cliente) (*model.Resultado, error) {
	inicio := time.Now()
	res := &model.Resultado{
		Fuente:    m.Slug(),
		URL:       m.URL(),
		Mecanismo: m.Mecanismo(),
	}
	terminar := func() { // se llama en todas las salidas para no olvidar métricas
		res.Duracion = time.Since(inicio).String()
		res.Obtenido = time.Now().UTC().Format(time.RFC3339)
		res.Total = len(res.Puntos)
	}

	var snap meSnapshot
	if err := c.GetJSON(ctx, meURLSnapshot, &snap); err != nil {
		res.OK = false
		res.Error = fmt.Sprintf("snapshot: %v", err)
		terminar()
		return res, fmt.Errorf("mapa-emergencia snapshot: %w", err)
	}

	corte := time.Now().UTC()
	puntos := make([]model.Punto, 0, len(snap.Puntos))
	var descartados int
	for _, crudo := range snap.Puntos {
		var mp mePunto
		if err := json.Unmarshal(crudo, &mp); err != nil {
			descartados++
			continue
		}
		if mp.ID == "" {
			descartados++
			continue
		}
		puntos = append(puntos, mePuntoAModelo(mp, crudo, corte))
	}
	res.Puntos = puntos
	res.OK = true // el snapshot mandó: los endpoints auxiliares son opcionales

	// Novedades y config son contexto útil, no datos críticos: si fallan se anota
	// la nota y se sigue, en vez de tumbar toda la fuente.
	var notas []string
	if descartados > 0 {
		notas = append(notas, fmt.Sprintf("%d puntos ilegibles descartados", descartados))
	}

	extra := map[string]any{"resumen": snap.Resumen}

	var novedades json.RawMessage
	if err := c.GetJSON(ctx, meURLNovedades, &novedades); err != nil {
		notas = append(notas, "novedades no disponibles: "+err.Error())
	} else {
		extra["novedades"] = novedades
	}

	var config json.RawMessage
	if err := c.GetJSON(ctx, meURLConfig, &config); err != nil {
		notas = append(notas, "config no disponible: "+err.Error())
	} else {
		extra["config"] = config
	}

	res.Extra = extra
	res.Nota = strings.Join(notas, "; ")
	terminar()
	return res, nil
}

func mePuntoAModelo(mp mePunto, crudo json.RawMessage, corte time.Time) model.Punto {
	p := model.Punto{
		ID:                model.HacerID("mapa-emergencia", mp.ID),
		Fuente:            "mapa-emergencia",
		FuenteURL:         meBase + "/#p=" + mp.ID,
		Tipo:              meTipos[mp.Tipo],
		Nombre:            mp.Nombre,
		Descripcion:       mp.Notas,
		Barrio:            mp.Barrio,
		Direccion:         mp.Direccion,
		Lat:               mp.Lat,
		Lon:               mp.Lng,
		Necesita:          mp.Necesidades,
		Estado:            meEstados[mp.Estado],
		VoluntariosHay:    mp.VoluntariosHay,
		VoluntariosFaltan: mp.VoluntariosFaltan,
		Contacto:          mp.Contacto,
		Organizacion:      mp.Autor,
		Verificado:        mp.Verificado != 0,
		Confirmaciones:    mp.Confirma,
		Desmentidos:       mp.Desmiente,
		Creado:            msATiempo(mp.Creado),
		Actualizado:       msATiempo(mp.Actualizado),
		Raw:               crudo,
	}
	if p.Tipo == "" {
		p.Tipo = model.TipoOtro
	}
	// Departamento/Municipio: la fuente no los trae y las coordenadas cruzan tres
	// departamentos, así que se dejan vacíos en vez de adivinar.
	switch mp.Estado {
	case "urgente":
		p.Prioridad = model.PrioridadCritica
	case "necesita":
		p.Prioridad = model.PrioridadAlta
	}
	// La fuente marca como "fresco" lo que sigue vigente según su propio reloj.
	// Republicar como URGENTE un punto que allá ya está vencido es la peor forma
	// de mentir: manda gente a un sitio que puede llevar horas resuelto, y de
	// paso quema la señal de urgencia para los puntos que sí lo están. Cuando la
	// fuente duda, nosotros dudamos.
	if !mp.Fresco && p.Estado == model.EstadoUrgente {
		p.Estado = model.EstadoPorConfirmar
		p.Prioridad = model.PrioridadAlta
		p.Avisos = append(p.Avisos,
			"el sitio de origen ya no lo tiene como vigente: confirma antes de ir")
	}

	model.Normalizar(&p, corte)
	return p
}

// msATiempo convierte epoch en milisegundos a *time.Time; 0 significa "sin dato".
func msATiempo(ms int64) *time.Time {
	if ms == 0 {
		return nil
	}
	t := time.UnixMilli(ms).UTC()
	return &t
}
