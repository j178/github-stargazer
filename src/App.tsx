import axios from 'axios'
import { type ChangeEvent, type FC, useEffect, useState } from 'react'
import {
  FiBell,
  FiBellOff,
  FiCheckCircle,
  FiExternalLink,
  FiFilter,
  FiGithub,
  FiRefreshCw,
  FiSave,
  FiSend,
  FiShield,
  FiTrash2,
  FiX,
} from 'react-icons/fi'
import { ToastContainer, toast } from 'react-toastify'
import styles from './App.module.css'
import {
  createEmptySettings,
  type Installation,
  type NotifySetting,
  normalizeSettings,
  type RepoInfo,
  type Settings,
} from './models'
import NotificationConfig from './NotificationConfig'
import RepoSelector from './RepoSelector'
import 'react-toastify/dist/ReactToastify.css'

enum ListMode {
  Allow = 'allow',
  Mute = 'mute',
}

const getErrorMessage = (error: unknown, fallback: string) => {
  if (axios.isAxiosError(error)) {
    const message =
      error.response?.data?.error || error.response?.data?.message || error.response?.data?.status || error.message
    if (message) {
      return `${fallback}: ${message}`
    }
  }

  return fallback
}

const buildSettingsPayload = (settings: Settings, selectedRepos: string[], listMode: ListMode): Settings => ({
  ...settings,
  allow_repos: listMode === ListMode.Allow ? selectedRepos : [],
  mute_repos: listMode === ListMode.Mute ? selectedRepos : [],
})

const normalizeNotifySettingSnapshot = (notifySetting: NotifySetting): NotifySetting =>
  Object.fromEntries(
    Object.entries(notifySetting)
      .filter(([, value]) => value)
      .sort(([left], [right]) => left.localeCompare(right))
  ) as NotifySetting

const serializeSettingsSnapshot = (settings: Settings) =>
  JSON.stringify({
    ...settings,
    notify_settings: [...settings.notify_settings]
      .map(normalizeNotifySettingSnapshot)
      .sort((left, right) =>
        JSON.stringify(normalizeNotifySettingSnapshot(left)).localeCompare(
          JSON.stringify(normalizeNotifySettingSnapshot(right))
        )
      ),
    allow_repos: [...settings.allow_repos].sort((left, right) => left.localeCompare(right)),
    mute_repos: [...settings.mute_repos].sort((left, right) => left.localeCompare(right)),
  } satisfies Settings)

const EMPTY_SETTINGS_SNAPSHOT = serializeSettingsSnapshot(createEmptySettings())

const deriveRepoSelection = (settings: Settings) => {
  if (settings.allow_repos.length > 0) {
    return {
      listMode: ListMode.Allow,
      repos: settings.allow_repos,
    }
  }

  return {
    listMode: ListMode.Mute,
    repos: settings.mute_repos,
  }
}

const AccountSelect: FC<{
  installations: Installation[]
  selectedAccount: Installation | null
  isRefreshing: boolean
  onChange: (event: ChangeEvent<HTMLSelectElement>) => void
  onInstallAnother: () => void
}> = ({ installations, selectedAccount, isRefreshing, onChange, onInstallAnother }) => {
  return (
    <div className={`${styles.accountControls} ${styles.workspaceAccountControls}`}>
      <label className={`${styles.selectField} ${styles.workspaceSelectField}`}>
        <span className={styles.fieldLabel}>GitHub account</span>
        <select className={styles.accountSelect} onChange={onChange} value={selectedAccount?.account ?? ''}>
          {installations.length === 0 ? (
            <option value=''>No installations available</option>
          ) : (
            <>
              <option value='' disabled>
                Select an account
              </option>
              {installations.map((installation) => (
                <option key={installation.id} value={installation.account}>
                  {installation.account} ({installation.account_type})
                </option>
              ))}
            </>
          )}
        </select>
      </label>
      <button
        className={`${styles.secondaryButton} ${styles.workspaceInstallButton}`}
        disabled={isRefreshing}
        onClick={onInstallAnother}
        type='button'
      >
        <FiExternalLink />
        Install on another account
      </button>
    </div>
  )
}

const Footer: FC = () => {
  return (
    <footer className={styles.footer}>
      <a href='https://github.com/apps/stars-notifier' rel='noopener noreferrer' target='_blank'>
        Powered by Star++
      </a>
    </footer>
  )
}

const App: FC = () => {
  const [isLoggedIn, setIsLoggedIn] = useState<boolean | null>(null)
  const [isInitializing, setIsInitializing] = useState(true)
  const [isRefreshingAccounts, setIsRefreshingAccounts] = useState(false)
  const [isLoadingAccountData, setIsLoadingAccountData] = useState(false)
  const [installations, setInstallations] = useState<Installation[]>([])
  const [repos, setRepos] = useState<RepoInfo[]>([])
  const [selectedAccount, setSelectedAccount] = useState<Installation | null>(null)
  const [settings, setSettings] = useState<Settings>(createEmptySettings())
  const [listMode, setListMode] = useState(ListMode.Mute)
  const [selectedRepos, setSelectedRepos] = useState<string[]>([])
  const [savedSettingsSnapshot, setSavedSettingsSnapshot] = useState(EMPTY_SETTINGS_SNAPSHOT)
  const [curPage, setPage] = useState(1)
  const [hasMore, setHasMore] = useState(true)
  const [isChecking, setIsChecking] = useState(false)
  const [isTesting, setIsTesting] = useState(false)
  const [isSaving, setIsSaving] = useState(false)
  const [isDeleting, setIsDeleting] = useState(false)

  useEffect(() => {
    let isMounted = true

    const fetchInstallations = async () => {
      setIsRefreshingAccounts(true)
      try {
        const response = await axios.get('/api/installations')
        if (!isMounted) {
          return
        }

        const nextInstallations = response.data as Installation[]
        setInstallations(nextInstallations)
        setIsLoggedIn(true)
        setSelectedAccount((current) => {
          if (current) {
            return (
              nextInstallations.find((installation) => installation.account === current.account) ??
              nextInstallations[0] ??
              null
            )
          }

          return nextInstallations[0] ?? null
        })
      } catch (error) {
        if (!isMounted) {
          return
        }

        if (axios.isAxiosError(error) && error.response?.status === 401) {
          setIsLoggedIn(false)
          setInstallations([])
          setSelectedAccount(null)
        } else {
          toast.error(getErrorMessage(error, 'Failed to load installations'))
        }
      } finally {
        if (isMounted) {
          setIsRefreshingAccounts(false)
          setIsInitializing(false)
        }
      }
    }

    void fetchInstallations()

    return () => {
      isMounted = false
    }
  }, [])

  useEffect(() => {
    if (!selectedAccount) {
      setSettings(createEmptySettings())
      setRepos([])
      setSelectedRepos([])
      setListMode(ListMode.Mute)
      setSavedSettingsSnapshot(EMPTY_SETTINGS_SNAPSHOT)
      setPage(1)
      setHasMore(true)
      return
    }

    let isMounted = true

    const fetchAccountData = async () => {
      setIsLoadingAccountData(true)
      setSettings(createEmptySettings())
      setRepos([])
      setSelectedRepos([])
      setListMode(ListMode.Mute)
      setSavedSettingsSnapshot(EMPTY_SETTINGS_SNAPSHOT)
      setPage(1)
      setHasMore(true)

      try {
        const [settingsResponse, reposResponse] = await Promise.all([
          axios.get(`/api/settings/${selectedAccount.account}`),
          axios.get(`/api/repos/${selectedAccount.id}`),
        ])

        if (!isMounted) {
          return
        }

        const normalizedSettings = normalizeSettings(settingsResponse.data)
        const selection = deriveRepoSelection(normalizedSettings)
        const nextRepos = (reposResponse.data ?? []) as RepoInfo[]

        setSettings(normalizedSettings)
        setSelectedRepos(selection.repos)
        setListMode(selection.listMode)
        setSavedSettingsSnapshot(serializeSettingsSnapshot(normalizedSettings))
        setRepos(nextRepos)
        setHasMore(nextRepos.length > 0)
      } catch (error) {
        if (!isMounted) {
          return
        }

        toast.error(getErrorMessage(error, `Failed to load configuration for ${selectedAccount.account}`))
      } finally {
        if (isMounted) {
          setIsLoadingAccountData(false)
        }
      }
    }

    void fetchAccountData()

    return () => {
      isMounted = false
    }
  }, [selectedAccount])

  const refreshInstallations = async () => {
    setIsRefreshingAccounts(true)
    try {
      const response = await axios.get('/api/installations')
      const nextInstallations = response.data as Installation[]
      setInstallations(nextInstallations)
      setIsLoggedIn(true)
      setSelectedAccount((current) => {
        if (current) {
          return (
            nextInstallations.find((installation) => installation.account === current.account) ??
            nextInstallations[0] ??
            null
          )
        }

        return nextInstallations[0] ?? null
      })
    } catch (error) {
      if (axios.isAxiosError(error) && error.response?.status === 401) {
        setIsLoggedIn(false)
      } else {
        toast.error(getErrorMessage(error, 'Failed to refresh installations'))
      }
    } finally {
      setIsRefreshingAccounts(false)
    }
  }

  const handleInstallAnotherAccount = () => {
    const popup = window.open(
      'https://github.com/apps/stars-notifier/installations/new',
      'github-installation',
      'width=720,height=780'
    )

    if (!popup) {
      window.location.href = 'https://github.com/apps/stars-notifier/installations/new'
      return
    }

    const timer = window.setInterval(() => {
      if (!popup.closed) {
        return
      }

      window.clearInterval(timer)
      void refreshInstallations()
    }, 1000)
  }

  const handleAccountChange = (event: ChangeEvent<HTMLSelectElement>) => {
    const nextAccount = installations.find((installation) => installation.account === event.target.value) ?? null
    setSelectedAccount(nextAccount)
  }

  const handleSelectRepo = (repoName: string) => {
    setSelectedRepos((current) => {
      if (current.includes(repoName)) {
        return current
      }

      return [...current, repoName]
    })
  }

  const handleUnselectRepo = (repoName: string) => {
    setSelectedRepos((current) => current.filter((repo) => repo !== repoName))
  }

  const loadMoreRepos = async () => {
    if (!selectedAccount) {
      return
    }

    const nextPage = curPage + 1
    try {
      const response = await axios.get(`/api/repos/${selectedAccount.id}?page=${nextPage}`)
      const nextRepos = (response.data ?? []) as RepoInfo[]
      if (nextRepos.length === 0) {
        setHasMore(false)
        return
      }

      setRepos((current) => [...current, ...nextRepos])
      setPage(nextPage)
      setHasMore(true)
    } catch (error) {
      toast.error(getErrorMessage(error, 'Failed to load more repositories'))
    }
  }

  const handleValidateSettings = async () => {
    setIsChecking(true)
    try {
      await axios.post('/api/settings/check', buildSettingsPayload(settings, selectedRepos, listMode))
      toast.success('Configuration is valid')
    } catch (error) {
      toast.error(getErrorMessage(error, 'Configuration validation failed'))
    } finally {
      setIsChecking(false)
    }
  }

  const handleTestSettings = async () => {
    setIsTesting(true)
    try {
      await axios.post('/api/settings/test', buildSettingsPayload(settings, selectedRepos, listMode))
      toast.success('Test notification sent')
    } catch (error) {
      toast.error(getErrorMessage(error, 'Failed to send test notification'))
    } finally {
      setIsTesting(false)
    }
  }

  const handleSaveSettings = async () => {
    if (!selectedAccount) {
      return
    }

    const nextSnapshot = serializeSettingsSnapshot(buildSettingsPayload(settings, selectedRepos, listMode))

    setIsSaving(true)
    try {
      await axios.post(
        `/api/settings/${selectedAccount.account}`,
        buildSettingsPayload(settings, selectedRepos, listMode)
      )
      setSavedSettingsSnapshot(nextSnapshot)
      toast.success('Configuration saved')
    } catch (error) {
      toast.error(getErrorMessage(error, 'Failed to save configuration'))
    } finally {
      setIsSaving(false)
    }
  }

  const handleDeleteSettings = async () => {
    if (!selectedAccount) {
      return
    }

    const confirmed = window.confirm(`Delete the saved configuration for ${selectedAccount.account}?`)
    if (!confirmed) {
      return
    }

    setIsDeleting(true)
    try {
      await axios.delete(`/api/settings/${selectedAccount.account}`)
      setSettings(createEmptySettings())
      setSelectedRepos([])
      setListMode(ListMode.Mute)
      setSavedSettingsSnapshot(EMPTY_SETTINGS_SNAPSHOT)
      toast.success('Configuration deleted')
    } catch (error) {
      toast.error(getErrorMessage(error, 'Failed to delete configuration'))
    } finally {
      setIsDeleting(false)
    }
  }

  const availableRepos = repos.filter((repo) => !selectedRepos.includes(repo.name))
  const currentSettings = buildSettingsPayload(settings, selectedRepos, listMode)
  const hasUnsavedChanges = serializeSettingsSnapshot(currentSettings) !== savedSettingsSnapshot
  const isBusy = isChecking || isTesting || isSaving || isDeleting
  const scopeModes = [
    {
      mode: ListMode.Mute,
      eyebrow: 'Recommended default',
      title: 'Mute exceptions',
      description: 'Most repositories can notify. Add only the ones you want to silence.',
      listedLabel: 'Listed repositories',
      listedValue: 'Muted',
      defaultLabel: 'Everything else',
      defaultValue: 'Notify',
    },
    {
      mode: ListMode.Allow,
      eyebrow: 'Tighter control',
      title: 'Allow only',
      description: 'Nothing notifies until it is explicitly added to the list.',
      listedLabel: 'Listed repositories',
      listedValue: 'Notify',
      defaultLabel: 'Everything else',
      defaultValue: 'Blocked',
    },
  ]
  const scopeMeta =
    listMode === ListMode.Allow
      ? {
          browserTitle: 'Choose repositories',
          browserDescription: '',
          listTitle: 'Allowed repositories',
          listDescription: 'Only repositories in this list can send notifications.',
          listBadge: 'Allow list',
          emptyTitle: 'No repositories selected yet',
          emptyText: '',
          listedValue: 'Notify',
          unlistedValue: 'Blocked',
        }
      : {
          browserTitle: 'Choose repositories to mute',
          browserDescription: '',
          listTitle: 'Muted repositories',
          listDescription: '',
          listBadge: 'Mute list',
          emptyTitle: 'No repositories muted',
          emptyText: 'Add repositories only if you want to silence a few noisy projects.',
          listedValue: 'Muted',
          unlistedValue: 'Notify',
        }

  return (
    <div className={styles.page}>
      <div className={styles.shell}>
        <header className={styles.hero}>
          <div className={styles.heroContent}>
            <p className={styles.heroEyebrow}>GitHub Stars Delivery</p>
            <h1 className={styles.heroTitle}>Notification settings for star activity.</h1>
            <p className={styles.heroText}>
              Choose delivery channels, define repository scope, and keep noisy events out of the way.
            </p>
            <div className={styles.heroBadges}>
              <span className={styles.heroBadge}>
                <FiGithub />
                {installations.length} account{installations.length === 1 ? '' : 's'}
              </span>
              <span className={styles.heroBadge}>
                <FiBell />
                {settings.notify_settings.length} channel{settings.notify_settings.length === 1 ? '' : 's'}
              </span>
              <span className={styles.heroBadge}>
                <FiFilter />
                {selectedRepos.length} repo rule{selectedRepos.length === 1 ? '' : 's'}
              </span>
            </div>
          </div>
          <aside className={`${styles.heroCard} ${styles.workspaceCard}`}>
            <div className={styles.workspaceCardTop}>
              <div>
                <span className={styles.heroCardLabel}>Current workspace</span>
                <strong className={styles.heroCardValue}>{selectedAccount?.account ?? 'No account selected'}</strong>
                {!selectedAccount ? (
                  <p className={styles.heroCardText}>
                    Pick a GitHub account or organization to start configuring notifications.
                  </p>
                ) : null}
              </div>
              {isLoggedIn ? (
                <button
                  aria-label='Refresh installations'
                  className={styles.workspaceRefreshButton}
                  disabled={isRefreshingAccounts}
                  onClick={() => void refreshInstallations()}
                  title='Refresh installations'
                  type='button'
                >
                  <FiRefreshCw />
                </button>
              ) : null}
            </div>
            {isLoggedIn ? (
              <AccountSelect
                installations={installations}
                isRefreshing={isRefreshingAccounts}
                onChange={handleAccountChange}
                onInstallAnother={handleInstallAnotherAccount}
                selectedAccount={selectedAccount}
              />
            ) : null}
          </aside>
        </header>

        <main className={styles.main}>
          {isInitializing ? (
            <section className={styles.panel}>
              <div className={styles.loadingState}>Loading configuration workspace…</div>
            </section>
          ) : !isLoggedIn ? (
            <section className={`${styles.panel} ${styles.loginPanel}`}>
              <div>
                <p className={styles.sectionEyebrow}>Authentication</p>
                <h2 className={styles.sectionTitle}>Sign in before editing configuration</h2>
                <p className={styles.sectionText}>
                  The configuration page uses your GitHub session to list installations and save account-specific rules.
                </p>
              </div>
              <a className={styles.primaryLink} href='/api/authorize'>
                <FiGithub />
                Log in with GitHub
              </a>
            </section>
          ) : !selectedAccount ? (
            <section className={styles.panel}>
              <div className={styles.emptyState}>
                <h2 className={styles.emptyStateTitle}>
                  {installations.length === 0 ? 'No installations found' : 'Select an account to continue'}
                </h2>
                <p className={styles.emptyStateText}>
                  {installations.length === 0
                    ? 'Install the GitHub app on a personal account or organization, then refresh the list.'
                    : 'The configuration editor appears once an installation is selected.'}
                </p>
              </div>
            </section>
          ) : (
            <div className={styles.workspaceGrid}>
              <section className={styles.panel}>
                <NotificationConfig settings={settings} setSettings={setSettings} />
              </section>

              <section className={styles.panel}>
                <div className={styles.panelHeader}>
                  <div>
                    <p className={styles.sectionEyebrow}>Repository scope</p>
                    <h2 className={styles.sectionTitle}>Design the delivery policy</h2>
                  </div>
                </div>

                {isLoadingAccountData ? (
                  <div className={styles.loadingState}>Loading repositories and existing rules…</div>
                ) : (
                  <>
                    <div className={styles.scopeModeGrid} role='tablist' aria-label='Repository delivery policy'>
                      {scopeModes.map((scopeMode) => {
                        const isActive = listMode === scopeMode.mode
                        return (
                          <button
                            aria-pressed={isActive}
                            className={
                              isActive ? `${styles.scopeModeCard} ${styles.scopeModeCardActive}` : styles.scopeModeCard
                            }
                            key={scopeMode.mode}
                            onClick={() => setListMode(scopeMode.mode)}
                            type='button'
                          >
                            <span className={styles.scopeModeEyebrow}>{scopeMode.eyebrow}</span>
                            <strong className={styles.scopeModeTitle}>{scopeMode.title}</strong>
                            <span className={styles.scopeModeText}>{scopeMode.description}</span>
                            <div className={styles.scopeModeMeta}>
                              <div className={styles.scopeModeStat}>
                                <span className={styles.scopeModeStatLabel}>{scopeMode.listedLabel}</span>
                                <strong
                                  className={
                                    scopeMode.mode === ListMode.Allow
                                      ? `${styles.scopeModeStatValue} ${styles.scopeModeStatValuePositive}`
                                      : `${styles.scopeModeStatValue} ${styles.scopeModeStatValueMuted}`
                                  }
                                >
                                  {scopeMode.listedValue}
                                </strong>
                              </div>
                              <div className={styles.scopeModeStat}>
                                <span className={styles.scopeModeStatLabel}>{scopeMode.defaultLabel}</span>
                                <strong
                                  className={
                                    scopeMode.mode === ListMode.Allow
                                      ? `${styles.scopeModeStatValue} ${styles.scopeModeStatValueBlocked}`
                                      : `${styles.scopeModeStatValue} ${styles.scopeModeStatValuePositive}`
                                  }
                                >
                                  {scopeMode.defaultValue}
                                </strong>
                              </div>
                            </div>
                          </button>
                        )
                      })}
                    </div>

                    <div className={styles.scopeWorkspace}>
                      <div className={`${styles.scopeBox} ${styles.scopeBrowserPanel}`}>
                        <div className={styles.scopeSectionHeader}>
                          <div>
                            <p className={styles.scopeSectionEyebrow}>Repository browser</p>
                            <h3 className={styles.subsectionTitle}>{scopeMeta.browserTitle}</h3>
                          </div>
                          <span className={styles.countBadge}>{availableRepos.length}</span>
                        </div>
                        {scopeMeta.browserDescription ? (
                          <p className={styles.scopeSectionText}>{scopeMeta.browserDescription}</p>
                        ) : null}
                        <RepoSelector
                          hasMore={hasMore}
                          loadMoreRepos={loadMoreRepos}
                          onSelect={handleSelectRepo}
                          repos={availableRepos}
                        />
                      </div>

                      <div className={`${styles.scopeBox} ${styles.scopeRulesPanel}`}>
                        <div className={styles.scopeSectionHeader}>
                          <div>
                            <p className={styles.scopeSectionEyebrow}>Current policy</p>
                            <h3 className={styles.subsectionTitle}>{scopeMeta.listTitle}</h3>
                          </div>
                          <span
                            className={
                              listMode === ListMode.Allow
                                ? `${styles.scopeListBadge} ${styles.scopeListBadgeAllow}`
                                : `${styles.scopeListBadge} ${styles.scopeListBadgeMute}`
                            }
                          >
                            {scopeMeta.listBadge}
                          </span>
                        </div>
                        <p className={styles.scopeSectionText}>{scopeMeta.listDescription}</p>
                        {selectedRepos.length === 0 ? (
                          <div className={styles.scopeEmptyState}>
                            <strong className={styles.scopeEmptyTitle}>{scopeMeta.emptyTitle}</strong>
                            <p className={styles.scopeEmptyText}>{scopeMeta.emptyText}</p>
                          </div>
                        ) : (
                          <div className={styles.scopeRuleList}>
                            {selectedRepos.map((repo, index) => (
                              <article className={styles.scopeRuleRow} key={repo}>
                                <div className={styles.scopeRuleIdentity}>
                                  <span className={styles.scopeRuleIndex}>{String(index + 1).padStart(2, '0')}</span>
                                  <strong className={styles.scopeRuleName}>{repo}</strong>
                                </div>
                                <button
                                  aria-label={`Remove ${repo}`}
                                  className={styles.scopeRuleRemove}
                                  onClick={() => handleUnselectRepo(repo)}
                                  title={`Remove ${repo}`}
                                  type='button'
                                >
                                  <FiX />
                                </button>
                              </article>
                            ))}
                          </div>
                        )}
                      </div>
                    </div>

                    <label className={styles.preferenceCard}>
                      <div className={styles.preferenceContent}>
                        <span className={styles.preferenceIcon}>
                          <FiBellOff />
                        </span>
                        <h3 className={styles.subsectionTitle}>Mute lost-star notifications</h3>
                        <p className={styles.sectionText}>
                          Keep delivery focused on new stars and ignore unstar events.
                        </p>
                      </div>
                      <input
                        checked={settings.mute_lost_stars}
                        className={styles.preferenceToggle}
                        onChange={(event) =>
                          setSettings((current) => ({
                            ...current,
                            mute_lost_stars: event.target.checked,
                          }))
                        }
                        type='checkbox'
                      />
                    </label>
                  </>
                )}
              </section>
              <aside className={`${styles.panel} ${styles.actionsPanel}`}>
                <div>
                  <p className={styles.sectionEyebrow}>Actions</p>
                  <h2 className={styles.sectionTitle}>Validate and save</h2>
                </div>
                <div className={styles.summaryRow}>
                  <span className={styles.summaryPill}>
                    <FiCheckCircle />
                    {currentSettings.notify_settings.length} destinations configured
                  </span>
                  <span className={styles.summaryPill}>
                    <FiFilter />
                    {currentSettings.allow_repos.length > 0
                      ? `${currentSettings.allow_repos.length} allow-listed repositories`
                      : `${currentSettings.mute_repos.length} muted repositories`}
                  </span>
                </div>
                <div className={styles.actionButtons}>
                  <button
                    className={styles.secondaryButton}
                    disabled={isBusy || isLoadingAccountData}
                    onClick={() => void handleValidateSettings()}
                    type='button'
                  >
                    <FiShield />
                    {isChecking ? 'Validating…' : 'Validate'}
                  </button>
                  <button
                    className={styles.secondaryButton}
                    disabled={isBusy || isLoadingAccountData}
                    onClick={() => void handleTestSettings()}
                    type='button'
                  >
                    <FiSend />
                    {isTesting ? 'Sending…' : 'Send test'}
                  </button>
                  <button
                    className={hasUnsavedChanges ? styles.primaryButton : styles.secondaryButton}
                    disabled={isBusy || isLoadingAccountData}
                    onClick={() => void handleSaveSettings()}
                    type='button'
                  >
                    <FiSave />
                    {isSaving ? 'Saving…' : 'Save configuration'}
                  </button>
                  <button
                    className={`${styles.dangerButton} ${styles.subtleDangerButton}`}
                    disabled={isBusy || isLoadingAccountData}
                    onClick={() => void handleDeleteSettings()}
                    type='button'
                  >
                    <FiTrash2 />
                    {isDeleting ? 'Deleting…' : 'Delete configuration'}
                  </button>
                </div>
              </aside>
            </div>
          )}
        </main>

        <Footer />
      </div>

      <ToastContainer autoClose={3500} closeOnClick newestOnTop position='top-right' />
    </div>
  )
}

export default App
