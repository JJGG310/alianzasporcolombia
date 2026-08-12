import { memo, useState } from 'react';

import Icono from './Icono';
import { useOrigen, distanciaDe } from './Filtros';
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

/** Estado del punto → sello del sistema de diseño. */
export function claseSello(estado?: string): string {
  switch (estado ?? 'por-confirmar') {
    case 'urgente':
      return 'sello sello--urgente';
    case 'activo':
      return 'sello sello--activo';
    case 'cubierto':
      return 'sello sello--cubierto';
    case 'cerrado':
    case 'descartado':
      return 'sello sello--cerrado';
    default:
      return 'sello sello--confirmar';
  }
}

/**
 * El primer celular colombiano que aparezca en un campo de texto libre.
 *
 * `contacto` trae de todo: "Cruz Roja - Marta 3001234567 / (2) 885 0001".
 * Acá solo interesa lo que se puede marcar de un toque, así que se exige el
 * formato real de un móvil: 10 dígitos empezando por 3, con o sin el +57.
 * Si no hay ninguno, no se inventa un botón que no va a servir.
 */
export function telefonoDe(txt?: string): string | null {
  if (!txt) return null;
  // Sin lookbehind a propósito: hay iPhones viejos allá afuera que no la
  // soportan y reventarían el módulo entero al parsear el archivo.
  const re = /(^|\D)((?:\+?57[ .-]?)?3\d{2}[ .-]?\d{3}[ .-]?\d{4})(?!\d)/g;
  let m: RegExpExecArray | null;
  while ((m = re.exec(txt)) !== null) {
    let d = m[2].replace(/\D/g, '');
    if (d.length === 12 && d.startsWith('57')) d = d.slice(2);
    if (d.length === 10 && d.startsWith('3')) return d;
  }
  return null;
}

/** Cómo llegar: con coordenadas si las hay, con la dirección si no. */
export function enlaceComoLlegar(p: Punto): string | null {
  if (typeof p.lat === 'number' && typeof p.lon === 'number') {
    return `https://www.google.com/maps/dir/?api=1&destination=${p.lat},${p.lon}`;
  }
  const consulta = [p.direccion, p.municipio, p.departamento].filter(Boolean).join(', ');
  if (!consulta) return null;
  return `https://www.google.com/maps/search/?api=1&query=${encodeURIComponent(consulta)}`;
}

/** Vocabulario cerrado de `recibe`/`no_recibe` en los puntos curados. */
const CATEGORIA: Record<string, string> = {
  agua: 'Agua',
  alimentos: 'Alimentos',
  aseo: 'Aseo e higiene',
  panales: 'Pañales',
  abrigo: 'Cobijas y colchonetas',
  ropa: 'Ropa',
  medicamentos: 'Medicamentos',
  mascotas: 'Comida de mascotas',
  sangre: 'Sangre',
  voluntarios: 'Voluntarios',
  rescate: 'Rescate y herramientas',
  bebes: 'Cosas de bebé',
  ortopedico: 'Ayudas ortopédicas',
};

export type ClaseEtq = 'necesita' | 'ofrece' | 'recibe' | 'no-recibe';

/**
 * Etiquetas de la ficha. `no-recibe` va de último pero es la que más ahorra:
 * evita que alguien cargue media cuadra con ropa a un punto que solo pide agua.
 */
export function etiquetasDe(p: Punto): { clase: ClaseEtq; txt: string }[] {
  return [
    ...(p.necesita ?? []).map((t) => ({ clase: 'necesita' as const, txt: t })),
    ...(p.ofrece ?? []).map((t) => ({ clase: 'ofrece' as const, txt: t })),
    ...(p.recibe ?? []).map((t) => ({ clase: 'recibe' as const, txt: CATEGORIA[t] ?? t })),
    ...(p.no_recibe ?? []).map((t) => ({ clase: 'no-recibe' as const, txt: CATEGORIA[t] ?? t })),
  ];
}

const ROTULO: Record<ClaseEtq, string> = {
  necesita: 'Necesita',
  ofrece: 'Ofrece',
  recibe: 'Recibe',
  'no-recibe': 'NO recibe',
};

/** Cuántas etiquetas se muestran antes de doblar el resto. */
const TOPE_ETIQUETAS = 4;

interface Props {
  punto: Punto;
  /** Origen para la distancia. Por defecto, el que compartió "Cerca de mí". */
  origen?: Coordenada | null;
}

/**
 * La unidad de información del sitio: lo que la gente realmente lee.
 *
 * Orden: qué tan grave está → cómo se llama → dónde queda (y a cuánto) → qué
 * necesita → cómo se contacta → qué hago con esto. Y al pie, siempre, de dónde
 * salió el dato y de cuándo es (principio #2 del proyecto).
 */
function TarjetaPunto({ punto: p, origen }: Props) {
  const [expandido, setExpandido] = useState(false);
  const origenStore = useOrigen();
  const desde = origen ?? origenStore;

  const t = tipoUI(p.tipo);
  const curado = esCurado(p);
  const enRevision = requiereRevision(p);
  const urgente = p.estado === 'urgente';
  const donde = [ubicacion(p), lugar(p)].filter(Boolean).join(' · ');
  const km = distanciaDe(p, desde);
  const tel = telefonoDe(p.contacto);
  const comoLlegar = enlaceComoLlegar(p);

  const etqs = etiquetasDe(p);
  // Lo que el punto rechaza nunca se poda: si queda detrás del "+N más" la
  // persona igual carga la donación hasta allá y se la devuelven.
  const rechaza = etqs.filter((e) => e.clase === 'no-recibe');
  const resto = etqs.filter((e) => e.clase !== 'no-recibe');
  const restoVisible = expandido ? resto : resto.slice(0, TOPE_ETIQUETAS);
  const visibles = [...restoVisible, ...rechaza];
  const ocultas = resto.length - restoVisible.length;

  // <article> y no <li>: la lista (<ul>) la arma App, que es quien pagina.
  // Así la tarjeta también sirve suelta —en el popup del mapa, por ejemplo—
  // sin dejar un <li> huérfano fuera de una lista.
  return (
    <article className={`tarjeta punto${urgente ? ' punto--urgente' : ''}`}>
      <div className="punto__sellos">
        <span className={claseSello(p.estado)}>{estadoLabel(p.estado)}</span>
        <span
          className={curado ? 'sello sello--verificado' : 'sello sello--agregado'}
          title={
            curado
              ? 'Revisado a mano por el equipo de Alianzas por Colombia.'
              : 'Tomado automáticamente de otro sitio ciudadano. Confirma antes de ir.'
          }
        >
          {curado ? 'Verificado a mano' : 'Agregado automáticamente'}
        </span>
        <span className="punto__tipo">{t.label}</span>
      </div>

      <h3 className="punto__nombre">{p.nombre}</h3>

      {donde && (
        <p className="punto__donde">
          {donde}
          {km !== null && (
            <>
              {' · '}
              <span className="punto__dist">
                <Icono nombre="ubicacion" tam="1em" />a {distanciaTexto(km)}
              </span>
            </>
          )}
        </p>
      )}
      {!donde && km !== null && (
        <p className="punto__donde">
          <span className="punto__dist">
            <Icono nombre="ubicacion" tam="1em" />a {distanciaTexto(km)}
          </span>
        </p>
      )}

      {enRevision && (
        <p className="punto__alerta">
          <Icono nombre="alerta" />
          <span>
            Requiere revisión humana: {p.politica_alerta?.join('; ')}. No lo compartas
            hasta confirmarlo.
          </span>
        </p>
      )}

      {p.descripcion && <p className="punto__desc">{p.descripcion}</p>}

      {etqs.length > 0 && (
        <ul className="punto__etqs">
          {visibles.map((e, i) => (
            <li key={`${e.clase}-${e.txt}`} className="punto__etq-item">
              {(i === 0 || visibles[i - 1].clase !== e.clase) && (
                <span className="punto__k">{ROTULO[e.clase]} </span>
              )}
              <span className={`punto__etq punto__etq--${e.clase}`}>{e.txt}</span>
            </li>
          ))}
          {ocultas > 0 && (
            <li>
              <button
                type="button"
                className="punto__mas"
                onClick={() => setExpandido(true)}
              >
                +{ocultas} más
              </button>
            </li>
          )}
        </ul>
      )}

      {(p.horario || p.contacto || p.organizacion) && (
        <p className="punto__meta">
          {p.horario && <span>{p.horario}</span>}
          {p.contacto && (
            <span>
              <Icono nombre="telefono" tam="1em" />
              {p.contacto}
            </span>
          )}
          {p.organizacion && <span>{p.organizacion}</span>}
        </p>
      )}

      <div className="punto__acciones">
        {comoLlegar && (
          <a
            className="btn btn--fuerte"
            href={comoLlegar}
            target="_blank"
            rel="noopener noreferrer"
          >
            <Icono nombre="ruta" />
            Cómo llegar
          </a>
        )}
        {tel && (
          <a className="btn" href={`tel:+57${tel}`}>
            <Icono nombre="telefono" />
            Llamar
          </a>
        )}
        <a
          className="btn"
          href={enlaceWhatsApp(p)}
          target="_blank"
          rel="noopener noreferrer"
          aria-label={`Compartir ${p.nombre} por WhatsApp`}
        >
          <Icono nombre="whatsapp" />
          Compartir
        </a>
      </div>

      <p className="punto__pie">
        <a href={p.fuente_url} target="_blank" rel="noopener noreferrer">
          <Icono nombre="enlace" tam="1em" />
          Fuente
        </a>
        <span>datos al {fechaCorta(p.fecha_corte)}</span>
        {!curado && p.fuente && <span>vía {p.fuente}</span>}
      </p>
    </article>
  );
}

export default memo(TarjetaPunto);

/**
 * Esqueletos con la forma de la tarjeta.
 *
 * Un spinner dice "espera"; esto dice "va a llegar una lista de sitios" y la
 * pantalla no salta cuando llega. Se exporta para que App lo use en vez del
 * "Cargando puntos de ayuda…".
 */
export function ListaEsqueleto({ cuantos = 6 }: { cuantos?: number }) {
  return (
    <ul className="esq-lista" aria-hidden="true">
      {Array.from({ length: cuantos }, (_, i) => (
        <li key={i} className="tarjeta esq-tarjeta">
          <span className="esqueleto esq-linea esq-linea--sello" />
          <span className="esqueleto esq-linea esq-linea--nombre" />
          <span className="esqueleto esq-linea" />
          <span className="esqueleto esq-linea esq-linea--corta" />
          <span className="esqueleto esq-linea esq-linea--accion" />
        </li>
      ))}
    </ul>
  );
}
