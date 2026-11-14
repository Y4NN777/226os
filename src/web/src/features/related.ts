import type { components } from '../api/schema';

type RelatedLink = components['schemas']['RelatedLink'];

/** Canonical frontend route for every content type returned by RelatedLink. */
export function relatedPath(link: Pick<RelatedLink, 'type' | 'id'>): string {
  switch (link.type) {
    case 'question':
      return `/questions/${link.id}`;
    case 'resource':
      return `/discover/${link.id}`;
    case 'project':
      return `/build/${link.id}`;
    case 'article':
      return `/apprendre/articles/${link.id}`;
    case 'series':
      return `/apprendre/formations/${link.id}`;
  }
}
