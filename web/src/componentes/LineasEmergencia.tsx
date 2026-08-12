// Siempre arriba y siempre visible: es lo primero que alguien necesita.
const LINEAS = [
  { numero: '123', marcar: '123', que: 'Emergencias (única nacional)' },
  { numero: '132', marcar: '132', que: 'Cruz Roja' },
  { numero: '119', marcar: '119', que: 'Bomberos' },
  { numero: '144', marcar: '144', que: 'Defensa Civil' },
  { numero: '125', marcar: '125', que: 'Emergencias en salud (Valle)' },
  { numero: '#767', marcar: '767', que: 'Estado de vías (INVÍAS)' },
];

export default function LineasEmergencia() {
  return (
    <section id="emergencias" aria-labelledby="t-emergencias">
      <h2 id="t-emergencias">Líneas de emergencia</h2>
      <ul className="lineas">
        {LINEAS.map((l) => (
          <li key={l.numero}>
            <a href={`tel:${l.marcar}`} aria-label={`Llamar al ${l.numero}: ${l.que}`}>
              <strong>{l.numero}</strong>
              <span>{l.que}</span>
            </a>
          </li>
        ))}
      </ul>
    </section>
  );
}
