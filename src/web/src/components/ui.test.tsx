import { Button } from '@astryxdesign/core/Button';
import { screen, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it, vi } from 'vitest';
import { renderWithProviders, setLanguages } from '../test/render';
import { Avatar } from './Avatar';
import { Breadcrumbs } from './Breadcrumbs';
import { CountryCode } from './CountryCode';
import { EmptyState } from './EmptyState';
import { ClockIcon } from './icons';
import { List, ListRow } from './List';
import { LoadingState } from './LoadingState';
import { Meta, MetaItem } from './Meta';
import { PageHeader } from './PageHeader';
import { ResultsMeta } from './ResultsMeta';
import { StatusBadge, type StatusKind } from './StatusBadge';
import { Tag, TagList } from './Tag';

describe('Breadcrumbs', () => {
  it('is a named navigation whose last item is the current page', () => {
    setLanguages(['fr-FR']);
    renderWithProviders(
      <Breadcrumbs
        items={[{ label: 'Découvrir', href: '/discover' }, { label: 'Sahel Pay Devkit' }]}
      />,
    );
    const nav = screen.getByRole('navigation', { name: 'Fil d’Ariane' });
    expect(within(nav).getByRole('link', { name: 'Découvrir' })).toHaveAttribute(
      'href',
      '/discover',
    );
    expect(within(nav).getByText('Sahel Pay Devkit').closest('[aria-current="page"]')).not.toBe(
      null,
    );
    expect(within(nav).queryByRole('link', { name: 'Sahel Pay Devkit' })).not.toBeInTheDocument();
  });
});

describe('PageHeader', () => {
  it('gives the page its one h1, its lead, breadcrumbs and actions', () => {
    setLanguages(['fr-FR']);
    renderWithProviders(
      <PageHeader
        title="Découvrir"
        lead="API, SDK et outils utiles en Afrique."
        breadcrumbs={[{ label: 'Accueil', href: '/' }, { label: 'Découvrir' }]}
        actions={<Button label="Proposer une ressource" />}
      />,
    );
    expect(screen.getAllByRole('heading', { level: 1 })).toHaveLength(1);
    expect(screen.getByRole('heading', { level: 1, name: 'Découvrir' })).toBeInTheDocument();
    expect(screen.getByText('API, SDK et outils utiles en Afrique.')).toBeInTheDocument();
    expect(screen.getByRole('navigation', { name: 'Fil d’Ariane' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Proposer une ressource' })).toBeInTheDocument();
  });

  it('orders the header as the prototype: breadcrumbs, then the title row with its actions, then the lead', () => {
    setLanguages(['fr-FR']);
    const { container } = renderWithProviders(
      <PageHeader
        title="Ressource d’exemple"
        lead="Une ressource du catalogue."
        breadcrumbs={[{ label: 'Découvrir', href: '/discover' }, { label: 'Ressource d’exemple' }]}
        actions={<Button label="Signaler une erreur" />}
      />,
    );
    const trail = screen.getByRole('navigation', { name: 'Fil d’Ariane' });
    const heading = screen.getByRole('heading', { level: 1, name: 'Ressource d’exemple' });
    const action = screen.getByRole('button', { name: 'Signaler une erreur' });
    const lead = screen.getByText('Une ressource du catalogue.');
    const follows = (a: Node, b: Node) =>
      Boolean(a.compareDocumentPosition(b) & Node.DOCUMENT_POSITION_FOLLOWING);
    expect(follows(trail, heading)).toBe(true);
    expect(follows(heading, action)).toBe(true);
    expect(follows(action, lead)).toBe(true);
    // The title and its actions share one row, as `.page-header-row` does in the prototype.
    expect(heading.parentElement).toBe(action.closest('.ui-page-header-actions')?.parentElement);
    // The trail links up and marks the current page as text.
    expect(within(trail).getByRole('link', { name: 'Découvrir' })).toHaveAttribute(
      'href',
      '/discover',
    );
    expect(within(trail).queryByRole('link', { name: 'Ressource d’exemple' })).toBeNull();
    expect(container.querySelector('.ui-page-header')).not.toBeNull();
  });

  it('renders only the title when nothing else is given', () => {
    setLanguages(['fr-FR']);
    renderWithProviders(<PageHeader title="Construire" />);
    expect(screen.getByRole('heading', { level: 1, name: 'Construire' })).toBeInTheDocument();
    expect(screen.queryByRole('navigation')).not.toBeInTheDocument();
  });
});

describe('ResultsMeta', () => {
  it('announces the count politely and keeps the controls next to it', () => {
    renderWithProviders(
      <ResultsMeta summary="12 ressources">
        <Button label="Trier" />
      </ResultsMeta>,
    );
    const summary = screen.getByText('12 ressources');
    expect(summary).toHaveAttribute('aria-live', 'polite');
    expect(screen.getByRole('button', { name: 'Trier' })).toBeInTheDocument();
  });
});

describe('List and ListRow', () => {
  it('lists rows whose title is a heading link to the item', () => {
    setLanguages(['fr-FR']);
    renderWithProviders(
      <List label="Ressources">
        <ListRow
          title="Sahel Pay Devkit"
          href="/discover/r-1"
          lead={<Avatar name="Sahel Pay" />}
          excerpt="Exemples d’intégration et jeu de données de test."
          tags={<TagList tags={['go', 'webhook']} />}
          meta={
            <Meta>
              <MetaItem>Par Sahel Pay</MetaItem>
            </Meta>
          }
          aside={<StatusBadge kind="verified" />}
        />
        <ListRow title="Nimba Maps Studio" href="/discover/r-2" />
      </List>,
    );
    const list = screen.getByRole('list', { name: 'Ressources' });
    const rows = within(list)
      .getAllByRole('listitem')
      .filter((li) => li.parentElement === list);
    expect(rows).toHaveLength(2);
    const first = rows[0] as HTMLElement;
    const heading = within(first).getByRole('heading', { level: 3, name: 'Sahel Pay Devkit' });
    expect(within(heading).getByRole('link', { name: 'Sahel Pay Devkit' })).toHaveAttribute(
      'href',
      '/discover/r-1',
    );
    expect(
      within(first).getByText('Exemples d’intégration et jeu de données de test.'),
    ).toBeVisible();
    expect(within(first).getByText('webhook')).toBeInTheDocument();
    expect(within(first).getByText('Par Sahel Pay')).toBeInTheDocument();
    expect(within(first).getByText('Vérifié')).toBeInTheDocument();
  });

  it('takes the heading level of the page outline', () => {
    renderWithProviders(
      <List label="Questions">
        <ListRow title="Erreur CORS sur Go" href="/questions/q-1" headingLevel={2} />
      </List>,
    );
    expect(
      screen.getByRole('heading', { level: 2, name: 'Erreur CORS sur Go' }),
    ).toBeInTheDocument();
  });

  it('marks the language of the user content on the title and the excerpt', () => {
    renderWithProviders(
      <List label="Questions">
        <ListRow
          title="CORS error in Go"
          href="/questions/q-2"
          excerpt="Preflight fails."
          lang="en"
        />
      </List>,
    );
    expect(screen.getByRole('heading', { name: 'CORS error in Go' })).toHaveAttribute('lang', 'en');
    expect(screen.getByText('Preflight fails.')).toHaveAttribute('lang', 'en');
  });
});

describe('StatusBadge', () => {
  it.each<[StatusKind, string]>([
    ['verified', 'Vérifié'],
    ['stale', 'À revérifier'],
    ['unverified', 'Non vérifié'],
    ['accepted', 'Acceptée'],
    ['answered', 'Répondue'],
    ['later', 'Bientôt'],
    ['demo', 'Démo'],
    ['open', 'Ouvert'],
  ])('%s carries its meaning in text, with a decorative icon', (kind, text) => {
    setLanguages(['fr-FR']);
    const { container } = renderWithProviders(<StatusBadge kind={kind} />);
    expect(screen.getByText(text)).toBeInTheDocument();
    const icon = container.querySelector('svg');
    expect(icon).not.toBeNull();
    expect(icon?.closest('[aria-hidden="true"]')).not.toBeNull();
  });

  it('accepts the page’s own wording', () => {
    setLanguages(['fr-FR']);
    renderWithProviders(<StatusBadge kind="accepted" label="Réponse acceptée" />);
    expect(screen.getByText('Réponse acceptée')).toBeInTheDocument();
    expect(screen.queryByText('Acceptée')).not.toBeInTheDocument();
  });

  it('explains demo data, as DemoBadge does', () => {
    setLanguages(['fr-FR']);
    renderWithProviders(<StatusBadge kind="demo" />);
    expect(screen.getByTitle('Données d’exemple.')).toBeInTheDocument();
  });

  it('speaks English too', () => {
    setLanguages(['en-US']);
    renderWithProviders(<StatusBadge kind="stale" />);
    expect(screen.getByText('Needs recheck')).toBeInTheDocument();
  });
});

describe('Meta', () => {
  it('lines up items with a decorative icon', () => {
    const { container } = renderWithProviders(
      <Meta>
        <MetaItem icon={ClockIcon}>il y a 2 jours</MetaItem>
        <MetaItem>3 réponses</MetaItem>
      </Meta>,
    );
    expect(screen.getByText('il y a 2 jours')).toBeInTheDocument();
    expect(screen.getByText('3 réponses')).toBeInTheDocument();
    expect(container.querySelectorAll('svg')).toHaveLength(1);
    expect(container.querySelector('svg')?.closest('[aria-hidden="true"]')).not.toBeNull();
  });
});

describe('Avatar', () => {
  it('stays out of the way of the name next to it', () => {
    const { container } = renderWithProviders(<Avatar name="Awa Traoré" />);
    // Astryx draws the initials itself; next to a name the picture is hidden from the tree.
    expect(container.querySelector('[aria-hidden="true"] .astryx-avatar')).not.toBeNull();
    expect(screen.queryByRole('img')).not.toBeInTheDocument();
  });

  it('is named by the person when it stands alone', () => {
    renderWithProviders(<Avatar name="Awa Traoré" isDecorative={false} />);
    expect(screen.getByRole('img', { name: 'Awa Traoré' })).toBeInTheDocument();
  });
});

describe('CountryCode', () => {
  it('shows the code and gives assistive technology the country name', () => {
    renderWithProviders(<CountryCode code="BF" name="Burkina Faso" />);
    expect(screen.getByText('BF')).toHaveAttribute('aria-hidden', 'true');
    expect(screen.getByText('Burkina Faso')).toBeInTheDocument();
  });
});

describe('Tag and TagList', () => {
  it('lists tags, as links when an address is given', () => {
    renderWithProviders(
      <TagList
        label="Technologies"
        tags={['go', 'postgresql']}
        hrefFor={(tag) => `/solve?tag=${tag}`}
      />,
    );
    const list = screen.getByRole('list', { name: 'Technologies' });
    expect(within(list).getAllByRole('listitem')).toHaveLength(2);
    expect(within(list).getByRole('link', { name: 'go' })).toHaveAttribute('href', '/solve?tag=go');
  });

  it('is plain text without an address', () => {
    renderWithProviders(<Tag label="webhook" />);
    expect(screen.getByText('webhook')).toBeInTheDocument();
    expect(screen.queryByRole('link')).not.toBeInTheDocument();
  });
});

describe('EmptyState', () => {
  it('names the empty area and offers the next step', async () => {
    const onReset = vi.fn();
    renderWithProviders(
      <EmptyState
        title="Aucune ressource ne correspond"
        description="Essayez un autre pays ou retirez un filtre."
        actions={<Button label="Effacer les filtres" onClick={onReset} />}
      />,
    );
    expect(
      screen.getByRole('heading', { name: 'Aucune ressource ne correspond' }),
    ).toBeInTheDocument();
    expect(screen.getByText('Essayez un autre pays ou retirez un filtre.')).toBeInTheDocument();
    await userEvent
      .setup({ delay: null })
      .click(screen.getByRole('button', { name: 'Effacer les filtres' }));
    expect(onReset).toHaveBeenCalledOnce();
  });
});

describe('LoadingState rows', () => {
  it('draws placeholder rows for a list and still announces loading', () => {
    setLanguages(['fr-FR']);
    const { container } = renderWithProviders(<LoadingState rows={3} />);
    // The providers own a live region too; this one is the list's.
    expect(screen.getByText('Chargement…').closest('[role="status"]')).not.toBeNull();
    expect(container.querySelectorAll('.ui-skeleton-row')).toHaveLength(3);
  });
});
