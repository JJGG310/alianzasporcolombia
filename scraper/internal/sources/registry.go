// Package sources tiene un adaptador por sitio de origen.
//
// Cada adaptador es responsable de UNA fuente y traduce su vocabulario al
// esquema de model.Punto. Si una fuente cambia de forma, solo se toca su archivo.
package sources

import (
	"context"

	"github.com/JJGG310/alianzasporcolombia/scraper/internal/fetch"
	"github.com/JJGG310/alianzasporcolombia/scraper/internal/model"
)

// Fuente es el contrato que implementa cada adaptador.
type Fuente interface {
	// Slug identifica la fuente en los archivos de salida. Estable para siempre.
	Slug() string
	// URL es el sitio público de origen (para citar).
	URL() string
	// Mecanismo describe cómo se obtienen los datos (para el README y auditoría).
	Mecanismo() string
	// Obtener trae y normaliza. Debe devolver puntos ya pasados por model.Normalizar.
	Obtener(ctx context.Context, c *fetch.Cliente) (*model.Resultado, error)
}

// Todas devuelve el registro completo de fuentes, en orden estable.
func Todas() []Fuente {
	return []Fuente{
		&MapaEmergencia{},
		&AquiHaceFalta{},
		&HaciendoComunidad{},
		&CaliSolidario{},
		&QueSeNecesita{},
		&PuntosCriticos{},
		&AidTrace{},
		&WindowOfHope{},
	}
}

// Por devuelve las fuentes cuyo slug esté en la lista. Lista vacía = todas.
func Por(slugs []string) []Fuente {
	if len(slugs) == 0 {
		return Todas()
	}
	quiero := map[string]bool{}
	for _, s := range slugs {
		quiero[s] = true
	}
	var out []Fuente
	for _, f := range Todas() {
		if quiero[f.Slug()] {
			out = append(out, f)
		}
	}
	return out
}
