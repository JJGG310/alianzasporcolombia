// Etiquetas, colores y utilidades de presentación compartidas por el mapa,
// la tabla y las tarjetas. Un solo lugar para que nada se desincronice.

import type { Punto } from '../types';
import { esCurado } from './carga';

export interface TipoUI {
  label: string;
  color: string;
  /** Grupo con el que se pinta en el mapa. Ver GRUPOS_MAPA. */
  grupo: GrupoMapa;
}

/**
 * Color categórico del mapa.
 *
 * Antes había 14 tipos con 14 hues escogidos a ojo. En un mapa cualquier par de
 * marcadores puede quedar pegado, y con 14 colores la mitad de los pares son
 * indistinguibles — sobre todo para quien tiene daltonismo, que es 1 de cada 12
 * hombres.
 *
 * Así que los 14 tipos se agrupan en 4 decisiones + un neutro. Cuatro es el
 * techo real: validado con los 10 pares posibles, estos 4 pasan separación CVD
 * (peor par ΔE 9,2 deutan) y el piso de visión normal (peor par ΔE 16,3), tanto
 * sobre las teselas claras como sobre las atenuadas del modo oscuro. Un quinto
 * hue rompe el piso de visión normal.
 *
 * El gris de "otro" queda fuera del set categórico a propósito: no compite como
 * identidad, marca ausencia de una.
 *
 * El rojo NO está acá. El rojo es urgencia y solo urgencia (ver tokens.css).
 */
export type GrupoMapa = 'albergue' | 'acopio' | 'atencion' | 'rescate' | 'otro';

export interface GrupoUI {
  label: string;
  color: string;
}

export const GRUPOS_MAPA: Record<GrupoMapa, GrupoUI> = {
  albergue: { label: 'Dónde dormir', color: '#2a78d6' },
  acopio: { label: 'Llevar o recoger', color: '#eb6834' },
  atencion: { label: 'Salud, agua y comida', color: '#1baf7a' },
  rescate: { label: 'Rescate y emergencia', color: '#4a3aa7' },
  otro: { label: 'Otro', color: '#3d4149' },
};

const G = GRUPOS_MAPA;

export const TIPOS_UI: Record<string, TipoUI> = {
  albergue: { label: 'Albergue', color: G.albergue.color, grupo: 'albergue' },

  acopio: { label: 'Punto de acopio', color: G.acopio.color, grupo: 'acopio' },
  'donacion-org': { label: 'Recibe donaciones', color: G.acopio.color, grupo: 'acopio' },
  oferta: { label: 'Ofrece ayuda', color: G.acopio.color, grupo: 'acopio' },

  salud: { label: 'Salud', color: G.atencion.color, grupo: 'atencion' },
  sangre: { label: 'Banco de sangre', color: G.atencion.color, grupo: 'atencion' },
  agua: { label: 'Agua', color: G.atencion.color, grupo: 'atencion' },
  olla: { label: 'Olla comunitaria', color: G.atencion.color, grupo: 'atencion' },
  psicologico: { label: 'Apoyo psicosocial', color: G.atencion.color, grupo: 'atencion' },
  veterinario: { label: 'Veterinario / mascotas', color: G.atencion.color, grupo: 'atencion' },

  rescate: { label: 'Rescate', color: G.rescate.color, grupo: 'rescate' },
  necesidad: { label: 'Necesidad reportada', color: G.rescate.color, grupo: 'rescate' },

  info: { label: 'Información', color: G.otro.color, grupo: 'otro' },
  otro: { label: 'Otro', color: G.otro.color, grupo: 'otro' },
};

export function tipoUI(tipo: string): TipoUI {
  return TIPOS_UI[tipo] ?? { label: tipo || 'Otro', color: G.otro.color, grupo: 'otro' };
}

/** Grupo de mapa al que pertenece un tipo. */
export function grupoDe(tipo: string): GrupoMapa {
  return tipoUI(tipo).grupo;
}

export const ESTADOS_UI: Record<string, string> = {
  urgente: 'Urgente',
  activo: 'Activo',
  cubierto: 'Cubierto',
  cerrado: 'Cerrado',
  'por-confirmar': 'Por confirmar',
  descartado: 'Descartado',
};

export function estadoLabel(estado?: string): string {
  if (!estado) return 'Por confirmar';
  return ESTADOS_UI[estado] ?? estado;
}

export function estadoClase(estado?: string): string {
  const e = estado ?? 'por-confirmar';
  return ESTADOS_UI[e] ? `estado estado-${e}` : 'estado estado-por-confirmar';
}

/** Fecha corta y legible en español. Tolera ISO completo o solo YYYY-MM-DD. */
export function fechaCorta(iso?: string): string {
  if (!iso) return 'sin fecha';
  const d = new Date(iso.length === 10 ? `${iso}T12:00:00` : iso);
  if (Number.isNaN(d.getTime())) return iso;
  return d.toLocaleDateString('es-CO', {
    day: 'numeric',
    month: 'short',
    year: 'numeric',
  });
}

/** "hace 2 h 15 min" — para las réplicas y la última corrida del recolector. */
export function haceCuanto(iso: string): string {
  const t = new Date(iso).getTime();
  if (Number.isNaN(t)) return '';
  const seg = Math.max(0, Math.floor((Date.now() - t) / 1000));
  if (seg < 60) return 'hace menos de un minuto';
  const min = Math.floor(seg / 60);
  if (min < 60) return `hace ${min} min`;
  const h = Math.floor(min / 60);
  if (h < 24) return `hace ${h} h ${min % 60} min`;
  const d = Math.floor(h / 24);
  return `hace ${d} día${d === 1 ? '' : 's'}`;
}

/** Dónde queda el punto, en una sola línea. */
export function ubicacion(p: Punto): string {
  return [p.direccion, p.barrio && `barrio ${p.barrio}`, p.comuna && `comuna ${p.comuna}`]
    .filter(Boolean)
    .join(' · ');
}

export function lugar(p: Punto): string {
  return [p.municipio, p.departamento].filter(Boolean).join(', ');
}

/** Qué ofrece / qué necesita, aplanado para vistas compactas. */
export function queHace(p: Punto): string {
  const partes: string[] = [];
  if (p.descripcion) partes.push(p.descripcion);
  if (p.ofrece?.length) partes.push(`Ofrece: ${p.ofrece.join(', ')}.`);
  if (p.necesita?.length) partes.push(`Necesita: ${p.necesita.join(', ')}.`);
  return partes.join(' ');
}

/**
 * Arma el texto listo para pegar en WhatsApp. En Colombia la distribución real
 * es WhatsApp; el sitio está pensado para ser citado por pedazos (PLAN.md #6).
 */
export function textoWhatsApp(p: Punto): string {
  const lineas: string[] = [];
  lineas.push(`*${p.nombre}* (${tipoUI(p.tipo).label})`);

  const donde = [ubicacion(p), lugar(p)].filter(Boolean).join(' — ');
  if (donde) lineas.push(`📍 ${donde}`);
  if (p.horario) lineas.push(`🕘 ${p.horario}`);

  if (p.descripcion) lineas.push(p.descripcion);
  if (p.necesita?.length) lineas.push(`Necesita: ${p.necesita.join(', ')}`);
  if (p.ofrece?.length) lineas.push(`Ofrece: ${p.ofrece.join(', ')}`);
  if (p.contacto) lineas.push(`☎️ ${p.contacto}`);

  lineas.push(`Estado: ${estadoLabel(p.estado)} · Datos al ${fechaCorta(p.fecha_corte)}`);
  lineas.push(
    esCurado(p)
      ? 'Verificado a mano por Alianzas por Colombia.'
      : 'Agregado automáticamente de otro sitio ciudadano: confirma antes de ir.',
  );
  if (p.fuente_url) lineas.push(`Fuente: ${p.fuente_url}`);
  lineas.push('Más en alianzasporcolombia.com — nunca entregues dinero sin validar.');

  return lineas.join('\n');
}

export function enlaceWhatsApp(p: Punto): string {
  return `https://wa.me/?text=${encodeURIComponent(textoWhatsApp(p))}`;
}

/** Texto plano sobre el que corre la búsqueda libre. */
export function textoBuscable(p: Punto): string {
  return [
    p.nombre,
    p.descripcion,
    p.direccion,
    p.barrio,
    p.comuna,
    p.municipio,
    p.departamento,
    p.organizacion,
    p.contacto,
    tipoUI(p.tipo).label,
    p.necesita?.join(' '),
    p.ofrece?.join(' '),
  ]
    .filter(Boolean)
    .join(' ')
    .toLowerCase()
    .normalize('NFD')
    .replace(/[\u0300-\u036f]/g, '');
}

export function normalizarBusqueda(txt: string): string {
  return txt
    .toLowerCase()
    .normalize('NFD')
    .replace(/[\u0300-\u036f]/g, '')
    .trim();
}

// ---------------------------------------------------------------------------
// Cercanía
//
// "¿Cuál me queda más cerca?" es la pregunta real de alguien parado en la
// calle con el celular en la mano. Contestarla bien vale más que cualquier
// filtro por departamento.
// ---------------------------------------------------------------------------

export interface Coordenada {
  lat: number;
  lon: number;
}

/** Distancia en kilómetros entre dos puntos (haversine). */
export function distanciaKm(a: Coordenada, b: Coordenada): number {
  const R = 6371;
  const rad = Math.PI / 180;
  const dLat = (b.lat - a.lat) * rad;
  const dLon = (b.lon - a.lon) * rad;
  const s =
    Math.sin(dLat / 2) ** 2 +
    Math.cos(a.lat * rad) * Math.cos(b.lat * rad) * Math.sin(dLon / 2) ** 2;
  return 2 * R * Math.asin(Math.sqrt(s));
}

/**
 * Distancia legible. Bajo un kilómetro se muestra en metros redondeados a 50:
 * "a 350 m" se camina, "a 0,35 km" hay que traducirlo mentalmente.
 */
export function distanciaTexto(km: number): string {
  if (km < 1) return `${Math.max(50, Math.round((km * 1000) / 50) * 50)} m`;
  if (km < 10) return `${km.toFixed(1).replace('.', ',')} km`;
  return `${Math.round(km)} km`;
}
