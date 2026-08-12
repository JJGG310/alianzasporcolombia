import { useEffect, useState } from 'react';
import { haceCuanto } from '../data/catalogo';
import Icono from './Icono';

// Feed público de USGS: CORS abierto, sin autenticación. Se consulta desde el
// navegador del visitante, así que siempre está fresco y no necesita backend.
const FEED =
  'https://earthquake.usgs.gov/earthquakes/feed/v1.0/summary/all_day.geojson';

// Colombia continental, la misma caja que usa el scraper (model.Normalizar).
const CAJA = { latMin: -5, latMax: 14, lonMin: -82, lonMax: -66 };

interface Sismo {
  id: string;
  magnitud: number | null;
  lugar: string;
  tiempo: number;
  profundidad: number | null;
  url: string;
}

interface RespuestaUSGS {
  features?: Array<{
    id: string;
    properties: { mag: number | null; place: string | null; time: number; url: string };
    geometry: { coordinates: [number, number, number] } | null;
  }>;
}

/**
 * Réplicas — franja de una línea.
 *
 * Esto era una tarjeta con borde rojo y tres párrafos. Es información valiosa
 * (es el antídoto contra las cadenas de «viene otro más fuerte») pero NO es a
 * lo que la gente entró: entró a buscar un albergue. Así que el dato duro va
 * primero, en una sola línea, y la explicación completa se despliega. No se
 * borró una sola palabra: se reordenó.
 *
 * La lógica del fetch al USGS no cambia.
 */
export default function Replicas() {
  const [sismos, setSismos] = useState<Sismo[] | null>(null);
  const [error, setError] = useState(false);

  useEffect(() => {
    const ac = new AbortController();
    fetch(FEED, { signal: ac.signal })
      .then((r) => {
        if (!r.ok) throw new Error(`HTTP ${r.status}`);
        return r.json() as Promise<RespuestaUSGS>;
      })
      .then((data) => {
        const enColombia = (data.features ?? [])
          .filter((f) => {
            const c = f.geometry?.coordinates;
            if (!c) return false;
            const [lon, lat] = c;
            return (
              lat >= CAJA.latMin &&
              lat <= CAJA.latMax &&
              lon >= CAJA.lonMin &&
              lon <= CAJA.lonMax
            );
          })
          .map<Sismo>((f) => ({
            id: f.id,
            magnitud: f.properties.mag,
            lugar: f.properties.place ?? 'Colombia',
            tiempo: f.properties.time,
            profundidad: f.geometry?.coordinates[2] ?? null,
            url: f.properties.url,
          }))
          .sort((a, b) => b.tiempo - a.tiempo);
        setSismos(enColombia);
      })
      .catch((e) => {
        if ((e as Error).name !== 'AbortError') setError(true);
      });
    return () => ac.abort();
  }, []);

  // Si el feed no carga, el widget desaparece en vez de mostrar datos falsos.
  if (error) return null;

  if (!sismos) {
    return (
      <section className="replicas-franja" aria-labelledby="t-replicas" aria-busy="true">
        <h2 id="t-replicas" className="solo-lectores">
          Réplicas en vivo
        </h2>
        <p className="replicas-franja__espera">
          <Icono nombre="refrescar" />
          <span className="esqueleto replicas-franja__cargando" aria-hidden="true" />
          <span className="solo-lectores">Consultando el feed público del USGS…</span>
        </p>
      </section>
    );
  }

  const ultima = sismos[0];

  return (
    <section className="replicas-franja" aria-labelledby="t-replicas">
      <h2 id="t-replicas" className="solo-lectores">
        Réplicas en vivo
      </h2>

      {/* Toda la franja es el resorte del desplegable: en un celular, con afán,
          un renglón entero se acierta y un enlace de dos palabras no. */}
      <details className="replicas-franja__caja">
        <summary className="replicas-franja__fila">
          <Icono nombre="refrescar" />
          {ultima ? (
            <>
              <span className="replicas-franja__rotulo">Última réplica:</span>
              <strong className="replicas-franja__mag">
                M {ultima.magnitud?.toFixed(1) ?? '—'}
              </strong>
              <span className="replicas-franja__sep" aria-hidden="true">
                ·
              </span>
              <span className="replicas-franja__cuando">
                {haceCuanto(new Date(ultima.tiempo).toISOString())}
              </span>
              <span className="replicas-franja__sep" aria-hidden="true">
                ·
              </span>
              <span className="replicas-franja__lugar" title={ultima.lugar}>
                {ultima.lugar}
              </span>
            </>
          ) : (
            <span className="replicas-franja__calma">
              Sin réplicas registradas en Colombia en 24 h
            </span>
          )}
          <span className="replicas-franja__por-que">
            por qué
            <Icono nombre="flecha-abajo" className="replicas-franja__flecha" />
          </span>
        </summary>

        <div className="replicas-franja__texto prosa">
          {ultima ? (
            <p>
              {sismos.length} sismo{sismos.length === 1 ? '' : 's'} registrado
              {sismos.length === 1 ? '' : 's'} en Colombia en las últimas 24 horas. El
              último: M {ultima.magnitud?.toFixed(1) ?? '—'} en {ultima.lugar}
              {ultima.profundidad !== null
                ? `, a ${Math.round(ultima.profundidad)} km de profundidad`
                : ''}
              .
            </p>
          ) : (
            <p>
              Sin sismos registrados en Colombia en las últimas 24 horas por el USGS.
            </p>
          )}
          <p>
            Las réplicas son normales y van bajando de frecuencia. Ninguna autoridad
            puede predecir un sismo: desconfía de las cadenas que anuncian «viene otro
            más fuerte».{' '}
            <a
              href="https://earthquake.usgs.gov/earthquakes/feed/v1.0/summary/all_day.geojson"
              target="_blank"
              rel="noopener noreferrer"
            >
              Fuente: USGS
            </a>{' '}
            · consultado desde tu navegador al abrir esta página. La red del SGC
            registra más réplicas:{' '}
            <a
              href="https://www.sgc.gov.co/sismos"
              target="_blank"
              rel="noopener noreferrer"
            >
              sgc.gov.co/sismos
            </a>
            .
          </p>
        </div>
      </details>
    </section>
  );
}
