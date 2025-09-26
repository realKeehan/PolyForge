import type { OptionDescriptor } from '../app/types';

type TemplateContext = Pick<OptionDescriptor, 'id' | 'title' | 'description' | 'requiresPath'>;

const HTML_ESCAPE_PATTERN = /[&<>"']/g;
const HTML_ESCAPE_LOOKUP: Record<string, string> = {
  '&': '&amp;',
  '<': '&lt;',
  '>': '&gt;',
  '"': '&quot;',
  "'": '&#39;',
};

const escapeHtml = (value: unknown): string => {
  const input = String(value ?? '');
  return input.replace(HTML_ESCAPE_PATTERN, (match) => HTML_ESCAPE_LOOKUP[match] ?? match);
};

const escapeAttribute = (value: unknown): string => escapeHtml(value).replace(/`/g, '&#96;');

export default function renderOption({
  id,
  title,
  description,
  requiresPath,
}: TemplateContext): string {
  const accessibleLabel = requiresPath ? `${title} (Path required)` : title;
  const requiresAttribute = requiresPath ? ' data-requires="true"' : '';
  const descriptionMarkup = description
    ? `<span class="option-card__description">${escapeHtml(description)}</span>`
    : '';
  const badgeMarkup = requiresPath
    ? '<span class="option-card__badge">Path required</span>'
    : '';

  return [
    '<button class="option-card" type="button"',
    ` data-id="${escapeAttribute(id)}"`,
    `${requiresAttribute}`,
    ` aria-label="${escapeAttribute(accessibleLabel)}">`,
    `<span class="option-card__title">${escapeHtml(title)}</span>`,
    descriptionMarkup,
    badgeMarkup,
    '</button>',
  ]
    .filter(Boolean)
    .join('');
}
