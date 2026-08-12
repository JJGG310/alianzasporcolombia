.PHONY: correr construir todas fmt vet limpiar husmear descubrir instalar-navegador

# Uso típico: `make correr` desde scraper/. Escribe en ../extracted-data/.
correr:
	go run ./cmd/scrape -salida ../extracted-data

# Binario portátil: cópialo a cualquier máquina del equipo y córrelo.
construir:
	go build -o bin/scrape ./cmd/scrape
	@echo "Listo: scraper/bin/scrape"

# Binarios para las tres plataformas del equipo, sin dependencias.
todas:
	GOOS=darwin  GOARCH=arm64 go build -o bin/scrape-mac-arm64   ./cmd/scrape
	GOOS=darwin  GOARCH=amd64 go build -o bin/scrape-mac-intel   ./cmd/scrape
	GOOS=linux   GOARCH=amd64 go build -o bin/scrape-linux-amd64 ./cmd/scrape
	GOOS=windows GOARCH=amd64 go build -o bin/scrape-windows.exe ./cmd/scrape
	@ls -lh bin/

fmt:
	gofmt -w ./cmd ./internal

vet:
	go vet ./...

# Diagnóstico sin navegador: de dónde saca sus datos un sitio.
#   make husmear URL=https://un-sitio-nuevo.com
husmear:
	go run ./cmd/husmear $(URL)

# Diagnóstico con navegador real (módulo aparte, ver descubrir/).
# Nota: la descarga del driver de Playwright está caída del lado de Microsoft.
instalar-navegador:
	cd descubrir && go mod tidy && go run . -instalar

descubrir:
	cd descubrir && go run . -salida ../../extracted-data/diagnostico-fuentes.json

limpiar:
	rm -rf bin/
