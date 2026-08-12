// Estado en vivo de los puntos de ayuda.
//
// El JSON de datos solo cambia con un commit; esto permite que el equipo marque
// "este albergue se llenó" o "aquí ya no reciben ropa" desde el celular y que se
// vea en el sitio en segundos, sin desplegar nada.
//
// GET  /api/estado  -> { actualizado, puntos: { <clave>: {...} } }   (público)
// POST /api/estado  -> escribe un punto. Requiere Authorization: Bearer <CLAVE_EQUIPO>
//
// Necesita en el proyecto de Pages:
//   - binding KV llamado ESTADO
//   - secreto CLAVE_EQUIPO

const LLAVE = 'overlay';
const ESTADOS_VALIDOS = ['abierto', 'lleno', 'cerrado'];

const json = (o, status = 200) =>
  new Response(JSON.stringify(o), {
    status,
    headers: { 'content-type': 'application/json; charset=utf-8', 'cache-control': 'no-store' },
  });

export async function onRequestGet({ env }) {
  if (!env.ESTADO) return json({ actualizado: null, puntos: {} });
  const guardado = await env.ESTADO.get(LLAVE);
  return new Response(guardado ?? '{"puntos":{}}', {
    headers: { 'content-type': 'application/json; charset=utf-8', 'cache-control': 'no-store' },
  });
}

export async function onRequestPost({ request, env }) {
  if (!env.ESTADO || !env.CLAVE_EQUIPO) return json({ error: 'falta configurar KV o la clave' }, 503);

  const auth = request.headers.get('authorization') || '';
  const enc = new TextEncoder();
  const dada = enc.encode(auth.startsWith('Bearer ') ? auth.slice(7) : '');
  const real = enc.encode(env.CLAVE_EQUIPO);
  // timingSafeEqual exige el mismo largo en bytes; se compara sobre bytes y no sobre
  // caracteres porque una clave con tildes ocupa más bytes que letras tiene.
  if (dada.byteLength !== real.byteLength || !crypto.timingSafeEqual(dada, real))
    return json({ error: 'clave incorrecta' }, 401);

  let cuerpo;
  try { cuerpo = await request.json(); } catch { return json({ error: 'JSON inválido' }, 400); }

  const clavePunto = String(cuerpo.punto || '').trim().slice(0, 160);
  if (!clavePunto) return json({ error: 'falta el punto' }, 400);
  if (cuerpo.estado && !ESTADOS_VALIDOS.includes(cuerpo.estado))
    return json({ error: `estado debe ser uno de: ${ESTADOS_VALIDOS.join(', ')}` }, 400);

  const actual = JSON.parse((await env.ESTADO.get(LLAVE)) || '{"puntos":{}}');
  actual.puntos = actual.puntos || {};

  if (cuerpo.borrar) {
    delete actual.puntos[clavePunto];
  } else {
    actual.puntos[clavePunto] = {
      estado: cuerpo.estado || 'abierto',
      urgente: String(cuerpo.urgente || '').trim().slice(0, 200),
      nota: String(cuerpo.nota || '').trim().slice(0, 300),
      por: String(cuerpo.por || '').trim().slice(0, 60),
      actualizado: new Date().toISOString(),
    };
  }
  actual.actualizado = new Date().toISOString();

  await env.ESTADO.put(LLAVE, JSON.stringify(actual));
  return json({ ok: true, puntos: Object.keys(actual.puntos).length });
}
