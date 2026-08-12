package model

import (
	"testing"
	"time"
)

func TestSanearRedactaDatosDePago(t *testing.T) {
	casos := []struct {
		nombre string
		texto  string
	}{
		{"cuenta de ahorros", "Ayuda con mercados. Cuenta de ahorros Bancolombia 123-456789-01"},
		{"nequi", "Necesitamos colchonetas, giren al Nequi 3001234567"},
		{"daviplata", "Daviplata 310 555 4433 para donaciones"},
		{"enlace de recaudo", "Apóyanos en https://vaki.co/vaki/ayudacali"},
		{"dirección cripto", "ETH: 0x1234567890abcdef1234567890abcdef12345678"},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			p := Punto{Descripcion: c.texto}
			Sanear(&p)
			if len(p.PoliticaAlerta) == 0 {
				t.Fatalf("no se marcó alerta para %q", c.texto)
			}
			if p.Descripcion == c.texto {
				t.Errorf("no se redactó: %q", p.Descripcion)
			}
		})
	}
}

func TestSanearNoTocaTextoLimpio(t *testing.T) {
	// Un teléfono de contacto suelto NO es un dato de pago: redactarlo dejaría
	// al punto sin la información que lo hace útil.
	limpios := []string{
		"Contacto: María Paula 314 7874022",
		"Se necesitan 20 palas y guantes de carnaza",
		"Albergue abierto 8:00 a.m. a 6:00 p.m., Comuna 21",
		// Nomenclatura de Cali: "Cta" es dirección, no "cuenta". Un caso real
		// que el saneador estaba borrando de un punto de rescate urgente.
		"Cta 98c 58 72 unidad Santa ana A",
		"Cuenta con 30 voluntarios en el punto",
	}
	for _, texto := range limpios {
		p := Punto{Descripcion: texto}
		Sanear(&p)
		if len(p.PoliticaAlerta) > 0 {
			t.Errorf("falso positivo en %q: %v", texto, p.PoliticaAlerta)
		}
		if p.Descripcion != texto {
			t.Errorf("modificó texto limpio: %q → %q", texto, p.Descripcion)
		}
	}
}

func TestNormalizarDescartaCoordenadasFueraDeColombia(t *testing.T) {
	lat, lon := 40.7128, -74.0060 // Nueva York
	p := Punto{Nombre: "Punto raro", Lat: &lat, Lon: &lon}
	Normalizar(&p, time.Now().UTC())
	if p.Lat != nil || p.Lon != nil {
		t.Error("no descartó coordenadas fuera de Colombia")
	}
	if len(p.Avisos) == 0 {
		t.Error("no dejó constancia del descarte")
	}
	if len(p.PoliticaAlerta) > 0 {
		t.Error("una coordenada mala no debe bloquear la publicación del punto")
	}
}

func TestHacerIDEsDeterminista(t *testing.T) {
	a := HacerID("fuente", "Coliseo El Pueblo", "Cali")
	b := HacerID("fuente", "  coliseo el pueblo  ", "CALI")
	if a != b {
		t.Errorf("el id debería ignorar mayúsculas y espacios: %s vs %s", a, b)
	}
	if c := HacerID("fuente", "Otro lugar"); c == a {
		t.Error("puntos distintos no deberían compartir id")
	}
}
