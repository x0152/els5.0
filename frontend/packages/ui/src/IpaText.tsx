import { cn } from './cn.ts'
import { splitIpa } from './phonemes.ts'
import { anchorOf, type PhonemeAnchor } from './PhonemePopover.tsx'

export interface IpaTextProps {
  ipa: string
  onSelect: (symbol: string, anchor: PhonemeAnchor) => void
  className?: string
}

export function IpaText({ ipa, onSelect, className }: IpaTextProps) {
  return (
    <span className={cn('font-mono', className)}>
      {splitIpa(ipa).map((t, i) =>
        t.symbol ? (
          <button
            key={i}
            type="button"
            onClick={(e) => onSelect(t.symbol!, anchorOf(e.currentTarget))}
            className="rounded px-px transition hover:bg-brand-100 hover:text-brand-700"
            title={`About /${t.symbol}/`}
          >
            {t.text}
          </button>
        ) : (
          <span key={i}>{t.text}</span>
        ),
      )}
    </span>
  )
}
