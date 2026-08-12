import { useCallback, useMemo, useState, useSyncExternalStore } from 'react';

import Icono from './Icono';
import { esCurado } from '../data/carga';
import {
  distanciaKm,
  estadoLabel,
  normalizarBusqueda,
  textoBuscable,
  tipoUI,
  type Coordenada,
} from '../data/catalogo';
import type { Filtro, Punto } from '../types';

// ---------------------------------------------------------------------------
// Origen de cercanía
//
// PRIVACIDAD: la ubicación vive en una variable de módulo y en ningún lado
// más. No se manda a ningún servidor, no se guarda en localStorage y no
// sobrevive a un recargue de la página. Si algún día alguien necesita
// persistirla, esto es lo que hay que discutir primero.
// ---------------------------------------------------------------------------

let origenActual: Coordenada | null = null;
const oyentes = new Set<() => void>();

function fijarOrigen(o: Coordenada | null): void {
  origenActual = o;
  for (const f of oyentes) f();
}

function suscribir(f: () => void): () => void {
  oyentes.add(f);
  return () => {
    oyentes.delete(f);
  };
}

function leerOrigen(): Coordenada | null {
  return origenActual;
}

/**
 * Dónde está la persona, o null si no lo sabemos (que es lo normal).
 *
 * Lo usan las tarjetas y la tabla para mostrar la distancia sin que App tenga
 * que pasar props, y App puede usarlo para ordenar la lista completa —que es
 * lo único que no se puede resolver desde acá abajo, porque la lista se
 * pagina arriba.
 */
export function useOrigen(): Coordenada | null {
  return useSyncExternalStore(suscribir, leerOrigen, leerOrigen);
}

/** Distancia del punto al origen, o null si al punto le faltan coordenadas. */
export function distanciaDe(p: Punto, origen: Coordenada | null): number | null {
  if (!origen || typeof p.lat !== 'number' || typeof p.lon !== 'number') return null;
  return distanciaKm(origen, { lat: p.lat, lon: p.lon });
}

/**
 * Ordena por cercanía dejando al final —nunca escondidos— los puntos sin
 * coordenadas: hoy son la mayoría, y un albergue sin lat/lon sigue siendo un
 * albergue.
 */
export function ordenarPorCercania(puntos: Punto[], origen: Coordenada | null): Punto[] {
  if (!origen) return puntos;
  const km = new Map<string, number>();
  for (const p of puntos) {
    const d = distanciaDe(p, origen);
    if (d !== null) km.set(p.id, d);
  }
  return [...puntos].sort(
    (a, b) => (km.get(a.id) ?? Infinity) - (km.get(b.id) ?? Infinity),
  );
}

// ---------------------------------------------------------------------------

type EstadoUbicacion = 'inactiva' | 'pidiendo' | 'lista' | 'negada' | 'error';

/** Los tipos que la gente busca de verdad, en el orden en que los busca. */
const CHIPS: { tipo: string; label: string }[] = [
  { tipo: 'albergue', label: 'Albergue' },
  { tipo: 'acopio', label: 'Acopio' },
  { tipo: 'agua', label: 'Agua' },
  { tipo: 'salud', label: 'Salud' },
  { tipo: 'sangre', label: 'Banco de sangre' },
  { tipo: 'olla', label: 'Olla comunitaria' },
  { tipo: 'rescate', label: 'Rescate' },
];

interface Props {
  filtro: Filtro;
  cambiar: (parcial: Partial<Filtro>) => void;
  limpiar: () => void;
  /** Opciones derivadas de los datos reales, no de la lista canónica completa. */
  tipos: string[];
  departamentos: string[];
  municipios: string[];
  estados: string[];
  /** Cuántos puntos hay escondidos por traer alerta de política. */
  totalEnRevision: number;
  mostrados: number;
  total: number;
  /**
   * NUEVA, opcional: el universo sobre el que App filtra. Sin ella los chips
   * funcionan pero no pueden mostrar su conteo, y saber que "Albergue (32)"
   * tiene 32 antes de tocarlo es justo lo que evita el callejón sin salida.
   */
  puntos?: Punto[];
}

export default function Filtros({
  filtro,
  cambiar,
  limpiar,
  tipos,
  departamentos,
  municipios,
  estados,
  totalEnRevision,
  mostrados,
  total,
  puntos,
}: Props) {
  const [ubicacion, setUbicacion] = useState<EstadoUbicacion>('inactiva');
  const origen = useOrigen();

  const hayFiltro =
    filtro.texto !== '' ||
    filtro.tipo !== '' ||
    filtro.departamento !== '' ||
    filtro.municipio !== '' ||
    filtro.estado !== '' ||
    filtro.soloVerificados ||
    filtro.requiereRevision;

  const finosActivos =
    (filtro.departamento ? 1 : 0) +
    (filtro.municipio ? 1 : 0) +
    (filtro.estado ? 1 : 0) +
    (filtro.soloVerificados ? 1 : 0) +
    (filtro.requiereRevision ? 1 : 0);

  // --- Conteos de los chips ------------------------------------------------
  // Se cuenta sobre el universo con TODOS los filtros aplicados menos el tipo:
  // el número del chip contesta "cuántos me quedan si toco esto", que es la
  // pregunta que evita el callejón sin salida.

  const indice = useMemo(() => {
    const m = new Map<string, string>();
    if (puntos) for (const p of puntos) m.set(p.id, textoBuscable(p));
    return m;
  }, [puntos]);

  const sinTipo = useMemo(() => {
    if (!puntos) return null;
    const terminos = normalizarBusqueda(filtro.texto).split(/\s+/).filter(Boolean);
    return puntos.filter((p) => {
      if (filtro.departamento && p.departamento !== filtro.departamento) return false;
      if (filtro.municipio && p.municipio !== filtro.municipio) return false;
      if (filtro.estado && (p.estado ?? 'por-confirmar') !== filtro.estado) return false;
      if (filtro.soloVerificados && !esCurado(p)) return false;
      if (terminos.length > 0) {
        const txt = indice.get(p.id) ?? '';
        if (!terminos.every((t) => txt.includes(t))) return false;
      }
      return true;
    });
  }, [
    puntos,
    indice,
    filtro.texto,
    filtro.departamento,
    filtro.municipio,
    filtro.estado,
    filtro.soloVerificados,
  ]);

  const conteos = useMemo(() => {
    const m = new Map<string, number>();
    if (sinTipo) for (const p of sinTipo) m.set(p.tipo, (m.get(p.tipo) ?? 0) + 1);
    return m;
  }, [sinTipo]);

  const urgentes = useMemo(() => {
    if (!sinTipo) return null;
    const lista = filtro.tipo ? sinTipo.filter((p) => p.tipo === filtro.tipo) : sinTipo;
    return lista.filter((p) => p.estado === 'urgente').length;
  }, [sinTipo, filtro.tipo]);

  // --- Ubicación -----------------------------------------------------------

  const pedirUbicacion = useCallback(() => {
    if (origen) {
      // Segundo toque: se apaga y se olvida.
      fijarOrigen(null);
      setUbicacion('inactiva');
      return;
    }
    if (typeof navigator === 'undefined' || !navigator.geolocation) {
      setUbicacion('error');
      return;
    }
    setUbicacion('pidiendo');
    navigator.geolocation.getCurrentPosition(
      (pos) => {
        fijarOrigen({ lat: pos.coords.latitude, lon: pos.coords.longitude });
        setUbicacion('lista');
      },
      (err) => {
        fijarOrigen(null);
        setUbicacion(err.code === err.PERMISSION_DENIED ? 'negada' : 'error');
      },
      // Sin alta precisión: pide menos batería y responde más rápido, y para
      // "cuál albergue me queda más cerca" 100 m de error no cambian nada.
      { enableHighAccuracy: false, timeout: 10000, maximumAge: 5 * 60 * 1000 },
    );
  }, [origen]);

  const cercaActiva = origen !== null;

  return (
    <div className="busqueda">
      <form
        role="search"
        aria-label="Buscar puntos de ayuda"
        onSubmit={(e) => e.preventDefault()}
      >
        <div className="buscador">
          <label className="solo-lectores" htmlFor="f-texto">
            Buscar por nombre, barrio, municipio o qué necesita
          </label>
          <Icono nombre="buscar" className="buscador__lupa" tam="1.25em" />
          <input
            id="f-texto"
            className="campo buscador__campo"
            type="search"
            autoComplete="off"
            placeholder="Buscar albergue, barrio, Cali, agua…"
            value={filtro.texto}
            onChange={(e) => cambiar({ texto: e.target.value })}
          />
          {filtro.texto !== '' && (
            <button
              type="button"
              className="buscador__limpiar"
              onClick={() => cambiar({ texto: '' })}
            >
              <Icono nombre="cerrar" titulo="Borrar la búsqueda" />
            </button>
          )}
        </div>
      </form>

      <div className="chips" role="group" aria-label="Filtros rápidos">
        <button
          type="button"
          className="chip chip--cerca"
          aria-pressed={cercaActiva}
          onClick={pedirUbicacion}
        >
          <Icono nombre="ubicacion" />
          {ubicacion === 'pidiendo' ? 'Buscando tu ubicación…' : 'Cerca de mí'}
        </button>

        {CHIPS.filter((c) => tipos.includes(c.tipo)).map((c) => {
          const activo = filtro.tipo === c.tipo;
          // Sin la prop `puntos` no hay conteo (undefined ≠ 0): un chip sin
          // número es un chip sin dato, no un chip vacío.
          const n = sinTipo ? (conteos.get(c.tipo) ?? 0) : undefined;
          return (
            <button
              key={c.tipo}
              type="button"
              className="chip"
              aria-pressed={activo}
              disabled={n === 0 && !activo}
              onClick={() => cambiar({ tipo: activo ? '' : c.tipo })}
            >
              {c.label}
              {n !== undefined && <span className="chip__conteo">{n}</span>}
            </button>
          );
        })}
      </div>

      {ubicacion !== 'inactiva' && (
        <p
          className={`ubic-aviso${
            ubicacion === 'negada' || ubicacion === 'error' ? ' ubic-aviso--problema' : ''
          }`}
          role="status"
        >
          <Icono
            nombre={ubicacion === 'lista' || ubicacion === 'pidiendo' ? 'ubicacion' : 'alerta'}
          />
          {ubicacion === 'pidiendo' && 'Buscando tu ubicación…'}
          {ubicacion === 'lista' &&
            'Ordenado por cercanía. Tu ubicación se queda en este teléfono: no la enviamos ni la guardamos.'}
          {ubicacion === 'negada' &&
            'Sin permiso de ubicación no podemos ordenar por cercanía. Puedes activarlo en el navegador, o buscar por barrio o municipio.'}
          {ubicacion === 'error' &&
            'No pudimos obtener tu ubicación en este momento. Prueba de nuevo o busca por barrio o municipio.'}
        </p>
      )}

      <details className="mas-filtros">
        <summary>
          <Icono nombre="filtro" />
          Más filtros
          {finosActivos > 0 && (
            <span className="mas-filtros__conteo">
              {finosActivos}
              <span className="solo-lectores"> filtros activos</span>
            </span>
          )}
          <Icono nombre="flecha-abajo" className="mas-filtros__flecha" />
        </summary>

        <div className="mas-filtros__cuerpo">
          <div>
            <label className="etiqueta" htmlFor="f-depto">
              Departamento
            </label>
            <select
              id="f-depto"
              className="campo"
              value={filtro.departamento}
              onChange={(e) => cambiar({ departamento: e.target.value, municipio: '' })}
            >
              <option value="">Todos</option>
              {departamentos.map((d) => (
                <option key={d} value={d}>
                  {d}
                </option>
              ))}
            </select>
          </div>

          <div>
            <label className="etiqueta" htmlFor="f-municipio">
              Municipio
            </label>
            <select
              id="f-municipio"
              className="campo"
              value={filtro.municipio}
              onChange={(e) => cambiar({ municipio: e.target.value })}
            >
              <option value="">Todos</option>
              {municipios.map((m) => (
                <option key={m} value={m}>
                  {m}
                </option>
              ))}
            </select>
          </div>

          <div>
            <label className="etiqueta" htmlFor="f-tipo">
              Tipo
            </label>
            <select
              id="f-tipo"
              className="campo"
              value={filtro.tipo}
              onChange={(e) => cambiar({ tipo: e.target.value })}
            >
              <option value="">Todos</option>
              {tipos.map((t) => (
                <option key={t} value={t}>
                  {tipoUI(t).label}
                </option>
              ))}
            </select>
          </div>

          <div>
            <label className="etiqueta" htmlFor="f-estado">
              Estado
            </label>
            <select
              id="f-estado"
              className="campo"
              value={filtro.estado}
              onChange={(e) => cambiar({ estado: e.target.value })}
            >
              <option value="">Todos</option>
              {estados.map((e) => (
                <option key={e} value={e}>
                  {estadoLabel(e)}
                </option>
              ))}
            </select>
          </div>

          <label className="casilla" htmlFor="f-verificados">
            <input
              id="f-verificados"
              type="checkbox"
              checked={filtro.soloVerificados}
              onChange={(e) => cambiar({ soloVerificados: e.target.checked })}
            />
            Solo verificados a mano
          </label>

          {totalEnRevision > 0 && (
            <label
              className="casilla"
              htmlFor="f-revision"
              title="Puntos donde el saneamiento automático detectó algo que un humano debe revisar (por ejemplo, posibles datos de pago). Están fuera de la vista pública."
            >
              <input
                id="f-revision"
                type="checkbox"
                checked={filtro.requiereRevision}
                onChange={(e) => cambiar({ requiereRevision: e.target.checked })}
              />
              Ver los que requieren revisión ({totalEnRevision})
            </label>
          )}
        </div>
      </details>

      <p className="resumen" aria-live="polite">
        <span>
          <strong>{mostrados.toLocaleString('es-CO')}</strong>{' '}
          {mostrados === 1 ? 'punto' : 'puntos'}
          {hayFiltro && total !== mostrados && ` de ${total.toLocaleString('es-CO')}`}
        </span>
        {urgentes !== null && urgentes > 0 && (
          <span className="resumen__urgentes">
            {urgentes} {urgentes === 1 ? 'urgente' : 'urgentes'}
          </span>
        )}
        {filtro.requiereRevision && (
          <span className="resumen__aviso">
            Estás viendo la cola de revisión: estos datos NO están verificados.
          </span>
        )}
        {hayFiltro && (
          <button type="button" className="btn btn--sutil btn--sm resumen__limpiar" onClick={limpiar}>
            Limpiar filtros
          </button>
        )}
      </p>

      {mostrados === 0 && (
        <SinResultados filtro={filtro} cambiar={cambiar} limpiar={limpiar} />
      )}
    </div>
  );
}

// ---------------------------------------------------------------------------

/**
 * Estado vacío con salida.
 *
 * "0 resultados" deja a la persona parada. Acá se nombra el filtro que está
 * estorbando —el más específico primero— y se ofrece quitarlo de un toque.
 */
function SinResultados({
  filtro,
  cambiar,
  limpiar,
}: {
  filtro: Filtro;
  cambiar: (parcial: Partial<Filtro>) => void;
  limpiar: () => void;
}) {
  const donde = filtro.municipio || filtro.departamento;

  // El título nombra TODO lo que está estorbando —"Ningún albergue en Quibdó
  // que coincida con «agua»"— porque quitar el filtro equivocado y seguir en
  // cero es peor que no haber intentado.
  const sujeto = filtro.tipo ? tipoUI(filtro.tipo).label.toLowerCase() : 'punto';
  const estado = filtro.estado ? ` ${estadoLabel(filtro.estado).toLowerCase()}` : '';
  const enDonde = donde ? ` en ${donde}` : '';
  const verificado = filtro.soloVerificados ? ' verificado a mano' : '';
  const conTexto = filtro.texto ? ` que coincida con «${filtro.texto}»` : '';
  const titulo = `Ningún ${sujeto}${estado}${verificado}${enDonde}${conTexto}.`;

  // La salida quita un solo filtro: el más estrecho primero.
  let salida: { texto: string; accion: () => void } | null = null;

  if (filtro.texto) {
    salida = { texto: 'Quitar la búsqueda', accion: () => cambiar({ texto: '' }) };
  } else if (filtro.tipo) {
    salida = {
      texto: donde ? `Ver todos los tipos en ${donde}` : 'Ver todos los tipos',
      accion: () => cambiar({ tipo: '' }),
    };
  } else if (filtro.municipio) {
    salida = {
      texto: filtro.departamento
        ? `Ver todo ${filtro.departamento}`
        : 'Ver todos los municipios',
      accion: () => cambiar({ municipio: '' }),
    };
  } else if (filtro.estado) {
    salida = { texto: 'Ver todos los estados', accion: () => cambiar({ estado: '' }) };
  } else if (filtro.soloVerificados) {
    salida = {
      texto: 'Incluir los agregados automáticamente',
      accion: () => cambiar({ soloVerificados: false }),
    };
  }

  return (
    <div className="vacio" role="status">
      <p className="vacio__titulo">{titulo}</p>
      <p className="vacio__nota">
        Los datos vienen de sitios ciudadanos y se quedan cortos en muchos
        municipios. Si buscas algo puntual, prueba con una palabra sola: el barrio,
        el municipio o lo que necesitas.
      </p>
      <div className="vacio__acciones">
        {salida && (
          <button type="button" className="btn btn--fuerte" onClick={salida.accion}>
            {salida.texto}
          </button>
        )}
        <button type="button" className="btn" onClick={limpiar}>
          Limpiar filtros
        </button>
      </div>
    </div>
  );
}
