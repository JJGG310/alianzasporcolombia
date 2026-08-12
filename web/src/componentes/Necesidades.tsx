import { useEffect, useMemo, useState } from 'react';

import { fechaCorta } from '../data/catalogo';
import Icono from './Icono';
import '../estilos/paridad.css';

/** Un pedido concreto, curado a mano. Espejo de public/datos/necesidades.json. */
interface Necesidad {
  lugar: string;
  municipio?: string;
  /** Puede traer varios separados por coma (corredores humanitarios). */
  departamento?: string;
  /** Texto de párrafo: la lista real de insumos, tal como la reportó la fuente. */
  necesita: string;
  fecha?: string;
  /** A veces trae dos URLs y texto alrededor; se extraen todas. */
  fuente?: string;
}

interface ArchivoNecesidades {
  actualizado?: string;
  items?: Necesidad[];
}

const RUTA = '/datos/necesidades.json';

/**
 * Largo a partir del cual `necesita` se recorta. Escogido para que la lista de
 * insumos corta (que es la mayoría) se lea entera sin tocar nada, y solo los
 * párrafos largos de verdad pidan un toque.
 */
const LARGO_MAX = 220;

/**
 * Recorta en el último espacio antes del límite y quita la puntuación que
 * quede colgando, para no producir un «desplegada.…». null si no hace falta.
 */
function recortar(txt: string): string | null {
  if (txt.length <= LARGO_MAX) return null;
  const corte = txt.lastIndexOf(' ', LARGO_MAX);
  const trozo = txt.slice(0, corte > LARGO_MAX / 2 ? corte : LARGO_MAX);
  return `${trozo.replace(/[\s.,;:—-]+$/, '')}…`;
}

const RE_URL = /https?:\/\/[^\s)<>"']+/g;

/**
 * Saca las URLs del campo `fuente`.
 *
 * En el JSON ese campo no siempre es una URL limpia: hay ítems con dos enlaces
 * unidos por «y» y otros con un paréntesis explicativo. El sitio en vivo mete
 * todo eso dentro de un href y produce un enlace roto; acá se extraen y se
 * enlaza cada una por separado.
 */
function urlsDe(fuente?: string): string[] {
  if (!fuente) return [];
  const halladas = fuente.match(RE_URL) ?? [];
  const limpias = halladas.map((u) => u.replace(/[.,;]+$/, ''));
  return [...new Set(limpias)];
}

/** Dominio legible para nombrar el enlace: dice a quién se está citando. */
function dominio(url: string): string {
  try {
    return new URL(url).hostname.replace(/^www\./, '');
  } catch {
    return 'fuente';
  }
}

/** Un ítem puede pertenecer a varios departamentos a la vez. */
function departamentosDe(n: Necesidad): string[] {
  return (n.departamento ?? '')
    .split(/\s*,\s*/)
    .map((d) => d.trim())
    .filter(Boolean);
}

/**
 * Qué se necesita ahora.
 *
 * Pedidos concretos y fechados: una nevera de banco de sangre en Quibdó, tejas
 * en San José del Palmar. Es la lista que convierte «quiero ayudar» en «llevo
 * esto a este lugar».
 *
 * Se filtra por departamento con chips y no se agrupa por encabezados: los
 * pedidos ya vienen ordenados por fecha, y agrupar obligaría a romper ese orden
 * —lo más reciente es lo más útil— para ganar una jerarquía que el filtro ya
 * da. Los chips además dicen de un vistazo cuántos pedidos hay por
 * departamento, cosa que un encabezado de grupo no hace hasta que lo alcanzas.
 *
 * Si el archivo no está o el fetch falla, no renderiza nada.
 */
export default function Necesidades() {
  const [datos, setDatos] = useState<ArchivoNecesidades | null>(null);
  const [error, setError] = useState(false);
  const [depto, setDepto] = useState('');
  const [abiertos, setAbiertos] = useState<Set<string>>(() => new Set());

  useEffect(() => {
    const ac = new AbortController();
    fetch(RUTA, { signal: ac.signal, headers: { Accept: 'application/json' } })
      .then((r) => {
        if (!r.ok) throw new Error(`HTTP ${r.status}`);
        return r.json() as Promise<ArchivoNecesidades>;
      })
      .then(setDatos)
      .catch((e) => {
        if ((e as Error).name !== 'AbortError') setError(true);
      });
    return () => ac.abort();
  }, []);

  const items = useMemo(
    () =>
      (datos?.items ?? []).filter(
        (n): n is Necesidad =>
          !!n && typeof n.lugar === 'string' && typeof n.necesita === 'string',
      ),
    [datos],
  );

  // Departamentos con su conteo, alfabéticos.
  const deptos = useMemo(() => {
    const cuenta = new Map<string, number>();
    for (const n of items) {
      for (const d of departamentosDe(n)) cuenta.set(d, (cuenta.get(d) ?? 0) + 1);
    }
    return [...cuenta.entries()].sort((a, b) => a[0].localeCompare(b[0], 'es'));
  }, [items]);

  const visibles = useMemo(
    () => (depto ? items.filter((n) => departamentosDe(n).includes(depto)) : items),
    [items, depto],
  );

  if (error) return null;

  if (!datos) {
    return (
      <section className="necesidades" aria-labelledby="t-necesidades" aria-busy="true">
        <h2 id="t-necesidades">Qué se necesita ahora</h2>
        <span className="solo-lectores">Cargando los pedidos reportados…</span>
        <div className="esqueleto necesidades__esqueleto" aria-hidden="true" />
        <div className="esqueleto necesidades__esqueleto" aria-hidden="true" />
      </section>
    );
  }

  if (items.length === 0) return null;

  return (
    <section className="necesidades" aria-labelledby="t-necesidades">
      <h2 id="t-necesidades">Qué se necesita ahora</h2>
      <p className="prosa necesidades__intro">
        Pedidos concretos reportados por hospitales, alcaldías y organizaciones de las
        zonas afectadas. Cada uno con su fuente y la fecha del reporte: llama o confirma
        antes de mover una carga.
        {datos.actualizado && (
          <span className="necesidades__corte"> Corte: {datos.actualizado}</span>
        )}
      </p>

      {deptos.length > 1 && (
        <div className="necesidades__chips" role="group" aria-label="Filtrar por departamento">
          <button
            type="button"
            className="chip"
            aria-pressed={depto === ''}
            onClick={() => setDepto('')}
          >
            Todos
            <span className="chip__conteo">{items.length}</span>
          </button>
          {deptos.map(([d, n]) => (
            <button
              key={d}
              type="button"
              className="chip"
              aria-pressed={depto === d}
              onClick={() => setDepto((actual) => (actual === d ? '' : d))}
            >
              {d}
              <span className="chip__conteo">{n}</span>
            </button>
          ))}
        </div>
      )}

      <ul className="necesidades__lista" aria-live="polite">
        {visibles.map((n, i) => {
          const clave = `${n.lugar}|${n.municipio ?? ''}|${i}`;
          const recorte = recortar(n.necesita);
          const abierto = abiertos.has(clave);
          const fuentes = urlsDe(n.fuente);

          return (
            <li key={clave} className="necesidades__item tarjeta">
              <h3 className="necesidades__lugar">{n.lugar}</h3>

              {(n.municipio || n.departamento) && (
                <p className="necesidades__donde">
                  <Icono nombre="ubicacion" />
                  <span>
                    {n.municipio}
                    {n.municipio && n.departamento ? ' · ' : ''}
                    {n.departamento}
                  </span>
                </p>
              )}

              <p className="necesidades__que">
                {recorte && !abierto ? recorte : n.necesita}
              </p>

              {recorte && (
                <button
                  type="button"
                  className="btn btn--sutil btn--sm necesidades__ver"
                  aria-expanded={abierto}
                  onClick={() =>
                    setAbiertos((prev) => {
                      const sig = new Set(prev);
                      if (sig.has(clave)) sig.delete(clave);
                      else sig.add(clave);
                      return sig;
                    })
                  }
                >
                  {abierto ? 'ver menos' : 'ver el pedido completo'}
                </button>
              )}

              <p className="necesidades__meta">
                {n.fecha && <span>Reporte del {fechaCorta(n.fecha)}</span>}
                {fuentes.map((u) => (
                  <a
                    key={u}
                    className="necesidades__fuente"
                    href={u}
                    target="_blank"
                    rel="noopener noreferrer"
                  >
                    <Icono nombre="enlace" />
                    {dominio(u)}
                  </a>
                ))}
              </p>
            </li>
          );
        })}
      </ul>
    </section>
  );
}
