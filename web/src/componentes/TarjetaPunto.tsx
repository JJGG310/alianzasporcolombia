import type { Punto } from '../types';
import { esCurado, requiereRevision } from '../data/carga';
import {
  enlaceWhatsApp,
  estadoClase,
  estadoLabel,
  fechaCorta,
  lugar,
  tipoUI,
  ubicacion,
} from '../data/catalogo';

interface Props {
  punto: Punto;
}

/** Vista principal en móvil. Cada tarjeta lleva SIEMPRE fuente + fecha de corte
 *  + estado (principio #2) y el distintivo verificado / agregado. */
export default function TarjetaPunto({ punto: p }: Props) {
  const t = tipoUI(p.tipo);
  const curado = esCurado(p);
  const enRevision = requiereRevision(p);
  const donde = ubicacion(p);

  return (
    <li
      className={`tarjeta ${curado ? 'es-curado' : 'es-agregado'}${enRevision ? ' en-revision' : ''}`}
    >
      <div className="tarjeta-sellos">
        <span className="pill" style={{ background: `${t.color}18`, color: t.color }}>
          {t.label}
        </span>
        <span className={estadoClase(p.estado)}>{estadoLabel(p.estado)}</span>
        <span
          className={curado ? 'sello sello-curado' : 'sello sello-agregado'}
          title={
            curado
              ? 'Revisado a mano por el equipo de Alianzas por Colombia.'
              : 'Tomado automáticamente de otro sitio ciudadano. Confirma antes de ir.'
          }
        >
          {curado ? '✓ verificado a mano' : '⟳ agregado automáticamente'}
        </span>
      </div>

      <h3>{p.nombre}</h3>

      {enRevision && (
        <p className="alerta-revision">
          ⚠️ Requiere revisión humana: {p.politica_alerta?.join('; ')}. No lo compartas
          hasta confirmarlo.
        </p>
      )}

      {p.descripcion && <p className="cuerpo">{p.descripcion}</p>}

      <dl>
        {donde && (
          <>
            <dt>Dónde</dt>
            <dd>{donde}</dd>
          </>
        )}
        {lugar(p) && (
          <>
            <dt>Municipio</dt>
            <dd>{lugar(p)}</dd>
          </>
        )}
        {p.horario && (
          <>
            <dt>Horario</dt>
            <dd>{p.horario}</dd>
          </>
        )}
        {p.contacto && (
          <>
            <dt>Contacto</dt>
            <dd>{p.contacto}</dd>
          </>
        )}
        {p.organizacion && (
          <>
            <dt>Organiza</dt>
            <dd>{p.organizacion}</dd>
          </>
        )}
        {typeof p.voluntarios_faltan === 'number' && p.voluntarios_faltan > 0 && (
          <>
            <dt>Voluntarios</dt>
            <dd>faltan {p.voluntarios_faltan}</dd>
          </>
        )}
      </dl>

      {(p.necesita?.length || p.ofrece?.length) && (
        <ul className="etiquetas">
          {p.necesita?.map((n) => (
            <li key={`n-${n}`}>necesita: {n}</li>
          ))}
          {p.ofrece?.map((o) => (
            <li key={`o-${o}`}>ofrece: {o}</li>
          ))}
        </ul>
      )}

      <div className="fuente">
        <span>
          <a href={p.fuente_url} target="_blank" rel="noopener noreferrer">
            Fuente
          </a>{' '}
          · datos al {fechaCorta(p.fecha_corte)}
          {!curado && p.fuente ? ` · vía ${p.fuente}` : ''}
        </span>
        <a
          className="boton boton-wa"
          href={enlaceWhatsApp(p)}
          target="_blank"
          rel="noopener noreferrer"
          aria-label={`Compartir ${p.nombre} por WhatsApp`}
        >
          Compartir por WhatsApp
        </a>
      </div>
    </li>
  );
}
