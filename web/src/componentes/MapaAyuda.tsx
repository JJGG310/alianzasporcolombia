import { useEffect, useMemo } from 'react';
import { CircleMarker, MapContainer, Popup, TileLayer, useMap } from 'react-leaflet';
import L from 'leaflet';
import 'leaflet/dist/leaflet.css';

import type { Punto } from '../types';
import { esCurado, tieneCoordenadas } from '../data/carga';
import {
  GRUPOS_MAPA,
  distanciaKm,
  distanciaTexto,
  enlaceWhatsApp,
  estadoLabel,
  fechaCorta,
  grupoDe,
  lugar,
  tipoUI,
  ubicacion,
  type GrupoMapa,
} from '../data/catalogo';
import Icono from './Icono';

interface Props {
  puntos: Punto[];
  /** Ubicación del visitante, si la concedió. Nunca sale del navegador. */
  miUbicacion?: { lat: number; lon: number } | null;
}

/** Centro por defecto: Valle del Cauca. */
const CENTRO: [number, number] = [3.9, -76.35];
const ZOOM = 7;

/** Reencuadra cuando cambian los puntos filtrados. */
function Encuadre({ puntos }: { puntos: Array<Punto & { lat: number; lon: number }> }) {
  const mapa = useMap();
  useEffect(() => {
    if (puntos.length === 0) return;
    const bordes = L.latLngBounds(puntos.map((p) => [p.lat, p.lon] as [number, number]));
    mapa.fitBounds(bordes, { padding: [28, 28], maxZoom: 14, animate: false });
  }, [mapa, puntos]);
  return null;
}

export default function MapaAyuda({ puntos, miUbicacion }: Props) {
  // Un punto con precision="aproximada" no tiene dirección geocodificada: su
  // coordenada es el centro del municipio. Dibujarlo miente dos veces — manda a
  // la persona a la plaza principal en vez de al lugar, y apila 9 marcadores
  // idénticos en Cali y 8 en Pereira, de los que solo uno se puede abrir.
  // Siguen en la lista, con su dirección escrita, que es como sí se llega.
  const conCoords = useMemo(
    () => puntos.filter(tieneCoordenadas).filter((p) => p.precision !== 'aproximada'),
    [puntos],
  );
  const sinCoords = puntos.length - conCoords.length;

  // La leyenda solo muestra lo que de verdad está en pantalla. Una leyenda con
  // grupos vacíos hace dudar de si el filtro funcionó.
  const gruposPresentes = useMemo(() => {
    const vistos = new Set<GrupoMapa>();
    for (const p of conCoords) vistos.add(grupoDe(p.tipo));
    return (Object.keys(GRUPOS_MAPA) as GrupoMapa[]).filter((g) => vistos.has(g));
  }, [conCoords]);

  const hayUrgentes = useMemo(
    () => conCoords.some((p) => p.estado === 'urgente'),
    [conCoords],
  );

  return (
    <>
      <div className="mapa">
        {conCoords.length === 0 ? (
          <div className="mapa__cargando" role="status">
            <p className="prosa" style={{ textAlign: 'center', padding: 'var(--e4)' }}>
              Ninguno de los {puntos.length} puntos filtrados tiene coordenadas
              confiables. Están todos abajo, en el listado.
            </p>
          </div>
        ) : (
          <MapContainer
            center={CENTRO}
            zoom={ZOOM}
            scrollWheelZoom={false}
            style={{ height: '100%', width: '100%' }}
            // El mapa complementa; el listado de abajo es la versión accesible
            // y la que se puede leer con lector de pantalla.
            aria-label="Mapa de puntos de ayuda"
          >
            <TileLayer
              url="https://tile.openstreetmap.org/{z}/{x}/{y}.png"
              attribution="&copy; OpenStreetMap"
              maxZoom={19}
            />
            <Encuadre puntos={conCoords} />

            {miUbicacion && (
              <CircleMarker
                center={[miUbicacion.lat, miUbicacion.lon]}
                radius={7}
                pathOptions={{
                  color: '#ffffff',
                  weight: 3,
                  fillColor: '#14459e',
                  fillOpacity: 1,
                }}
              >
                <Popup>Estás aquí (aproximado)</Popup>
              </CircleMarker>
            )}

            {conCoords.map((p) => {
              const grupo = GRUPOS_MAPA[grupoDe(p.tipo)];
              const curado = esCurado(p);
              const urgente = p.estado === 'urgente';
              const t = tipoUI(p.tipo);
              const km = miUbicacion
                ? distanciaKm(miUbicacion, { lat: p.lat, lon: p.lon })
                : null;

              return (
                <CircleMarker
                  key={p.id}
                  center={[p.lat, p.lon]}
                  radius={urgente ? 8 : 6}
                  pathOptions={{
                    // Anillo blanco de 2px: separa marcadores encimados, que en
                    // Cali centro son decenas. Sin él el mapa es una mancha.
                    color: urgente ? 'var(--urgente)' : '#ffffff',
                    weight: 2,
                    fillColor: grupo.color,
                    // El relleno tenue es el segundo canal que distingue lo
                    // verificado de lo agregado, además del sello en la ficha.
                    fillOpacity: curado ? 0.95 : 0.55,
                  }}
                >
                  <Popup>
                    <strong style={{ fontSize: 'var(--txt-base)' }}>{p.nombre}</strong>
                    <p style={{ marginTop: 'var(--e1)', color: 'var(--tinta-500)' }}>
                      {t.label}
                      {urgente ? ' · urgente' : ''}
                      {curado ? ' · verificado a mano' : ' · agregado automáticamente'}
                      {km != null ? ` · a ${distanciaTexto(km)}` : ''}
                    </p>
                    {ubicacion(p) && <p>{ubicacion(p)}</p>}
                    <p>{lugar(p)}</p>
                    {p.contacto && <p>{p.contacto}</p>}
                    <p style={{ marginTop: 'var(--e2)' }}>
                      <a
                        href={`https://www.google.com/maps/dir/?api=1&destination=${p.lat},${p.lon}`}
                        target="_blank"
                        rel="noopener noreferrer"
                      >
                        Cómo llegar
                      </a>
                      {' · '}
                      <a href={enlaceWhatsApp(p)} target="_blank" rel="noopener noreferrer">
                        Compartir
                      </a>
                    </p>
                    <p style={{ fontSize: 'var(--txt-xs)', color: 'var(--tinta-500)' }}>
                      {estadoLabel(p.estado)} · datos al {fechaCorta(p.fecha_corte)} ·{' '}
                      <a href={p.fuente_url} target="_blank" rel="noopener noreferrer">
                        fuente
                      </a>
                    </p>
                  </Popup>
                </CircleMarker>
              );
            })}
          </MapContainer>
        )}
      </div>

      {/* La leyenda no es opcional: varios de estos colores quedan por debajo de
          3:1 contra las teselas, y el texto es lo que sostiene la identidad. */}
      <ul className="mapa-pie" aria-label="Qué significa cada color">
        {gruposPresentes.map((g) => (
          <li key={g} className="mapa-pie__clave" style={{ color: GRUPOS_MAPA[g].color }}>
            <span className="mapa-pie__punto" />
            <span style={{ color: 'var(--tinta-500)' }}>{GRUPOS_MAPA[g].label}</span>
          </li>
        ))}
        {hayUrgentes && (
          <li className="mapa-pie__clave" style={{ color: 'var(--urgente)' }}>
            <span className="mapa-pie__punto" />
            <span style={{ color: 'var(--tinta-500)' }}>Urgente (borde rojo)</span>
          </li>
        )}
        {conCoords.length > 0 && (
          <li className="mapa-pie__clave">
            <span style={{ color: 'var(--tinta-500)' }}>
              Relleno tenue = agregado de otro sitio
            </span>
          </li>
        )}
        {sinCoords > 0 && (
          <li className="mapa-pie__clave">
            <Icono nombre="alerta" />
            <span>
              {sinCoords} punto{sinCoords === 1 ? '' : 's'} sin coordenadas: solo en el
              listado.
            </span>
          </li>
        )}
      </ul>
    </>
  );
}
