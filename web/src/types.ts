// Tipos del dominio. `Punto` refleja EXACTAMENTE model.Punto de
// scraper/internal/model/punto.go (nombres json incluidos). Si ese archivo
// cambia, este también.

/** Tipos canónicos (model.Tipo*). */
export const TIPOS = [
  'albergue',
  'acopio',
  'sangre',
  'olla',
  'donacion-org',
  'rescate',
  'salud',
  'agua',
  'veterinario',
  'psicologico',
  'necesidad',
  'oferta',
  'info',
  'otro',
] as const;
export type Tipo = (typeof TIPOS)[number];

/** Estados canónicos (model.Estado*). */
export const ESTADOS = [
  'urgente',
  'activo',
  'cubierto',
  'cerrado',
  'por-confirmar',
  'descartado',
] as const;
export type Estado = (typeof ESTADOS)[number];

/** Prioridades canónicas (model.Prioridad*). */
export const PRIORIDADES = ['critica', 'alta', 'media', 'baja'] as const;
export type Prioridad = (typeof PRIORIDADES)[number];

/**
 * Un ítem de ayuda normalizado, venga del scraper o del JSON curado a mano.
 * Espejo de model.Punto. Los campos opcionales lo son porque en Go llevan
 * `omitempty`: pueden no venir en el JSON.
 */
export interface Punto {
  id: string;
  /** slug del sitio de origen; `curado` para los 91 verificados a mano. */
  fuente: string;
  fuente_url: string;

  tipo: Tipo | string;
  nombre: string;
  descripcion?: string;

  departamento?: string;
  municipio?: string;
  comuna?: string;
  barrio?: string;
  direccion?: string;
  lat?: number | null;
  lon?: number | null;

  necesita?: string[];
  ofrece?: string[];

  estado?: Estado | string;
  prioridad?: Prioridad | string;

  voluntarios_hay?: number;
  voluntarios_faltan?: number;

  contacto?: string;
  organizacion?: string;
  horario?: string;

  verificado: boolean;
  confirmaciones?: number;
  desmentidos?: number;

  /** ISO8601 */
  creado?: string;
  /** ISO8601 */
  actualizado?: string;
  /** ISO8601 del momento del scrape. Obligatorio: principio #2 del proyecto. */
  fecha_corte: string;

  /** BLOQUEANTE: si trae algo, un humano debe revisarlo antes de publicarlo. */
  politica_alerta?: string[];

  /**
   * Informativo y NO bloqueante: cosas que conviene saber del dato (p. ej.
   * coordenadas descartadas por caer fuera de Colombia). Va separado a
   * propósito — un punto publicable no debe esconderse por tener mal el lat/lon.
   */
  avisos?: string[];

  raw?: unknown;
}

/** Espejo de model.Resultado (sin los puntos, que van aparte en el archivo). */
export interface FuenteResultado {
  fuente: string;
  url: string;
  mecanismo: string;
  ok: boolean;
  error?: string;
  nota?: string;
  duracion?: string;
  obtenido: string;
  total: number;
}

/** Forma de /extracted-data/puntos.json (lo genera el scraper en Go). */
export interface ArchivoScrapeado {
  generado: string;
  licencia: string;
  politica: string;
  total: number;
  fuentes?: FuenteResultado[];
  puntos: Punto[];
  /** Solo lo trae el archivo de muestra de desarrollo. */
  muestra?: boolean;
  nota?: string;
}

// ---------------------------------------------------------------------------
// Esquema actual de datos/ayuda.json (los 91 curados a mano).
// ---------------------------------------------------------------------------

export interface ItemCurado {
  tipo: Tipo | string;
  nombre: string;
  departamento?: string;
  municipio?: string;
  direccion?: string;
  /** Descripción libre: se mapea a Punto.descripcion. */
  que?: string;
  contacto?: string;
  /** URL citable: se mapea a Punto.fuente_url. */
  fuente: string;
  fecha_corte: string;
  lat?: number | null;
  lon?: number | null;
  // Campos que el JSON curado todavía no usa pero podría empezar a usar.
  estado?: Estado | string;
  horario?: string;
  organizacion?: string;
  barrio?: string;
  comuna?: string;
}

/** Forma de /datos/ayuda.json. */
export interface ArchivoCurado {
  actualizado: string;
  licencia: string;
  politica: string;
  items: ItemCurado[];
}

// ---------------------------------------------------------------------------

/** Estado de los filtros de la vista de ayuda. */
export interface Filtro {
  texto: string;
  tipo: string;
  departamento: string;
  municipio: string;
  estado: string;
  /** Solo puntos verificados a mano. */
  soloVerificados: boolean;
  /** Apagado por defecto: muestra los puntos con `politica_alerta`. */
  requiereRevision: boolean;
}

export const FILTRO_VACIO: Filtro = {
  texto: '',
  tipo: '',
  departamento: '',
  municipio: '',
  estado: '',
  soloVerificados: false,
  requiereRevision: false,
};

/** Lo que la app tiene cargado después de fusionar las dos fuentes. */
export interface DatosApp {
  puntos: Punto[];
  /** Puntos escondidos por traer `politica_alerta`. */
  enRevision: Punto[];
  totalCurados: number;
  totalAgregados: number;
  /** Texto legible de `actualizado` en datos/ayuda.json. */
  actualizadoCurado: string;
  /** ISO8601 de la última corrida del scraper, o null si no hay archivo. */
  generadoScrape: string | null;
  fuentes: FuenteResultado[];
  licencia: string;
  politica: string;
  /** Aviso discreto para la UI cuando algo no cargó (p. ej. 404 del scraper). */
  aviso: string | null;
  /** true cuando /extracted-data/puntos.json es el archivo de muestra. */
  esMuestra: boolean;
}
