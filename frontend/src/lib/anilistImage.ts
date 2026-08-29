export function anilistExtraLargeCover(coverUrl: string): string {
  if (!coverUrl.includes('/cover/large/')) {
    return coverUrl
  }
  return coverUrl.replace('/cover/large/', '/cover/extraLarge/')
}

export function libraryHeroBackgroundImage(
  watchingItem: {bannerImage?: string; coverImage: string} | null,
  show: {coverImage: string} | null | undefined,
): string {
  if (watchingItem?.bannerImage) {
    return watchingItem.bannerImage
  }
  const coverUrl = show?.coverImage ?? watchingItem?.coverImage ?? ''
  return anilistExtraLargeCover(coverUrl)
}
