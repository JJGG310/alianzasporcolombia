package sources

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/JJGG310/alianzasporcolombia/scraper/internal/fetch"
	"github.com/JJGG310/alianzasporcolombia/scraper/internal/model"
)

// CaliSolidario lee el tablero ciudadano de necesidades y ofertas.
//
// Es una app Next.js (App Router) sin API pública: los datos llegan renderizados
// en el servidor dentro del stream RSC, partido en trozos self.__next_f.push.
// Se reconstruye ese texto y se saca el arreglo "posts" que ya viene en JSON.
type CaliSolidario struct{}

const (
	csSlug         = "calisolidario"
	csSitio        = "https://calisolidario.triadaaliados.com/"
	csURLNecesidad = csSitio + "necesidades"
	csURLOferta    = csSitio + "ofertas"
	csAviso        = csSitio + "aviso/"
	csMecanismo    = "payload RSC de Next.js embebido en el HTML de /necesidades y /ofertas"
	csMarcaPush    = "self.__next_f.push("
	csMarcaPosts   = `"posts":[`
	csMarcaEscapa  = `\"posts\":[`
)

func (s *CaliSolidario) Slug() string      { return csSlug }
func (s *CaliSolidario) URL() string       { return csSitio }
func (s *CaliSolidario) Mecanismo() string { return csMecanismo }

type csPost struct {
	ID           string `json:"id"`
	Kind         string `json:"kind"`
	Category     string `json:"category"`
	Title        string `json:"title"`
	Description  string `json:"description"`
	QuantityText string `json:"quantity_text"`
	Comuna       string `json:"comuna"`
	Barrio       string `json:"barrio"`
	Address      string `json:"address"`
	Status       string `json:"status"`
	WarningCount int    `json:"warning_count"`
	CreatedAt    string `json:"created_at"`
}

var csCategorias = map[string]string{
	"mano_de_obra":      "Mano de obra",
	"alimentos":         "Alimentos",
	"agua":              "Agua",
	"salud":             "Salud",
	"albergue":          "Albergue",
	"transporte":        "Transporte",
	"herramientas":      "Herramientas",
	"aseo":              "Aseo e higiene",
	"cobijas_colchones": "Cobijas y colchones",
	"ropa":              "Ropa",
	"medicamentos":      "Medicamentos",
	"panales":           "Pañales",
	"mascotas":          "Mascotas",
	"dinero":            "Donaciones en efectivo",
	"otro":              "Otro",
}

var csEstados = map[string]string{
	"open":      model.EstadoActivo,
	"closed":    model.EstadoCerrado,
	"fulfilled": model.EstadoCubierto,
}

// Municipios del Valle que, nombrados en barrio/dirección, no dejan duda.
var csMunicipios = map[string]string{
	"jamundi":      "Jamundí",
	"jamundí":      "Jamundí",
	"roldanillo":   "Roldanillo",
	"yumbo":        "Yumbo",
	"palmira":      "Palmira",
	"buenaventura": "Buenaventura",
	"candelaria":   "Candelaria",
	"florida":      "Florida",
	"pradera":      "Pradera",
	"dagua":        "Dagua",
	"tulua":        "Tuluá",
	"tuluá":        "Tuluá",
	"buga":         "Buga",
	"cartago":      "Cartago",
	"zarzal":       "Zarzal",
	"vijes":        "Vijes",
	"yotoco":       "Yotoco",
	"restrepo":     "Restrepo",
	"la cumbre":    "La Cumbre",
}

// Barrios inconfundiblemente caleños: si el aviso no trae comuna, sirven para
// ubicarlo. Lista corta a propósito; ante duda el municipio queda vacío.
var csBarriosCali = []string{
	"valle del lili", "pampalinda", "el caney", "ciudad pacifica", "ciudad pacífica",
	"melendez", "meléndez", "limonar", "siloe", "siloé", "aguablanca", "desepaz",
	"chipichape", "ciudad jardin", "ciudad jardín", "el ingenio", "juanchito",
}

func (s *CaliSolidario) Obtener(ctx context.Context, c *fetch.Cliente) (*model.Resultado, error) {
	inicio := time.Now()
	res := &model.Resultado{
		Fuente:    s.Slug(),
		URL:       s.URL(),
		Mecanismo: s.Mecanismo(),
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

	corte := time.Now().UTC()
	var (
		puntos       []model.Punto
		notas        []string
		porKind      = map[string]int{}
		porCategoria = map[string]int{}
		porPagina    = map[string]int{}
		vistos       = map[string]bool{}
	)

	for _, pagina := range []string{csURLNecesidad, csURLOferta} {
		html, err := c.Get(ctx, pagina)
		if err != nil {
			notas = append(notas, fmt.Sprintf("%s no se pudo descargar: %v", pagina, err))
			continue
		}
		crudos, err := csExtraerPosts(string(html))
		if err != nil {
			notas = append(notas, fmt.Sprintf("%s: %v", pagina, err))
			continue
		}
		for _, crudo := range crudos {
			var p csPost
			if err := json.Unmarshal(crudo, &p); err != nil || p.ID == "" {
				continue
			}
			if vistos[p.ID] { // el mismo aviso puede salir en las dos vistas
				continue
			}
			vistos[p.ID] = true
			porKind[p.Kind]++
			porCategoria[p.Category]++
			porPagina[pagina]++
			puntos = append(puntos, csAModelo(p, crudo, corte))
		}
	}

	if len(puntos) == 0 {
		motivo := strings.Join(notas, "; ")
		if motivo == "" {
			motivo = "no apareció el arreglo posts en el payload RSC"
		}
		return fallar(fmt.Errorf("calisolidario sin avisos: el formato RSC de Next.js cambió o el sitio no respondió (%s)", motivo))
	}

	res.Puntos = puntos
	res.OK = true
	res.Nota = strings.Join(notas, "; ")
	res.Extra = map[string]any{
		"por_kind":      porKind,
		"por_categoria": porCategoria,
		"por_pagina":    porPagina,
	}
	terminar()
	return res, nil
}

func csAModelo(p csPost, crudo json.RawMessage, corte time.Time) model.Punto {
	pt := model.Punto{
		ID:          model.HacerID(csSlug, p.ID),
		Fuente:      csSlug,
		FuenteURL:   csAviso + p.ID,
		Nombre:      p.Title,
		Descripcion: p.Description,
		Comuna:      p.Comuna,
		Barrio:      p.Barrio,
		Direccion:   p.Address,
		Creado:      csFecha(p.CreatedAt),
		Raw:         crudo,
	}

	detalle := csDetalle(p)
	if strings.EqualFold(p.Kind, "offer") {
		pt.Tipo = model.TipoOferta
		pt.Ofrece = detalle
	} else {
		pt.Tipo = model.TipoNecesidad
		pt.Necesita = detalle
	}
	if pt.Nombre == "" { // el título es lo que se ve en la tarjeta: nunca vacío
		pt.Nombre = csRespaldoNombre(p)
	}

	if e, ok := csEstados[strings.ToLower(strings.TrimSpace(p.Status))]; ok {
		pt.Estado = e
	} else {
		pt.Estado = model.EstadoPorConfirmar
	}

	// Un aviso reportado por la comunidad no se borra, se degrada a por-confirmar
	// y se deja la alerta visible para quien modere.
	if p.WarningCount > 0 {
		pt.Estado = model.EstadoPorConfirmar
		pt.PoliticaAlerta = append(pt.PoliticaAlerta,
			fmt.Sprintf("reportado %d veces por la comunidad en el sitio origen", p.WarningCount))
	}

	if muni := csMunicipio(p); muni != "" {
		pt.Municipio = muni
		pt.Departamento = "Valle del Cauca"
	}

	// La fuente no publica coordenadas: se deja sin geolocalizar antes que
	// geocodificar a ciegas.
	model.Normalizar(&pt, corte)
	return pt
}

func csDetalle(p csPost) []string {
	var out []string
	if q := strings.TrimSpace(p.QuantityText); q != "" && q != "." {
		out = append(out, q)
	}
	if cat := csCategoriaLegible(p.Category); cat != "" {
		out = append(out, cat)
	}
	return out
}

func csCategoriaLegible(cat string) string {
	cat = strings.ToLower(strings.TrimSpace(cat))
	if cat == "" {
		return ""
	}
	if v, ok := csCategorias[cat]; ok {
		return v
	}
	return ahfCapitalizar(cat)
}

func csRespaldoNombre(p csPost) string {
	que := csCategoriaLegible(p.Category)
	if que == "" {
		que = "Ayuda"
	}
	verbo := "Necesita"
	if strings.EqualFold(p.Kind, "offer") {
		verbo = "Ofrece"
	}
	if b := strings.TrimSpace(p.Barrio); b != "" {
		return verbo + " " + strings.ToLower(que) + " — " + b
	}
	return verbo + " " + strings.ToLower(que)
}

// csMunicipio prefiere lo que diga el texto: un aviso de Roldanillo no es de
// Cali aunque quien lo escribió haya llenado el campo comuna.
func csMunicipio(p csPost) string {
	texto := strings.ToLower(p.Barrio + " " + p.Address + " " + p.Title)
	for clave, nombre := range csMunicipios {
		if csMenciona(texto, clave) {
			return nombre
		}
	}
	if strings.TrimSpace(p.Comuna) != "" {
		return "Cali"
	}
	if csMenciona(texto, "cali") {
		return "Cali"
	}
	for _, b := range csBarriosCali {
		if strings.Contains(texto, b) {
			return "Cali"
		}
	}
	return ""
}

// csMenciona busca la palabra completa: "cali" no debe pegar dentro de "calima".
func csMenciona(texto, palabra string) bool {
	desde := 0
	for {
		i := strings.Index(texto[desde:], palabra)
		if i < 0 {
			return false
		}
		i += desde
		fin := i + len(palabra)
		rAntes, _ := utf8.DecodeLastRuneInString(texto[:i])
		rDespues, _ := utf8.DecodeRuneInString(texto[fin:])
		antes := i == 0 || !csAlfanumerico(rAntes)
		despues := fin >= len(texto) || !csAlfanumerico(rDespues)
		if antes && despues {
			return true
		}
		desde = i + 1
	}
}

func csAlfanumerico(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r)
}

func csFecha(s string) *time.Time {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	for _, f := range []string{time.RFC3339Nano, time.RFC3339, "2006-01-02 15:04:05.999999-07:00"} {
		if t, err := time.Parse(f, s); err == nil {
			u := t.UTC()
			return &u
		}
	}
	return nil
}

// ---------- extracción del payload RSC ----------

// csExtraerPosts saca el arreglo "posts" del HTML. Plan A: reconstruir el stream
// RSC desde los trozos self.__next_f.push y buscarlo ya des-escapado. Plan B: si
// el stream no se pudo rearmar, buscar el fragmento escapado en el HTML crudo.
func csExtraerPosts(html string) ([]json.RawMessage, error) {
	if texto := csReconstruirRSC(html); texto != "" {
		if crudo, err := csPostsEn(texto); err == nil {
			var posts []json.RawMessage
			if err := json.Unmarshal(crudo, &posts); err == nil {
				return posts, nil
			}
		}
	}
	crudo, err := csPostsEscapados(html)
	if err != nil {
		return nil, err
	}
	var posts []json.RawMessage
	if err := json.Unmarshal(crudo, &posts); err != nil {
		return nil, fmt.Errorf("arreglo posts ilegible: %w", err)
	}
	return posts, nil
}

// csReconstruirRSC concatena, en orden, el contenido de cada
// self.__next_f.push([n,"…"]) tras des-escapar su literal JSON.
func csReconstruirRSC(html string) string {
	var b strings.Builder
	i := 0
	for {
		k := strings.Index(html[i:], csMarcaPush)
		if k < 0 {
			break
		}
		p := i + k + len(csMarcaPush)
		i = p
		if p >= len(html) || html[p] != '[' {
			continue
		}
		p++
		// El primer elemento es el índice del trozo; el segundo, el texto.
		tope := min(p+16, len(html))
		coma := strings.IndexByte(html[p:tope], ',')
		if coma < 0 {
			continue
		}
		p += coma + 1
		for p < len(html) && (html[p] == ' ' || html[p] == '\n' || html[p] == '\t') {
			p++
		}
		if p >= len(html) || html[p] != '"' { // p. ej. push([2,null])
			continue
		}
		fin, ok := csFinCadena(html, p)
		if !ok {
			continue
		}
		var trozo string
		if err := json.Unmarshal([]byte(html[p:fin]), &trozo); err == nil {
			b.WriteString(trozo)
		}
		i = fin
	}
	return b.String()
}

// csFinCadena devuelve el índice justo después de la comilla que cierra el
// literal JSON que empieza en ini.
func csFinCadena(s string, ini int) (int, bool) {
	for i := ini + 1; i < len(s); i++ {
		switch s[i] {
		case '\\':
			i++
		case '"':
			return i + 1, true
		}
	}
	return 0, false
}

func csPostsEn(texto string) (json.RawMessage, error) {
	k := strings.Index(texto, csMarcaPosts)
	if k < 0 {
		return nil, fmt.Errorf("no aparece %s en el payload RSC", csMarcaPosts)
	}
	frag, err := csArreglo(texto, k+len(csMarcaPosts)-1)
	if err != nil {
		return nil, err
	}
	return json.RawMessage(frag), nil
}

// csPostsEscapados es el plan B: el arreglo sigue escapado dentro del HTML, así
// que se recorre en ese estado y luego se des-escapa como literal JSON.
func csPostsEscapados(html string) (json.RawMessage, error) {
	k := strings.Index(html, csMarcaEscapa)
	if k < 0 {
		return nil, fmt.Errorf("no aparece el arreglo posts (ni des-escapado ni escapado) en el HTML")
	}
	frag, err := csArregloEscapado(html, k+len(csMarcaEscapa)-1)
	if err != nil {
		return nil, err
	}
	var plano string
	if err := json.Unmarshal([]byte(`"`+frag+`"`), &plano); err != nil {
		return nil, fmt.Errorf("no se pudo des-escapar el arreglo posts: %w", err)
	}
	return json.RawMessage(plano), nil
}

// csArreglo recorre desde el '[' contando llaves y corchetes, ignorando los que
// van dentro de cadenas. Un regex no sirve aquí: los textos traen corchetes.
func csArreglo(s string, ini int) (string, error) {
	if ini < 0 || ini >= len(s) || s[ini] != '[' {
		return "", fmt.Errorf("no hay '[' donde empieza el arreglo posts")
	}
	prof := 0
	enCadena := false
	for i := ini; i < len(s); i++ {
		c := s[i]
		if enCadena {
			switch c {
			case '\\':
				i++
			case '"':
				enCadena = false
			}
			continue
		}
		switch c {
		case '"':
			enCadena = true
		case '[', '{':
			prof++
		case ']', '}':
			prof--
			if prof == 0 {
				return s[ini : i+1], nil
			}
		}
	}
	return "", fmt.Errorf("el arreglo posts quedó sin cerrar")
}

// csArregloEscapado hace lo mismo sobre texto todavía escapado, donde las
// comillas de las cadenas son \" y las barras literales son \\.
func csArregloEscapado(s string, ini int) (string, error) {
	if ini < 0 || ini >= len(s) || s[ini] != '[' {
		return "", fmt.Errorf("no hay '[' donde empieza el arreglo posts escapado")
	}
	prof := 0
	enCadena := false
	for i := ini; i < len(s); {
		if s[i] == '\\' && i+1 < len(s) {
			if s[i+1] == '"' {
				enCadena = !enCadena
			}
			i += 2
			continue
		}
		if !enCadena {
			switch s[i] {
			case '[', '{':
				prof++
			case ']', '}':
				prof--
				if prof == 0 {
					return s[ini : i+1], nil
				}
			}
		}
		i++
	}
	return "", fmt.Errorf("el arreglo posts escapado quedó sin cerrar")
}
