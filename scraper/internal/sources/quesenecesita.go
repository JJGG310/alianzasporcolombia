package sources

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/JJGG310/alianzasporcolombia/scraper/internal/fetch"
	"github.com/JJGG310/alianzasporcolombia/scraper/internal/model"
)

// QueSeNecesita lee el inventario de necesidades de quesenecesita.org, que se
// publica como una hoja de cálculo de Google exportada a CSV (la URL sale de la
// constante SHEET_CSV_URL incrustada en el HTML del sitio).
type QueSeNecesita struct{}

const (
	qsnSitio  = "https://quesenecesita.org/"
	qsnCSVURL = "https://docs.google.com/spreadsheets/d/e/2PACX-1vTgKZXN8HuIpF-d2qbJuZ6lL8txaDFNvM6nFPyoSO5F1W1i-2-3oO9V21NvpZV23MNYiAOH21kJoagj/pub?gid=371359520&single=true&output=csv"
)

func (q *QueSeNecesita) Slug() string { return "quesenecesita" }
func (q *QueSeNecesita) URL() string  { return qsnSitio }
func (q *QueSeNecesita) Mecanismo() string {
	return "Google Sheets publicado como CSV (descubierto en SHEET_CSV_URL del HTML)"
}

// Municipios del Valle del Cauca que aparecen en la fuente. Solo se completa el
// departamento para estos; para el resto se deja vacío antes que inventarlo.
var qsnValle = map[string]bool{
	"cali": true, "santiago de cali": true,
	"palmira": true, "yumbo": true,
	"jamundi": true, "jamundí": true,
	"buga": true, "guadalajara de buga": true,
	"tulua": true, "tuluá": true,
	"buenaventura": true, "roldanillo": true,
	"zarzal": true, "cartago": true,
	"candelaria": true, "florida": true,
	"pradera": true, "dagua": true,
	"vijes": true, "yotoco": true,
}

// qsnSitioAcum junta las filas (una por ítem) que comparten sitio_id.
type qsnSitioAcum struct {
	id       string
	filas    []map[string]string
	items    []string
	maxFecha time.Time
	urgente  bool
	critica  bool
}

func (q *QueSeNecesita) Obtener(ctx context.Context, c *fetch.Cliente) (*model.Resultado, error) {
	inicio := time.Now()
	res := &model.Resultado{
		Fuente:    q.Slug(),
		URL:       q.URL(),
		Mecanismo: q.Mecanismo(),
	}
	fallar := func(err error) (*model.Resultado, error) {
		res.OK = false
		res.Error = err.Error()
		res.Duracion = time.Since(inicio).String()
		res.Obtenido = time.Now().UTC().Format(time.RFC3339)
		res.Total = len(res.Puntos)
		return res, err
	}

	cuerpo, err := c.Get(ctx, qsnCSVURL)
	if err != nil {
		return fallar(fmt.Errorf("descargando csv: %w", err))
	}

	r := csv.NewReader(bytes.NewReader(cuerpo))
	r.FieldsPerRecord = -1 // la hoja puede ganar o perder columnas sin avisar
	r.LazyQuotes = true

	cab, err := r.Read()
	if err != nil {
		return fallar(fmt.Errorf("leyendo cabecera: %w", err))
	}
	idx := map[string]int{}
	for i, nombre := range cab {
		idx[strings.ToLower(strings.TrimSpace(nombre))] = i
	}
	if _, ok := idx["sitio_id"]; !ok {
		return fallar(errors.New("el csv no trae columna sitio_id: cambió el esquema"))
	}

	// Se conserva el orden de aparición para que la salida sea estable entre corridas.
	acums := map[string]*qsnSitioAcum{}
	var orden []string
	var filasMalas int

	for {
		fila, err := r.Read()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			filasMalas++
			continue
		}
		get := func(col string) string {
			i, ok := idx[col]
			if !ok || i >= len(fila) {
				return ""
			}
			return strings.TrimSpace(fila[i])
		}

		sitioID := get("sitio_id")
		if sitioID == "" {
			filasMalas++
			continue
		}

		a := acums[sitioID]
		if a == nil {
			a = &qsnSitioAcum{id: sitioID}
			acums[sitioID] = a
			orden = append(orden, sitioID)
		}

		registro := map[string]string{}
		for col, i := range idx {
			if i < len(fila) {
				registro[col] = fila[i]
			}
		}
		a.filas = append(a.filas, registro)

		if item := qsnItem(get("item"), get("cantidad"), get("unidad")); item != "" {
			a.items = append(a.items, item)
		}
		switch strings.ToLower(get("urgencia")) {
		case "critica", "crítica":
			a.critica, a.urgente = true, true
		case "alta":
			a.urgente = true
		}
		if t, err := time.Parse(time.RFC3339, get("actualizado")); err == nil && t.After(a.maxFecha) {
			a.maxFecha = t
		}
	}

	corte := time.Now().UTC()
	puntos := make([]model.Punto, 0, len(orden))
	for _, id := range orden {
		puntos = append(puntos, qsnPunto(acums[id], corte))
	}

	res.Puntos = puntos
	res.Total = len(puntos)
	res.OK = true
	res.Extra = map[string]any{
		"csv_url":     qsnCSVURL,
		"filas_items": contarFilas(acums),
		"sitios":      len(puntos),
	}
	if filasMalas > 0 {
		res.Nota = fmt.Sprintf("%d filas ilegibles o sin sitio_id descartadas", filasMalas)
	}
	res.Duracion = time.Since(inicio).String()
	res.Obtenido = time.Now().UTC().Format(time.RFC3339)
	return res, nil
}

func qsnPunto(a *qsnSitioAcum, corte time.Time) model.Punto {
	// Los campos del sitio se repiten en cada fila: basta el primer valor no vacío.
	campo := func(col string) string {
		for _, f := range a.filas {
			if v := strings.TrimSpace(f[col]); v != "" {
				return v
			}
		}
		return ""
	}

	ciudad := campo("ciudad")
	p := model.Punto{
		ID:           model.HacerID("quesenecesita", a.id),
		Fuente:       "quesenecesita",
		FuenteURL:    qsnSitio + "#" + a.id,
		Tipo:         model.TipoAcopio, // son puntos que reciben/necesitan insumos
		Nombre:       campo("sitio"),
		Municipio:    ciudad,
		Direccion:    campo("direccion"),
		Organizacion: campo("operador"),
		Contacto:     campo("contacto"),
		Necesita:     a.items,
	}
	if qsnValle[strings.ToLower(strings.TrimSpace(ciudad))] {
		p.Departamento = "Valle del Cauca"
	}

	switch strings.ToLower(campo("estado")) {
	case "abierta", "abierto":
		p.Estado = model.EstadoActivo
	case "cerrada", "cerrado":
		p.Estado = model.EstadoCerrado
	default:
		p.Estado = model.EstadoPorConfirmar
	}
	// Una sola necesidad urgente marca todo el sitio: es lo que un voluntario
	// necesita ver en la lista.
	if a.urgente && p.Estado != model.EstadoCerrado {
		p.Estado = model.EstadoUrgente
	}
	switch {
	case a.critica:
		p.Prioridad = model.PrioridadCritica
	case a.urgente:
		p.Prioridad = model.PrioridadAlta
	}

	if !a.maxFecha.IsZero() {
		t := a.maxFecha.UTC()
		p.Actualizado = &t
	}
	// Lat/Lon quedan nil: el CSV no publica coordenadas.

	if crudo, err := json.Marshal(a.filas); err == nil {
		p.Raw = crudo
	}

	model.Normalizar(&p, corte)
	return p
}

// qsnItem arma "Agua potable (200 botellas)" cuando hay cantidad y/o unidad.
func qsnItem(item, cantidad, unidad string) string {
	item = strings.TrimSpace(item)
	if item == "" {
		return ""
	}
	detalle := strings.TrimSpace(cantidad + " " + unidad)
	if detalle == "" {
		return item
	}
	return item + " (" + detalle + ")"
}

func contarFilas(acums map[string]*qsnSitioAcum) int {
	n := 0
	for _, a := range acums {
		n += len(a.filas)
	}
	return n
}
