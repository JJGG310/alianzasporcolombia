import Icono from './Icono';

interface Props {
  licencia?: string;
  politica?: string;
}

const REPO = 'https://github.com/JJGG310/alianzasporcolombia';

/**
 * Pie de página — tres columnas en escritorio, apiladas en celular.
 *
 * Era un bloque de párrafos sueltos donde lo importante (no pedimos plata) y lo
 * administrativo (licencias) pesaban igual. Acá lo primero encabeza su propia
 * columna y lo demás se ordena por para-qué-sirve.
 *
 * Los datos abiertos son la API pública del proyecto y hasta ahora no estaban
 * enlazados en ninguna parte del sitio: si alguien quiere replicar esto en otra
 * ciudad mañana, este pie es el único sitio donde va a encontrar los archivos.
 */
export default function PieDePagina({ licencia, politica }: Props) {
  return (
    <footer className="pie-marco">
      <div className="pie-marco__interior contenedor">
        <section className="pie-marco__col pie-marco__col--ancha">
          <h2 className="pie-marco__titulo">Alianzas por Colombia</h2>
          <p className="pie-marco__texto">
            Iniciativa ciudadana sin ánimo de lucro. No recibimos donaciones
            directamente ni publicamos cuentas: enlazamos a organizaciones verificadas —{' '}
            <strong>valida siempre con ellas antes de entregar dinero</strong>.
          </p>
          {politica && <p className="pie-marco__texto">{politica}</p>}
          <p className="pie-marco__texto">
            Cada punto y cada cifra lleva su fuente y su fecha de corte. Si encuentras un
            dato vencido o equivocado, escríbenos: corregirlo rápido es la mitad del
            trabajo.
          </p>
          <p className="pie-marco__texto">
            📶 <strong>¿Poca señal?</strong>{' '}
            <a href="/lite.html">Versión liviana de solo texto (27&nbsp;KB)</a> — carga
            hasta en 2G y se puede guardar para leer sin conexión.
          </p>
        </section>

        <section className="pie-marco__col">
          <h2 className="pie-marco__titulo">Datos abiertos</h2>
          <ul className="pie-marco__enlaces">
            <li>
              <a href="/datos/ayuda.json">
                <span>Puntos curados</span>
                <code>/datos/ayuda.json</code>
              </a>
            </li>
            <li>
              <a href="/extracted-data/puntos.json">
                <span>Puntos agregados</span>
                <code>/extracted-data/puntos.json</code>
              </a>
            </li>
            <li>
              <a href="/extracted-data/puntos.geojson">
                <span>Puntos en GeoJSON</span>
                <code>/extracted-data/puntos.geojson</code>
              </a>
            </li>
          </ul>
        </section>

        <section className="pie-marco__col">
          <h2 className="pie-marco__titulo">Código y licencias</h2>
          <ul className="pie-marco__enlaces">
            <li>
              <a href={REPO} target="_blank" rel="noopener noreferrer">
                <span>
                  Código fuente <Icono nombre="enlace" tam="0.9em" />
                </span>
                <code>github.com/JJGG310/alianzasporcolombia</code>
              </a>
            </li>
          </ul>
          <p className="pie-marco__texto">Código: MIT.</p>
          <p className="pie-marco__texto">Datos: {licencia ?? 'CC BY 4.0'}</p>
        </section>
      </div>
    </footer>
  );
}
