# extracted-data

Datos agregados automáticamente de los sitios ciudadanos que surgieron por el terremoto del 10 de agosto de 2026. **Todo lo de esta carpeta lo genera [`../scraper/`](../scraper/) — no lo edites a mano.**

Para regenerarlo desde cualquier computador del equipo:

```bash
cd scraper && make correr
```

Después revisa el diff de git y súbelo. El diff **es** el control de calidad: la salida es determinista, así que lo que cambie en el diff es lo que cambió en la realidad.

## Archivos

| Archivo | Qué es |
|---|---|
| `puntos.json` | Lo que consume el sitio: todo normalizado, fusionado y con el informe de la corrida |
| `puntos.geojson` | Solo los puntos con coordenadas. Se abre en QGIS, Google My Maps, geojson.io o cualquier mapa |
| `fuentes.json` | Auditoría: qué fuente respondió, cuántos puntos dio, cuánto tardó, qué falló |
| `raw/<fuente>.json` | La respuesta original de cada sitio, tal cual llegó |

`raw/` existe para poder reprocesar sin volver a golpear los sitios de origen: si cambiamos cómo mapeamos un campo, se recalcula desde ahí.

## Cómo leer un punto

```json
{
  "id": "mapa-emergencia-a1b2c3d4e5",
  "fuente": "mapa-emergencia",
  "fuente_url": "https://mapa-emergencia.artefactofilms.workers.dev/#p=sn367eapw7nt",
  "tipo": "rescate",
  "nombre": "Bomberos Zarzal",
  "necesita": ["Agua", "Palas", "Guantes de obra"],
  "estado": "urgente",
  "lat": 4.39, "lon": -76.07,
  "fecha_corte": "2026-08-12T19:40:00Z",
  "verificado": false,
  "tambien_en": [{"fuente": "haciendo-comunidad", "id": "...", "url": "..."}]
}
```

Tres campos hacen el trabajo pesado:

- **`fuente_url`** — de dónde salió. Todo punto lo lleva. Si no lo puedes citar, no lo publicas.
- **`verificado`** — `true` solo si alguien del equipo lo confirmó (los de `datos/ayuda.json`). Lo agregado automáticamente es `false`: está bien mostrarlo, pero no como si lo hubiéramos comprobado.
- **`tambien_en`** — otras fuentes que reportan el mismo punto. Que dos sitios independientes coincidan es la mejor señal de confianza que hay sin llamar por teléfono.

## `politica_alerta`: no publicar sin mirar

Un punto con `politica_alerta` **no se publica sin revisión humana**. Se marca cuando:

- el saneamiento detectó algo que parece dato de pago (cuenta bancaria, Nequi/Daviplata, enlace de recaudo, dirección cripto) y lo redactó;
- el sitio de origen lo tiene marcado como desactualizado o reportado como abuso;
- las coordenadas caían fuera de Colombia y se descartaron.

Agregamos texto que escribe cualquiera. No podemos confiar en que venga limpio, y el principio #1 del proyecto es que aquí nunca aparece una cuenta bancaria.

## Lo que no está aquí, a propósito

**Datos de personas desaparecidas.** Ni de estos sitios ni de ningún otro. Es el principio #2 del proyecto: son datos personales con carga legal (Ley 1581) y humana que no necesitamos asumir, y ya existe un estándar de facto —Colombia Te Busca— al que hay que enlazar, no duplicar. Al menos dos de las fuentes agregadas tienen su propia base de desaparecidos; los adaptadores la saltan explícitamente.

## Licencia

**CC BY 4.0.** Úsalo libremente citando la fuente de cada ítem — cada punto trae la suya en `fuente_url`.

Estos datos vienen de proyectos ciudadanos de terceros. No están verificados uno por uno por Alianzas por Colombia: por eso cada punto cita su origen, para que cualquiera pueda confirmarlo.
