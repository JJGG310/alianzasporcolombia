import { useEffect, useState } from 'react';

import { fechaCorta } from '../data/catalogo';
import Icono from './Icono';
import '../estilos/paridad.css';

/** Un aviso operativo curado a mano. Espejo de public/datos/avisos.json. */
interface Aviso {
  texto: string;
  /** ISO corto (YYYY-MM-DD). */
  fecha?: string;
  /** URL citable. Principio #2: todo lleva fuente y fecha. */
  fuente?: string;
}

interface ArchivoAvisos {
  actualizado?: string;
  items?: Aviso[];
}

const RUTA = '/datos/avisos.json';

/** Cuántos se ven sin tocar nada. Los demás quedan tras «ver los otros N». */
const VISIBLES = 2;

/**
 * Avisos operativos — toque de queda, aeropuertos cerrados, vías.
 *
 * Es lo que cambia la decisión de alguien de salir a la calle esta noche, así
 * que va arriba del todo, antes del mapa y antes de las réplicas.
 *
 * TRATAMIENTO VISUAL, y por qué:
 * NO usa `--urgente`. El rojo es de un punto que necesita algo YA; un toque de
 * queda es contexto, y si el contexto se pinta de rojo, el rojo deja de
 * significar algo. Tampoco es una franja ámbar de borde a borde (el sitio en
 * vivo la tiene y app.css ya dejó dicho por qué no: una barra amarilla a todo
 * lo ancho enseña a ignorar los avisos).
 *
 * Lo que hace es separar las dos capas: el BLOQUE es una superficie de
 * contraste (`--superficie` sobre el `--fondo` cálido de la página, con borde y
 * sombra) —eso es lo que dice «esto es aparte»— y el ÁMBAR (`--confirmar`)
 * queda reducido al rótulo y al icono, del tamaño de un sello. Así el aviso se
 * distingue sin gritar, y el texto —que es lo que hay que leer— se queda con
 * todo el contraste de `--tinta-900`.
 *
 * Si el archivo no está o el fetch falla, el componente no renderiza nada: un
 * error nuestro no puede robarle espacio en pantalla a quien está buscando un
 * albergue.
 */
export default function Avisos() {
  const [items, setItems] = useState<Aviso[] | null>(null);
  const [error, setError] = useState(false);

  useEffect(() => {
    const ac = new AbortController();
    fetch(RUTA, { signal: ac.signal, headers: { Accept: 'application/json' } })
      .then((r) => {
        if (!r.ok) throw new Error(`HTTP ${r.status}`);
        return r.json() as Promise<ArchivoAvisos>;
      })
      .then((d) => {
        const limpios = (d.items ?? []).filter(
          (a): a is Aviso => !!a && typeof a.texto === 'string' && a.texto.trim() !== '',
        );
        setItems(limpios);
      })
      .catch((e) => {
        if ((e as Error).name !== 'AbortError') setError(true);
      });
    return () => ac.abort();
  }, []);

  // Sin datos no hay componente. Nunca un error visible por esto.
  if (error) return null;

  if (!items) {
    return (
      <section className="avisos-op" aria-labelledby="t-avisos" aria-busy="true">
        <h2 id="t-avisos" className="avisos-op__rotulo">
          <Icono nombre="alerta" />
          Avisos operativos
        </h2>
        <span className="solo-lectores">Cargando avisos operativos…</span>
        {/* Reserva el alto de dos renglones para que la página no salte. */}
        <span className="esqueleto avisos-op__esqueleto" aria-hidden="true" />
      </section>
    );
  }

  if (items.length === 0) return null;

  const primeros = items.slice(0, VISIBLES);
  const resto = items.slice(VISIBLES);

  return (
    <section className="avisos-op" aria-labelledby="t-avisos" aria-live="polite">
      <h2 id="t-avisos" className="avisos-op__rotulo">
        <Icono nombre="alerta" />
        Avisos operativos
      </h2>

      <ul className="avisos-op__lista">
        {primeros.map((a) => (
          <ItemAviso key={a.texto} aviso={a} />
        ))}

        {resto.length > 0 && (
          <li>
            <details className="avisos-op__mas">
              <summary className="avisos-op__resorte">
                ver {resto.length === 1 ? 'el otro aviso' : `los otros ${resto.length}`}
                <Icono nombre="flecha-abajo" className="avisos-op__flecha" />
              </summary>
              <ul className="avisos-op__lista avisos-op__lista--anidada">
                {resto.map((a) => (
                  <ItemAviso key={a.texto} aviso={a} />
                ))}
              </ul>
            </details>
          </li>
        )}
      </ul>
    </section>
  );
}

function ItemAviso({ aviso }: { aviso: Aviso }) {
  return (
    <li className="avisos-op__item">
      <p className="avisos-op__texto">{aviso.texto}</p>
      <p className="avisos-op__meta">
        {aviso.fecha && <span>{fechaCorta(aviso.fecha)}</span>}
        {aviso.fuente && (
          <a
            className="avisos-op__fuente"
            href={aviso.fuente}
            target="_blank"
            rel="noopener noreferrer"
          >
            <Icono nombre="enlace" />
            fuente
          </a>
        )}
      </p>
    </li>
  );
}
