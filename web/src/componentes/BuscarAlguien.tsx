import Icono from './Icono';
import '../estilos/paridad.css';

/**
 * ¿Estás buscando a alguien? — una PUERTA, no una base de datos.
 *
 * Este componente no hace fetch de nada y no muestra ni un nombre. Es a
 * propósito y es el principio #2 del proyecto:
 *
 *   1. Los datos de una persona desaparecida son datos personales sensibles
 *      (Ley 1581 de 2012). Publicarlos o almacenarlos nos obliga a cosas que no
 *      podemos garantizar: consentimiento, corrección, borrado.
 *   2. Ya existe un estándar de facto —Colombia Te Busca— que las autoridades
 *      señalan como canal. Una base más no ayuda: parte la búsqueda de una
 *      persona en cinco registros que no se hablan, y hace que un «ya apareció»
 *      llegue a uno solo de ellos.
 *
 * Por eso: explicamos los tres caminos, decimos para qué sirve cada uno, y
 * enlazamos. Nada más.
 *
 * Las URLs y los teléfonos salen de RECURSOS.md y del sitio en vivo
 * (public/index.html). Ninguno es inventado.
 */

interface Camino {
  n: number;
  nombre: string;
  url: string;
  /** Ciudadano u oficial: dice qué esperar del otro lado. */
  clase: 'ciudadano' | 'oficial';
  /** Qué hace la herramienta. */
  hace: string;
  /** En qué caso conviene usar esta y no otra. */
  cuando: string;
  /** Contactos directos, solo cuando el camino es un canal telefónico. */
  contactos?: Array<{ texto: string; href: string }>;
}

const CAMINOS: Camino[] = [
  {
    n: 1,
    nombre: 'Colombia Te Busca',
    url: 'https://www.colombiatebusca.com',
    clase: 'ciudadano',
    hace: 'El registro ciudadano más grande, con más de 5.100 personas reportadas. Buscas por nombre o documento y, si no aparece, la reportas: queda una ficha que puedes compartir por WhatsApp.',
    cuando:
      'Empieza aquí. Ten a mano nombre completo, edad, documento, el último lugar donde se le vio y una foto.',
  },
  {
    n: 2,
    nombre: 'Encontrados.co',
    url: 'https://encontrados.co',
    clase: 'ciudadano',
    hace: 'Cruza con inteligencia artificial las fotos de personas ya rescatadas contra los reportes de desaparecidos.',
    cuando:
      'Útil si la persona pudo haber sido llevada sin identificar a un hospital o a un albergue.',
  },
  {
    n: 3,
    nombre: 'Cruz Roja — Restablecimiento de Contactos Familiares',
    url: 'https://www.cruzrojacolombiana.org/cruz-roja-colombiana-despliega-capacidades-para-dar-respuesta-tras-las-afectaciones-por-sismo-en-colombia/',
    clase: 'oficial',
    hace: 'El canal humanitario oficial para restablecer el contacto entre familiares separados por la emergencia.',
    cuando:
      'Cuando necesitas que una institución gestione el caso, no solo un registro en línea.',
    contactos: [
      { texto: 'Línea 132', href: 'tel:132' },
      { texto: 'WhatsApp 321 213 9525', href: 'https://wa.me/573212139525' },
      { texto: 'rcf@cruzrojacolombiana.org', href: 'mailto:rcf@cruzrojacolombiana.org' },
      { texto: 'Personería de Cali, 24/7: 318 335 5722', href: 'tel:3183355722' },
    ],
  },
];

export default function BuscarAlguien() {
  return (
    <section id="buscar-alguien" className="busca" aria-labelledby="t-busca">
      <h2 id="t-busca">¿Estás buscando a alguien?</h2>

      <p className="prosa busca__intro">
        Tres caminos, cada uno sirve para algo distinto. Puedes usar los tres a la vez.
      </p>

      {/* La razón de ser del componente, dicha en voz alta y no escondida en un
          comentario del código: quien llega acá merece saber por qué lo estamos
          mandando a otro lado. */}
      <p className="prosa busca__porque">
        <strong>Aquí no guardamos ni publicamos datos de personas desaparecidas.</strong>{' '}
        Enlazamos a los registros que ya existen en vez de abrir uno más, para que la
        búsqueda de una persona no quede partida en cinco bases distintas que no se
        hablan entre sí —y para que un «ya apareció» le llegue a todo el mundo.
      </p>

      <ol className="busca__lista">
        {CAMINOS.map((c) => (
          <li key={c.url} className="busca__item tarjeta">
            <p className="busca__cabeza">
              <a
                className="busca__enlace"
                href={c.url}
                target="_blank"
                rel="noopener noreferrer"
              >
                <span className="busca__num" aria-hidden="true">
                  {c.n}
                </span>
                {c.nombre}
                <Icono nombre="enlace" />
              </a>
              <span className={`sello sello--${c.clase === 'oficial' ? 'cubierto' : 'agregado'}`}>
                {c.clase === 'oficial' ? 'canal oficial' : 'ciudadano'}
              </span>
            </p>

            <p className="busca__hace">{c.hace}</p>
            <p className="busca__cuando">{c.cuando}</p>

            {c.contactos && (
              <ul className="busca__contactos">
                {c.contactos.map((k) => (
                  <li key={k.href}>
                    <a className="btn btn--sm" href={k.href} target={k.href.startsWith('http') ? '_blank' : undefined} rel="noopener noreferrer">
                      <Icono nombre={k.href.startsWith('tel:') ? 'telefono' : k.href.startsWith('https://wa.me') ? 'whatsapp' : 'enlace'} />
                      {k.texto}
                    </a>
                  </li>
                ))}
              </ul>
            )}
          </li>
        ))}
      </ol>

      <p className="prosa busca__cifras">
        <strong>Sobre las cifras que ves en las noticias:</strong> los reportes oficiales
        hablan de unos 195 desaparecidos y los registros ciudadanos de varios miles. No se
        contradicen: los primeros son casos verificados por las autoridades y los segundos
        son reportes abiertos, donde muchas personas ya aparecieron pero nadie ha cerrado
        el reporte. Si reportaste a alguien y ya apareció, márcalo como localizado: eso
        ayuda a todos.
      </p>
    </section>
  );
}
