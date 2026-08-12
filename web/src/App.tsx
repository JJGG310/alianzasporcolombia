import { Suspense, lazy, useCallback, useEffect, useMemo, useState } from 'react';

import Avisos from './componentes/Avisos';
import BuscarAlguien from './componentes/BuscarAlguien';
import Directorio from './componentes/Directorio';
import Necesidades from './componentes/Necesidades';
import Encabezado from './componentes/Encabezado';
import Filtros, { ordenarPorCercania, useOrigen } from './componentes/Filtros';
import PieDePagina from './componentes/PieDePagina';
import Replicas from './componentes/Replicas';
import TablaPuntos from './componentes/TablaPuntos';
import TarjetaPunto from './componentes/TarjetaPunto';
import Icono from './componentes/Icono';

import { cargarDatos, esCurado } from './data/carga';
import { fechaCorta, haceCuanto, normalizarBusqueda, textoBuscable } from './data/catalogo';
import { FILTRO_VACIO, type DatosApp, type Filtro, type Punto } from './types';

// Leaflet pesa: se carga aparte para que lo crítico pinte primero en una red mala.
const MapaAyuda = lazy(() => import('./componentes/MapaAyuda'));

/** Cuántos puntos se pintan de entrada. El resto, bajo demanda. */
const PAGINA = 60;

/**
 * Las cinco secciones, en el orden en que sirven durante una emergencia.
 *
 * Se navegan por pestañas y no por scroll: la página completa mide varios
 * metros en un celular, y quien busca dónde donar no tiene por qué pasar por
 * 979 puntos de rescate para llegar. La estructura viene del sitio en vivo,
 * que ya la validó con gente usándolo.
 */
const SECCIONES = [
  { id: 'ahora', titulo: 'Ahora' },
  { id: 'ayuda', titulo: 'Ayuda cerca' },
  { id: 'buscar', titulo: 'Buscar a alguien' },
  { id: 'donar', titulo: 'Donar' },
  { id: 'fuentes', titulo: 'Fuentes' },
] as const;

type Seccion = (typeof SECCIONES)[number]['id'];

/** La sección se refleja en el hash para poder compartir un enlace directo. */
function seccionDelHash(): Seccion {
  const h = typeof window !== 'undefined' ? window.location.hash.replace('#', '') : '';
  return (SECCIONES.find((s) => s.id === h)?.id ?? 'ahora') as Seccion;
}

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
  const [seccion, setSeccion] = useState<Seccion>(seccionDelHash);
  const anchoGrande = useAnchoGrande();

  useEffect(() => {
    if (seccion === 'donar') {
      setFiltro((f) => (f.tipo === 'donacion-org' ? f : { ...f, tipo: 'donacion-org' }));
    } else if (seccion === 'ayuda') {
      setFiltro((f) => (f.tipo === 'donacion-org' ? { ...f, tipo: '' } : f));
    }
  }, [seccion]);

  // El botón "atrás" del celular debe devolver a la pestaña anterior, no sacar
  // del sitio: es el gesto que la gente usa sin pensar.
  useEffect(() => {
    const alCambiar = () => setSeccion(seccionDelHash());
    window.addEventListener('hashchange', alCambiar);
    return () => window.removeEventListener('hashchange', alCambiar);
  }, []);

  // La ubicación que la persona compartió con "Cerca de mí". Vive solo en
  // memoria del navegador: no se guarda ni se manda a ningún lado.
  const origen = useOrigen();

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

  // El universo cambia si se pide ver la cola de revisión.
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
    const base = universo.filter((p) => {
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
    // Con ubicación compartida, lo más cerca manda sobre cualquier otro orden:
    // es la única pregunta que importa cuando estás parado en la calle.
    return origen ? ordenarPorCercania(base, origen) : base;
  }, [universo, filtro, indice, origen]);

  const opciones = useMemo(() => {
    const tipos = new Set<string>();
    const departamentos = new Set<string>();
    const municipios = new Set<string>();
    const estados = new Set<string>();
    for (const p of universo) {
      tipos.add(p.tipo);
      if (p.departamento) departamentos.add(p.departamento);
      if (p.estado) estados.add(p.estado);
      // Los municipios se acotan al departamento elegido.
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
  const urgentes = useMemo(
    () => filtrados.filter((p) => p.estado === 'urgente').length,
    [filtrados],
  );

  return (
    <div className="app">
      <Encabezado actualizado={datos?.actualizadoCurado ?? ''} />

      <main className="app__cuerpo">
        <div className="contenedor">
          <Avisos />

          <nav className="pestanas" aria-label="Secciones del sitio">
            {SECCIONES.map((s) => (
              <button
                key={s.id}
                type="button"
                className="chip"
                aria-pressed={seccion === s.id}
                onClick={() => {
                  setSeccion(s.id);
                  // replaceState y no push: no queremos 20 entradas de historial
                  // por tocar pestañas, pero sí un enlace compartible.
                  history.replaceState(null, '', `#${s.id}`);
                }}
              >
                {s.titulo}
              </button>
            ))}
          </nav>

          {seccion === 'ahora' && (
            <>
              <Replicas />
              <Necesidades />
            </>
          )}

          {seccion === 'buscar' && <BuscarAlguien />}

          {error && (
            <div className="aviso aviso--error" role="alert">
              <Icono nombre="alerta" />
              <div>
                <p>{error}</p>
                <p>
                  Los datos siguen abiertos y los puedes ver directo:{' '}
                  <a href="/extracted-data/puntos.json">todos los puntos</a>.
                </p>
              </div>
            </div>
          )}

          {datos?.aviso && (
            <div className="aviso" role="status">
              <Icono nombre="alerta" />
              <p>{datos.aviso}</p>
            </div>
          )}

          {(seccion === 'ayuda' || seccion === 'donar') && (
          <section id="herramienta" className="herramienta" aria-label="Buscar ayuda">
            {datos && (
              <p className="procedencia">
                <span className="procedencia__dato">
                  <Icono nombre="verificado" />
                  <span className="procedencia__num">{datos.totalCurados}</span> verificados
                  a mano
                </span>
                <span className="procedencia__dato">
                  <Icono nombre="refrescar" />
                  <span className="procedencia__num">{datos.totalAgregados}</span> agregados
                  de otros sitios
                  {datos.generadoScrape
                    ? ` · ${haceCuanto(datos.generadoScrape)} (${fechaCorta(datos.generadoScrape)})`
                    : ''}
                </span>
              </p>
            )}

            {!datos && !error && (
              <div role="status" aria-live="polite">
                <span className="solo-lectores">Cargando puntos de ayuda…</span>
                <div className="esqueleto" style={{ height: 52, marginBottom: 'var(--e3)' }} />
                <div className="esqueleto" style={{ height: 38, marginBottom: 'var(--e4)' }} />
                <div className="esqueleto" style={{ height: '46vh', marginBottom: 'var(--e4)' }} />
                <div className="esqueleto" style={{ height: 120, marginBottom: 'var(--e3)' }} />
                <div className="esqueleto" style={{ height: 120 }} />
              </div>
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
                  puntos={universo}
                />

                {filtrados.length > 0 && (
                  <Suspense
                    fallback={
                      <div className="mapa">
                        <div className="mapa__cargando" role="status">
                          Cargando el mapa…
                        </div>
                      </div>
                    }
                  >
                    <MapaAyuda puntos={filtrados} miUbicacion={origen} />
                  </Suspense>
                )}

                {urgentes > 0 && (
                  <p className="solo-lectores" aria-live="polite">
                    {urgentes} de los {filtrados.length} puntos están marcados como urgentes.
                  </p>
                )}

                {filtrados.length > 0 &&
                  (anchoGrande ? (
                    <TablaPuntos puntos={aMostrar} origen={origen} />
                  ) : (
                    <ul className="lista-puntos">
                      {aMostrar.map((p) => (
                        <li key={p.id}>
                          <TarjetaPunto punto={p} origen={origen} />
                        </li>
                      ))}
                    </ul>
                  ))}

                {filtrados.length > aMostrar.length && (
                  <div className="cargar-mas">
                    <button
                      type="button"
                      className="btn"
                      onClick={() => setVisibles((v) => v + PAGINA)}
                    >
                      Mostrar {Math.min(PAGINA, filtrados.length - aMostrar.length)} más
                      <span style={{ color: 'var(--tinta-400)' }}>
                        {aMostrar.length} de {filtrados.length}
                      </span>
                    </button>
                  </div>
                )}
              </>
            )}
          </section>
          )}

          {seccion === 'donar' && (
          <div className="explicacion">
            <h2>Qué es esto</h2>
            <p className="prosa">
              Información ciudadana centralizada sobre el terremoto M 7,4 del 10 de agosto
              de 2026. Empezamos por Cali y el Valle del Cauca en las primeras horas y
              crecimos por regiones a todo el país. Recopilamos y enlazamos: las cifras y
              las órdenes de evacuación las emiten las autoridades, no nosotros.
            </p>
            <p className="prosa">
              Cada punto lleva su fuente y su fecha de corte, y dice si lo verificamos a
              mano o si lo agregamos automáticamente de otro sitio ciudadano. Lo agregado no
              está confirmado uno por uno: por eso cita su origen, para que puedas
              comprobarlo.
            </p>
            <p className="regla-dinero">
              <strong>Nunca publicamos cuentas bancarias.</strong> Ninguna. Si vas a donar,
              contacta directamente a la organización y valida con ella antes de entregar
              dinero. Si un punto de este sitio te pide plata, no es nuestro.
            </p>
          </div>
          )}

          {seccion === 'fuentes' && <Directorio />}
        </div>
      </main>

      <PieDePagina licencia={datos?.licencia} politica={datos?.politica} />
    </div>
  );
}
