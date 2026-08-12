import { Suspense, lazy, useCallback, useEffect, useMemo, useState } from 'react';

import Directorio from './componentes/Directorio';
import Encabezado from './componentes/Encabezado';
import Filtros from './componentes/Filtros';
import LineasEmergencia from './componentes/LineasEmergencia';
import PieDePagina from './componentes/PieDePagina';
import Replicas from './componentes/Replicas';
import TablaPuntos from './componentes/TablaPuntos';
import TarjetaPunto from './componentes/TarjetaPunto';

import { cargarDatos, esCurado } from './data/carga';
import { fechaCorta, haceCuanto, normalizarBusqueda, textoBuscable } from './data/catalogo';
import { FILTRO_VACIO, type DatosApp, type Filtro, type Punto } from './types';

// Leaflet pesa: se carga aparte para que el contenido crítico (líneas de
// emergencia, listado) pinte primero en una red mala.
const MapaAyuda = lazy(() => import('./componentes/MapaAyuda'));

/** Cuántos puntos se pintan de entrada. El resto, bajo demanda. */
const PAGINA = 60;

function useAnchoGrande(): boolean {
  const consulta = '(min-width: 900px)';
  const [grande, setGrande] = useState(
    () => typeof window !== 'undefined' && window.matchMedia(consulta).matches,
  );
  useEffect(() => {
    const mq = window.matchMedia(consulta);
    const alCambiar = (e: MediaQueryListEvent) => setGrande(e.matches);
    mq.addEventListener('change', alCambiar);
    return () => mq.removeEventListener('change', alCambiar);
  }, []);
  return grande;
}

export default function App() {
  const [datos, setDatos] = useState<DatosApp | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [filtro, setFiltro] = useState<Filtro>(FILTRO_VACIO);
  const [visibles, setVisibles] = useState(PAGINA);
  const anchoGrande = useAnchoGrande();

  useEffect(() => {
    const ac = new AbortController();
    cargarDatos(ac.signal)
      .then(setDatos)
      .catch((e: unknown) => {
        if ((e as Error).name === 'AbortError') return;
        setError((e as Error).message);
      });
    return () => ac.abort();
  }, []);

  const cambiar = useCallback((parcial: Partial<Filtro>) => {
    setFiltro((f) => ({ ...f, ...parcial }));
    setVisibles(PAGINA);
  }, []);

  const limpiar = useCallback(() => {
    setFiltro(FILTRO_VACIO);
    setVisibles(PAGINA);
  }, []);

  // El universo cambia si el usuario pide ver la cola de revisión.
  const universo: Punto[] = useMemo(() => {
    if (!datos) return [];
    return filtro.requiereRevision ? datos.enRevision : datos.puntos;
  }, [datos, filtro.requiereRevision]);

  // Índice de búsqueda: se calcula una vez por conjunto, no por tecla.
  const indice = useMemo(() => {
    const m = new Map<string, string>();
    for (const p of universo) m.set(p.id, textoBuscable(p));
    return m;
  }, [universo]);

  const filtrados = useMemo(() => {
    // Búsqueda por palabras sueltas y en cualquier orden: quien escribe
    // "sangre cali" espera los bancos de sangre de Cali, no una frase exacta.
    const terminos = normalizarBusqueda(filtro.texto).split(/\s+/).filter(Boolean);
    return universo.filter((p) => {
      if (filtro.tipo && p.tipo !== filtro.tipo) return false;
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
  }, [universo, filtro, indice]);

  const opciones = useMemo(() => {
    const tipos = new Set<string>();
    const departamentos = new Set<string>();
    const municipios = new Set<string>();
    const estados = new Set<string>();
    for (const p of universo) {
      tipos.add(p.tipo);
      if (p.departamento) departamentos.add(p.departamento);
      if (p.estado) estados.add(p.estado);
      // Los municipios se acotan al departamento elegido, como en el sitio viejo.
      if (p.municipio && (!filtro.departamento || p.departamento === filtro.departamento)) {
        municipios.add(p.municipio);
      }
    }
    const orden = (a: string, b: string) => a.localeCompare(b, 'es');
    return {
      tipos: [...tipos].sort(orden),
      departamentos: [...departamentos].sort(orden),
      municipios: [...municipios].sort(orden),
      estados: [...estados].sort(orden),
    };
  }, [universo, filtro.departamento]);

  const aMostrar = filtrados.slice(0, visibles);

  return (
    <>
      <Encabezado actualizado={datos?.actualizadoCurado ?? ''} />

      <main>
        <div className="columna-texto">
          <LineasEmergencia />
          <Replicas />
        </div>

        <section id="ayuda" className="ancha" aria-labelledby="t-ayuda">
          <h2 id="t-ayuda">Mapa y directorio de ayuda</h2>
          <p className="desc">
            Albergues, puntos de acopio, bancos de sangre, ollas comunitarias y
            organizaciones que gestionan donaciones. Cada punto lleva su fuente y su fecha
            de corte, y dice si lo verificamos a mano o si lo agregamos automáticamente de
            otro sitio ciudadano.{' '}
            <strong>Nunca publicamos cuentas bancarias:</strong> contacta directamente a la
            organización y valida con ella antes de entregar dinero.
          </p>

          {datos && (
            <p className="meta">
              {datos.totalCurados} puntos verificados a mano (corte:{' '}
              {datos.actualizadoCurado}) · {datos.totalAgregados} agregados
              automáticamente
              {datos.generadoScrape
                ? ` (última recolección ${haceCuanto(datos.generadoScrape)}, ${fechaCorta(
                    datos.generadoScrape,
                  )})`
                : ''}
              .
            </p>
          )}

          {datos?.aviso && (
            <p className="meta" role="status">
              ℹ️ {datos.aviso}
            </p>
          )}

          {error && (
            <div className="error" role="alert">
              <p>{error}</p>
              <p>
                Mientras tanto puedes abrir los datos abiertos directamente:{' '}
                <a href="/datos/ayuda.json">/datos/ayuda.json</a>.
              </p>
            </div>
          )}

          {!datos && !error && (
            <p className="cargando" role="status">
              Cargando puntos de ayuda…
            </p>
          )}

          {datos && (
            <>
              <Filtros
                filtro={filtro}
                cambiar={cambiar}
                limpiar={limpiar}
                tipos={opciones.tipos}
                departamentos={opciones.departamentos}
                municipios={opciones.municipios}
                estados={opciones.estados}
                totalEnRevision={datos.enRevision.length}
                mostrados={filtrados.length}
                total={universo.length}
              />

              <Suspense
                fallback={
                  <div className="mapa mapa-vacio" role="status">
                    <p>Cargando el mapa…</p>
                  </div>
                }
              >
                <MapaAyuda puntos={filtrados} />
              </Suspense>

              {filtrados.length === 0 ? (
                <p className="sin-resultados" role="status">
                  Ningún punto coincide con esos filtros. Prueba con menos filtros o busca
                  por municipio.
                </p>
              ) : anchoGrande ? (
                <TablaPuntos puntos={aMostrar} />
              ) : (
                <ul className="lista-tarjetas">
                  {aMostrar.map((p) => (
                    <TarjetaPunto key={p.id} punto={p} />
                  ))}
                </ul>
              )}

              {filtrados.length > aMostrar.length && (
                <div className="mas">
                  <button
                    type="button"
                    className="boton"
                    onClick={() => setVisibles((v) => v + PAGINA)}
                  >
                    Mostrar {Math.min(PAGINA, filtrados.length - aMostrar.length)} más (van{' '}
                    {aMostrar.length} de {filtrados.length})
                  </button>
                </div>
              )}
            </>
          )}
        </section>

        <div className="columna-texto">
          <Directorio />
        </div>
      </main>

      <PieDePagina licencia={datos?.licencia} politica={datos?.politica} />
    </>
  );
}
