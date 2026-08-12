package sources

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/JJGG310/alianzasporcolombia/scraper/internal/fetch"
	"github.com/JJGG310/alianzasporcolombia/scraper/internal/model"
)

// PuntosCriticos lee el tablero ciudadano de puntos críticos del sismo de Cali.
//
// El sitio es un único HTML con React UMD + @babel/standalone y los datos
// escritos a mano como arreglos JS dentro de <script type="text/babel">: no hay
// API ni JSON servido. Toca recortar cada literal del HTML y traducirlo a JSON,
// porque son literales de JavaScript (llaves sin comillas, identificadores
// desnudos como HORA_ACT) que json.Unmarshal no acepta.
type PuntosCriticos struct{}

const (
	pcSlug      = "puntos-criticos"
	pcSitio     = "https://terremoto-cali-puntos-criticos.netlify.app/"
	pcMecanismo = "arreglos JS embebidos en el HTML (React vía Babel standalone, sin API)"
)

func (p *PuntosCriticos) Slug() string      { return pcSlug }
func (p *PuntosCriticos) URL() string       { return pcSitio }
func (p *PuntosCriticos) Mecanismo() string { return pcMecanismo }

func (p *PuntosCriticos) Obtener(ctx context.Context, c *fetch.Cliente) (*model.Resultado, error) {
	inicio := time.Now()
	res := &model.Resultado{
		Fuente:    pcSlug,
		URL:       pcSitio,
		Mecanismo: pcMecanismo,
	}
	terminar := func() {
		res.Duracion = time.Since(inicio).String()
		res.Obtenido = time.Now().UTC().Format(time.RFC3339)
		res.Total = len(res.Puntos)
	}

	crudo, err := c.Get(ctx, pcSitio)
	if err != nil {
		res.OK = false
		res.Error = fmt.Sprintf("descargando el HTML: %v", err)
		terminar()
		return res, fmt.Errorf("puntos-criticos: %w", err)
	}

	doc := string(crudo)
	// Las constantes de texto (HORA_ACT, ALERTA…) se resuelven primero porque los
	// arreglos las usan como identificadores desnudos.
	consts := pcConstantesTexto(doc)
	corte := time.Now().UTC()

	var notas []string
	acum := pcNuevoAcum()
	conteos := map[string]int{}

	// PUNTOS y ALTA_PRIORIDAD son el mismo universo (puntos donde se rescata y se
	// recibe apoyo), por eso comparten tipo e id: si un punto aparece en los dos
	// arreglos se fusiona en vez de duplicarse.
	if items, err := pcObjetos(doc, "PUNTOS", consts); err != nil {
		notas = append(notas, "PUNTOS: "+err.Error())
	} else {
		conteos["PUNTOS"] = len(items)
		for _, it := range items {
			if pt, ok := pcPuntoRescate(it, corte, false); ok {
				acum.agregar(pt, corte)
			}
		}
	}

	if items, err := pcObjetos(doc, "ALTA_PRIORIDAD", consts); err != nil {
		notas = append(notas, "ALTA_PRIORIDAD: "+err.Error())
	} else {
		conteos["ALTA_PRIORIDAD"] = len(items)
		for _, it := range items {
			if pt, ok := pcPuntoRescate(it, corte, true); ok {
				acum.agregar(pt, corte)
			}
		}
	}

	if items, err := pcObjetos(doc, "ALBERGUES", consts); err != nil {
		notas = append(notas, "ALBERGUES: "+err.Error())
	} else {
		conteos["ALBERGUES"] = len(items)
		contacto := consts["ALBERGUES_CONTACTO"]
		for _, it := range items {
			if pt, ok := pcPuntoAlbergue(it, contacto, corte); ok {
				acum.agregar(pt, corte)
			}
		}
	}

	if nombres, err := pcCadenas(doc, "CLINICAS", consts); err != nil {
		notas = append(notas, "CLINICAS: "+err.Error())
	} else {
		conteos["CLINICAS"] = len(nombres)
		for _, n := range nombres {
			if pt, ok := pcPuntoClinica(n, corte); ok {
				acum.agregar(pt, corte)
			}
		}
	}

	// AYUDAS es "ayuda disponible pendiente por asignar": no son sitios que
	// reciban insumos sino gente y empresas que OFRECEN recursos (maquinaria,
	// hielo, ollas comunitarias). Eso es donación de un tercero, no acopio.
	if items, err := pcObjetos(doc, "AYUDAS", consts); err != nil {
		notas = append(notas, "AYUDAS: "+err.Error())
	} else {
		conteos["AYUDAS"] = len(items)
		for _, it := range items {
			if pt, ok := pcPuntoAyuda(it, corte); ok {
				acum.agregar(pt, corte)
			}
		}
	}

	extra := map[string]any{}
	if v := consts["ACTUALIZADO"]; v != "" {
		extra["actualizado"] = v
	}
	if v := consts["ALERTA"]; v != "" {
		extra["alerta"] = v
	}
	if v := consts["NOTA_NO_DESPLAZARSE"]; v != "" {
		extra["nota_no_desplazarse"] = v
	}
	if v := consts["ALBERGUES_CONTACTO"]; v != "" {
		extra["albergues_contacto"] = v
	}

	// Estas tres listas no son puntos, pero son lo más valioso del sitio: qué
	// hace falta en general, dónde NO ir, y la cronología de reportes.
	if lista, err := pcCadenas(doc, "NECESIDADES_GENERALES", consts); err != nil {
		notas = append(notas, "NECESIDADES_GENERALES: "+err.Error())
	} else {
		conteos["NECESIDADES_GENERALES"] = len(lista)
		extra["necesidades_generales"] = lista
	}

	if items, err := pcObjetos(doc, "NO_DESPLAZARSE", consts); err != nil {
		notas = append(notas, "NO_DESPLAZARSE: "+err.Error())
	} else {
		conteos["NO_DESPLAZARSE"] = len(items)
		avisos := make([]map[string]string, 0, len(items))
		for _, it := range items {
			avisos = append(avisos, map[string]string{
				"sector": pcTexto(it.m, "sector"),
				"nota":   pcTexto(it.m, "nota"),
			})
		}
		extra["no_desplazarse"] = avisos
	}

	if items, err := pcObjetos(doc, "ACTUALIZACIONES", consts); err != nil {
		notas = append(notas, "ACTUALIZACIONES: "+err.Error())
	} else {
		conteos["ACTUALIZACIONES"] = len(items)
		crono := make([]map[string]string, 0, len(items))
		for _, it := range items {
			crono = append(crono, map[string]string{
				"fecha":    pcTexto(it.m, "fecha"),
				"punto":    pcTexto(it.m, "punto"),
				"contacto": pcTexto(it.m, "contacto"),
				"texto":    pcTexto(it.m, "texto"),
			})
		}
		extra["actualizaciones"] = crono
	}

	if len(conteos) == 0 {
		res.OK = false
		res.Error = "no se pudo extraer ningún arreglo JS del HTML: el sitio cambió de forma (se esperaban PUNTOS, ALBERGUES, CLINICAS, AYUDAS, ALTA_PRIORIDAD, NECESIDADES_GENERALES, NO_DESPLAZARSE, ACTUALIZACIONES)"
		terminar()
		return res, fmt.Errorf("puntos-criticos: %s", res.Error)
	}

	extra["conteo_por_arreglo"] = conteos
	res.Extra = extra
	res.Puntos = acum.puntos
	res.OK = true
	if acum.fusionados > 0 {
		notas = append(notas, fmt.Sprintf("%d puntos repetidos entre arreglos fusionados por nombre", acum.fusionados))
	}
	res.Nota = strings.Join(notas, "; ")
	terminar()
	return res, nil
}

// ---------------------------------------------------------------------------
// Acumulador con deduplicación por id
// ---------------------------------------------------------------------------

type pcAcum struct {
	puntos     []model.Punto
	indice     map[string]int
	fusionados int
}

func pcNuevoAcum() *pcAcum { return &pcAcum{indice: map[string]int{}} }

// agregar suma un punto o, si ya existía ese id (mismo lugar reportado en dos
// arreglos), fusiona ambos registros conservando lo más completo de cada uno.
func (a *pcAcum) agregar(p model.Punto, corte time.Time) {
	if i, ok := a.indice[p.ID]; ok {
		pcFusionar(&a.puntos[i], p)
		model.Normalizar(&a.puntos[i], corte)
		a.fusionados++
		return
	}
	a.indice[p.ID] = len(a.puntos)
	a.puntos = append(a.puntos, p)
}

// pcFusionar completa los huecos de dst con lo que traiga src y se queda con el
// estado más grave de los dos.
func pcFusionar(dst *model.Punto, src model.Punto) {
	for _, par := range []struct{ d, s *string }{
		{&dst.Direccion, &src.Direccion},
		{&dst.Contacto, &src.Contacto},
		{&dst.Organizacion, &src.Organizacion},
		{&dst.Horario, &src.Horario},
		{&dst.Municipio, &src.Municipio},
		{&dst.Departamento, &src.Departamento},
	} {
		if *par.d == "" {
			*par.d = *par.s
		}
	}
	if src.Descripcion != "" && !strings.Contains(dst.Descripcion, src.Descripcion) {
		dst.Descripcion = strings.TrimSpace(dst.Descripcion + " · " + src.Descripcion)
		dst.Descripcion = strings.TrimPrefix(dst.Descripcion, "· ")
	}
	dst.Necesita = append(dst.Necesita, src.Necesita...)
	dst.Ofrece = append(dst.Ofrece, src.Ofrece...)
	if src.Estado == model.EstadoUrgente {
		dst.Estado = model.EstadoUrgente
	}
	if src.Prioridad == model.PrioridadCritica {
		dst.Prioridad = model.PrioridadCritica
	}
	// Raw se queda con el registro más largo, que es el más informativo.
	if len(src.Raw) > len(dst.Raw) {
		dst.Raw = src.Raw
	}
}

// ---------------------------------------------------------------------------
// Mapeo de cada arreglo a model.Punto
// ---------------------------------------------------------------------------

// pcPuntoRescate sirve tanto para PUNTOS como para ALTA_PRIORIDAD: los dos
// describen sitios de rescate/apoyo, solo cambia de dónde sale la urgencia.
func pcPuntoRescate(it pcItem, corte time.Time, urgentePorArreglo bool) (model.Punto, bool) {
	nombre := pcTexto(it.m, "lugar", "nombre", "punto")
	if nombre == "" {
		return model.Punto{}, false
	}
	direccion := pcTexto(it.m, "direccion")
	p := model.Punto{
		ID:          model.HacerID(pcSlug, nombre),
		Fuente:      pcSlug,
		FuenteURL:   pcSitio,
		Tipo:        model.TipoRescate,
		Nombre:      nombre,
		Descripcion: pcTexto(it.m, "nota"),
		Direccion:   direccion,
		Necesita:    pcLista(it.m, "necesidades"),
		Contacto:    pcContacto(it.m),
		Estado:      model.EstadoActivo,
		Raw:         it.crudo,
	}
	// `hora` es la hora del último reporte del punto, no un horario de atención:
	// va en la descripción como señal de frescura, no en Horario.
	if h := pcTexto(it.m, "hora"); h != "" {
		if p.Descripcion == "" {
			p.Descripcion = "Reporte de las " + h
		} else {
			p.Descripcion += " (reporte de las " + h + ")"
		}
	}
	if urgentePorArreglo || pcBool(it.m, "prioridad") {
		p.Estado = model.EstadoUrgente
		p.Prioridad = model.PrioridadCritica
	}
	p.Municipio, p.Departamento = pcUbicacion(nombre + " " + direccion + " " + p.Descripcion)
	model.Normalizar(&p, corte)
	return p, true
}

func pcPuntoAlbergue(it pcItem, contacto string, corte time.Time) (model.Punto, bool) {
	nombre := pcTexto(it.m, "nombre", "lugar")
	if nombre == "" {
		return model.Punto{}, false
	}
	direccion := pcTexto(it.m, "direccion")
	p := model.Punto{
		ID:        model.HacerID(pcSlug, nombre),
		Fuente:    pcSlug,
		FuenteURL: pcSitio,
		Tipo:      model.TipoAlbergue,
		Nombre:    nombre,
		Direccion: direccion,
		// La fuente publica un solo contacto para toda la red de albergues.
		Contacto: contacto,
		Estado:   model.EstadoActivo,
		Raw:      it.crudo,
	}
	p.Municipio, p.Departamento = pcUbicacion(nombre + " " + direccion)
	model.Normalizar(&p, corte)
	return p, true
}

func pcPuntoClinica(nombre string, corte time.Time) (model.Punto, bool) {
	nombre = strings.TrimSpace(nombre)
	if nombre == "" {
		return model.Punto{}, false
	}
	crudo, _ := json.Marshal(map[string]string{"clinica": nombre})
	p := model.Punto{
		ID:          model.HacerID(pcSlug, nombre),
		Fuente:      pcSlug,
		FuenteURL:   pcSitio,
		Tipo:        model.TipoSalud,
		Nombre:      nombre,
		Descripcion: "Reportada como afectada y en necesidad de evacuación",
		// Evacuar un hospital es urgente por definición, aunque el arreglo sea
		// solo una lista de nombres sin campo de estado.
		Estado:    model.EstadoUrgente,
		Prioridad: model.PrioridadAlta,
		Raw:       crudo,
	}
	p.Municipio, p.Departamento = pcUbicacion(nombre)
	model.Normalizar(&p, corte)
	return p, true
}

func pcPuntoAyuda(it pcItem, corte time.Time) (model.Punto, bool) {
	recurso := pcTexto(it.m, "recurso")
	if recurso == "" {
		return model.Punto{}, false
	}
	destino := pcTexto(it.m, "destino")
	var desc []string
	if destino != "" {
		desc = append(desc, "Destino: "+destino)
	}
	if obs := pcTexto(it.m, "obs"); obs != "" {
		desc = append(desc, obs)
	}
	p := model.Punto{
		ID:          model.HacerID(pcSlug, "ayuda", recurso, destino),
		Fuente:      pcSlug,
		FuenteURL:   pcSitio,
		Tipo:        model.TipoDonacionOrg,
		Nombre:      recurso,
		Descripcion: strings.Join(desc, " · "),
		Ofrece:      []string{recurso},
		Contacto:    pcContacto(it.m),
		Estado:      model.EstadoActivo,
		Raw:         it.crudo,
	}
	p.Municipio, p.Departamento = pcUbicacion(recurso + " " + destino + " " + p.Descripcion)
	model.Normalizar(&p, corte)
	return p, true
}

// pcContacto junta el nombre de quien coordina con su teléfono, que la fuente
// guarda en campos separados.
func pcContacto(m map[string]any) string {
	coord := pcTexto(m, "coordinador")
	tel := pcTexto(m, "contacto")
	switch {
	case coord != "" && tel != "":
		return coord + " — " + tel
	case coord != "":
		return coord
	default:
		return tel
	}
}

// pcUbicacion: el sitio es explícitamente de Cali, así que ese es el valor por
// defecto; solo se cambia si el texto nombra otro municipio.
func pcUbicacion(texto string) (municipio, departamento string) {
	for _, l := range pcLugares {
		if l.re.MatchString(texto) {
			return l.municipio, l.departamento
		}
	}
	return "Cali", "Valle del Cauca"
}

type pcLugar struct {
	re                      *regexp.Regexp
	municipio, departamento string
}

var pcLugares = pcCompilarLugares([][3]string{
	{"jamund[íi]", "Jamundí", "Valle del Cauca"},
	{"yumbo", "Yumbo", "Valle del Cauca"},
	{"palmira", "Palmira", "Valle del Cauca"},
	{"buenaventura", "Buenaventura", "Valle del Cauca"},
	{"dagua", "Dagua", "Valle del Cauca"},
	{"candelaria", "Candelaria", "Valle del Cauca"},
	{"pradera", "Pradera", "Valle del Cauca"},
	{"tulu[áa]", "Tuluá", "Valle del Cauca"},
	{"buga", "Guadalajara de Buga", "Valle del Cauca"},
	{"cartago", "Cartago", "Valle del Cauca"},
	{"el cairo", "El Cairo", "Valle del Cauca"},
	{"restrepo", "Restrepo", "Valle del Cauca"},
	{"la cumbre", "La Cumbre", "Valle del Cauca"},
	{"el cerrito", "El Cerrito", "Valle del Cauca"},
	{"zarzal", "Zarzal", "Valle del Cauca"},
	{"roldanillo", "Roldanillo", "Valle del Cauca"},
	{"quibd[óo]", "Quibdó", "Chocó"},
	{"istmina", "Istmina", "Chocó"},
	{"condoto", "Condoto", "Chocó"},
	{"tad[óo]", "Tadó", "Chocó"},
})

// pcCompilarLugares exige frontera de letra a cada lado para no confundir
// "Buga" dentro de "Residencias Bugatier".
func pcCompilarLugares(defs [][3]string) []pcLugar {
	out := make([]pcLugar, 0, len(defs))
	for _, d := range defs {
		out = append(out, pcLugar{
			re:           regexp.MustCompile(`(?i)(^|[^\p{L}])` + d[0] + `([^\p{L}]|$)`),
			municipio:    d[1],
			departamento: d[2],
		})
	}
	return out
}

// ---------------------------------------------------------------------------
// Lectura tipada de los mapas que salen del parser
// ---------------------------------------------------------------------------

// pcItem es un elemento del arreglo: el mapa ya decodificado y su JSON crudo,
// que se guarda tal cual en Punto.Raw.
type pcItem struct {
	m     map[string]any
	crudo json.RawMessage
}

// pcTexto devuelve el primer campo no vacío de los pedidos.
func pcTexto(m map[string]any, claves ...string) string {
	for _, k := range claves {
		if s, ok := m[k].(string); ok && strings.TrimSpace(s) != "" {
			return strings.TrimSpace(s)
		}
	}
	return ""
}

func pcBool(m map[string]any, clave string) bool {
	b, _ := m[clave].(bool)
	return b
}

func pcLista(m map[string]any, clave string) []string {
	arr, ok := m[clave].([]any)
	if !ok {
		return nil
	}
	out := make([]string, 0, len(arr))
	for _, v := range arr {
		if s, ok := v.(string); ok && strings.TrimSpace(s) != "" {
			out = append(out, strings.TrimSpace(s))
		}
	}
	return out
}

// ---------------------------------------------------------------------------
// Mini-parser de literales JS (sin dependencias: los datos no son JSON válido)
// ---------------------------------------------------------------------------

// pcObjetos extrae `const NOMBRE = [ {...}, ... ]` y lo devuelve como mapas.
func pcObjetos(html, nombre string, consts map[string]string) ([]pcItem, error) {
	elems, err := pcElementos(html, nombre, consts)
	if err != nil {
		return nil, err
	}
	out := make([]pcItem, 0, len(elems))
	for _, e := range elems {
		var m map[string]any
		if err := json.Unmarshal(e, &m); err != nil {
			continue // un objeto ilegible no debe tumbar el arreglo entero
		}
		out = append(out, pcItem{m: m, crudo: e})
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("%s no trajo objetos legibles", nombre)
	}
	return out, nil
}

// pcCadenas extrae un arreglo de strings (CLINICAS, NECESIDADES_GENERALES).
func pcCadenas(html, nombre string, consts map[string]string) ([]string, error) {
	elems, err := pcElementos(html, nombre, consts)
	if err != nil {
		return nil, err
	}
	out := make([]string, 0, len(elems))
	for _, e := range elems {
		var s string
		if err := json.Unmarshal(e, &s); err != nil {
			continue
		}
		if s = strings.TrimSpace(s); s != "" {
			out = append(out, s)
		}
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("%s no trajo cadenas legibles", nombre)
	}
	return out, nil
}

func pcElementos(html, nombre string, consts map[string]string) ([]json.RawMessage, error) {
	literal, ok := pcRecortarArreglo(html, nombre)
	if !ok {
		return nil, fmt.Errorf("no se encontró el arreglo %s en el HTML", nombre)
	}
	var elems []json.RawMessage
	if err := json.Unmarshal([]byte(pcJSAJSON(literal, consts)), &elems); err != nil {
		return nil, fmt.Errorf("%s no se pudo convertir a JSON: %w", nombre, err)
	}
	return elems, nil
}

var pcReConst = regexp.MustCompile(`\bconst\s+([A-Za-z_$][A-Za-z0-9_$]*)\s*=\s*`)

// pcRecortarArreglo devuelve el literal `[...]` completo de una constante,
// contando corchetes y llaves y saltándose cadenas y comentarios.
func pcRecortarArreglo(html, nombre string) (string, bool) {
	re := regexp.MustCompile(`\bconst\s+` + regexp.QuoteMeta(nombre) + `\s*=\s*\[`)
	loc := re.FindStringIndex(html)
	if loc == nil {
		return "", false
	}
	inicio := loc[1] - 1 // posición del '['
	fin, ok := pcFinDelLiteral(html, inicio)
	if !ok {
		return "", false
	}
	return html[inicio:fin], true
}

// pcFinDelLiteral devuelve el índice justo después del cierre del literal que
// empieza en i (un '[' o un '{').
func pcFinDelLiteral(s string, i int) (int, bool) {
	profundidad := 0
	for i < len(s) {
		switch c := s[i]; c {
		case '[', '{':
			profundidad++
			i++
		case ']', '}':
			profundidad--
			i++
			if profundidad == 0 {
				return i, true
			}
		case '"', '\'', '`':
			_, fin, ok := pcLeerCadena(s, i)
			if !ok {
				return 0, false
			}
			i = fin
		case '/':
			if salto, ok := pcSaltarComentario(s, i); ok {
				i = salto
			} else {
				i++
			}
		default:
			i++
		}
	}
	return 0, false
}

// pcConstantesTexto resuelve todas las `const NOMBRE = "texto";` del HTML, que
// es lo que permite sustituir identificadores desnudos como HORA_ACT.
func pcConstantesTexto(html string) map[string]string {
	out := map[string]string{}
	for _, m := range pcReConst.FindAllStringSubmatchIndex(html, -1) {
		nombre := html[m[2]:m[3]]
		i := m[1]
		if i >= len(html) {
			continue
		}
		if c := html[i]; c != '"' && c != '\'' && c != '`' {
			continue
		}
		if valor, _, ok := pcLeerCadena(html, i); ok {
			out[nombre] = valor
		}
	}
	return out
}

// pcJSAJSON traduce un literal de objeto/arreglo JS a JSON: entrecomilla las
// llaves desnudas, pasa las cadenas de comilla simple o backtick a comillas
// dobles, borra comentarios y comas colgantes, y reemplaza cada identificador
// suelto por su constante de texto (o por null si no se puede resolver).
func pcJSAJSON(src string, consts map[string]string) string {
	out := make([]byte, 0, len(src)+len(src)/8)
	for i := 0; i < len(src); {
		c := src[i]
		switch {
		case c == '"' || c == '\'' || c == '`':
			valor, fin, ok := pcLeerCadena(src, i)
			if !ok {
				return string(out)
			}
			b, _ := json.Marshal(valor)
			out = append(out, b...)
			i = fin

		case c == '/':
			if salto, ok := pcSaltarComentario(src, i); ok {
				i = salto
			} else {
				out = append(out, c)
				i++
			}

		case c == ']' || c == '}':
			out = pcQuitarComaFinal(out)
			out = append(out, c)
			i++

		case c >= '0' && c <= '9':
			j := i
			for j < len(src) && (src[j] >= '0' && src[j] <= '9' || src[j] == '.') {
				j++
			}
			if j < len(src) && (src[j] == 'e' || src[j] == 'E') { // notación científica
				k := j + 1
				if k < len(src) && (src[k] == '+' || src[k] == '-') {
					k++
				}
				for k < len(src) && src[k] >= '0' && src[k] <= '9' {
					k++
				}
				j = k
			}
			out = append(out, src[i:j]...)
			i = j

		case pcEsInicioIdent(c):
			j := i
			for j < len(src) && pcEsIdent(src[j]) {
				j++
			}
			ident := src[i:j]
			i = j
			if k, ok := pcSiguienteUtil(src, i); ok && src[k] == ':' {
				// Es una llave sin comillas: `lugar:` -> `"lugar":`
				b, _ := json.Marshal(ident)
				out = append(out, b...)
				continue
			}
			out = append(out, pcValorDeIdent(ident, consts)...)
			i = pcSaltarAcceso(src, i)

		default:
			out = append(out, c)
			i++
		}
	}
	return string(out)
}

// pcValorDeIdent decide qué poner donde el JS tenía un identificador desnudo.
func pcValorDeIdent(ident string, consts map[string]string) []byte {
	switch ident {
	case "true", "false", "null":
		return []byte(ident)
	case "undefined", "NaN", "Infinity":
		return []byte("null")
	}
	if v, ok := consts[ident]; ok {
		b, _ := json.Marshal(v)
		return b
	}
	// Sin valor conocido: null es honesto, inventar un texto no lo sería.
	return []byte("null")
}

// pcSaltarAcceso descarta lo que cuelgue de un identificador ya resuelto
// (`.algo`, `(...)`), que de otro modo dejaría JSON inválido.
func pcSaltarAcceso(s string, i int) int {
	for {
		j, ok := pcSiguienteUtil(s, i)
		if !ok {
			return i
		}
		switch s[j] {
		case '.':
			k := j + 1
			for k < len(s) && pcEsIdent(s[k]) {
				k++
			}
			i = k
		case '(':
			fin, ok := pcFinDeParentesis(s, j)
			if !ok {
				return i
			}
			i = fin
		default:
			return i
		}
	}
}

func pcFinDeParentesis(s string, i int) (int, bool) {
	profundidad := 0
	for i < len(s) {
		switch s[i] {
		case '(':
			profundidad++
			i++
		case ')':
			profundidad--
			i++
			if profundidad == 0 {
				return i, true
			}
		case '"', '\'', '`':
			_, fin, ok := pcLeerCadena(s, i)
			if !ok {
				return 0, false
			}
			i = fin
		default:
			i++
		}
	}
	return 0, false
}

// pcSiguienteUtil devuelve el índice del próximo carácter que no sea espacio ni
// comentario.
func pcSiguienteUtil(s string, i int) (int, bool) {
	for i < len(s) {
		switch c := s[i]; {
		case c == ' ' || c == '\t' || c == '\n' || c == '\r':
			i++
		case c == '/':
			salto, ok := pcSaltarComentario(s, i)
			if !ok {
				return i, true
			}
			i = salto
		default:
			return i, true
		}
	}
	return 0, false
}

func pcSaltarComentario(s string, i int) (int, bool) {
	if i+1 >= len(s) {
		return 0, false
	}
	switch s[i+1] {
	case '/':
		j := strings.IndexByte(s[i:], '\n')
		if j < 0 {
			return len(s), true
		}
		return i + j + 1, true
	case '*':
		j := strings.Index(s[i+2:], "*/")
		if j < 0 {
			return len(s), true
		}
		return i + 2 + j + 2, true
	}
	return 0, false
}

// pcQuitarComaFinal borra la coma colgante antes de un cierre (JS la permite,
// JSON no).
func pcQuitarComaFinal(out []byte) []byte {
	j := len(out)
	for j > 0 && (out[j-1] == ' ' || out[j-1] == '\t' || out[j-1] == '\n' || out[j-1] == '\r') {
		j--
	}
	if j > 0 && out[j-1] == ',' {
		return out[:j-1]
	}
	return out
}

// pcLeerCadena decodifica la cadena JS que empieza en i (comilla simple, doble o
// backtick) y devuelve su valor y el índice siguiente al cierre.
func pcLeerCadena(s string, i int) (string, int, bool) {
	comilla := s[i]
	var b strings.Builder
	for j := i + 1; j < len(s); {
		c := s[j]
		switch c {
		case comilla:
			return b.String(), j + 1, true
		case '\\':
			if j+1 >= len(s) {
				return "", 0, false
			}
			e := s[j+1]
			switch e {
			case 'n':
				b.WriteByte('\n')
			case 't':
				b.WriteByte('\t')
			case 'r':
				b.WriteByte('\r')
			case 'b':
				b.WriteByte('\b')
			case 'f':
				b.WriteByte('\f')
			case '0':
				b.WriteByte(0)
			case 'u', 'x':
				ancho := 4
				if e == 'x' {
					ancho = 2
				}
				if j+2+ancho <= len(s) {
					if r, err := pcHexARuna(s[j+2 : j+2+ancho]); err == nil {
						b.WriteRune(r)
						j += 2 + ancho
						continue
					}
				}
				b.WriteByte(e)
			case '\n': // continuación de línea: no aporta carácter
			default:
				b.WriteByte(e)
			}
			j += 2
		default:
			b.WriteByte(c)
			j++
		}
	}
	return "", 0, false
}

func pcHexARuna(h string) (rune, error) {
	var v rune
	for i := 0; i < len(h); i++ {
		c := h[i]
		var d rune
		switch {
		case c >= '0' && c <= '9':
			d = rune(c - '0')
		case c >= 'a' && c <= 'f':
			d = rune(c-'a') + 10
		case c >= 'A' && c <= 'F':
			d = rune(c-'A') + 10
		default:
			return 0, fmt.Errorf("hex inválido: %q", h)
		}
		v = v*16 + d
	}
	return v, nil
}

func pcEsInicioIdent(c byte) bool {
	return c == '_' || c == '$' || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}

func pcEsIdent(c byte) bool {
	return pcEsInicioIdent(c) || (c >= '0' && c <= '9')
}
