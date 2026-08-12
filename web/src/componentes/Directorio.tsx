import { RECURSOS } from '../data/recursos';

/** Secciones de enlaces verificados, igual que en el sitio original.
 *  Editar src/data/recursos.ts ES actualizar la página. */
export default function Directorio() {
  return (
    <>
      {RECURSOS.secciones.map((sec) => (
        <section key={sec.id} id={sec.id} aria-labelledby={`t-${sec.id}`}>
          <h2 id={`t-${sec.id}`}>{sec.titulo}</h2>
          <p className="desc">{sec.desc}</p>

          {sec.id === 'donaciones' && (
            <p className="aviso" role="note" style={{ borderRadius: 'var(--radio)' }}>
              💸 <strong>Nunca publicamos cuentas bancarias ni datos de pago.</strong>{' '}
              Contacta a la organización y <strong>valida con ella</strong> antes de
              entregar dinero. Alianzas por Colombia no recibe donaciones.
            </p>
          )}

          <ul className="recursos">
            {sec.items.map((item) => (
              <li key={item.url}>
                <a href={item.url} target="_blank" rel="noopener noreferrer">
                  {item.nombre}
                </a>
                {item.badge && (
                  <span className={`badge ${item.badge.replace(/\s+/g, '-')}`}>
                    {item.badge}
                  </span>
                )}
                <p>{item.desc}</p>
              </li>
            ))}
          </ul>
        </section>
      ))}
    </>
  );
}
