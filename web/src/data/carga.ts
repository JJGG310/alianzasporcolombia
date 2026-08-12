// Carga y fusión de las dos fuentes de datos.
//
//   /datos/ayuda.json            91 puntos curados a mano  → siempre debe existir
//   /extracted-data/puntos.json  puntos scrapeados          → puede dar 404
//
// La app nunca se rompe porque falte el segundo: muestra solo los curados y un
// aviso discreto.

import type {
  ArchivoCurado,
  FuenteResultado,
  ArchivoScrapeado,
  DatosApp,
  ItemCurado,
  Punto,
} from '../types';

/** Slug que marca un punto verificado a mano. */
export const FUENTE_CURADO = 'curado';

export const RUTA_CURADOS = '/datos/ayuda.json';
export const RUTA_SCRAPEADOS = '/extracted-data/puntos.json';

/** ¿Este punto lo revisó un humano de Alianzas por Colombia? */
export function esCurado(p: Punto): boolean {
  return p.fuente === FUENTE_CURADO;
}

/**
 * Un punto "requiere revisión" si el saneamiento del scraper le dejó una
 * alerta de política (posible dato bancario, coordenada descartada...).
 * Estos NO salen en la vista pública normal.
 */
export function requiereRevision(p: Punto): boolean {
  return (p.politica_alerta?.length ?? 0) > 0;
}

/** ¿Tiene coordenadas usables para el mapa? */
export function tieneCoordenadas(
  p: Punto,
): p is Punto & { lat: number; lon: number } {
  return typeof p.lat === 'number' && typeof p.lon === 'number';
}

/** Hash determinista (djb2) para dar ids estables a los curados. */
function hash(txt: string): string {
  let h = 5381;
  for (let i = 0; i < txt.length; i++) {
    h = ((h << 5) + h + txt.charCodeAt(i)) | 0;
  }
  return (h >>> 0).toString(36);
}

function limpiar(s: string | undefined | null): string {
  return (s ?? '').replace(/\s+/g, ' ').trim();
}

/** Clave de deduplicación: mismo nombre + mismo municipio = mismo punto. */
function clave(p: Pick<Punto, 'nombre' | 'municipio'>): string {
  const norm = (s: string) =>
    s
      .toLowerCase()
      .normalize('NFD')
      .replace(/[\u0300-\u036f]/g, '')
      .replace(/[^a-z0-9]+/g, ' ')
      .trim();
  return `${norm(p.nombre ?? '')}|${norm(p.municipio ?? '')}`;
}

/**
 * Normaliza un ítem del JSON curado al mismo tipo que los scrapeados.
 * Mapeos: `que` → descripcion, `fuente` → fuente_url, fuente = "curado",
 * verificado = true.
 */
export function normalizarCurado(item: ItemCurado): Punto {
  const nombre = limpiar(item.nombre);
  const municipio = limpiar(item.municipio);
  return {
    id: `${FUENTE_CURADO}-${hash(`${item.tipo}|${nombre}|${municipio}|${item.fuente}`)}`,
    fuente: FUENTE_CURADO,
    fuente_url: item.fuente,
    tipo: item.tipo || 'otro',
    nombre,
    descripcion: limpiar(item.que) || undefined,
    departamento: limpiar(item.departamento) || undefined,
    municipio: municipio || undefined,
    comuna: limpiar(item.comuna) || undefined,
    barrio: limpiar(item.barrio) || undefined,
    direccion: limpiar(item.direccion) || undefined,
    lat: typeof item.lat === 'number' ? item.lat : null,
    lon: typeof item.lon === 'number' ? item.lon : null,
    estado: item.estado || 'activo',
    contacto: limpiar(item.contacto) || undefined,
    organizacion: limpiar(item.organizacion) || undefined,
    horario: limpiar(item.horario) || undefined,
    // Los 91 pasaron por revisión humana: ese es justo el punto del archivo.
    verificado: true,
    fecha_corte: item.fecha_corte,
  };
}

/** Rellena lo que el scraper pudo haber omitido, sin inventar datos. */
function normalizarScrapeado(p: Punto, i: number): Punto {
  return {
    ...p,
    id: p.id || `agregado-${i}`,
    fuente: p.fuente || 'agregado',
    tipo: p.tipo || 'otro',
    estado: p.estado || 'por-confirmar',
    nombre: limpiar(p.nombre),
    lat: typeof p.lat === 'number' ? p.lat : null,
    lon: typeof p.lon === 'number' ? p.lon : null,
    verificado: p.verificado === true,
  };
}

async function traerJSON<T>(ruta: string, señal?: AbortSignal): Promise<T> {
  const r = await fetch(ruta, { signal: señal, headers: { Accept: 'application/json' } });
  if (!r.ok) throw new Error(`${ruta}: HTTP ${r.status}`);
  return (await r.json()) as T;
}

/**
 * Carga las dos fuentes en paralelo y las fusiona.
 * Si falla el archivo curado la promesa rechaza (sin él no hay sitio).
 * Si falla el scrapeado, se devuelve solo lo curado más un aviso.
 */
export async function cargarDatos(señal?: AbortSignal): Promise<DatosApp> {
  const [curadoRes, scrapeRes] = await Promise.allSettled([
    traerJSON<ArchivoCurado>(RUTA_CURADOS, señal),
    traerJSON<ArchivoScrapeado>(RUTA_SCRAPEADOS, señal),
  ]);

  if (curadoRes.status === 'rejected') {
    throw new Error(
      'No pudimos cargar los datos de ayuda. Revisa tu conexión y vuelve a intentar.',
    );
  }

  const curado = curadoRes.value;
  const curados = (curado.items ?? []).map(normalizarCurado);

  let agregados: Punto[] = [];
  let generadoScrape: string | null = null;
  let fuentes: FuenteResultado[] = [];
  let aviso: string | null = null;
  let esMuestra = false;

  if (scrapeRes.status === 'fulfilled') {
    const scrape = scrapeRes.value;
    agregados = (scrape.puntos ?? []).map(normalizarScrapeado);
    generadoScrape = scrape.generado ?? null;
    fuentes = scrape.fuentes ?? [];
    esMuestra = scrape.muestra === true;
    if (esMuestra) {
      aviso =
        'Los puntos agregados automáticamente son datos de MUESTRA para desarrollo. Todavía no corrió el recolector.';
    }
  } else {
    aviso =
      'Todavía no cargamos los puntos agregados de otros sitios ciudadanos. Estás viendo solo los verificados a mano.';
  }

  // Los curados mandan: si el scraper trajo el mismo lugar, se descarta.
  const vistos = new Set(curados.map(clave));
  const agregadosUnicos = agregados.filter((p) => {
    const k = clave(p);
    if (vistos.has(k)) return false;
    vistos.add(k);
    return true;
  });

  const todos = [...curados, ...agregadosUnicos];

  // Los que traen alerta de política salen del listado público.
  const publicos = todos.filter((p) => !requiereRevision(p));
  const enRevision = todos.filter(requiereRevision);

  return {
    puntos: ordenar(publicos),
    enRevision: ordenar(enRevision),
    totalCurados: curados.length,
    totalAgregados: agregadosUnicos.length,
    actualizadoCurado: curado.actualizado,
    generadoScrape,
    fuentes,
    licencia: curado.licencia,
    politica: curado.politica,
    aviso,
    esMuestra,
  };
}

const PESO_ESTADO: Record<string, number> = {
  urgente: 0,
  activo: 1,
  'por-confirmar': 2,
  cubierto: 3,
  cerrado: 4,
  descartado: 5,
};

/** Verificados primero; dentro de eso, por urgencia y luego alfabético. */
function ordenar(puntos: Punto[]): Punto[] {
  return [...puntos].sort((a, b) => {
    const ca = esCurado(a) ? 0 : 1;
    const cb = esCurado(b) ? 0 : 1;
    if (ca !== cb) return ca - cb;
    const ea = PESO_ESTADO[a.estado ?? ''] ?? 3;
    const eb = PESO_ESTADO[b.estado ?? ''] ?? 3;
    if (ea !== eb) return ea - eb;
    return a.nombre.localeCompare(b.nombre, 'es');
  });
}
