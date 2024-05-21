export interface Installation {
  id: number
  account: string
  account_type: string
}

export interface RepoInfo {
  id: number
  owner: string
  name: string
  full_name: string
  private: boolean
  description: string
  fork: boolean
  html_url: string
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
