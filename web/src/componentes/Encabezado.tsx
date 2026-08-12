import Icono from './Icono';
import LineasEmergencia from './LineasEmergencia';

interface Props {
  /** Fecha de corte legible de los datos curados. Puede venir vacía. */
  actualizado: string;
}

/**
 * Encabezado — barra fija de 52 px.
 *
 * Antes era un bloque rojo con dos líneas de subtítulo que, sumado al aviso
 * ámbar, se comía el primer golpe de vista de un celular: el mapa arrancaba
 * fuera de pantalla. Esto es una herramienta, no una portada, así que la marca
 * ocupa lo que ocupa un nombre y el resto del alto es del contenido.
 *
 * La barra es fija porque el 123 tiene que estar a un toque en cualquier punto
 * del scroll, no solo arriba del todo. La marca va en tinta: el rojo de este
 * sistema significa URGENTE, y un logotipo no es una urgencia.
 *
 * Lo explicativo (qué es esto, de qué terremoto hablamos, el aviso de fuentes
 * oficiales) baja a una tira compacta debajo de la barra, que sí hace scroll.
 */
export default function Encabezado({ actualizado }: Props) {
  return (
    <>
      <a className="saltar" href="#herramienta">
        Saltar al mapa de ayuda
      </a>

      <header className="barra">
        <div className="barra__interior contenedor">
          <h1 className="marca">
            Alianzas<span className="marca__resto"> por Colombia</span>
          </h1>
          <LineasEmergencia />
        </div>
      </header>

      <div className="contexto contenedor">
        <p className="contexto__lede">
          Información ciudadana centralizada · Terremoto M 7,4 — 10 de agosto de 2026 ·
          Del Valle del Cauca para todo el país
        </p>
        <p className="contexto__aviso" role="note">
          <Icono nombre="alerta" />
          <span>
            Verifica siempre con las fuentes oficiales. Este sitio recopila y enlaza; las
            cifras y órdenes de evacuación las emiten las autoridades.
            {actualizado ? (
              <span className="contexto__corte"> Datos actualizados: {actualizado}</span>
            ) : null}
          </span>
        </p>
      </div>
    </>
  );
}
