/**
 * Iconos del sitio.
 *
 * Dibujados a mano sobre una retícula de 24 px con trazo de 1,5 y remates
 * redondos, siguiendo el lenguaje de Iconsax. Van inline y no como paquete:
 * son ocho iconos, y una librería entera para ocho iconos son 40 KB que en una
 * red mala se sienten.
 *
 * Son decorativos por defecto (`aria-hidden`). Cuando un icono ES el botón,
 * pásale `titulo` y se vuelve una imagen con nombre accesible.
 */

export type NombreIcono =
  | 'buscar'
  | 'ubicacion'
  | 'telefono'
  | 'ruta'
  | 'whatsapp'
  | 'alerta'
  | 'verificado'
  | 'refrescar'
  | 'filtro'
  | 'cerrar'
  | 'flecha-abajo'
  | 'enlace';

const TRAZOS: Record<NombreIcono, JSX.Element> = {
  buscar: (
    <>
      <circle cx="11" cy="11" r="7" />
      <path d="m20 20-3.5-3.5" />
    </>
  ),
  ubicacion: (
    <>
      <path d="M12 21s7-5.5 7-11a7 7 0 1 0-14 0c0 5.5 7 11 7 11Z" />
      <circle cx="12" cy="10" r="2.5" />
    </>
  ),
  telefono: (
    <path d="M6.5 3.5h3l1.5 4-2 1.5a12 12 0 0 0 6 6l1.5-2 4 1.5v3a2 2 0 0 1-2.2 2A17 17 0 0 1 4.5 5.7 2 2 0 0 1 6.5 3.5Z" />
  ),
  ruta: (
    <>
      <path d="M12 3 21 21 12 17 3 21 12 3Z" />
    </>
  ),
  whatsapp: (
    <>
      <path d="M3.5 20.5 5 16.2A8.2 8.2 0 1 1 8 19.2l-4.5 1.3Z" />
      <path d="M9 9.2c.3 2.6 3.2 5.5 5.8 5.8.7.1 1.4-.5 1.4-1.2v-.6l-2-.8-.9 1a7 7 0 0 1-2.7-2.7l1-.9-.8-2h-.6c-.7 0-1.3.7-1.2 1.4Z" />
    </>
  ),
  alerta: (
    <>
      <path d="M12 3.5 21.5 20H2.5L12 3.5Z" />
      <path d="M12 10v4" />
      <path d="M12 17.2v.1" />
    </>
  ),
  verificado: (
    <>
      <path d="M12 3l2.2 1.6 2.7-.2.8 2.6 2.2 1.6-1 2.5 1 2.5-2.2 1.6-.8 2.6-2.7-.2L12 21l-2.2-1.6-2.7.2-.8-2.6L4.1 15l1-2.5-1-2.5 2.2-1.6.8-2.6 2.7.2L12 3Z" />
      <path d="m9 12 2 2 4-4" />
    </>
  ),
  refrescar: (
    <>
      <path d="M20 12a8 8 0 1 1-2.6-5.9" />
      <path d="M20 4v4h-4" />
    </>
  ),
  filtro: (
    <>
      <path d="M4 6h16" />
      <path d="M7 12h10" />
      <path d="M10 18h4" />
    </>
  ),
  cerrar: (
    <>
      <path d="m6 6 12 12" />
      <path d="m18 6-12 12" />
    </>
  ),
  'flecha-abajo': <path d="m6 9 6 6 6-6" />,
  enlace: (
    <>
      <path d="M14 4h6v6" />
      <path d="M20 4 11 13" />
      <path d="M18 14v5a1.5 1.5 0 0 1-1.5 1.5h-11A1.5 1.5 0 0 1 4 19V8a1.5 1.5 0 0 1 1.5-1.5H10" />
    </>
  ),
};

interface Props {
  nombre: NombreIcono;
  /** Tamaño en píxeles. Por defecto sigue al texto (1em). */
  tam?: number | string;
  /** Si el icono carga el significado (un botón sin texto), su nombre va acá. */
  titulo?: string;
  className?: string;
}

export default function Icono({ nombre, tam = '1.15em', titulo, className }: Props) {
  return (
    <svg
      className={className}
      width={tam}
      height={tam}
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth={1.6}
      strokeLinecap="round"
      strokeLinejoin="round"
      role={titulo ? 'img' : undefined}
      aria-hidden={titulo ? undefined : true}
      aria-label={titulo}
      focusable="false"
      style={{ flex: 'none' }}
    >
      {titulo ? <title>{titulo}</title> : null}
      {TRAZOS[nombre]}
    </svg>
  );
}
