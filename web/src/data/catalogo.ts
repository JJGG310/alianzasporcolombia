// Etiquetas, colores y utilidades de presentación compartidas por el mapa,
// la tabla y las tarjetas. Un solo lugar para que nada se desincronice.

import type { Punto } from '../types';
import { esCurado } from './carga';

export interface TipoUI {
  label: string;
  color: string;
}

/** Colores heredados del index.html original + los tipos nuevos del scraper. */
export const TIPOS_UI: Record<string, TipoUI> = {
  albergue: { label: 'Albergue', color: '#c0182b' },
  acopio: { label: 'Punto de acopio', color: '#1d6b2f' },
  sangre: { label: 'Banco de sangre', color: '#8b0f57' },
  olla: { label: 'Olla comunitaria', color: '#b45f06' },
  'donacion-org': { label: 'Recibe donaciones', color: '#0b57d0' },
  rescate: { label: 'Rescate', color: '#7a2f00' },
  salud: { label: 'Salud', color: '#0f6e6e' },
  agua: { label: 'Agua', color: '#1069a8' },
  veterinario: { label: 'Veterinario / mascotas', color: '#6b4f1d' },
  psicologico: { label: 'Apoyo psicosocial', color: '#4b2fa8' },
  necesidad: { label: 'Necesidad reportada', color: '#a5201f' },
  oferta: { label: 'Ofrece ayuda', color: '#2b6b1d' },
  info: { label: 'Información', color: '#555a66' },
  otro: { label: 'Otro', color: '#555a66' },
};

export function tipoUI(tipo: string): TipoUI {
  return TIPOS_UI[tipo] ?? { label: tipo || 'Otro', color: '#555a66' };
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
