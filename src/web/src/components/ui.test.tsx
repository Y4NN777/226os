import { Button } from '@astryxdesign/core/Button';
import { screen, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it, vi } from 'vitest';
import { renderWithProviders, setLanguages } from '../test/render';
import { Avatar, initials, tone } from './Avatar';
import { Breadcrumbs } from './Breadcrumbs';
import { CountryCode } from './CountryCode';
import { EmptyState } from './EmptyState';
import { ClockIcon } from './icons';
import { LangCode } from './LangCode';
import { LinkedItems } from './LinkedItems';
import { List, ListRow } from './List';
import { Section } from './Section';
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

  it('is the page header (not a banner landmark), with the item’s tile, eyebrow, subtitle and facts', () => {
    setLanguages(['fr-FR']);
    renderWithProviders(
      <PageHeader
        title="Wave Pay"
        titleLang="fr"
        headingId="resource-title"
        media={<Avatar name="Wave Pay" variant="tile" size="lg" />}
        eyebrow="API · Paiements"
        subtitle="Wave"
        meta={<StatusBadge kind="verified" />}
      />,
    );
    const heading = screen.getByRole('heading', { level: 1, name: 'Wave Pay' });
    expect(heading).toHaveAttribute('id', 'resource-title');
    expect(heading.closest('[lang]')).toHaveAttribute('lang', 'fr');
    const header = heading.closest('.ui-page-header');
    if (header === null) throw new Error('the title is not inside the page header');
    expect(screen.queryByRole('banner')).toBeNull();
    expect(within(header).getByText('API · Paiements')).toBeInTheDocument();
    expect(within(header).getByText('Wave')).toBeInTheDocument();
    expect(within(header).getByText('Vérifié')).toBeInTheDocument();
    expect(within(header).getByText('WP')).toHaveAttribute('aria-hidden', 'true');
  });
});

describe('Section', () => {
  it('is a region named by its title, with the title at the given level', () => {
    renderWithProviders(
      <Section title="Sources" isPanel>
        <p>Une source</p>
      </Section>,
    );
    const region = screen.getByRole('region', { name: 'Sources' });
    expect(within(region).getByRole('heading', { level: 2, name: 'Sources' })).toBeInTheDocument();
    expect(within(region).getByText('Une source')).toBeInTheDocument();
  });

  it('puts its one-sentence lede under the title', () => {
    renderWithProviders(
      <Section title="Préférences" lede="Thème et langue de l’interface.">
        <p>Contenu</p>
      </Section>,
    );
    const region = screen.getByRole('region', { name: 'Préférences' });
    expect(within(region).getByText('Thème et langue de l’interface.')).toBeInTheDocument();
  });

  it('can be a complementary block with its own heading level and a note beside the title', () => {
    renderWithProviders(
      <Section as="aside" title="Contexte" headingLevel={3} titleAside={<Tag label="FR" />}>
        <p>Burkina Faso</p>
      </Section>,
    );
    const aside = screen.getByRole('complementary', { name: 'Contexte' });
    expect(within(aside).getByRole('heading', { level: 3, name: 'Contexte' })).toBeInTheDocument();
    expect(within(aside).getByText('FR')).toBeInTheDocument();
  });
});

describe('LinkedItems', () => {
  it('lists linked content as tiles whose link is named by the title alone', () => {
    renderWithProviders(
      <LinkedItems
        label="À voir aussi"
        items={[
          { key: 'q-1', kind: 'Question', title: 'Webhook non reçu', href: '/questions/q-1' },
          { key: 'p-1', kind: 'Projet', title: 'kora-sdk', href: '/build/p-1', lang: 'en' },
        ]}
      />,
    );
    const list = screen.getByRole('list', { name: 'À voir aussi' });
    const link = within(list).getByRole('link', { name: 'Webhook non reçu' });
    expect(link).toHaveAttribute('href', '/questions/q-1');
    expect(link.closest('li')).toHaveTextContent('Question');
    expect(within(list).getByRole('link', { name: 'kora-sdk' }).closest('[lang]')).toHaveAttribute(
      'lang',
      'en',
    );
  });
});

describe('LangCode', () => {
  it('shows the content language code, with its name as a hint', () => {
    setLanguages(['fr-FR']);
    renderWithProviders(<LangCode lang="fr" />);
    expect(screen.getByText('FR')).toHaveAttribute('title', 'Langue : français');
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

  it('carries the content language on the title and a subtitle line under it', () => {
    renderWithProviders(
      <List label="Personnes">
        <ListRow title="Awa Traoré" href="/profils/awa" titleLang="fr" subtitle="@awa" />
      </List>,
    );
    const link = screen.getByRole('link', { name: 'Awa Traoré' });
    expect(link.closest('[lang]')).toHaveAttribute('lang', 'fr');
    expect(link.closest('li')).toHaveTextContent('@awa');
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
    ['closed', 'Fermé'],
    ['claimed', 'Déclaré'],
    ['disputed', 'Contesté'],
    ['pending', 'En attente'],
    ['private', 'Privé'],
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
    expect(screen.getByTitle('Donnée d’exemple, pas une activité réelle.')).toBeInTheDocument();
  });

  it('speaks English too', () => {
    setLanguages(['en-US']);
    renderWithProviders(<StatusBadge kind="stale" />);
    expect(screen.getByText('Needs re-checking')).toBeInTheDocument();
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

  it('draws an item as a decorative monogram tile', () => {
    renderWithProviders(<Avatar name="Kalimba Pay API" variant="tile" />);
    expect(screen.getByText('KP')).toHaveAttribute('aria-hidden', 'true');
    expect(screen.queryByRole('img')).not.toBeInTheDocument();
  });
});

// Moved from features/discover/ResourceThumb.test.ts with the tile itself.
describe('Avatar tile initials and tone', () => {
  it('takes the first letter of the first two words', () => {
    expect(initials('Kalimba Pay API')).toBe('KP');
    expect(initials('geojson-lint')).toBe('G');
    expect(initials('Éclair SMS')).toBe('ÉS');
  });

  it('ignores a trailing parenthesis such as « (démo) »', () => {
    expect(initials('Nimba Maps Studio (démo)')).toBe('NM');
    expect(tone('Nimba Maps Studio (démo)')).toBe(tone('Nimba Maps Studio'));
  });

  it('falls back to a placeholder when the name has no letter or digit', () => {
    expect(initials('—')).toBe('?');
  });

  it('keeps the tone within the five token pairs', () => {
    for (const name of ['A', 'Wave Pay', 'SDK SMS Baobab', '']) {
      expect(tone(name)).toBeGreaterThanOrEqual(0);
      expect(tone(name)).toBeLessThan(5);
    }
  });
});

describe('CountryCode', () => {
  it('shows the code and gives assistive technology the country name', () => {
    renderWithProviders(<CountryCode code="BF" name="Burkina Faso" />);
    expect(screen.getByText('BF')).toHaveAttribute('aria-hidden', 'true');
    expect(screen.getByText('Burkina Faso')).toBeInTheDocument();
  });

  it('names the country in the UI language when no name is given, shown once when visible', () => {
    setLanguages(['fr-FR']);
    renderWithProviders(<CountryCode code="CI" withName />);
    expect(screen.getByText('CI')).toHaveAttribute('aria-hidden', 'true');
    expect(screen.getAllByText('Côte d’Ivoire')).toHaveLength(1);
    expect(screen.getByText('Côte d’Ivoire')).toBeVisible();
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

  it('says how many tags it left out', () => {
    renderWithProviders(
      <TagList tags={['go', 'sql']} more={{ count: 2, label: '2 autres technologies' }} />,
    );
    expect(screen.getByText('+2')).toHaveAttribute('aria-hidden', 'true');
    expect(screen.getByText('2 autres technologies')).toBeInTheDocument();
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
