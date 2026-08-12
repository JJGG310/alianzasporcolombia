interface Props {
  actualizado: string;
}

export default function Encabezado({ actualizado }: Props) {
  return (
    <>
      <a className="saltar" href="#ayuda">
        Saltar al mapa de ayuda
      </a>
      <header className="encabezado">
        <div className="wrap">
          <h1>Alianzas por Colombia</h1>
          <p>
            Información ciudadana centralizada · Terremoto M 7,4 — 10 de agosto de 2026 ·
            Del Valle del Cauca para todo el país
          </p>
        </div>
      </header>
      <p className="aviso" role="note">
        ⚠️ Verifica siempre con las fuentes oficiales. Este sitio recopila y enlaza; las
        cifras y órdenes de evacuación las emiten las autoridades.
        {actualizado ? ` Datos actualizados: ${actualizado}` : ''}
      </p>
    </>
  );
}
