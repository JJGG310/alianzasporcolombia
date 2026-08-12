interface Props {
  licencia?: string;
  politica?: string;
}

export default function PieDePagina({ licencia, politica }: Props) {
  return (
    <footer className="pie">
      <p>
        Alianzas por Colombia es una iniciativa ciudadana sin ánimo de lucro. No recibimos
        donaciones directamente ni publicamos cuentas: enlazamos a organizaciones
        verificadas —{' '}
        <strong>valida siempre con ellas antes de entregar dinero</strong>.
      </p>
      {politica && <p>{politica}</p>}
      <p>
        Cada punto y cada cifra lleva su fuente y su fecha de corte. Si encuentras un dato
        vencido o equivocado, escríbenos: corregirlo rápido es la mitad del trabajo.
      </p>
      <p>Datos abiertos {licencia ?? 'CC BY 4.0'} · Código MIT.</p>
      <ul>
        <li>
          <a href="/datos/ayuda.json">Datos curados (JSON)</a>
        </li>
        <li>
          <a href="/extracted-data/puntos.json">Datos agregados (JSON)</a>
        </li>
        <li>
          <a
            href="https://github.com/JJGG310/alianzasporcolombia"
            target="_blank"
            rel="noopener noreferrer"
          >
            Código fuente
          </a>
        </li>
      </ul>
    </footer>
  );
}
