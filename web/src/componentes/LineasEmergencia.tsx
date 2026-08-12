import { useCallback, useEffect, useRef, useState } from 'react';

import Icono from './Icono';

/**
 * Líneas de emergencia — panel desplegable del encabezado.
 *
 * Antes eran seis tarjetas apiladas en el cuerpo de la página: ocupaban una
 * pantalla entera de un celular y empujaban el mapa —lo que la gente vino a
 * buscar— fuera del primer golpe de vista. Ahora viven detrás de un botón que
 * siempre está a la mano porque la barra es fija, y el 123 queda escrito en el
 * botón: quien no puede leer un menú igual ve el número.
 *
 * `<details>` nativo y no un popover con estado: pesa cero, ya funciona con
 * teclado y con lector de pantalla, y sigue abriendo si el JS falla. Lo único
 * que le agregamos es lo que no trae de fábrica: cerrar con Escape, cerrar al
 * tocar afuera y cerrar al salir con Tab.
 */
const LINEAS = [
  { numero: '123', marcar: '123', que: 'Emergencias (única nacional)' },
  { numero: '132', marcar: '132', que: 'Cruz Roja' },
  { numero: '119', marcar: '119', que: 'Bomberos' },
  { numero: '144', marcar: '144', que: 'Defensa Civil' },
  { numero: '125', marcar: '125', que: 'Emergencias en salud (Valle)' },
  { numero: '#767', marcar: '767', que: 'Estado de vías (INVÍAS)' },
];

export default function LineasEmergencia() {
  const caja = useRef<HTMLDetailsElement>(null);
  const [abierto, setAbierto] = useState(false);

  const cerrar = useCallback((devolverFoco = false) => {
    const d = caja.current;
    if (!d) return;
    d.open = false;
    setAbierto(false);
    if (devolverFoco) d.querySelector('summary')?.focus();
  }, []);

  useEffect(() => {
    if (!abierto) return;

    const alTecla = (e: KeyboardEvent) => {
      // Escape devuelve el foco al botón: si no, el foco se queda en la nada y
      // quien navega con teclado pierde el sitio.
      if (e.key === 'Escape') cerrar(true);
    };
    const afuera = (destino: EventTarget | null) =>
      !(destino instanceof Node) || !caja.current?.contains(destino);
    const alTocar = (e: PointerEvent) => {
      if (afuera(e.target)) cerrar();
    };
    const alEnfocar = (e: FocusEvent) => {
      if (afuera(e.target)) cerrar();
    };

    document.addEventListener('keydown', alTecla);
    document.addEventListener('pointerdown', alTocar);
    document.addEventListener('focusin', alEnfocar);
    return () => {
      document.removeEventListener('keydown', alTecla);
      document.removeEventListener('pointerdown', alTocar);
      document.removeEventListener('focusin', alEnfocar);
    };
  }, [abierto, cerrar]);

  return (
    <details
      className="emergencias"
      ref={caja}
      onToggle={(e) => setAbierto(e.currentTarget.open)}
    >
      <summary className="btn emergencias__boton" aria-label="Líneas de emergencia">
        <Icono nombre="telefono" />
        <span className="emergencias__num">123</span>
        <span className="emergencias__rotulo">Emergencias</span>
        <Icono nombre="flecha-abajo" className="emergencias__flecha" />
      </summary>

      <div className="emergencias__panel aparece">
        <p className="emergencias__titulo" id="t-emergencias">
          Líneas de emergencia
        </p>
        <ul className="emergencias__lista" aria-labelledby="t-emergencias">
          {LINEAS.map((l) => (
            <li key={l.numero}>
              <a
                href={`tel:${l.marcar}`}
                aria-label={`Llamar al ${l.numero}: ${l.que}`}
                onClick={() => cerrar()}
              >
                <span className="emergencias__linea-num">{l.numero}</span>
                <span className="emergencias__linea-que">{l.que}</span>
              </a>
            </li>
          ))}
        </ul>
      </div>
    </details>
  );
}
