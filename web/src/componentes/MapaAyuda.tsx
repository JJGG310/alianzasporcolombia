import { useEffect, useMemo } from 'react';
import { CircleMarker, MapContainer, Popup, TileLayer, useMap } from 'react-leaflet';
import L from 'leaflet';
import 'leaflet/dist/leaflet.css';

import type { Punto } from '../types';
import { esCurado, tieneCoordenadas } from '../data/carga';
import {
  enlaceWhatsApp,
  estadoLabel,
  fechaCorta,
  lugar,
  tipoUI,
  ubicacion,
} from '../data/catalogo';

interface Props {
  puntos: Punto[];
}

/** Centro por defecto: Valle del Cauca, como en el sitio original. */
const CENTRO: [number, number] = [3.9, -76.35];
const ZOOM = 7;

/** Reencuadra el mapa cuando cambian los puntos filtrados. */
function Encuadre({ puntos }: { puntos: Array<Punto & { lat: number; lon: number }> }) {
  const mapa = useMap();
  useEffect(() => {
    if (puntos.length === 0) return;
    const bordes = L.latLngBounds(puntos.map((p) => [p.lat, p.lon] as [number, number]));
    mapa.fitBounds(bordes, { padding: [28, 28], maxZoom: 14, animate: false });
  }, [mapa, puntos]);
  return null;
}

export default function MapaAyuda({ puntos }: Props) {
  const conCoords = useMemo(() => puntos.filter(tieneCoordenadas), [puntos]);
  const sinCoords = puntos.length - conCoords.length;

  const tiposPresentes = useMemo(
    () => [...new Set(conCoords.map((p) => p.tipo))].sort(),
    [conCoords],
  );

  return (
    <div>
      {conCoords.length === 0 ? (
        <div className="mapa mapa-vacio" role="status">
          <p>
            Ninguno de los {puntos.length} puntos filtrados tiene coordenadas confiables.
            Siguen todos abajo, en el listado.
          </p>
        </div>
      ) : (
        <MapContainer
          center={CENTRO}
          zoom={ZOOM}
          scrollWheelZoom={false}
          className="mapa"
          // El mapa es un complemento: el listado de abajo es la fuente accesible.
          aria-label="Mapa de puntos de ayuda"
        >
          <TileLayer
            url="https://tile.openstreetmap.org/{z}/{x}/{y}.png"
            attribution="&copy; OpenStreetMap"
            maxZoom={19}
          />
          <Encuadre puntos={conCoords} />
          {conCoords.map((p) => {
            const t = tipoUI(p.tipo);
            const curado = esCurado(p);
            return (
              <CircleMarker
                key={p.id}
                center={[p.lat, p.lon]}
                radius={curado ? 9 : 7}
                pathOptions={{
                  color: t.color,
                  weight: curado ? 3 : 1,
                  dashArray: curado ? undefined : '2 3',
                  fillColor: t.color,
                  fillOpacity: curado ? 0.8 : 0.45,
                }}
              >
                <Popup>
                  <h4>{p.nombre}</h4>
                  <p>
                    <span className="pill" style={{ background: `${t.color}18`, color: t.color }}>
                      {t.label}
                    </span>{' '}
                    <span className={curado ? 'sello sello-curado' : 'sello sello-agregado'}>
                      {curado ? '✓ verificado a mano' : '⟳ agregado automáticamente'}
                    </span>
                  </p>
                  {ubicacion(p) && <p>{ubicacion(p)}</p>}
                  <p>{lugar(p)}</p>
                  {p.descripcion && <p>{p.descripcion}</p>}
                  {p.contacto && <p>☎ {p.contacto}</p>}
                  <p>
                    {estadoLabel(p.estado)} · datos al {fechaCorta(p.fecha_corte)} ·{' '}
                    <a href={p.fuente_url} target="_blank" rel="noopener noreferrer">
                      fuente
                    </a>
                  </p>
                  <p>
                    <a
                      className="boton boton-wa"
                      href={enlaceWhatsApp(p)}
                      target="_blank"
                      rel="noopener noreferrer"
                    >
                      Compartir por WhatsApp
                    </a>
                  </p>
                </Popup>
              </CircleMarker>
            );
          })}
        </MapContainer>
      )}

      <ul className="leyenda-mapa">
        {tiposPresentes.map((t) => (
          <li key={t}>
            <span className="punto-leyenda" style={{ background: tipoUI(t).color }} />
            {tipoUI(t).label}
          </li>
        ))}
        {conCoords.length > 0 && (
          <li>
            <span
              className="punto-leyenda"
              style={{ background: '#fff', border: '2px dashed #9aa1b1' }}
            />
            Relleno tenue y borde punteado = agregado automáticamente
          </li>
        )}
        {sinCoords > 0 && (
          <li>
            {sinCoords} punto{sinCoords === 1 ? '' : 's'} sin coordenadas confiables: solo
            en el listado.
          </li>
        )}
      </ul>
    </div>
  );
}
