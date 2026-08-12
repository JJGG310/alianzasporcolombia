// Estado en vivo publicado por el equipo desde /equipo.html.
//
// El JSON de datos solo cambia con un commit y un despliegue. Esto es la vía
// corta: alguien llama al albergue, marca "lleno" en el celular y el sitio lo
// refleja en un minuto. Quien acaba de llamar sabe más que un archivo de ayer,
// así que este estado manda sobre el publicado.
//
// Si el endpoint no está configurado (falta el KV o el secreto), todo esto es
// silencioso y el sitio se comporta exactamente como antes.

import type { Punto } from '../types';

export const RUTA_ESTADO = '/api/estado';

export interface EstadoVivo {
  estado?: 'abierto' | 'lleno' | 'cerrado' | string;
  urgente?: string;
  nota?: string;
  por?: string;
  /** ISO8601 */
  actualizado?: string;
}

export type MapaEnVivo = Record<string, EstadoVivo>;

/** Trae el estado en vivo. Nunca lanza: sin endpoint devuelve vacío. */
export async function traerEnVivo(señal?: AbortSignal): Promise<MapaEnVivo> {
  try {
    const r = await fetch(RUTA_ESTADO, { signal: señal, cache: 'no-store' });
    if (!r.ok) return {};
    const d = (await r.json()) as { puntos?: MapaEnVivo };
    return d.puntos ?? {};
  } catch {
    return {};
  }
}

/** El estado que reporta un humano gana sobre el del archivo. */
const A_ESTADO_UI: Record<string, string> = {
  abierto: 'activo',
  lleno: 'cubierto',
  cerrado: 'cerrado',
};

/**
 * Aplica los reportes sobre los puntos. Se cruza por nombre porque es lo único
 * estable entre el archivo curado y lo que ve quien reporta: los ids se
 * recalculan cuando cambia cualquier campo del punto.
 */
export function aplicarEnVivo(puntos: Punto[], vivo: MapaEnVivo): Punto[] {
  if (!vivo || Object.keys(vivo).length === 0) return puntos;
  return puntos.map((p) => {
    const v = vivo[p.nombre];
    if (!v) return p;
    return {
      ...p,
      estado: A_ESTADO_UI[v.estado ?? ''] ?? p.estado,
      necesita: v.urgente ? [v.urgente] : p.necesita,
      en_vivo: v,
    };
  });
}
