import { useEffect, useState } from 'react';

import Icono from './Icono';
import '../estilos/paridad.css';

/** Una métrica del balance oficial. Espejo de public/datos/cifras.json. */
interface Metrica {
  nombre: string;
  valor: string;
  /** Desglose o contexto (p. ej. fallecidos por departamento). */
  detalle?: string;
  /** Quién lo dijo y cuándo. Principio #2: sin fuente y hora, la cifra no va. */
  fuente: string;
  url: string;
}

interface ArchivoCifras {
  actualizado?: string;
  nota?: string;
  metricas?: Metrica[];
}

const RUTA = '/datos/cifras.json';

/**
 * Cifras oficiales — el balance con fuente y hora en cada tarjeta.
 *
 * Nadie más en el ecosistema hace esto: los medios publican «239 muertos» a
 * secas y cada uno trae un número distinto, porque consolidan a ritmos
 * distintos. Aquí cada número dice QUIÉN lo dijo y CUÁNDO, y la pareja
 * desaparecidos-oficiales / reportes-ciudadanos va explicada en el detalle —
 * la brecha entre 287 y 4.155 asusta solo si nadie te cuenta que cuentan
 * cosas distintas.
 *
 * El valor va grande y el nombre debajo, como en un tablero: quien entra con
 * afán lee los números; quien duda de un número tiene la fuente a un toque.
 */
export default function Cifras() {
  const [archivo, setArchivo] = useState<ArchivoCifras | null>(null);
  const [error, setError] = useState(false);

  useEffect(() => {
    const ac = new AbortController();
    fetch(RUTA, { cache: 'no-store', signal: ac.signal })
      .then((r) => {
        if (!r.ok) throw new Error(`HTTP ${r.status}`);
        return r.json() as Promise<ArchivoCifras>;
      })
      .then(setArchivo)
      .catch((e) => {
        if ((e as Error).name !== 'AbortError') setError(true);
      });
    return () => ac.abort();
  }, []);

  // Sin archivo no hay sección: mejor ausente que con datos viejos inventados.
  if (error) return null;

  if (!archivo) {
    return (
      <section className="cifras" aria-labelledby="t-cifras" aria-busy="true">
        <h2 id="t-cifras" className="cifras__titulo">
          Cifras oficiales
        </h2>
        <div className="cifras__rejilla" aria-hidden="true">
          <span className="esqueleto cifras__cargando" />
          <span className="esqueleto cifras__cargando" />
          <span className="esqueleto cifras__cargando" />
        </div>
      </section>
    );
  }

  const metricas = archivo.metricas ?? [];
  if (!metricas.length) return null;

  return (
    <section className="cifras" aria-labelledby="t-cifras">
      <h2 id="t-cifras" className="cifras__titulo">
        Cifras oficiales
      </h2>
      {archivo.nota && (
        <p className="cifras__nota">
          {archivo.nota}
          {archivo.actualizado ? ` Corte: ${archivo.actualizado}.` : ''}
        </p>
      )}
      <div className="cifras__rejilla">
        {metricas.map((m) => (
          <article key={m.nombre} className="cifras__tarjeta">
            <strong className="cifras__valor">{m.valor}</strong>
            <span className="cifras__nombre">{m.nombre}</span>
            {m.detalle && <p className="cifras__detalle">{m.detalle}</p>}
            <p className="cifras__fuente">
              {m.fuente} ·{' '}
              <a href={m.url} target="_blank" rel="noopener noreferrer">
                fuente <Icono nombre="enlace" tam="0.8em" />
              </a>
            </p>
          </article>
        ))}
      </div>
    </section>
  );
}
