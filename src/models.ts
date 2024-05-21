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
  service: string
  [key: string]: any
}

export interface Settings {
  notify_settings: NotifySetting[]
  allow_repos: string[]
  mute_repos: string[]
  mute_lost_stars: boolean
}
