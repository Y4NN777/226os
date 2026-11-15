import { Avatar as AstryxAvatar } from '@astryxdesign/core/Avatar';

/**
 * A person's initials in a circle (prototype `.avatar`); no remote picture. Next to the
 * person's name (the usual case) it is decorative and hidden from assistive technology; alone,
 * pass `isDecorative={false}` and it is an image named by the person.
 */
export function Avatar({
  name,
  size = 'md',
  isDecorative = true,
}: {
  name: string;
  size?: 'sm' | 'md' | 'lg';
  isDecorative?: boolean;
}) {
  const avatar = <AstryxAvatar name={name} size={size} tooltip={false} className="ui-avatar" />;
  return isDecorative ? (
    <span aria-hidden="true" className="ui-avatar-wrap">
      {avatar}
    </span>
  ) : (
    avatar
  );
}
