package model

import (
	"regexp"
	"strings"
)

// Principio #1 del proyecto: nunca publicamos cuentas bancarias ni datos de pago.
// Al agregar datos de terceros no controlamos qué escribe la gente en un campo
// libre, así que redactamos lo que parezca dato de pago y marcamos el punto para
// revisión humana en vez de publicarlo tal cual.

var (
	// "cuenta de ahorros 123456789", "cuenta bancolombia ..."
	//
	// Ojo con abreviar de más: "Cta 98c 58 72" es una DIRECCIÓN de Cali, no una
	// cuenta. Por eso no se acepta "cta" suelto y por eso todo candidato pasa
	// después por suficientesDigitos: una cuenta real tiene 8+ dígitos y una
	// nomenclatura de calle no.
	reCuenta = regexp.MustCompile(`(?i)\b(cuenta|ahorros|corriente|bancolombia|davivienda|bbva|banco de bogot[áa]|av\.? villas|scotiabank|colpatria|banco caja social)\b[^.\n]{0,40}?\d[\d\s.\-]{5,}`)
	// Nequi / Daviplata / Movii: el número es un celular, pero anunciado como
	// destino de plata deja de ser "contacto" y pasa a ser dato de pago.
	reBilletera = regexp.MustCompile(`(?i)\b(nequi|daviplata|davi ?plata|movii|powwi|tpaga|bancolombia a la mano)\b[^.\n]{0,30}?\d[\d\s.\-]{6,}`)
	// Enlaces de recaudo.
	reRecaudo = regexp.MustCompile(`(?i)https?://\S*(paypal|gofundme|vaki\.co|donadora|mercadopago|wompi|payu|epayco|bold\.co|kickstarter)\S*`)
	// Cripto.
	reCripto = regexp.MustCompile(`\b(0x[a-fA-F0-9]{40}|bc1[a-z0-9]{25,60}|[13][a-km-zA-HJ-NP-Z1-9]{25,34})\b`)
)

const redactado = "[dato de pago retirado — política del sitio]"

// Sanear redacta posibles datos de pago en los campos de texto libre y anota la
// alerta para que una persona revise antes de publicar.
func Sanear(p *Punto) {
	campos := []*string{&p.Descripcion, &p.Contacto, &p.Nombre, &p.Direccion, &p.Organizacion, &p.Horario}
	nombres := []string{"descripcion", "contacto", "nombre", "direccion", "organizacion", "horario"}

	for i, c := range campos {
		limpio, motivo := redactar(*c)
		if motivo != "" {
			*c = limpio
			p.PoliticaAlerta = append(p.PoliticaAlerta,
				"posible dato de pago en '"+nombres[i]+"' ("+motivo+"): redactado, revisar antes de publicar")
		}
	}

	for i, s := range p.Necesita {
		if limpio, motivo := redactar(s); motivo != "" {
			p.Necesita[i] = limpio
			p.PoliticaAlerta = append(p.PoliticaAlerta, "posible dato de pago en 'necesita' ("+motivo+"): redactado")
		}
	}
	for i, s := range p.Ofrece {
		if limpio, motivo := redactar(s); motivo != "" {
			p.Ofrece[i] = limpio
			p.PoliticaAlerta = append(p.PoliticaAlerta, "posible dato de pago en 'ofrece' ("+motivo+"): redactado")
		}
	}
}

func redactar(s string) (string, string) {
	if s == "" {
		return s, ""
	}
	var motivos []string
	for _, r := range []struct {
		re          *regexp.Regexp
		motivo      string
		pideDigitos bool
	}{
		{reCuenta, "cuenta bancaria", true},
		{reBilletera, "billetera digital", true},
		{reRecaudo, "enlace de recaudo", false},
		{reCripto, "dirección cripto", false},
	} {
		coincidencias := r.re.FindAllString(s, -1)
		reemplazo := 0
		for _, c := range coincidencias {
			// Un número de cuenta o de celular tiene 8+ dígitos. Una dirección
			// como "Cta 98c 58 72" no llega, y así no borramos un punto de ayuda
			// legítimo por parecerse a un dato de pago.
			if r.pideDigitos && !suficientesDigitos(c, 8) {
				continue
			}
			s = strings.Replace(s, c, redactado, 1)
			reemplazo++
		}
		if reemplazo > 0 {
			motivos = append(motivos, r.motivo)
		}
	}
	return s, strings.Join(motivos, ", ")
}

func suficientesDigitos(s string, minimo int) bool {
	n := 0
	for _, r := range s {
		if r >= '0' && r <= '9' {
			n++
		}
	}
	return n >= minimo
}
