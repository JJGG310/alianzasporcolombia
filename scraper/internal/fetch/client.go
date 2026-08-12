// Package fetch es el cliente HTTP compartido: timeouts, reintentos y un
// User-Agent honesto que dice quiénes somos y cómo contactarnos.
package fetch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const UserAgent = "AlianzasPorColombiaBot/1.0 (+https://alianzasporcolombia.com; agregador ciudadano de ayuda humanitaria; contacto: hola@alianzasporcolombia.com)"

// Cliente hace peticiones con reintento exponencial ante 5xx y errores de red.
type Cliente struct {
	HTTP       *http.Client
	Reintentos int
	EsperaBase time.Duration
}

func Nuevo() *Cliente {
	return &Cliente{
		HTTP:       &http.Client{Timeout: 45 * time.Second},
		Reintentos: 3,
		EsperaBase: 800 * time.Millisecond,
	}
}

// Get descarga una URL y devuelve el cuerpo.
func (c *Cliente) Get(ctx context.Context, url string) ([]byte, error) {
	return c.hacer(ctx, http.MethodGet, url, nil, nil)
}

// PostJSON envía un cuerpo JSON y devuelve la respuesta cruda.
func (c *Cliente) PostJSON(ctx context.Context, url string, cuerpo any) ([]byte, error) {
	b, err := json.Marshal(cuerpo)
	if err != nil {
		return nil, fmt.Errorf("serializando cuerpo: %w", err)
	}
	return c.hacer(ctx, http.MethodPost, url, b, map[string]string{"Content-Type": "application/json"})
}

// GetJSON descarga y deserializa en destino.
func (c *Cliente) GetJSON(ctx context.Context, url string, destino any) error {
	b, err := c.Get(ctx, url)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(b, destino); err != nil {
		return fmt.Errorf("json inválido de %s: %w", url, err)
	}
	return nil
}

func (c *Cliente) hacer(ctx context.Context, metodo, url string, cuerpo []byte, cabeceras map[string]string) ([]byte, error) {
	var ultimo error
	for intento := 0; intento <= c.Reintentos; intento++ {
		if intento > 0 {
			espera := c.EsperaBase * time.Duration(1<<(intento-1))
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(espera):
			}
		}

		var lector io.Reader
		if cuerpo != nil {
			lector = bytes.NewReader(cuerpo)
		}
		req, err := http.NewRequestWithContext(ctx, metodo, url, lector)
		if err != nil {
			return nil, err
		}
		req.Header.Set("User-Agent", UserAgent)
		req.Header.Set("Accept-Language", "es-CO,es;q=0.9")
		for k, v := range cabeceras {
			req.Header.Set(k, v)
		}

		resp, err := c.HTTP.Do(req)
		if err != nil {
			ultimo = err
			continue
		}
		b, err := io.ReadAll(io.LimitReader(resp.Body, 32<<20))
		resp.Body.Close()
		if err != nil {
			ultimo = err
			continue
		}
		if resp.StatusCode >= 500 {
			ultimo = fmt.Errorf("%s devolvió %d", url, resp.StatusCode)
			continue
		}
		if resp.StatusCode >= 400 {
			// 4xx no se reintenta: no va a mejorar solo.
			return b, fmt.Errorf("%s devolvió %d: %s", url, resp.StatusCode, primeros(b, 200))
		}
		return b, nil
	}
	return nil, fmt.Errorf("tras %d intentos: %w", c.Reintentos+1, ultimo)
}

func primeros(b []byte, n int) string {
	if len(b) > n {
		return string(b[:n]) + "…"
	}
	return string(b)
}
