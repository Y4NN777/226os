import { ARCS, CITIES, DOTS, MAP } from './globeData';

/**
 * The hero's map: Africa as dots (Natural Earth land, generated offline into `globeData.ts`),
 * with a dense network of arcs between 24 cities. A free map: no sphere, outline or disc.
 * Purely decorative, hidden from assistive technology. Motion is CSS (see landing.css): a comet
 * runs along each arc, with its own duration and start so something is always moving; it stops
 * with prefers-reduced-motion and while the hero is off screen.
 */
export function Globe() {
  return (
    <svg
      className="lp-globe"
      viewBox={`${String(MAP.x)} ${String(MAP.y)} ${String(MAP.width)} ${String(MAP.height)}`}
      aria-hidden="true"
      focusable="false"
    >
      <path className="lp-dots" d={DOTS} fill="none" strokeLinecap="round" strokeWidth="4" />
      <g className="lp-arcs" fill="none" strokeLinecap="round">
        {ARCS.map((d, i) => (
          <g
            key={d}
            style={{
              ['--i' as string]: i,
              // 1.0 to 1.4 s per arc, spread so the comets never beat in step.
              ['--d' as string]: `${String(1000 + ((i * 137) % 401))}ms`,
            }}
          >
            <path className="lp-arc" d={d} pathLength="1" />
            <path className="lp-comet" d={d} pathLength="1" />
          </g>
        ))}
      </g>
      <g className="lp-cities">
        {CITIES.map(([x, y], i) => (
          <g key={`${String(x)}-${String(y)}`} style={{ ['--i' as string]: i }}>
            <circle className="lp-ring" cx={x} cy={y} r="5" />
            <circle className="lp-city" cx={x} cy={y} r="3" />
          </g>
        ))}
      </g>
    </svg>
  );
}
