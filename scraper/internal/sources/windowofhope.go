package sources

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/JJGG310/alianzasporcolombia/scraper/internal/fetch"
	"github.com/JJGG310/alianzasporcolombia/scraper/internal/model"
)

// WindowOfHope es la cuenta regresiva de la "ventana de vida" (las 72 horas
// tras el sismo). No tiene dataset ni hace peticiones de datos: es una pieza
// editorial. Se conserva como fuente de información, no de puntos de ayuda.
type WindowOfHope struct{}

const (
	wohSlug       = "window-of-hope"
	wohSitio      = "https://window-of-hope-countdown.lovable.app"
	wohMecanismo  = "página informativa estática (sin dataset; no hace peticiones de datos)"
	wohNota       = "sin dataset estructurado; se conserva como fuente informativa/editorial"
	wohResumenMax = 600
)

func (w *WindowOfHope) Slug() string      { return wohSlug }
func (w *WindowOfHope) URL() string       { return wohSitio }
func (w *WindowOfHope) Mecanismo() string { return wohMecanismo }

func (w *WindowOfHope) Obtener(ctx context.Context, c *fetch.Cliente) (*model.Resultado, error) {
	inicio := time.Now()
	res := &model.Resultado{
		Fuente:    wohSlug,
		URL:       wohSitio,
		Mecanismo: wohMecanismo,
	}
	terminar := func() {
		res.Duracion = time.Since(inicio).String()
		res.Obtenido = time.Now().UTC().Format(time.RFC3339)
		res.Total = len(res.Puntos)
	}

	cuerpo, err := c.Get(ctx, wohSitio)
	if err != nil {
		res.OK = false
		res.Error = fmt.Sprintf("descargando el HTML: %v", err)
		res.Nota = wohNota
		terminar()
		return res, fmt.Errorf("window-of-hope: %w", err)
	}

	html := string(cuerpo)
	titulo := wohTitulo(html)
	texto := wohTextoVisible(html)
	if titulo == "" {
		titulo = wohRecortar(texto, 80)
	}
	if titulo == "" {
		titulo = "Ventana de Vida"
	}

	corte := time.Now().UTC()
	crudo, _ := json.Marshal(map[string]string{"titulo": titulo, "texto": texto})
	p := model.Punto{
		ID:          model.HacerID(wohSlug, wohSitio),
		Fuente:      wohSlug,
		FuenteURL:   wohSitio,
		Tipo:        model.TipoInfo,
		Nombre:      titulo,
		Descripcion: wohRecortar(texto, wohResumenMax),
		Estado:      model.EstadoActivo,
		Raw:         crudo,
	}
	// El departamento solo se afirma si la propia página lo dice.
	if wohReChoco.MatchString(titulo) || wohReChoco.MatchString(texto) {
		p.Departamento = "Chocó"
	}
	model.Normalizar(&p, corte)

	res.Puntos = []model.Punto{p}
	res.OK = true
	res.Nota = wohNota
	res.Extra = map[string]any{"caracteres_texto": len(texto)}
	terminar()
	return res, nil
}

var (
	wohReTitulo   = regexp.MustCompile(`(?is)<title[^>]*>(.*?)</title>`)
	wohReOGTitulo = regexp.MustCompile(`(?is)<meta[^>]+property=["']og:title["'][^>]+content=["']([^"']*)["']`)
	wohReChoco    = regexp.MustCompile(`(?i)choc[óo]`)
	// Bloques cuyo contenido no es texto para leer. Va uno por uno porque RE2 no
	// tiene retrorreferencias para cerrar la misma etiqueta que abrió.
	wohReBloques    = regexp.MustCompile(`(?is)<script\b[^>]*>.*?</script>|<style\b[^>]*>.*?</style>|<svg\b[^>]*>.*?</svg>|<noscript\b[^>]*>.*?</noscript>|<template\b[^>]*>.*?</template>|<head\b[^>]*>.*?</head>`)
	wohReComentario = regexp.MustCompile(`(?s)<!--.*?-->`)
	wohReEtiqueta   = regexp.MustCompile(`(?s)<[^>]*>`)
	wohReEntidad    = regexp.MustCompile(`&#?[0-9A-Za-z]{1,8};`)
)

func wohTitulo(html string) string {
	if m := wohReTitulo.FindStringSubmatch(html); m != nil {
		if t := wohLimpiar(wohEntidades(m[1])); t != "" {
			return t
		}
	}
	if m := wohReOGTitulo.FindStringSubmatch(html); m != nil {
		return wohLimpiar(wohEntidades(m[1]))
	}
	return ""
}

// wohTextoVisible deja solo el texto que un lector vería: sin scripts, estilos,
// svg ni etiquetas, con entidades resueltas y espacios colapsados.
func wohTextoVisible(html string) string {
	s := wohReBloques.ReplaceAllString(html, " ")
	s = wohReComentario.ReplaceAllString(s, " ")
	s = wohReEtiqueta.ReplaceAllString(s, " ")
	return wohLimpiar(wohEntidades(s))
}

var wohEntidadesNombradas = map[string]string{
	"amp": "&", "lt": "<", "gt": ">", "quot": "\"", "apos": "'",
	"nbsp": " ", "ndash": "–", "mdash": "—", "hellip": "…",
	"laquo": "«", "raquo": "»", "ldquo": "“", "rdquo": "”",
	"lsquo": "‘", "rsquo": "’", "middot": "·", "deg": "°", "euro": "€",
}

func wohEntidades(s string) string {
	return wohReEntidad.ReplaceAllStringFunc(s, func(e string) string {
		cuerpo := strings.Trim(e, "&;")
		if v, ok := wohEntidadesNombradas[strings.ToLower(cuerpo)]; ok {
			return v
		}
		if strings.HasPrefix(cuerpo, "#") {
			base, digitos := 10, cuerpo[1:]
			if len(digitos) > 1 && (digitos[0] == 'x' || digitos[0] == 'X') {
				base, digitos = 16, digitos[1:]
			}
			if n, err := strconv.ParseInt(digitos, base, 32); err == nil && n > 0 {
				return string(rune(n))
			}
		}
		return e
	})
}

// wohLimpiar colapsa espacios y descarta caracteres de control (el HTML servido
// trae algún byte de control que ensuciaría el texto).
func wohLimpiar(s string) string {
	s = strings.Map(func(r rune) rune {
		if r < 0x20 || r == 0x7f || r == ' ' {
			return ' '
		}
		return r
	}, s)
	return strings.TrimSpace(strings.Join(strings.Fields(s), " "))
}

// wohRecortar corta en el último espacio antes del límite para no partir palabras.
func wohRecortar(s string, n int) string {
	if len([]rune(s)) <= n {
		return s
	}
	r := []rune(s)[:n]
	if i := strings.LastIndex(string(r), " "); i > n/2 {
		return strings.TrimSpace(string(r)[:i]) + "…"
	}
	return strings.TrimSpace(string(r)) + "…"
}
