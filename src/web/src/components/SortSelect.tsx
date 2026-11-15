import { Selector } from '@astryxdesign/core/Selector';

/**
 * The sort control of a results line (prototype `select.select` beside the count). The label is
 * hidden; the first option is the default, which pages keep out of the address.
 */
export function SortSelect<T extends string>({
  label,
  options,
  value,
  onChange,
}: {
  label: string;
  options: readonly { value: T; label: string }[];
  value: T;
  onChange: (value: T) => void;
}) {
  return (
    <Selector
      label={label}
      isLabelHidden
      size="sm"
      options={[...options]}
      value={value}
      onChange={(next) => {
        const known = options.find((option) => option.value === next);
        if (known) onChange(known.value);
      }}
    />
  );
}
