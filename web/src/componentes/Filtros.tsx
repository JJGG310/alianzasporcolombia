import type { Filtro } from '../types';
import { estadoLabel, tipoUI } from '../data/catalogo';

interface Props {
  filtro: Filtro;
  cambiar: (parcial: Partial<Filtro>) => void;
  limpiar: () => void;
  /** Opciones derivadas de los datos reales, no de la lista canónica completa. */
  tipos: string[];
  departamentos: string[];
  municipios: string[];
  estados: string[];
  /** Cuántos puntos hay escondidos por traer alerta de política. */
  totalEnRevision: number;
  mostrados: number;
  total: number;
}

export default function Filtros({
  filtro,
  cambiar,
  limpiar,
  tipos,
  departamentos,
  municipios,
  estados,
  totalEnRevision,
  mostrados,
  total,
}: Props) {
  const hayFiltro =
    filtro.texto !== '' ||
    filtro.tipo !== '' ||
    filtro.departamento !== '' ||
    filtro.municipio !== '' ||
    filtro.estado !== '' ||
    filtro.soloVerificados ||
    filtro.requiereRevision;

  return (
    <div className="filtros-caja">
      <form
        className="filtros"
        role="search"
        aria-label="Filtrar puntos de ayuda"
        onSubmit={(e) => e.preventDefault()}
      >
        <label className="sr-solo" htmlFor="f-texto">
          Buscar por nombre, barrio, dirección o qué ofrece
        </label>
        <input
          id="f-texto"
          type="search"
          placeholder="Buscar: albergue, Cali, agua, Comuna 2…"
          value={filtro.texto}
          onChange={(e) => cambiar({ texto: e.target.value })}
        />

        <label className="sr-solo" htmlFor="f-tipo">
          Tipo de punto
        </label>
        <select
          id="f-tipo"
          value={filtro.tipo}
          onChange={(e) => cambiar({ tipo: e.target.value })}
        >
          <option value="">Todos los tipos</option>
          {tipos.map((t) => (
            <option key={t} value={t}>
              {tipoUI(t).label}
            </option>
          ))}
        </select>

        <label className="sr-solo" htmlFor="f-depto">
          Departamento
        </label>
        <select
          id="f-depto"
          value={filtro.departamento}
          onChange={(e) => cambiar({ departamento: e.target.value, municipio: '' })}
        >
          <option value="">Todos los departamentos</option>
          {departamentos.map((d) => (
            <option key={d} value={d}>
              {d}
            </option>
          ))}
        </select>

        <label className="sr-solo" htmlFor="f-municipio">
          Municipio
        </label>
        <select
          id="f-municipio"
          value={filtro.municipio}
          onChange={(e) => cambiar({ municipio: e.target.value })}
        >
          <option value="">Todos los municipios</option>
          {municipios.map((m) => (
            <option key={m} value={m}>
              {m}
            </option>
          ))}
        </select>

        <label className="sr-solo" htmlFor="f-estado">
          Estado
        </label>
        <select
          id="f-estado"
          value={filtro.estado}
          onChange={(e) => cambiar({ estado: e.target.value })}
        >
          <option value="">Todos los estados</option>
          {estados.map((e) => (
            <option key={e} value={e}>
              {estadoLabel(e)}
            </option>
          ))}
        </select>

        <label className="casilla" htmlFor="f-verificados">
          <input
            id="f-verificados"
            type="checkbox"
            checked={filtro.soloVerificados}
            onChange={(e) => cambiar({ soloVerificados: e.target.checked })}
          />
          Solo verificados a mano
        </label>

        {totalEnRevision > 0 && (
          <label
            className="casilla"
            htmlFor="f-revision"
            title="Puntos donde el saneamiento automático detectó algo que un humano debe revisar (por ejemplo, posibles datos de pago). Están fuera de la vista pública."
          >
            <input
              id="f-revision"
              type="checkbox"
              checked={filtro.requiereRevision}
              onChange={(e) => cambiar({ requiereRevision: e.target.checked })}
            />
            Ver los que requieren revisión ({totalEnRevision})
          </label>
        )}

        {hayFiltro && (
          <button type="button" className="boton" onClick={limpiar}>
            Limpiar filtros
          </button>
        )}
      </form>

      <p className="resumen-filtros" aria-live="polite">
        <span>
          <strong>{mostrados}</strong> de {total} puntos
        </span>
        {filtro.requiereRevision && (
          <span>Estás viendo la cola de revisión: estos datos NO están verificados.</span>
        )}
      </p>
    </div>
  );
}
