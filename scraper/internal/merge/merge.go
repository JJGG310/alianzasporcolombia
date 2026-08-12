// Package merge fusiona los puntos de todas las fuentes en una sola lista.
//
// El problema real: los sitios ciudadanos se copian entre sí. El mismo albergue
// aparece en tres tableros con el nombre escrito distinto. Si publicamos todo
// sin fusionar, el mapa miente sobre cuántos puntos hay.
//
// El criterio es deliberadamente conservador: preferimos dejar dos puntos
// separados (ruido visible) a fusionar dos puntos distintos (dato perdido).
package merge

import (
	"math"
	"regexp"
	"sort"
	"strings"

	"github.com/JJGG310/alianzasporcolombia/scraper/internal/model"
)

// Prioridad de fuentes: ante conflicto, gana la de más arriba. Los datos
// curados a mano mandan; después, la fuente con más estructura y confirmación.
var prioridadFuente = map[string]int{
	"curado":             0,
	"mapa-emergencia":    1,
	"aqui-hace-falta":    2,
	"haciendo-comunidad": 3,
	"quesenecesita":      4,
	"puntos-criticos":    5,
	"calisolidario":      6,
	"aidtrace":           7,
	"window-of-hope":     8,
}

// Fusionar recibe los puntos de todas las fuentes y devuelve la lista fusionada
// más un informe de cuántos se colapsaron.
func Fusionar(entrada []model.Punto) ([]model.Punto, Informe) {
	inf := Informe{Entrada: len(entrada)}

	// Orden estable por prioridad de fuente: el primero de cada grupo será el
	// canónico, así que queremos que sea el de la fuente más confiable.
	puntos := make([]model.Punto, len(entrada))
	copy(puntos, entrada)
	sort.SliceStable(puntos, func(i, j int) bool {
		pi, pj := prio(puntos[i].Fuente), prio(puntos[j].Fuente)
		if pi != pj {
			return pi < pj
		}
		return puntos[i].ID < puntos[j].ID
	})

	var canon []model.Punto
	// Índice por clave exacta (nombre+dirección normalizados) para el caso fácil.
	porClave := map[string]int{}

	for _, p := range puntos {
		if idx, ok := encontrarGemelo(p, canon, porClave); ok {
			absorber(&canon[idx], p)
			inf.Fusionados++
			continue
		}
		canon = append(canon, p)
		if k := claveExacta(p); k != "" {
			porClave[k] = len(canon) - 1
		}
	}

	inf.Salida = len(canon)
	inf.PorFuente = map[string]int{}
	inf.PorTipo = map[string]int{}
	for _, p := range canon {
		inf.PorFuente[p.Fuente]++
		inf.PorTipo[p.Tipo]++
		if p.Lat != nil && p.Lon != nil {
			inf.ConCoordenadas++
		}
		if len(p.PoliticaAlerta) > 0 {
			inf.RequierenRevision++
		}
		if len(p.TambienEn) > 0 {
			inf.Corroborados++
		}
	}
	return canon, inf
}

// Informe resume qué pasó al fusionar. Se publica junto a los datos: si el
// número de puntos baja de golpe entre corridas, esto dice por qué.
type Informe struct {
	Entrada           int            `json:"entrada"`
	Salida            int            `json:"salida"`
	Fusionados        int            `json:"fusionados"`
	ConCoordenadas    int            `json:"con_coordenadas"`
	RequierenRevision int            `json:"requieren_revision"`
	Corroborados      int            `json:"corroborados_por_2_o_mas_fuentes"`
	PorFuente         map[string]int `json:"por_fuente"`
	PorTipo           map[string]int `json:"por_tipo"`
}

func prio(fuente string) int {
	if p, ok := prioridadFuente[fuente]; ok {
		return p
	}
	return 99
}

func encontrarGemelo(p model.Punto, canon []model.Punto, porClave map[string]int) (int, bool) {
	// Nunca fusionamos avisos ciudadanos ("necesito X", "ofrezco Y"): son
	// mensajes de personas distintas, no lugares. Dos vecinos del mismo barrio
	// pidiendo agua son dos pedidos, no uno.
	if p.Tipo == model.TipoNecesidad || p.Tipo == model.TipoOferta {
		return 0, false
	}

	if k := claveExacta(p); k != "" {
		if idx, ok := porClave[k]; ok && compatible(p, canon[idx]) {
			return idx, true
		}
	}

	// Caso difícil: mismo lugar, escrito distinto. Exigimos cercanía física
	// (< 120 m) Y nombres parecidos. Solo aplica si ambos tienen coordenadas.
	if p.Lat == nil || p.Lon == nil {
		return 0, false
	}
	for i := range canon {
		c := &canon[i]
		if c.Lat == nil || c.Lon == nil || !compatible(p, *c) {
			continue
		}
		if metros(*p.Lat, *p.Lon, *c.Lat, *c.Lon) > 120 {
			continue
		}
		if parecidos(p.Nombre, c.Nombre) {
			return i, true
		}
	}
	return 0, false
}

// compatible evita fusionar cosas que solo comparten dirección: un albergue y
// un banco de sangre en la misma cuadra siguen siendo dos puntos.
func compatible(a, b model.Punto) bool {
	if a.Fuente == b.Fuente {
		return false // dentro de una misma fuente confiamos en sus propios ids
	}
	return a.Tipo == b.Tipo
}

var noAlfanum = regexp.MustCompile(`[^a-z0-9 ]+`)

func normalizar(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = quitarTildes(s)
	s = noAlfanum.ReplaceAllString(s, " ")
	return strings.Join(strings.Fields(s), " ")
}

func quitarTildes(s string) string {
	r := strings.NewReplacer(
		"á", "a", "é", "e", "í", "i", "ó", "o", "ú", "u", "ü", "u", "ñ", "n",
		"Á", "a", "É", "e", "Í", "i", "Ó", "o", "Ú", "u", "Ü", "u", "Ñ", "n",
	)
	return r.Replace(s)
}

func claveExacta(p model.Punto) string {
	n, d := normalizar(p.Nombre), normalizar(p.Direccion)
	if n == "" || d == "" {
		return "" // sin ambos, no arriesgamos una fusión por clave exacta
	}
	return p.Tipo + "|" + n + "|" + d
}

// parecidos usa solapamiento de palabras significativas: barato, explicable y
// suficientemente estricto para nombres de lugares.
func parecidos(a, b string) bool {
	pa, pb := palabras(a), palabras(b)
	if len(pa) == 0 || len(pb) == 0 {
		return false
	}
	comunes := 0
	for w := range pa {
		if pb[w] {
			comunes++
		}
	}
	menor := len(pa)
	if len(pb) < menor {
		menor = len(pb)
	}
	return float64(comunes)/float64(menor) >= 0.6
}

// Palabras que aparecen en medio Cali y no distinguen nada.
var vacias = map[string]bool{
	"de": true, "del": true, "la": true, "el": true, "los": true, "las": true,
	"y": true, "en": true, "punto": true, "centro": true, "cali": true,
	"barrio": true, "calle": true, "carrera": true, "cra": true, "cll": true,
}

func palabras(s string) map[string]bool {
	out := map[string]bool{}
	for _, w := range strings.Fields(normalizar(s)) {
		if len(w) < 3 || vacias[w] {
			continue
		}
		out[w] = true
	}
	return out
}

// metros da la distancia haversine entre dos coordenadas.
func metros(lat1, lon1, lat2, lon2 float64) float64 {
	const R = 6371000.0
	rad := math.Pi / 180
	dLat := (lat2 - lat1) * rad
	dLon := (lon2 - lon1) * rad
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1*rad)*math.Cos(lat2*rad)*math.Sin(dLon/2)*math.Sin(dLon/2)
	return 2 * R * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
}

// absorber mete en el canónico lo que el duplicado tenga y a él le falte.
// El canónico gana en los conflictos; el duplicado solo rellena vacíos.
func absorber(canon *model.Punto, otro model.Punto) {
	canon.TambienEn = append(canon.TambienEn, model.Referencia{
		Fuente: otro.Fuente, ID: otro.ID, URL: otro.FuenteURL,
	})

	rellenar(&canon.Descripcion, otro.Descripcion)
	rellenar(&canon.Direccion, otro.Direccion)
	rellenar(&canon.Barrio, otro.Barrio)
	rellenar(&canon.Comuna, otro.Comuna)
	rellenar(&canon.Municipio, otro.Municipio)
	rellenar(&canon.Departamento, otro.Departamento)
	rellenar(&canon.Contacto, otro.Contacto)
	rellenar(&canon.Organizacion, otro.Organizacion)
	rellenar(&canon.Horario, otro.Horario)

	if canon.Lat == nil && otro.Lat != nil {
		canon.Lat, canon.Lon = otro.Lat, otro.Lon
	}
	canon.Necesita = unir(canon.Necesita, otro.Necesita)
	canon.Ofrece = unir(canon.Ofrece, otro.Ofrece)
	canon.PoliticaAlerta = unir(canon.PoliticaAlerta, otro.PoliticaAlerta)
	canon.Avisos = unir(canon.Avisos, otro.Avisos)

	canon.Confirmaciones += otro.Confirmaciones
	canon.Desmentidos += otro.Desmentidos
	if otro.VoluntariosFaltan > canon.VoluntariosFaltan {
		canon.VoluntariosFaltan = otro.VoluntariosFaltan
	}

	// El estado más urgente manda: preferimos alarmar de más que de menos.
	if urgencia(otro.Estado) > urgencia(canon.Estado) {
		canon.Estado = otro.Estado
	}
	// La actualización más reciente manda.
	if otro.Actualizado != nil && (canon.Actualizado == nil || otro.Actualizado.After(*canon.Actualizado)) {
		canon.Actualizado = otro.Actualizado
	}
}

func urgencia(estado string) int {
	switch estado {
	case model.EstadoUrgente:
		return 4
	case model.EstadoActivo:
		return 3
	case model.EstadoPorConfirmar:
		return 2
	case model.EstadoCubierto:
		return 1
	default:
		return 0
	}
}

func rellenar(destino *string, valor string) {
	if *destino == "" {
		*destino = valor
	}
}

func unir(a, b []string) []string {
	visto := map[string]bool{}
	out := make([]string, 0, len(a)+len(b))
	for _, s := range append(append([]string{}, a...), b...) {
		k := strings.ToLower(strings.TrimSpace(s))
		if k == "" || visto[k] {
			continue
		}
		visto[k] = true
		out = append(out, s)
	}
	sort.Strings(out)
	return out
}
