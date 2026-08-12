package sources

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/JJGG310/alianzasporcolombia/scraper/internal/fetch"
	"github.com/JJGG310/alianzasporcolombia/scraper/internal/model"
)

// AidTrace rastrea donaciones y entregas con respaldo en blockchain.
//
// Su único endpoint público (/api/timeline) está a medias: la mayoría de las
// veces responde 500 porque no encuentra el recibo de una transacción, y de vez
// en cuando devuelve 200 con los eventos del contrato. El Supabase del sitio
// exige autenticación (401), así que no hay otra puerta pública. Por eso se
// intenta en cada corrida y, cuando falla, se reporta el motivo sin inventar.
type AidTrace struct{}

const (
	atSlug        = "aidtrace"
	atSitio       = "https://aidtrace-rastroayuda.vercel.app"
	atURLTimeline = atSitio + "/api/timeline?limit=100&cursor=0"
	atURLConfig   = atSitio + "/api/config"
	atMecanismo   = "API propia /api/timeline (actualmente caída) — el Supabase del sitio exige autenticación"
	atNotaCaida   = "endpoint público caído (500, error de blockchain); el Supabase del sitio exige autenticación (401). Sin datos públicos disponibles — se reintenta en cada corrida."
	// El endpoint alterna 200 con datos y 500: los reintentos del cliente son los
	// que lo hacen utilizable, así que conviene decirlo en la salida.
	atNotaIntermitente = "endpoint intermitente: alterna 200 con datos y 500 (error de blockchain, viem); el Supabase del sitio exige autenticación (401)"
)

func (a *AidTrace) Slug() string      { return atSlug }
func (a *AidTrace) URL() string       { return atSitio }
func (a *AidTrace) Mecanismo() string { return atMecanismo }

// atConfig declara a propósito solo dos campos: /api/config también devuelve
// supabaseAnonKey, privyAppId y turnstileSiteKey, y esas claves no se copian
// jamás a nuestra salida.
type atConfig struct {
	App          string `json:"app"`
	AuthRequired bool   `json:"authRequired"`
}

func (a *AidTrace) Obtener(ctx context.Context, c *fetch.Cliente) (*model.Resultado, error) {
	inicio := time.Now()
	res := &model.Resultado{
		Fuente:    atSlug,
		URL:       atSitio,
		Mecanismo: atMecanismo,
	}
	extra := map[string]any{}
	terminar := func() {
		if len(extra) > 0 {
			res.Extra = extra
		}
		res.Duracion = time.Since(inicio).String()
		res.Obtenido = time.Now().UTC().Format(time.RFC3339)
		res.Total = len(res.Puntos)
	}

	var cfg atConfig
	if err := c.GetJSON(ctx, atURLConfig, &cfg); err == nil {
		extra["app"] = cfg.App
		extra["authRequired"] = cfg.AuthRequired
	}

	cuerpo, err := c.Get(ctx, atURLTimeline)
	if err != nil {
		res.OK = false
		res.Error = fmt.Sprintf("timeline: %v", err)
		res.Nota = atNotaCaida
		terminar()
		// El error va de vuelta junto con el resultado: el runner lo registra
		// pero no aborta el resto de las fuentes.
		return res, fmt.Errorf("aidtrace: %w", err)
	}

	var doc any
	if err := json.Unmarshal(cuerpo, &doc); err != nil {
		res.OK = false
		res.Error = fmt.Sprintf("timeline devolvió algo que no es JSON: %v", err)
		res.Nota = atNotaCaida
		terminar()
		return res, fmt.Errorf("aidtrace: %w", err)
	}

	// Responde 200 con {"ok":false,...} cuando falla por dentro: un 200 no basta.
	if raiz, ok := doc.(map[string]any); ok {
		if bien, hay := raiz["ok"].(bool); hay && !bien {
			detalle := atTexto(raiz, "error", "message", "detail")
			res.OK = false
			res.Error = "timeline respondió 200 pero con ok=false: " + detalle
			res.Nota = atNotaCaida
			terminar()
			return res, fmt.Errorf("aidtrace: %s", res.Error)
		}
	}

	corte := time.Now().UTC()
	items := atItems(doc)
	if len(items) == 0 {
		res.OK = true
		res.Nota = "el endpoint respondió, pero no se reconoció ninguna lista de eventos en el esquema (claves vistas: " +
			strings.Join(atClaves(doc), ", ") + "); revisar el adaptador"
		terminar()
		return res, nil
	}

	puntos := make([]model.Punto, 0, len(items))
	var descartados int
	for _, m := range items {
		p, ok := atPunto(m, corte)
		if !ok {
			descartados++
			continue
		}
		puntos = append(puntos, p)
	}
	res.Puntos = puntos
	res.OK = true
	notas := []string{atNotaIntermitente}
	if descartados > 0 {
		notas = append(notas, fmt.Sprintf("%d eventos sin texto reconocible descartados", descartados))
	}
	if atHayMas(doc) {
		notas = append(notas, "hay más páginas (hasMore=true): solo se lee la primera de 100 eventos")
	}
	res.Nota = strings.Join(notas, "; ")
	terminar()
	return res, nil
}

// atHayMas mira la paginación sin exigir que exista.
func atHayMas(doc any) bool {
	raiz, ok := doc.(map[string]any)
	if !ok {
		return false
	}
	pag, ok := raiz["pagination"].(map[string]any)
	if !ok {
		return false
	}
	mas, _ := pag["hasMore"].(bool)
	return mas
}

// atItems busca la lista de eventos sin conocer el esquema: puede venir suelta
// o envuelta en cualquiera de los nombres habituales.
func atItems(doc any) []map[string]any {
	switch v := doc.(type) {
	case []any:
		return atMapas(v)
	case map[string]any:
		for _, k := range []string{"items", "data", "timeline", "results", "events", "eventos", "entries", "rows", "list", "records"} {
			switch hijo := v[k].(type) {
			case []any:
				if m := atMapas(hijo); len(m) > 0 {
					return m
				}
			case map[string]any:
				if m := atItems(hijo); len(m) > 0 { // p. ej. {"data":{"items":[...]}}
					return m
				}
			}
		}
	}
	return nil
}

func atMapas(v []any) []map[string]any {
	out := make([]map[string]any, 0, len(v))
	for _, e := range v {
		if m, ok := e.(map[string]any); ok {
			out = append(out, m)
		}
	}
	return out
}

func atClaves(doc any) []string {
	m, ok := doc.(map[string]any)
	if !ok {
		return []string{fmt.Sprintf("%T", doc)}
	}
	claves := make([]string, 0, len(m))
	for k := range m {
		claves = append(claves, k)
	}
	return claves
}

// atPunto traduce un evento a model.Punto. El esquema no está documentado y el
// endpoint va y viene, así que se lee por nombres alternativos de campo y todo
// el registro queda en Raw para poder reprocesar sin volver a pedirlo.
func atPunto(m map[string]any, corte time.Time) (model.Punto, bool) {
	detalle := atTexto(m, "details", "detail", "detalle", "description", "descripcion", "message", "mensaje", "texto", "text", "notes", "nota", "summary")
	accion := atAccion(atTexto(m, "actionType", "action", "accion", "evento", "event"))
	lote := atTexto(m, "batchId", "batch", "lote")

	nombre := atTexto(m, "title", "titulo", "nombre", "name", "label", "asunto")
	if nombre == "" {
		nombre = atRecortar(detalle, 120)
	}
	if nombre == "" {
		nombre = accion
	}
	if nombre == "" {
		return model.Punto{}, false
	}

	var partes []string
	if accion != "" {
		partes = append(partes, accion)
	}
	if lote != "" {
		partes = append(partes, "lote "+lote)
	}
	if detalle != "" && detalle != nombre {
		partes = append(partes, detalle)
	}

	crudo, err := json.Marshal(m)
	if err != nil {
		crudo = nil
	}
	id := atTexto(m, "id", "_id", "uuid", "hash", "txHash", "tx_hash", "transactionHash", "slug")
	if id == "" {
		id = nombre + "|" + lote
	}
	// Cada evento tiene su transacción pública: esa es la URL citable del ítem.
	fuenteURL := atTexto(m, "txUrl", "url", "link", "qrLink")
	if fuenteURL == "" {
		fuenteURL = atSitio
	}

	p := model.Punto{
		ID:           model.HacerID(atSlug, id),
		Fuente:       atSlug,
		FuenteURL:    fuenteURL,
		Tipo:         atTipo(m),
		Nombre:       nombre,
		Descripcion:  strings.Join(partes, " · "),
		Direccion:    atTexto(m, "address", "direccion", "location", "lugar", "ubicacion"),
		Municipio:    atTexto(m, "city", "ciudad", "municipio"),
		Departamento: atTexto(m, "state", "departamento", "region", "depto"),
		Lat:          atNumero(m, "lat", "latitude", "latitud"),
		Lon:          atNumero(m, "lon", "lng", "longitude", "longitud"),
		Contacto:     atTexto(m, "contact", "contacto", "phone", "telefono", "whatsapp"),
		Organizacion: atTexto(m, "organization", "organizacion", "org", "entity", "donor", "donante", "ong", "beneficiario", "recipient"),
		Estado:       atEstado(atTexto(m, "status", "estado", "state")),
		Creado:       atFecha(m, "blockTimestamp", "created_at", "createdAt", "fecha", "date", "timestamp", "time"),
		Raw:          crudo,
	}
	model.Normalizar(&p, corte)
	return p, true
}

// atAccion traduce el vocabulario de acciones del contrato (mayúsculas en
// inglés) a algo legible en español; lo desconocido se deja tal cual.
func atAccion(a string) string {
	switch strings.ToUpper(strings.TrimSpace(a)) {
	case "":
		return ""
	case "CREATED":
		return "Lote creado"
	case "PICKUP":
		return "Recogida"
	case "TRANSFER", "IN_TRANSIT":
		return "En traslado"
	case "ARRIVED":
		return "Llegada al destino"
	case "DELIVERED":
		return "Entregado"
	case "REVIEW":
		return "Revisión"
	case "CANCELLED", "CANCELED":
		return "Cancelado"
	default:
		return a
	}
}

// atTipo: todo lo de AidTrace es model.TipoInfo, a propósito.
//
// Lo que publica no son puntos de ayuda sino eventos de trazabilidad de un lote
// ("Recibido 20 cajas de ibuprofeno", "Llegada al destino"): no tienen dirección
// ni coordenadas, y no hay a dónde ir. Marcarlos como donacion-org los mezclaría
// en la lista de organizaciones a las que sí puedes acudir —eran 21 y pasaban a
// 121— y volvería inútil esa sección. Se conservan porque la trazabilidad de
// donaciones es información valiosa, pero como cronología, no como lugares.
func atTipo(map[string]any) string {
	return model.TipoInfo
}

func atEstado(s string) string {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "urgent", "urgente", "critical", "critico", "crítico":
		return model.EstadoUrgente
	case "active", "activo", "open", "abierto", "pending", "pendiente", "in_progress", "en_curso",
		"sent_to_relayer", "confirmed", "mined", "onchain": // estados de la transacción

		return model.EstadoActivo
	case "delivered", "entregado", "completed", "completado", "done", "cubierto", "fulfilled":
		return model.EstadoCubierto
	case "closed", "cerrado", "cancelled", "canceled", "cancelado":
		return model.EstadoCerrado
	default:
		return "" // Normalizar lo deja en "por-confirmar"
	}
}

func atTexto(m map[string]any, claves ...string) string {
	for _, k := range claves {
		switch v := m[k].(type) {
		case string:
			if s := strings.TrimSpace(v); s != "" {
				return s
			}
		case float64:
			return atNumeroTexto(v)
		}
	}
	return ""
}

func atNumero(m map[string]any, claves ...string) *float64 {
	for _, k := range claves {
		if v, ok := m[k].(float64); ok {
			f := v
			return &f
		}
	}
	return nil
}

// atFecha acepta ISO8601 o epoch (segundos o milisegundos), que es lo que suelen
// devolver estas APIs.
func atFecha(m map[string]any, claves ...string) *time.Time {
	for _, k := range claves {
		switch v := m[k].(type) {
		case string:
			for _, formato := range []string{time.RFC3339, "2006-01-02T15:04:05.000Z", "2006-01-02 15:04:05", "2006-01-02"} {
				if t, err := time.Parse(formato, strings.TrimSpace(v)); err == nil {
					t = t.UTC()
					return &t
				}
			}
		case float64:
			if v <= 0 {
				continue
			}
			ms := int64(v)
			if ms < 1e11 { // venía en segundos
				ms *= 1000
			}
			t := time.UnixMilli(ms).UTC()
			return &t
		}
	}
	return nil
}

func atRecortar(s string, n int) string {
	s = strings.TrimSpace(s)
	if len(s) <= n {
		return s
	}
	corte := strings.LastIndex(s[:n], " ")
	if corte < n/2 {
		corte = n
	}
	return strings.TrimSpace(s[:corte]) + "…"
}

// atNumeroTexto imprime números sin notación científica ni ".0" sobrante.
func atNumeroTexto(v float64) string {
	s := fmt.Sprintf("%.6f", v)
	s = strings.TrimRight(s, "0")
	return strings.TrimSuffix(s, ".")
}
