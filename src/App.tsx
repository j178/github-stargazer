import axios from 'axios'
import { type ChangeEvent, type FC, useEffect, useState } from 'react'
import {
  FiBell,
  FiCheckCircle,
  FiExternalLink,
  FiFilter,
  FiGithub,
  FiRefreshCw,
  FiSave,
  FiSend,
  FiShield,
  FiTrash2,
} from 'react-icons/fi'
import { ToastContainer, toast } from 'react-toastify'
import styles from './App.module.css'
import { createEmptySettings, type Installation, normalizeSettings, type RepoInfo, type Settings } from './models'
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
    <div className={styles.accountControls}>
      <label className={styles.selectField}>
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
      <button className={styles.secondaryButton} disabled={isRefreshing} onClick={onInstallAnother} type='button'>
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

    setIsSaving(true)
    try {
      await axios.post(
        `/api/settings/${selectedAccount.account}`,
        buildSettingsPayload(settings, selectedRepos, listMode)
      )
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
      toast.success('Configuration deleted')
    } catch (error) {
      toast.error(getErrorMessage(error, 'Failed to delete configuration'))
    } finally {
      setIsDeleting(false)
    }
  }

  const availableRepos = repos.filter((repo) => !selectedRepos.includes(repo.name))
  const currentSettings = buildSettingsPayload(settings, selectedRepos, listMode)
  const isBusy = isChecking || isTesting || isSaving || isDeleting
  const repoScopeHint =
    listMode === ListMode.Allow ? 'Only selected repositories can notify.' : 'Selected repositories are muted.'
  const selectedRepoTitle = listMode === ListMode.Allow ? 'Allowed repositories' : 'Muted repositories'
  const selectedRepoStateLabel = listMode === ListMode.Allow ? 'Allowed' : 'Muted'
  const selectedRepoEmptyMessage =
    listMode === ListMode.Allow
      ? 'No repositories have been allow-listed yet. Add items from the list above.'
      : 'No repositories are muted yet. Add items from the list above if you want to silence them.'

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
          <aside className={styles.heroCard}>
            <span className={styles.heroCardLabel}>Current workspace</span>
            <strong className={styles.heroCardValue}>{selectedAccount?.account ?? 'No account selected'}</strong>
            <p className={styles.heroCardText}>
              {selectedAccount
                ? 'Changes on this page apply only to the selected installation.'
                : 'Pick a GitHub account or organization to start configuring notifications.'}
            </p>
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
          ) : (
            <>
              <section className={styles.panel}>
                <div className={styles.panelHeader}>
                  <div>
                    <p className={styles.sectionEyebrow}>Account</p>
                    <h2 className={styles.sectionTitle}>Choose an installation</h2>
                    <p className={styles.sectionText}>
                      Each personal account and organization keeps an independent notification policy.
                    </p>
                  </div>
                  <button
                    className={styles.ghostButton}
                    disabled={isRefreshingAccounts}
                    onClick={() => void refreshInstallations()}
                    type='button'
                  >
                    <FiRefreshCw />
                    Refresh
                  </button>
                </div>
                <AccountSelect
                  installations={installations}
                  isRefreshing={isRefreshingAccounts}
                  onChange={handleAccountChange}
                  onInstallAnother={handleInstallAnotherAccount}
                  selectedAccount={selectedAccount}
                />
              </section>

              {!selectedAccount ? (
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
                <>
                  <div className={styles.workspaceGrid}>
                    <section className={styles.panel}>
                      <NotificationConfig settings={settings} setSettings={setSettings} />
                    </section>

                    <section className={styles.panel}>
                      <div className={styles.panelHeader}>
                        <div>
                          <p className={styles.sectionEyebrow}>Repository scope</p>
                          <h2 className={styles.sectionTitle}>Choose which repositories can notify</h2>
                          <p className={styles.sectionText}>
                            Use a mute list to silence specific repositories, or switch to an allow list for a tighter
                            setup.
                          </p>
                        </div>
                        <div className={styles.modeSwitch} role='tablist' aria-label='Repository list mode'>
                          <button
                            aria-pressed={listMode === ListMode.Mute}
                            className={listMode === ListMode.Mute ? styles.modeButtonActive : styles.modeButton}
                            onClick={() => setListMode(ListMode.Mute)}
                            type='button'
                          >
                            Mute list
                          </button>
                          <button
                            aria-pressed={listMode === ListMode.Allow}
                            className={listMode === ListMode.Allow ? styles.modeButtonActive : styles.modeButton}
                            onClick={() => setListMode(ListMode.Allow)}
                            type='button'
                          >
                            Allow list
                          </button>
                        </div>
                      </div>

                      {isLoadingAccountData ? (
                        <div className={styles.loadingState}>Loading repositories and existing rules…</div>
                      ) : (
                        <>
                          <div
                            className={
                              listMode === ListMode.Allow
                                ? `${styles.scopeHint} ${styles.scopeHintAllow}`
                                : `${styles.scopeHint} ${styles.scopeHintMute}`
                            }
                          >
                            <span
                              className={
                                listMode === ListMode.Allow
                                  ? `${styles.selectedRepoState} ${styles.selectedRepoStateAllow}`
                                  : `${styles.selectedRepoState} ${styles.selectedRepoStateMute}`
                              }
                            >
                              {selectedRepoStateLabel}
                            </span>
                            <span className={styles.scopeHintText}>{repoScopeHint}</span>
                          </div>

                          <RepoSelector
                            hasMore={hasMore}
                            loadMoreRepos={loadMoreRepos}
                            onSelect={handleSelectRepo}
                            repos={availableRepos}
                          />

                          <div className={styles.selectedReposSection}>
                            <div className={styles.selectedReposHeader}>
                              <h3 className={styles.subsectionTitle}>{selectedRepoTitle}</h3>
                              <span className={styles.countBadge}>{selectedRepos.length}</span>
                            </div>
                            {selectedRepos.length === 0 ? (
                              <div className={styles.inlineEmptyState}>{selectedRepoEmptyMessage}</div>
                            ) : (
                              <div className={styles.selectedReposList}>
                                {selectedRepos.map((repo) => (
                                  <article
                                    className={
                                      listMode === ListMode.Allow
                                        ? `${styles.selectedRepoCard} ${styles.selectedRepoCardAllow}`
                                        : `${styles.selectedRepoCard} ${styles.selectedRepoCardMute}`
                                    }
                                    key={repo}
                                  >
                                    <div className={styles.selectedRepoInfo}>
                                      <strong className={styles.selectedRepoName}>{repo}</strong>
                                    </div>
                                    <div className={styles.selectedRepoMeta}>
                                      <span
                                        className={
                                          listMode === ListMode.Allow
                                            ? `${styles.selectedRepoState} ${styles.selectedRepoStateAllow}`
                                            : `${styles.selectedRepoState} ${styles.selectedRepoStateMute}`
                                        }
                                      >
                                        {selectedRepoStateLabel}
                                      </span>
                                      <button
                                        aria-label={`Remove ${repo}`}
                                        className={styles.repoChipButton}
                                        onClick={() => handleUnselectRepo(repo)}
                                        type='button'
                                      >
                                        Remove
                                      </button>
                                    </div>
                                  </article>
                                ))}
                              </div>
                            )}
                          </div>

                          <label className={styles.preferenceCard}>
                            <div>
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
                  </div>

                  <section className={`${styles.panel} ${styles.actionsPanel}`}>
                    <div>
                      <p className={styles.sectionEyebrow}>Actions</p>
                      <h2 className={styles.sectionTitle}>Validate, test, and save</h2>
                      <p className={styles.sectionText}>
                        Validate checks the structure, test sends a sample notification, and save stores the current
                        account policy.
                      </p>
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
                        className={styles.primaryButton}
                        disabled={isBusy || isLoadingAccountData}
                        onClick={() => void handleSaveSettings()}
                        type='button'
                      >
                        <FiSave />
                        {isSaving ? 'Saving…' : 'Save configuration'}
                      </button>
                      <button
                        className={styles.dangerButton}
                        disabled={isBusy || isLoadingAccountData}
                        onClick={() => void handleDeleteSettings()}
                        type='button'
                      >
                        <FiTrash2 />
                        {isDeleting ? 'Deleting…' : 'Delete saved configuration'}
                      </button>
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
                  </section>
                </>
              )}
            </>
          )}
        </main>

        <Footer />
      </div>

      <ToastContainer autoClose={3500} closeOnClick newestOnTop position='top-right' />
    </div>
  )
}

export default App
