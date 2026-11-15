import { Heading } from '@astryxdesign/core/Heading';
import type { ReactNode } from 'react';
import './identity.css';

/**
 * The prototype's account card head (prototypes/html/pages/connexion.html `.acc-head`): one h1 at
 * card size, then a short lead in the secondary tone.
 */
export function CardHead({ title, lede }: { title: ReactNode; lede?: ReactNode }) {
  return (
    <header className="acc-head">
      <Heading level={1}>{title}</Heading>
      {lede ? <p className="acc-lede">{lede}</p> : null}
    </header>
  );
}

/** The centred line under a form: "No account yet? Create one", "Back to sign in". */
export function SwitchLine({ children }: { children: ReactNode }) {
  return <p className="acc-switch">{children}</p>;
}
