export type NotificationService = 'telegram' | 'discord_webhook' | 'discord_bot' | 'bark' | 'webhook'

export interface Installation {
  id: number
  account: string
  account_type: string
}

export interface RepoInfo {
  id: number
  name: string
  description: string
  fork: boolean
}

export interface NotifySetting {
  service: NotificationService
  [key: string]: string | undefined
}

export interface Settings {
  notify_settings: NotifySetting[]
  allow_repos: string[]
  mute_repos: string[]
  mute_lost_stars: boolean
}

export const notificationServices: NotificationService[] = [
  'telegram',
  'discord_webhook',
  'discord_bot',
  'bark',
  'webhook',
]

const isRecord = (value: unknown): value is Record<string, unknown> =>
  typeof value === 'object' && value !== null && !Array.isArray(value)

const normalizeStringArray = (value: unknown) => {
  if (!Array.isArray(value)) {
    return []
  }

  return value.map((item) => String(item)).filter(Boolean)
}

export const createEmptySettings = (): Settings => ({
  notify_settings: [],
  allow_repos: [],
  mute_repos: [],
  mute_lost_stars: false,
})

export const normalizeNotifySetting = (value: unknown): NotifySetting | null => {
  if (!isRecord(value)) {
    return null
  }

  const service = value.service
  if (typeof service !== 'string' || !(notificationServices as string[]).includes(service)) {
    return null
  }

  const next: NotifySetting = { service: service as NotificationService }

  for (const [key, item] of Object.entries(value)) {
    if (key === 'service' || item == null) {
      continue
    }

    next[key] = String(item)
  }

  return next
}

export const normalizeSettings = (value: unknown): Settings => {
  if (!isRecord(value)) {
    return createEmptySettings()
  }

  return {
    notify_settings: Array.isArray(value.notify_settings)
      ? value.notify_settings
          .map(normalizeNotifySetting)
          .filter((setting): setting is NotifySetting => setting !== null)
      : [],
    allow_repos: normalizeStringArray(value.allow_repos),
    mute_repos: normalizeStringArray(value.mute_repos),
    mute_lost_stars: Boolean(value.mute_lost_stars),
  }
}
