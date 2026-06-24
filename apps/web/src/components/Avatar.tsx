type AvatarSize = 'xs' | 'sm' | 'md' | 'lg'

const sizeClass: Record<AvatarSize, string> = {
  xs: 'h-5 w-5 text-[9px]',
  sm: 'h-7 w-7 text-[11px]',
  md: 'h-10 w-10 text-sm',
  lg: 'h-14 w-14 text-lg',
}

const palette = [
  'bg-brand text-white',
  'bg-grass text-white',
  'bg-tangerine text-white',
  'bg-grape text-white',
  'bg-rose text-white',
]

function hashString(seed: string): number {
  let hash = 0
  for (let index = 0; index < seed.length; index += 1) {
    hash = (hash << 5) - hash + seed.charCodeAt(index)
    hash |= 0
  }
  return Math.abs(hash)
}

function initialOf(name: string): string {
  const trimmed = name.trim()
  if (!trimmed) {
    return '?'
  }
  return [...trimmed][0]!.toUpperCase()
}

export type AvatarProps = {
  name: string
  seed?: string
  src?: string | null
  size?: AvatarSize
  className?: string
  ring?: boolean
}

export function Avatar({
  name,
  seed,
  src,
  size = 'md',
  className = '',
  ring = false,
}: AvatarProps) {
  const ringClass = ring ? 'ring-2 ring-surface' : ''
  if (src) {
    return (
      <img
        alt={name}
        className={`${sizeClass[size]} ${ringClass} rounded-full object-cover ${className}`}
        src={src}
      />
    )
  }
  const color = palette[hashString(seed ?? name) % palette.length]
  return (
    <span
      aria-label={name}
      className={`${sizeClass[size]} ${color} ${ringClass} inline-flex items-center justify-center rounded-full font-display font-semibold ${className}`}
      role="img"
    >
      {initialOf(name)}
    </span>
  )
}

export type AvatarPerson = {
  id: string
  name: string
  src?: string | null
}

export type AvatarStackProps = {
  people: AvatarPerson[]
  max?: number
  size?: AvatarSize
}

export function AvatarStack({ people, max = 4, size = 'sm' }: AvatarStackProps) {
  if (people.length === 0) {
    return null
  }
  const shown = people.slice(0, max)
  const overflow = people.length - shown.length
  return (
    <div className="flex items-center">
      <div className="flex -space-x-2">
        {shown.map((person) => (
          <Avatar
            key={person.id}
            name={person.name}
            seed={person.id}
            size={size}
            src={person.src}
            ring
          />
        ))}
      </div>
      {overflow > 0 ? (
        <span className="ml-2 text-xs font-semibold text-muted">
          +{overflow}
        </span>
      ) : null}
    </div>
  )
}
