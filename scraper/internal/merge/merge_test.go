package merge

import (
	"testing"

	"github.com/JJGG310/alianzasporcolombia/scraper/internal/model"
)

func f(v float64) *float64 { return &v }

func TestFusionaMismoLugarDeFuentesDistintas(t *testing.T) {
	entrada := []model.Punto{
		{ID: "a", Fuente: "mapa-emergencia", Tipo: model.TipoAlbergue,
			Nombre: "Coliseo El Pueblo", Direccion: "Calle 5 #42-00",
			Lat: f(3.4200), Lon: f(-76.5400), Estado: model.EstadoActivo},
		{ID: "b", Fuente: "calisolidario", Tipo: model.TipoAlbergue,
			Nombre: "Coliseo el Pueblo", Direccion: "Calle 5 #42-00",
			Contacto: "3001234567", Estado: model.EstadoUrgente},
	}
	salida, inf := Fusionar(entrada)

	if len(salida) != 1 {
		t.Fatalf("esperaba 1 punto fusionado, hay %d", len(salida))
	}
	if inf.Fusionados != 1 {
		t.Errorf("el informe no contó la fusión: %+v", inf)
	}
	p := salida[0]
	if len(p.TambienEn) != 1 || p.TambienEn[0].Fuente != "calisolidario" {
		t.Errorf("no dejó rastro de la otra fuente: %+v", p.TambienEn)
	}
	if p.Contacto != "3001234567" {
		t.Error("no absorbió el contacto que solo tenía la otra fuente")
	}
	if p.Estado != model.EstadoUrgente {
		t.Errorf("debía quedarse con el estado más urgente, quedó %q", p.Estado)
	}
}

func TestNoFusionaTiposDistintosEnLaMismaDireccion(t *testing.T) {
	entrada := []model.Punto{
		{ID: "a", Fuente: "mapa-emergencia", Tipo: model.TipoAlbergue,
			Nombre: "Sede comunal", Direccion: "Calle 5 #42-00"},
		{ID: "b", Fuente: "calisolidario", Tipo: model.TipoSangre,
			Nombre: "Sede comunal", Direccion: "Calle 5 #42-00"},
	}
	if salida, _ := Fusionar(entrada); len(salida) != 2 {
		t.Errorf("un albergue y un banco de sangre en la misma cuadra son 2 puntos, quedaron %d", len(salida))
	}
}

func TestNoFusionaAvisosCiudadanos(t *testing.T) {
	// Dos vecinos pidiendo agua son dos pedidos, no uno.
	entrada := []model.Punto{
		{ID: "a", Fuente: "calisolidario", Tipo: model.TipoNecesidad,
			Nombre: "Necesito agua", Direccion: "Comuna 21"},
		{ID: "b", Fuente: "aqui-hace-falta", Tipo: model.TipoNecesidad,
			Nombre: "Necesito agua", Direccion: "Comuna 21"},
	}
	if salida, _ := Fusionar(entrada); len(salida) != 2 {
		t.Errorf("los avisos ciudadanos no se fusionan, quedaron %d", len(salida))
	}
}

func TestNoFusionaDentroDeLaMismaFuente(t *testing.T) {
	entrada := []model.Punto{
		{ID: "a", Fuente: "mapa-emergencia", Tipo: model.TipoAcopio,
			Nombre: "Punto de acopio", Direccion: "Calle 5 #42-00"},
		{ID: "b", Fuente: "mapa-emergencia", Tipo: model.TipoAcopio,
			Nombre: "Punto de acopio", Direccion: "Calle 5 #42-00"},
	}
	if salida, _ := Fusionar(entrada); len(salida) != 2 {
		t.Errorf("dentro de una fuente confiamos en sus ids, quedaron %d", len(salida))
	}
}

func TestFusionaPorCercaniaYNombreParecido(t *testing.T) {
	entrada := []model.Punto{
		{ID: "a", Fuente: "mapa-emergencia", Tipo: model.TipoAcopio,
			Nombre: "Plazoleta Jairo Varela", Lat: f(3.45160), Lon: f(-76.53200)},
		{ID: "b", Fuente: "quesenecesita", Tipo: model.TipoAcopio,
			Nombre: "Plazoleta Jairo Varela (acopio)", Lat: f(3.45165), Lon: f(-76.53205)},
	}
	if salida, _ := Fusionar(entrada); len(salida) != 1 {
		t.Errorf("mismo lugar a pocos metros con nombre parecido, quedaron %d", len(salida))
	}
}

func TestNoFusionaLugaresLejanosConMismoNombre(t *testing.T) {
	// "Bomberos" hay en cada municipio del Valle.
	entrada := []model.Punto{
		{ID: "a", Fuente: "mapa-emergencia", Tipo: model.TipoRescate,
			Nombre: "Bomberos", Lat: f(3.4516), Lon: f(-76.5320)},
		{ID: "b", Fuente: "quesenecesita", Tipo: model.TipoRescate,
			Nombre: "Bomberos", Lat: f(4.3928), Lon: f(-76.0733)},
	}
	if salida, _ := Fusionar(entrada); len(salida) != 2 {
		t.Errorf("dos bomberos a 100 km son 2 puntos, quedaron %d", len(salida))
	}
}

func TestElCuradoGanaComoCanonico(t *testing.T) {
	entrada := []model.Punto{
		{ID: "auto", Fuente: "calisolidario", Tipo: model.TipoAlbergue,
			Nombre: "Coliseo El Pueblo", Direccion: "Calle 5 #42-00", Verificado: false},
		{ID: "manual", Fuente: "curado", Tipo: model.TipoAlbergue,
			Nombre: "Coliseo El Pueblo", Direccion: "Calle 5 #42-00", Verificado: true},
	}
	salida, _ := Fusionar(entrada)
	if len(salida) != 1 {
		t.Fatalf("esperaba 1, hay %d", len(salida))
	}
	if salida[0].Fuente != "curado" || !salida[0].Verificado {
		t.Errorf("el dato verificado a mano debe mandar, quedó %+v", salida[0].Fuente)
	}
}
