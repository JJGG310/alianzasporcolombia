// Comando scrape: recorre todas las fuentes, normaliza, fusiona y escribe
// extracted-data/. Un binario sin dependencias externas — se compila y se corre
// en cualquier computador del equipo.
//
//	go build -o bin/scrape ./cmd/scrape
//	./bin/scrape -salida ../extracted-data
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/JJGG310/alianzasporcolombia/scraper/internal/fetch"
	"github.com/JJGG310/alianzasporcolombia/scraper/internal/merge"
	"github.com/JJGG310/alianzasporcolombia/scraper/internal/model"
	"github.com/JJGG310/alianzasporcolombia/scraper/internal/output"
	"github.com/JJGG310/alianzasporcolombia/scraper/internal/sources"
)

func main() {
	var (
		salida    = flag.String("salida", "../extracted-data", "carpeta donde escribir los datos")
		curados   = flag.String("curados", "../datos/ayuda.json", "JSON curado a mano que se fusiona con lo scrapeado (vacío para omitirlo)")
		soloStr   = flag.String("fuentes", "", "lista separada por comas de fuentes a correr (vacío = todas)")
		listar    = flag.Bool("listar", false, "listar las fuentes disponibles y salir")
		timeout   = flag.Duration("timeout", 3*time.Minute, "tiempo máximo total")
		paralelas = flag.Int("paralelas", 4, "cuántas fuentes consultar a la vez")
		seco      = flag.Bool("seco", false, "correr sin escribir nada a disco")
	)
	flag.Parse()

	if *listar {
		for _, f := range sources.Todas() {
			fmt.Printf("%-18s %s\n%18s %s\n", f.Slug(), f.URL(), "", f.Mecanismo())
		}
		return
	}

	var solo []string
	if *soloStr != "" {
		for _, s := range strings.Split(*soloStr, ",") {
			if s = strings.TrimSpace(s); s != "" {
				solo = append(solo, s)
			}
		}
	}
	fuentes := sources.Por(solo)
	if len(fuentes) == 0 {
		fatal(fmt.Errorf("ninguna fuente coincide con %q (usa -listar)", *soloStr))
	}

	// Ctrl-C debe cortar limpio: mejor no dejar extracted-data/ a medias.
	ctx, cancelar := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancelar()
	ctx, cancelarTimeout := context.WithTimeout(ctx, *timeout)
	defer cancelarTimeout()

	inicio := time.Now()
	cliente := fetch.Nuevo()

	fmt.Fprintf(os.Stderr, "Alianzas por Colombia · scraper\nConsultando %d fuentes…\n\n", len(fuentes))

	resultados := correr(ctx, cliente, fuentes, *paralelas)

	// Los puntos curados a mano entran a la fusión como una fuente más, con la
	// prioridad más alta: si un dato verificado por teléfono choca con uno
	// scrapeado, gana el verificado.
	todos := []model.Punto{}
	if *curados != "" {
		cur, err := cargarCurados(*curados, inicio)
		if err != nil {
			fmt.Fprintf(os.Stderr, "  aviso: no se pudieron cargar los curados (%v)\n\n", err)
		} else {
			fmt.Fprintf(os.Stderr, "  curado             %4d puntos verificados a mano\n\n", len(cur))
			todos = append(todos, cur...)
		}
	}
	for _, r := range resultados {
		if r != nil {
			todos = append(todos, r.Puntos...)
		}
	}

	fusionados, inf := merge.Fusionar(todos)
	imprimirResumen(resultados, inf, time.Since(inicio))

	if *seco {
		fmt.Fprintln(os.Stderr, "\n(-seco: no se escribió nada)")
		return
	}

	if err := output.Escribir(*salida, resultados, fusionados, inf, inicio); err != nil {
		fatal(err)
	}
	abs, _ := filepath.Abs(*salida)
	fmt.Fprintf(os.Stderr, "\nEscrito en %s\n  puntos.json  puntos.geojson  fuentes.json  raw/*.json\n", abs)

	// Si TODAS las fuentes fallaron, salimos con error: así una tarea
	// programada avisa en vez de sobrescribir con silencio.
	fallaron := 0
	for _, r := range resultados {
		if r == nil || !r.OK {
			fallaron++
		}
	}
	if fallaron == len(resultados) {
		fatal(fmt.Errorf("todas las fuentes fallaron"))
	}
}

// correr consulta las fuentes en paralelo con un límite de concurrencia: no
// queremos golpear ocho sitios de emergencia al tiempo desde una sola IP.
func correr(ctx context.Context, c *fetch.Cliente, fuentes []sources.Fuente, limite int) []*model.Resultado {
	if limite < 1 {
		limite = 1
	}
	resultados := make([]*model.Resultado, len(fuentes))
	sem := make(chan struct{}, limite)
	var wg sync.WaitGroup
	var mu sync.Mutex

	for i, f := range fuentes {
		wg.Add(1)
		go func(i int, f sources.Fuente) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			// Un adaptador con pánico no puede tumbar la corrida entera.
			defer func() {
				if r := recover(); r != nil {
					resultados[i] = &model.Resultado{
						Fuente: f.Slug(), URL: f.URL(), Mecanismo: f.Mecanismo(),
						OK: false, Error: fmt.Sprintf("pánico en el adaptador: %v", r),
						Obtenido: time.Now().UTC().Format(time.RFC3339),
					}
					mu.Lock()
					fmt.Fprintf(os.Stderr, "  %-18s PÁNICO  %v\n", f.Slug(), r)
					mu.Unlock()
				}
			}()

			res, err := f.Obtener(ctx, c)
			if res == nil {
				res = &model.Resultado{Fuente: f.Slug(), URL: f.URL(), Mecanismo: f.Mecanismo()}
			}
			if err != nil {
				res.OK = false
				if res.Error == "" {
					res.Error = err.Error()
				}
			}
			if res.Obtenido == "" {
				res.Obtenido = time.Now().UTC().Format(time.RFC3339)
			}
			res.Total = len(res.Puntos)
			resultados[i] = res

			mu.Lock()
			if res.OK {
				fmt.Fprintf(os.Stderr, "  %-18s %4d puntos  %s\n", f.Slug(), res.Total, res.Duracion)
			} else {
				fmt.Fprintf(os.Stderr, "  %-18s FALLÓ    %s\n", f.Slug(), recortar(res.Error, 90))
			}
			mu.Unlock()
		}(i, f)
	}
	wg.Wait()
	return resultados
}

// curado refleja el esquema de datos/ayuda.json, el archivo que el equipo
// mantiene a mano.
type curado struct {
	Actualizado string `json:"actualizado"`
	Items       []struct {
		Tipo         string   `json:"tipo"`
		Nombre       string   `json:"nombre"`
		Departamento string   `json:"departamento"`
		Municipio    string   `json:"municipio"`
		Direccion    string   `json:"direccion"`
		Que          string   `json:"que"`
		Contacto     string   `json:"contacto"`
		Fuente       string   `json:"fuente"`
		FechaCorte   string   `json:"fecha_corte"`
		Lat          *float64 `json:"lat"`
		Lon          *float64 `json:"lon"`
		Estado       string   `json:"estado"`
	} `json:"items"`
}

func cargarCurados(ruta string, corte time.Time) ([]model.Punto, error) {
	b, err := os.ReadFile(ruta)
	if err != nil {
		return nil, err
	}
	var c curado
	if err := json.Unmarshal(b, &c); err != nil {
		return nil, err
	}
	out := make([]model.Punto, 0, len(c.Items))
	for _, it := range c.Items {
		crudo, _ := json.Marshal(it)
		p := model.Punto{
			ID:           model.HacerID("curado", it.Nombre, it.Municipio, it.Direccion),
			Fuente:       "curado",
			FuenteURL:    it.Fuente,
			Tipo:         it.Tipo,
			Nombre:       it.Nombre,
			Descripcion:  it.Que,
			Departamento: it.Departamento,
			Municipio:    it.Municipio,
			Direccion:    it.Direccion,
			Contacto:     it.Contacto,
			Lat:          it.Lat,
			Lon:          it.Lon,
			Estado:       it.Estado,
			FechaCorte:   it.FechaCorte,
			Verificado:   true, // por definición: alguien del equipo lo confirmó
			Raw:          crudo,
		}
		if p.Estado == "" {
			p.Estado = model.EstadoActivo
		}
		model.Normalizar(&p, corte)
		out = append(out, p)
	}
	return out, nil
}

func imprimirResumen(resultados []*model.Resultado, inf merge.Informe, dur time.Duration) {
	fmt.Fprintf(os.Stderr, "\n%d puntos en bruto → %d tras fusionar (%d duplicados entre sitios)\n",
		inf.Entrada, inf.Salida, inf.Fusionados)
	fmt.Fprintf(os.Stderr, "  %d con coordenadas · %d corroborados por 2+ fuentes · %d requieren revisión humana\n",
		inf.ConCoordenadas, inf.Corroborados, inf.RequierenRevision)

	fmt.Fprintln(os.Stderr, "\nPor tipo:")
	for _, k := range ordenar(inf.PorTipo) {
		fmt.Fprintf(os.Stderr, "  %-14s %4d\n", k, inf.PorTipo[k])
	}
	fmt.Fprintf(os.Stderr, "\nTotal en %s\n", dur.Round(time.Millisecond))
}

func ordenar(m map[string]int) []string {
	ks := make([]string, 0, len(m))
	for k := range m {
		ks = append(ks, k)
	}
	sort.Slice(ks, func(i, j int) bool {
		if m[ks[i]] != m[ks[j]] {
			return m[ks[i]] > m[ks[j]]
		}
		return ks[i] < ks[j]
	})
	return ks
}

func recortar(s string, n int) string {
	s = strings.Join(strings.Fields(s), " ")
	if len(s) > n {
		return s[:n] + "…"
	}
	return s
}

func fatal(err error) {
	fmt.Fprintf(os.Stderr, "\nerror: %v\n", err)
	os.Exit(1)
}
