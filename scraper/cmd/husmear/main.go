// Comando husmear: averigua de dónde saca sus datos un sitio, sin navegador.
//
// Descarga el HTML y los bundles JS que enlaza, y busca las huellas de los
// backends que usa la gente para montar estas apps a las carreras: Convex,
// Firestore, Supabase, Google Sheets publicados, rutas /api/…, payloads de
// Next.js embebidos, arreglos JS escritos a mano.
//
// Es el hermano pobre de ../descubrir (que abre un navegador real y ve TODAS
// las peticiones), pero no depende de nada: se compila y corre en cualquier
// parte. Con estos ocho sitios acertó en casi todos; el único que se le escapa
// es el que habla solo por WebSocket sin dejar la URL en el bundle.
//
//	go run ./cmd/husmear https://un-sitio-nuevo.com
package main

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"

	"github.com/JJGG310/alianzasporcolombia/scraper/internal/fetch"
)

// Cada pista dice qué encontró y qué hacer con eso.
var pistas = []struct {
	nombre  string
	re      *regexp.Regexp
	quePasa string
}{
	{"Convex", regexp.MustCompile(`https://[a-z0-9-]+\.convex\.(cloud|site)`),
		"Los datos van por WebSocket, pero se pueden pedir por HTTP:\n     POST https://<deployment>.convex.cloud/api/query\n     {\"path\":\"modulo:funcion\",\"args\":{},\"format\":\"json\"}\n     El nombre de la función no está en el bundle: sale espiando el WebSocket con ../descubrir, o parchando WebSocket.prototype.send en la consola del navegador y mirando el mensaje ModifyQuerySet."},
	{"Firebase / Firestore", regexp.MustCompile(`(?i)projectId"?\s*:\s*"([a-z0-9-]+)"`),
		"Si la base permite lectura anónima, se lee por REST:\n     https://firestore.googleapis.com/v1/projects/<projectId>/databases/(default)/documents/<coleccion>?key=<apiKey>&pageSize=300\n     La apiKey está en el mismo archivo. Busca también el mapa de colecciones."},
	{"Supabase", regexp.MustCompile(`https://[a-z0-9]+\.supabase\.co`),
		"Prueba la REST con la anon key del bundle:\n     https://<proyecto>.supabase.co/rest/v1/<tabla>?select=*  (cabecera apikey)\n     Un 401 significa RLS activo: no hay datos públicos."},
	{"Google Sheets publicado", regexp.MustCompile(`https://docs\.google\.com/spreadsheets/[^"'\s)]+`),
		"Es la fuente entera y ya viene en CSV. Descárgala directo."},
	{"Next.js (payload RSC embebido)", regexp.MustCompile(`self\.__next_f\.push`),
		"Los datos vienen dentro del HTML, en el stream RSC. Hay que reconstruir\n     los trozos y sacar el arreglo JSON de adentro (ver internal/sources/calisolidario.go)."},
	{"Arreglos JS escritos a mano", regexp.MustCompile(`const\s+[A-Z_]{3,}\s*=\s*\[`),
		"No hay API: los datos están escritos en el código. Hay que parsear los\n     literales JS (ver internal/sources/puntoscriticos.go)."},
	{"Airtable", regexp.MustCompile(`https://api\.airtable\.com/[^"'\s)]+`), "API de Airtable: necesita token."},
	{"Cloudflare Worker", regexp.MustCompile(`https://[a-z0-9-]+\.workers\.dev`), "Suele traer su propia API REST. Mira las rutas /api/ de abajo."},
}

var reRutasAPI = regexp.MustCompile(`["'](/api/[a-zA-Z0-9/_-]{2,60})["']`)
var reScripts = regexp.MustCompile(`(?i)<script[^>]+src=["']([^"']+)["']`)
var reModulos = regexp.MustCompile(`(?i)from\s*['"](\.{1,2}/[^'"]+\.js)['"]`)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "uso: husmear <url> [url…]")
		os.Exit(2)
	}
	c := fetch.Nuevo()
	ctx := context.Background()

	for _, url := range os.Args[1:] {
		fmt.Printf("\n═══ %s\n", url)
		if err := husmear(ctx, c, url); err != nil {
			fmt.Printf("  error: %v\n", err)
		}
	}
}

func husmear(ctx context.Context, c *fetch.Cliente, url string) error {
	html, err := c.Get(ctx, url)
	if err != nil {
		return err
	}
	fuentes := map[string]string{"(html)": string(html)}

	// Los bundles JS son donde de verdad están las URLs del backend. Y hay que
	// seguir los imports en cascada: en una app de módulos ES el config con las
	// claves suele estar dos o tres saltos abajo del script de entrada
	// (app.js → mapa.js → config.js), no en el que enlaza el HTML.
	base := strings.TrimSuffix(url, "/")
	pendientes := []string{}
	for _, m := range reScripts.FindAllStringSubmatch(string(html), -1) {
		pendientes = append(pendientes, resolver(base, m[1]))
	}

	visitados := map[string]bool{}
	for len(pendientes) > 0 && len(fuentes) < 40 {
		src := pendientes[0]
		pendientes = pendientes[1:]
		if src == "" || visitados[src] || esCDN(src) {
			continue
		}
		visitados[src] = true

		js, err := c.Get(ctx, src)
		if err != nil {
			continue
		}
		fuentes[corto(src)] = string(js)

		for _, mm := range reModulos.FindAllStringSubmatch(string(js), -1) {
			pendientes = append(pendientes, resolver(recortarArchivo(src), mm[1]))
		}
	}

	fmt.Printf("  %d archivos analizados\n", len(fuentes))

	encontrado := false
	for _, p := range pistas {
		var donde []string
		vistos := map[string]bool{}
		for nombre, cuerpo := range fuentes {
			for _, m := range p.re.FindAllString(cuerpo, 3) {
				clave := nombre + m
				if vistos[clave] {
					continue
				}
				vistos[clave] = true
				donde = append(donde, fmt.Sprintf("%s → %s", nombre, recorta(m, 110)))
			}
		}
		if len(donde) == 0 {
			continue
		}
		encontrado = true
		sort.Strings(donde)
		fmt.Printf("\n  ● %s\n", p.nombre)
		for _, d := range donde {
			fmt.Printf("     %s\n", d)
		}
		fmt.Printf("     ⇒ %s\n", p.quePasa)
	}

	// Rutas de API propias del sitio.
	rutas := map[string]bool{}
	for _, cuerpo := range fuentes {
		for _, m := range reRutasAPI.FindAllStringSubmatch(cuerpo, -1) {
			rutas[m[1]] = true
		}
	}
	if len(rutas) > 0 {
		encontrado = true
		lista := make([]string, 0, len(rutas))
		for r := range rutas {
			lista = append(lista, r)
		}
		sort.Strings(lista)
		fmt.Printf("\n  ● Rutas de API propias\n")
		for _, r := range lista {
			fmt.Printf("     %s%s\n", base, r)
		}
		fmt.Printf("     ⇒ Pruébalas con curl. Las que devuelvan JSON son la fuente.\n")
	}

	if !encontrado {
		fmt.Println("\n  Nada reconocible. Puede que los datos estén embebidos en el HTML,\n" +
			"  o que el sitio hable solo por WebSocket. Corre ../descubrir con navegador real.")
	}
	return nil
}

func resolver(base, ruta string) string {
	switch {
	case strings.HasPrefix(ruta, "http"):
		return ruta
	case strings.HasPrefix(ruta, "//"):
		return "https:" + ruta
	case strings.HasPrefix(ruta, "/"):
		return raiz(base) + ruta
	case strings.HasPrefix(ruta, "./"):
		return base + "/" + strings.TrimPrefix(ruta, "./")
	case strings.HasPrefix(ruta, "../"):
		return base + "/" + ruta // aproximado; suficiente para husmear
	default:
		return base + "/" + ruta
	}
}

// esCDN salta las librerías de terceros: no aportan pistas y pesan mucho.
func esCDN(u string) bool {
	for _, h := range []string{"unpkg.com", "cdnjs.", "jsdelivr", "gstatic.com", "googletagmanager", "clarity.ms", "cloudflare.com/ajax"} {
		if strings.Contains(u, h) {
			return true
		}
	}
	return false
}

func raiz(u string) string {
	i := strings.Index(u, "://")
	if i < 0 {
		return u
	}
	if j := strings.Index(u[i+3:], "/"); j >= 0 {
		return u[:i+3+j]
	}
	return u
}

func recortarArchivo(u string) string {
	if i := strings.LastIndex(u, "/"); i > 8 {
		return u[:i]
	}
	return u
}

func corto(u string) string {
	if i := strings.LastIndex(u, "/"); i >= 0 && i < len(u)-1 {
		return u[i+1:]
	}
	return u
}

func recorta(s string, n int) string {
	s = strings.Join(strings.Fields(s), " ")
	if len(s) > n {
		return s[:n] + "…"
	}
	return s
}
