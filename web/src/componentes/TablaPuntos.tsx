import type { Punto } from '../types';
import { esCurado, requiereRevision } from '../data/carga';
import {
  enlaceWhatsApp,
  estadoClase,
  estadoLabel,
  fechaCorta,
  lugar,
  queHace,
  tipoUI,
  ubicacion,
} from '../data/catalogo';

interface Props {
  puntos: Punto[];
}

/** Vista de escritorio. Misma información que las tarjetas, más densa.
 *  Los puntos sin coordenadas viven aquí: nunca se esconden por no ser mapeables. */
export default function TablaPuntos({ puntos }: Props) {
  return (
    <div className="tabla-wrap">
      <table className="tabla-puntos">
        <caption className="sr-solo">
          Puntos de ayuda filtrados, con su fuente, fecha de corte y estado.
        </caption>
        <colgroup>
          <col className="c-tipo" />
          <col className="c-nombre" />
          <col className="c-donde" />
          <col className="c-que" />
          <col className="c-contacto" />
          <col className="c-estado" />
          <col className="c-origen" />
          <col className="c-fuente" />
          <col className="c-compartir" />
        </colgroup>
        <thead>
          <tr>
            <th scope="col">Tipo</th>
            <th scope="col">Nombre</th>
            <th scope="col">Dónde</th>
            <th scope="col">Qué</th>
            <th scope="col">Contacto</th>
            <th scope="col">Estado</th>
            <th scope="col">Origen</th>
            <th scope="col">Fuente y corte</th>
            <th scope="col">
              <span className="sr-solo">Compartir</span>
            </th>
          </tr>
        </thead>
        <tbody>
          {puntos.map((p) => {
            const t = tipoUI(p.tipo);
            const curado = esCurado(p);
            return (
              <tr key={p.id} className={curado ? 'fila-curado' : 'fila-agregado'}>
                <td>
                  <span
                    className="pill"
                    style={{ background: `${t.color}18`, color: t.color }}
                  >
                    {t.label}
                  </span>
                </td>
                <td className="nombre">
                  {p.nombre}
                  {requiereRevision(p) && (
                    <>
                      <br />
                      <small>⚠️ requiere revisión: {p.politica_alerta?.join('; ')}</small>
                    </>
                  )}
                </td>
                <td>
                  {ubicacion(p)}
                  {ubicacion(p) && <br />}
                  {lugar(p)}
                  {p.horario && (
                    <>
                      <br />
                      <small>{p.horario}</small>
                    </>
                  )}
                </td>
                <td>{queHace(p)}</td>
                <td>{p.contacto}</td>
                <td>
                  <span className={estadoClase(p.estado)}>{estadoLabel(p.estado)}</span>
                </td>
                <td>
                  <span className={curado ? 'sello sello-curado' : 'sello sello-agregado'}>
                    {curado ? '✓ a mano' : '⟳ automático'}
                  </span>
                  {!curado && p.fuente && (
                    <>
                      <br />
                      <small>{p.fuente}</small>
                    </>
                  )}
                </td>
                <td>
                  <a href={p.fuente_url} target="_blank" rel="noopener noreferrer">
                    ver
                  </a>
                  <br />
                  <small>{fechaCorta(p.fecha_corte)}</small>
                </td>
                <td className="col-acciones">
                  <a
                    className="boton boton-wa"
                    href={enlaceWhatsApp(p)}
                    target="_blank"
                    rel="noopener noreferrer"
                    aria-label={`Compartir ${p.nombre} por WhatsApp`}
                    title="Compartir por WhatsApp"
                  >
                    WA
                  </a>
                </td>
              </tr>
            );
          })}
        </tbody>
      </table>
    </div>
  );
}
