// Comando descubrir: abre cada sitio con un navegador real y registra TODAS las
// peticiones de datos que hace (fetch/XHR, WebSocket, documentos), para saber de
// dónde saca su información.
//
// Por qué existe: los adaptadores de ../internal/sources apuntan a endpoints
// concretos que descubrimos así. Cuando un sitio cambia de forma —y en una
// emergencia cambian cada día— el adaptador se rompe. En vez de adivinar, se
// corre esto y se ve el nuevo endpoint.
//
// Va en su propio módulo Go a propósito: Playwright arrastra dependencias y la
// descarga de un navegador. El scraper de producción (../cmd/scrape) no depende
// de nada externo y se compila en cualquier máquina; esto es la herramienta de
// diagnóstico que se corre cuando algo se rompe.
//
// Uso:
//
//	go run . -instalar          # una sola vez: baja Chromium
//	go run .                    # audita todos los sitios conocidos
//	go run . -url https://…     # audita uno suelto
//	go run . -salida informe.json
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/playwright-community/playwright-go"
)

// Los sitios ciudadanos que agregamos. Mantener sincronizado con
// ../internal/sources/registry.go.
var sitios = []struct{ Slug, URL string }{
	{"mapa-emergencia", "https://mapa-emergencia.artefactofilms.workers.dev/"},
	{"aqui-hace-falta", "https://aqui-hace-falta.web.app/"},
	{"haciendo-comunidad", "https://personnofound.github.io/HaciendoComunidad/"},
	{"calisolidario", "https://calisolidario.triadaaliados.com/necesidades"},
	{"quesenecesita", "https://quesenecesita.org/"},
	{"sismoinfo", "https://sismoinfo.co/"},
	{"puntos-criticos", "https://terremoto-cali-puntos-criticos.netlify.app/"},
	{"aidtrace", "https://aidtrace-rastroayuda.vercel.app/"},
	{"window-of-hope", "https://window-of-hope-countdown.lovable.app/"},
}

// Ruido que no aporta nada al diagnóstico.
var ruido = regexp.MustCompile(`(?i)(google-analytics|googletagmanager|clarity\.ms|cloudflareinsights|doubleclick|facebook\.|hotjar|sentry\.io|zgo\.at|fonts\.(googleapis|gstatic)|/og\.png|favicon)`)

// Extensiones de recursos estáticos: tampoco son fuentes de datos.
var estatico = regexp.MustCompile(`(?i)\.(js|css|png|jpe?g|gif|svg|webp|woff2?|ttf|ico|map)(\?|$)`)

type peticion struct {
	Metodo     string `json:"metodo"`
	URL        string `json:"url"`
	Estado     int    `json:"estado,omitempty"`
	TipoCuerpo string `json:"tipo_contenido,omitempty"`
	Bytes      int    `json:"bytes,omitempty"`
	Muestra    string `json:"muestra,omitempty"`
}

type informeSitio struct {
	Slug        string     `json:"slug"`
	URL         string     `json:"url"`
	Titulo      string     `json:"titulo"`
	Error       string     `json:"error,omitempty"`
	Datos       []peticion `json:"peticiones_de_datos"`
	WebSockets  []string   `json:"websockets,omitempty"`
	MensajesWS  []string   `json:"mensajes_websocket,omitempty"`
	Diagnostico string     `json:"diagnostico"`
}

func main() {
	var (
		instalar = flag.Bool("instalar", false, "descargar el navegador de Playwright y salir")
		una      = flag.String("url", "", "auditar una sola URL en vez de la lista conocida")
		salida   = flag.String("salida", "", "escribir el informe JSON a este archivo")
		espera   = flag.Duration("espera", 6*time.Second, "cuánto esperar tras cargar, para que la app haga sus peticiones")
		visible  = flag.Bool("visible", false, "mostrar el navegador (útil para depurar)")
	)
	flag.Parse()

	if *instalar {
		if err := playwright.Install(&playwright.RunOptions{Browsers: []string{"chromium"}}); err != nil {
			fatal(fmt.Errorf("instalando navegador: %w", err))
		}
		fmt.Println("Chromium instalado. Ya puedes correr: go run .")
		return
	}

	pw, err := playwright.Run()
	if err != nil {
		fatal(fmt.Errorf("iniciando Playwright (¿corriste `go run . -instalar`?): %w", err))
	}
	defer pw.Stop()

	navegador, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(!*visible),
	})
	if err != nil {
		fatal(fmt.Errorf("abriendo Chromium: %w", err))
	}
	defer navegador.Close()

	objetivos := sitios
	if *una != "" {
		objetivos = []struct{ Slug, URL string }{{"adhoc", *una}}
	}

	informes := make([]informeSitio, 0, len(objetivos))
	for _, s := range objetivos {
		fmt.Fprintf(os.Stderr, "→ %s\n", s.URL)
		inf := auditar(navegador, s.Slug, s.URL, *espera)
		informes = append(informes, inf)
		imprimir(inf)
	}

	if *salida != "" {
		b, _ := json.MarshalIndent(informes, "", "  ")
		if err := os.WriteFile(*salida, append(b, '\n'), 0o644); err != nil {
			fatal(err)
		}
		fmt.Fprintf(os.Stderr, "\nInforme escrito en %s\n", *salida)
	}
}

func auditar(navegador playwright.Browser, slug, url string, espera time.Duration) informeSitio {
	inf := informeSitio{Slug: slug, URL: url}

	ctx, err := navegador.NewContext(playwright.BrowserNewContextOptions{
		UserAgent: playwright.String("AlianzasPorColombiaBot/1.0 (+https://alianzasporcolombia.com; diagnóstico de fuentes)"),
	})
	if err != nil {
		inf.Error = err.Error()
		return inf
	}
	defer ctx.Close()

	pagina, err := ctx.NewPage()
	if err != nil {
		inf.Error = err.Error()
		return inf
	}

	vistos := map[string]bool{}

	// Las respuestas nos dicen qué endpoint devolvió datos y de qué tamaño.
	pagina.On("response", func(r playwright.Response) {
		u := r.URL()
		if ruido.MatchString(u) || estatico.MatchString(u) {
			return
		}
		ct := r.Headers()["content-type"]
		// Solo nos interesan respuestas de datos, no HTML de navegación.
		esDatos := strings.Contains(ct, "json") || strings.Contains(ct, "csv") ||
			strings.Contains(ct, "text/plain") || strings.Contains(ct, "geo+json")
		if !esDatos && r.Request().ResourceType() != "fetch" && r.Request().ResourceType() != "xhr" {
			return
		}
		clave := r.Request().Method() + " " + u
		if vistos[clave] {
			return
		}
		vistos[clave] = true

		p := peticion{Metodo: r.Request().Method(), URL: u, Estado: r.Status(), TipoCuerpo: ct}
		if cuerpo, err := r.Body(); err == nil {
			p.Bytes = len(cuerpo)
			p.Muestra = recortar(string(cuerpo), 400)
		}
		inf.Datos = append(inf.Datos, p)
	})

	// WebSocket: así es como hablan Convex, Firebase y Supabase Realtime. Sin
	// esto un sitio parece "sin API" cuando en realidad tiene toda su data ahí.
	pagina.OnWebSocket(func(ws playwright.WebSocket) {
		inf.WebSockets = append(inf.WebSockets, ws.URL())
		ws.OnFrameSent(func(datos []byte) {
			if t := string(datos); t != "" && len(inf.MensajesWS) < 25 {
				inf.MensajesWS = append(inf.MensajesWS, "→ "+recortar(t, 300))
			}
		})
		ws.OnFrameReceived(func(datos []byte) {
			if t := string(datos); t != "" && len(inf.MensajesWS) < 25 {
				inf.MensajesWS = append(inf.MensajesWS, "← "+recortar(t, 300))
			}
		})
	})

	if _, err := pagina.Goto(url, playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateNetworkidle,
		Timeout:   playwright.Float(45000),
	}); err != nil {
		// networkidle falla en sitios con polling o sockets abiertos; no es fatal.
		inf.Error = "aviso al cargar: " + err.Error()
	}

	pagina.WaitForTimeout(float64(espera.Milliseconds()))

	if t, err := pagina.Title(); err == nil {
		inf.Titulo = t
	}

	sort.Slice(inf.Datos, func(i, j int) bool { return inf.Datos[i].Bytes > inf.Datos[j].Bytes })
	inf.Diagnostico = diagnosticar(inf)
	return inf
}

// diagnosticar traduce lo observado a una recomendación de cómo scrapear.
func diagnosticar(inf informeSitio) string {
	for _, ws := range inf.WebSockets {
		switch {
		case strings.Contains(ws, "convex"):
			return "Convex por WebSocket. Los datos se pueden pedir por HTTP: POST https://<deployment>.convex.cloud/api/query con {\"path\":\"modulo:funcion\",\"args\":{},\"format\":\"json\"}. El nombre de la función sale en los mensajes ModifyQuerySet de abajo (campo udfPath)."
		case strings.Contains(ws, "firestore") || strings.Contains(ws, "firebaseio"):
			return "Firebase/Firestore por WebSocket. Revisa si el proyecto permite lectura anónima por la API REST de Firestore."
		case strings.Contains(ws, "supabase"):
			return "Supabase Realtime. Prueba la API REST (/rest/v1/<tabla>) con la anon key del bundle; si da 401, hay RLS y no es accesible."
		}
	}
	if len(inf.Datos) > 0 {
		var urls []string
		for _, p := range inf.Datos {
			if p.Estado >= 200 && p.Estado < 300 && p.Bytes > 500 {
				urls = append(urls, p.Metodo+" "+p.URL)
			}
		}
		if len(urls) > 0 {
			return "API HTTP directa. Endpoints con datos: " + strings.Join(urls, " · ")
		}
		return "Hace peticiones de datos pero ninguna devolvió contenido útil (¿caída, o requiere sesión?)."
	}
	return "Sin peticiones de datos: los datos vienen embebidos en el HTML (SSR o arreglos JS en el bundle). Hay que parsear el documento."
}

func imprimir(inf informeSitio) {
	fmt.Printf("\n%s — %s\n", inf.Slug, inf.Titulo)
	if inf.Error != "" {
		fmt.Printf("  %s\n", inf.Error)
	}
	for _, p := range inf.Datos {
		fmt.Printf("  [%d] %s %s (%d bytes, %s)\n", p.Estado, p.Metodo, p.URL, p.Bytes, p.TipoCuerpo)
	}
	for _, ws := range inf.WebSockets {
		fmt.Printf("  [WS] %s\n", ws)
	}
	for _, m := range inf.MensajesWS {
		fmt.Printf("       %s\n", m)
	}
	fmt.Printf("  ⇒ %s\n", inf.Diagnostico)
}

func recortar(s string, n int) string {
	s = strings.TrimSpace(s)
	if len(s) > n {
		return s[:n] + "…"
	}
	return s
}

func fatal(err error) {
	fmt.Fprintf(os.Stderr, "error: %v\n", err)
	os.Exit(1)
}
