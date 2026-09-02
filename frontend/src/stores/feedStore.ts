import {create} from 'zustand'
import {
  AddRSSFeed,
  ListRSSFeedItems,
  ListRSSFeeds,
  MarkAllRSSFeedItemsSeen,
  MarkRSSFeedItemsSeen,
  PollRSSFeedsNow,
  RemoveRSSFeed,
  SetRSSFeedEnabled,
} from '../../wailsjs/go/main/App'
import {errorMessage} from '../lib/format'
import type {RSSFeedItemView, RSSFeedView} from '../lib/types'

type NoticeFn = (message: string, isError?: boolean) => void

export const FEED_PAGE_SIZE = 20

type FeedState = {
  feeds: RSSFeedView[]
  items: RSSFeedItemView[]
  showNewOnly: boolean
  query: string
  page: number
  total: number
  loading: boolean
  error: string
  setShowNewOnly: (showNewOnly: boolean) => Promise<void>
  setQuery: (query: string) => Promise<void>
  setPage: (page: number) => Promise<void>
  reload: () => Promise<void>
  addFeed: (
    feedURL: string,
    feedTitle: string,
    notice: NoticeFn,
  ) => Promise<boolean>
  removeFeed: (feed: RSSFeedView, notice: NoticeFn) => Promise<void>
  toggleFeed: (feed: RSSFeedView, notice: NoticeFn) => Promise<void>
  pollNow: (notice: NoticeFn) => Promise<void>
  markAllSeen: (notice: NoticeFn) => Promise<void>
  markSeen: (item: RSSFeedItemView, notice: NoticeFn) => Promise<void>
}

export const useFeedStore = create<FeedState>((set, get) => ({
  feeds: [],
  items: [],
  showNewOnly: true,
  query: '',
  page: 1,
  total: 0,
  loading: true,
  error: '',

  setShowNewOnly: async (showNewOnly) => {
    set({showNewOnly, page: 1})
    await get().reload()
  },

  setQuery: async (query) => {
    set({query, page: 1})
    await get().reload()
  },

  setPage: async (page) => {
    set({page})
    await get().reload()
  },

  reload: async () => {
    set({error: ''})
    const {showNewOnly, query, page} = get()
    const offset = (page - 1) * FEED_PAGE_SIZE
    try {
      const [loadedFeeds, pageData] = await Promise.all([
        ListRSSFeeds(),
        ListRSSFeedItems(showNewOnly, query, FEED_PAGE_SIZE, offset),
      ])
      const items = pageData?.items ?? []
      const total = pageData?.total ?? items.length
      if (items.length === 0 && page > 1 && total > 0) {
        set({feeds: loadedFeeds ?? [], page: 1})
        await get().reload()
        return
      }
      set({
        feeds: loadedFeeds ?? [],
        items,
        total,
        loading: false,
      })
    } catch (err) {
      set({error: errorMessage(err), loading: false})
    }
  },

  addFeed: async (feedURL, feedTitle, notice) => {
    const trimmedURL = feedURL.trim()
    if (!trimmedURL) {
      notice('Enter an RSS feed URL', true)
      return false
    }
    try {
      await AddRSSFeed(trimmedURL, feedTitle.trim())
      notice('Feed subscribed')
      await get().reload()
      return true
    } catch (err) {
      notice(errorMessage(err), true)
      return false
    }
  },

  removeFeed: async (feed, notice) => {
    try {
      await RemoveRSSFeed(feed.id)
      notice('Feed removed')
      await get().reload()
    } catch (err) {
      notice(errorMessage(err), true)
    }
  },

  toggleFeed: async (feed, notice) => {
    try {
      await SetRSSFeedEnabled(feed.id, !feed.enabled)
      await get().reload()
    } catch (err) {
      notice(errorMessage(err), true)
    }
  },

  pollNow: async (notice) => {
    try {
      await PollRSSFeedsNow()
      notice('Feeds polled')
      await get().reload()
    } catch (err) {
      notice(errorMessage(err), true)
    }
  },

  markAllSeen: async (notice) => {
    try {
      await MarkAllRSSFeedItemsSeen()
      await get().reload()
    } catch (err) {
      notice(errorMessage(err), true)
    }
  },

  markSeen: async (item, notice) => {
    try {
      await MarkRSSFeedItemsSeen([item.id])
      await get().reload()
    } catch (err) {
      notice(errorMessage(err), true)
    }
  },
}))
