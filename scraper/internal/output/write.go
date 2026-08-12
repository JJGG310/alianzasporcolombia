// Package output escribe extracted-data/ de forma determinista.
//
// Determinista importa: estos archivos van a git. Si dos corridas seguidas
// producen bytes distintos sin que los datos hayan cambiado, el diff deja de
// servir para revisar qué cambió de verdad — y revisar el diff es justamente
// el control humano que exige el proyecto.
package output

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/JJGG310/alianzasporcolombia/scraper/internal/merge"
	"github.com/JJGG310/alianzasporcolombia/scraper/internal/model"
)

const (
	Licencia = "CC BY 4.0 — datos abiertos de Alianzas por Colombia. Cita la fuente de cada ítem."
	Politica = "Nunca publicamos cuentas bancarias ni datos de pago. Contacta a la organización y valida con ella antes de entregar dinero."
)

// Consolidado es extracted-data/puntos.json: lo que consume el sitio.
type Consolidado struct {
	Generado string          `json:"generado"`
	Licencia string          `json:"licencia"`
	Politica string          `json:"politica"`
	Aviso    string          `json:"aviso"`
	Total    int             `json:"total"`
	Informe  merge.Informe   `json:"informe"`
	Fuentes  []ResumenFuente `json:"fuentes"`
	Puntos   []model.Punto   `json:"puntos"`
}

// ResumenFuente es la ficha de auditoría de cada sitio de origen.
type ResumenFuente struct {
	Fuente    string `json:"fuente"`
	URL       string `json:"url"`
	Mecanismo string `json:"mecanismo"`
	OK        bool   `json:"ok"`
	Total     int    `json:"total"`
	Obtenido  string `json:"obtenido"`
	Duracion  string `json:"duracion"`
	Nota      string `json:"nota,omitempty"`
	Error     string `json:"error,omitempty"`
}

// Escribir vuelca todo a disco bajo dir (normalmente extracted-data/).
func Escribir(dir string, resultados []*model.Resultado, puntos []model.Punto, inf merge.Informe, generado time.Time) error {
	if err := os.MkdirAll(filepath.Join(dir, "raw"), 0o755); err != nil {
		return fmt.Errorf("creando %s: %w", dir, err)
	}

	// 1) Crudo por fuente: permite reprocesar sin volver a golpear los sitios.
	for _, r := range resultados {
		if r == nil {
			continue
		}
		ruta := filepath.Join(dir, "raw", r.Fuente+".json")
		if err := escribirJSON(ruta, r); err != nil {
			return err
		}
	}

	// 2) Consolidado normalizado y fusionado.
	resumenes := make([]ResumenFuente, 0, len(resultados))
	for _, r := range resultados {
		if r == nil {
			continue
		}
		resumenes = append(resumenes, ResumenFuente{
			Fuente: r.Fuente, URL: r.URL, Mecanismo: r.Mecanismo, OK: r.OK,
			Total: r.Total, Obtenido: r.Obtenido, Duracion: r.Duracion,
			Nota: r.Nota, Error: r.Error,
		})
	}
	sort.Slice(resumenes, func(i, j int) bool { return resumenes[i].Fuente < resumenes[j].Fuente })

	// Orden estable de puntos: por fuente y luego por id. Sin esto el diff de
	// git sería ilegible aunque nada hubiera cambiado.
	ordenados := make([]model.Punto, len(puntos))
	copy(ordenados, puntos)
	sort.SliceStable(ordenados, func(i, j int) bool {
		if ordenados[i].Fuente != ordenados[j].Fuente {
			return ordenados[i].Fuente < ordenados[j].Fuente
		}
		return ordenados[i].ID < ordenados[j].ID
	})

	// El consolidado publicable no lleva Raw: pesa mucho y ya está en raw/.
	sinCrudo := make([]model.Punto, len(ordenados))
	copy(sinCrudo, ordenados)
	for i := range sinCrudo {
		sinCrudo[i].Raw = nil
	}

	cons := Consolidado{
		Generado: generado.UTC().Format(time.RFC3339),
		Licencia: Licencia,
		Politica: Politica,
		Aviso:    "Datos agregados automáticamente de sitios ciudadanos de terceros. No están verificados uno por uno por Alianzas por Colombia: cada punto cita su fuente para que puedas confirmarlo. Los puntos con 'politica_alerta' requieren revisión humana antes de publicarse.",
		Total:    len(sinCrudo),
		Informe:  inf,
		Fuentes:  resumenes,
		Puntos:   sinCrudo,
	}
	if err := escribirJSON(filepath.Join(dir, "puntos.json"), cons); err != nil {
		return err
	}

	// 3) Ficha corta de fuentes: útil para el README y para el panel del sitio.
	if err := escribirJSON(filepath.Join(dir, "fuentes.json"), map[string]any{
		"generado": cons.Generado,
		"fuentes":  resumenes,
		"informe":  inf,
	}); err != nil {
		return err
	}

	// 4) Solo lo mapeable, como GeoJSON: para que prensa y aliados lo consuman
	//    sin escribir código.
	return escribirJSON(filepath.Join(dir, "puntos.geojson"), geojson(sinCrudo, cons.Generado))
}

func escribirJSON(ruta string, v any) error {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Errorf("serializando %s: %w", ruta, err)
	}
	b = append(b, '\n')
	// Escritura atómica: si el proceso muere a mitad, no dejamos un JSON
	// truncado que rompa el sitio en producción.
	tmp := ruta + ".tmp"
	if err := os.WriteFile(tmp, b, 0o644); err != nil {
		return fmt.Errorf("escribiendo %s: %w", tmp, err)
	}
	return os.Rename(tmp, ruta)
}

func geojson(puntos []model.Punto, generado string) map[string]any {
	features := []map[string]any{}
	for _, p := range puntos {
		if p.Lat == nil || p.Lon == nil {
			continue
		}
		features = append(features, map[string]any{
			"type": "Feature",
			"geometry": map[string]any{
				"type":        "Point",
				"coordinates": []float64{*p.Lon, *p.Lat},
			},
			"properties": map[string]any{
				"id": p.ID, "nombre": p.Nombre, "tipo": p.Tipo, "estado": p.Estado,
				"direccion": p.Direccion, "barrio": p.Barrio, "municipio": p.Municipio,
				"departamento": p.Departamento, "necesita": p.Necesita,
				"contacto": p.Contacto, "fuente": p.Fuente, "fuente_url": p.FuenteURL,
				"fecha_corte": p.FechaCorte, "verificado": p.Verificado,
			},
		})
	}
	return map[string]any{
		"type":     "FeatureCollection",
		"licencia": Licencia,
		"generado": generado,
		"features": features,
	}
}
