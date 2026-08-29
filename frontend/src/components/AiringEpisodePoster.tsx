type Props = {
  coverImage: string
  size: 'compact' | 'standard' | 'medium' | 'large' | 'xlarge' | 'xxlarge'
}

const sizeClasses = {
  compact: 'h-14 w-10',
  standard: 'h-[4.25rem] w-12',
  medium: 'h-20 w-14',
  large: 'h-24 w-16',
  xlarge: 'h-30 w-20',
  xxlarge: 'h-36 w-24',
} as const

const sizeDimensions = {
  compact: {width: 40, height: 56},
  standard: {width: 48, height: 68},
  medium: {width: 56, height: 80},
  large: {width: 64, height: 96},
  xlarge: {width: 80, height: 120},
  xxlarge: {width: 96, height: 144},
} as const

export function AiringEpisodePoster({coverImage, size}: Props) {
  const className = sizeClasses[size]
  const dimensions = sizeDimensions[size]

  if (coverImage) {
    return (
      <img
        src={coverImage}
        alt=""
        width={dimensions.width}
        height={dimensions.height}
        className={`${className} shrink-0 object-cover`}
      />
    )
  }

  return <span className={`${className} shrink-0 bg-muted`} aria-hidden="true" />
}
