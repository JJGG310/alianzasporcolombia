import Icono from './Icono';
import { useOrigen, distanciaDe } from './Filtros';
import { claseSello, enlaceComoLlegar, etiquetasDe, telefonoDe } from './TarjetaPunto';
import { esCurado, requiereRevision } from '../data/carga';
import {
  distanciaTexto,
  enlaceWhatsApp,
  estadoLabel,
  fechaCorta,
  lugar,
  tipoUI,
  ubicacion,
  type Coordenada,
} from '../data/catalogo';
import type { Punto } from '../types';

interface Props {
  puntos: Punto[];
  /** Origen para la columna de distancia. Por defecto, el de "Cerca de mí". */
  origen?: Coordenada | null;
}

/**
 * Vista de escritorio (>= 900 px).
 *
 * Una tabla es lo correcto para barrer 878 filas con los ojos: columnas
 * alineadas, encabezado que no se va al hacer scroll y una fila por punto. Los
 * puntos sin coordenadas viven aquí igual que los demás: nunca se esconden por
 * no ser mapeables.
 */
export default function TablaPuntos({ puntos, origen }: Props) {
  const origenStore = useOrigen();
  const desde = origen ?? origenStore;
  const conDistancia = desde !== null;

  return (
    <div className="tabla-caja">
      <table className="tabla">
        <caption className="solo-lectores">
          Puntos de ayuda filtrados, con su estado, ubicación, qué necesitan, fuente y
          fecha de corte.
        </caption>
        {/* Anchos fijos: sin esto la columna de "qué necesita" empuja los
            botones fuera de la pantalla, y compartir es prioridad del proyecto. */}
        <colgroup>
          <col style={{ width: conDistancia ? '9%' : '10%' }} />
          <col style={{ width: conDistancia ? '18%' : '19%' }} />
          <col style={{ width: conDistancia ? '16%' : '18%' }} />
          {conDistancia && <col style={{ width: '7%' }} />}
          <col style={{ width: conDistancia ? '20%' : '21%' }} />
          <col style={{ width: '14%' }} />
          <col style={{ width: conDistancia ? '8%' : '9%' }} />
          <col style={{ width: '9%' }} />
        </colgroup>
        <thead>
          <tr>
            <th scope="col">Estado</th>
            <th scope="col">Punto</th>
            <th scope="col">Dónde</th>
            {conDistancia && (
              <th scope="col" className="tabla__num">
                Distancia
              </th>
            )}
            <th scope="col">Necesita / ofrece</th>
            <th scope="col">Contacto</th>
            <th scope="col">Fuente</th>
            <th scope="col">
              <span className="solo-lectores">Acciones</span>
            </th>
          </tr>
        </thead>
        <tbody>
          {puntos.map((p) => {
            const curado = esCurado(p);
            const km = distanciaDe(p, desde);
            const tel = telefonoDe(p.contacto);
            const comoLlegar = enlaceComoLlegar(p);
            const etqs = etiquetasDe(p);
            const donde = ubicacion(p);

            return (
              <tr key={p.id} className={p.estado === 'urgente' ? 'fila--urgente' : undefined}>
                <td>
                  <span className={claseSello(p.estado)}>{estadoLabel(p.estado)}</span>
                </td>

                <td>
                  <span className="tabla__nombre">{p.nombre}</span>
                  <span className="tabla__sub">{tipoUI(p.tipo).label}</span>
                  <span className="tabla__sub">
                    {curado ? 'Verificado a mano' : `Agregado · ${p.fuente}`}
                  </span>
                  {requiereRevision(p) && (
                    <span className="tabla__sub">
                      Requiere revisión: {p.politica_alerta?.join('; ')}
                    </span>
                  )}
                </td>

                <td>
                  {donde}
                  {donde && lugar(p) && <br />}
                  {lugar(p)}
                  {p.horario && <span className="tabla__sub">{p.horario}</span>}
                </td>

                {conDistancia && (
                  <td className="tabla__num">{km === null ? '—' : distanciaTexto(km)}</td>
                )}

                <td>
                  {etqs.length > 0 ? (
                    <ul className="punto__etqs">
                      {etqs.slice(0, 4).map((e, i) => (
                        <li key={`${e.clase}-${e.txt}`} className="punto__etq-item">
                          {(i === 0 || etqs[i - 1].clase !== e.clase) && (
                            <span className="punto__k">
                              {e.clase === 'necesita' ? 'Necesita' : 'Ofrece'}{' '}
                            </span>
                          )}
                          <span className="punto__etq">{e.txt}</span>
                        </li>
                      ))}
                      {etqs.length > 4 && (
                        <li className="punto__k">+{etqs.length - 4}</li>
                      )}
                    </ul>
                  ) : (
                    <span className="tabla__desc">{p.descripcion}</span>
                  )}
                </td>

                <td>{p.contacto}</td>

                <td>
                  <a href={p.fuente_url} target="_blank" rel="noopener noreferrer">
                    ver
                  </a>
                  <span className="tabla__sub">{fechaCorta(p.fecha_corte)}</span>
                </td>

                <td>
                  <span className="tabla__acciones">
                    {comoLlegar && (
                      <a
                        className="btn btn--sm"
                        href={comoLlegar}
                        target="_blank"
                        rel="noopener noreferrer"
                        title={`Cómo llegar a ${p.nombre}`}
                      >
                        <Icono nombre="ruta" titulo={`Cómo llegar a ${p.nombre}`} />
                      </a>
                    )}
                    {tel && (
                      <a className="btn btn--sm" href={`tel:+57${tel}`} title="Llamar">
                        <Icono nombre="telefono" titulo={`Llamar a ${p.nombre}`} />
                      </a>
                    )}
                    <a
                      className="btn btn--sm"
                      href={enlaceWhatsApp(p)}
                      target="_blank"
                      rel="noopener noreferrer"
                      title="Compartir por WhatsApp"
                    >
                      <Icono
                        nombre="whatsapp"
                        titulo={`Compartir ${p.nombre} por WhatsApp`}
                      />
                    </a>
                  </span>
                </td>
              </tr>
            );
          })}
        </tbody>
      </table>
    </div>
  );
}
