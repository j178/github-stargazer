import axios, { AxiosResponse } from 'axios'
import React, { useEffect, useState } from 'react'
import { FaTimes } from 'react-icons/fa'
import { GoArrowDown, GoRepo, GoRepoForked } from 'react-icons/go'
import { ToastContainer, toast } from 'react-toastify'
import { Tooltip } from 'react-tooltip'

import NotificationConfig from './NotificationConfig'
import RepoSelector from './RepoSelector'
import { Installation, RepoInfo, Settings } from './models'

import styles from './App.module.css'
import 'react-toastify/dist/ReactToastify.css'
import 'react-tooltip/dist/react-tooltip.css'

type ListMode = "allow" | "mute";

const AccountSelect: React.FC<{
  installations: Installation[]
  selectedAccount: Installation | null
  handleAccountChange: (event: React.ChangeEvent<HTMLSelectElement>) => void
  updateAccountState: () => void
}> = ({ installations, selectedAccount, handleAccountChange, updateAccountState }) => {
  const [, setPopup] = useState<Window | null>(null)

  const handleAddAccount = () => {
    const newPopup = window.open(
      'https://github.com/apps/stars-notifier/installations/new',
      'popup',
      'width=600,height=600'
    )!
    setPopup(newPopup)

    const checkPopup = setInterval(() => {
      if (newPopup.closed) {
        clearInterval(checkPopup)
        setPopup(null)
        // Call a function to update the account state after the user installs the app
        updateAccountState()
      }
    }, 1000)
  }

  const handleChange = (event: React.ChangeEvent<HTMLSelectElement>) => {
    const value = event.target.value
    if (value === 'add-account') {
      handleAddAccount()
    } else {
      handleAccountChange(event)
    }
  }

  return (
    <select id='account-select' onChange={handleChange} value={selectedAccount?.account}>
      <option value='' disabled>
        Select an account
      </option>
      {installations.map((installation) => (
        <option key={installation.id} value={installation.account}>
          {installation.account} ({installation.account_type})
        </option>
      ))}
      <option value='add-account'>Add GitHub Account</option>
    </select>
  )
}

const Header: React.FC = () => {
  return (
    <header>
      <h1 className={styles.title}>Star++ Configuration</h1>
    </header>
  )
}

const Footer: React.FC = () => {
  return (
    <footer>
      <a href='https://github.com/apps/stars-notifier' target='_blank' rel='noopener noreferrer'>
        Powered by <img src='/avatar.png' alt='Star++' className={styles.logo} />
      </a>
    </footer>
  )
}

const App: React.FC = () => {
  const [isLoggedIn, setIsLoggedIn] = useState(true)
  const [installations, setInstallations] = useState<Installation[]>([])
  const [repos, setRepos] = useState<RepoInfo[]>([])
  const [selectedAccount, setSelectedAccount] = useState<Installation | null>(null)
  const [settings, setSettings] = useState<Settings | null>(null)
  const [listMode, setListMode] = useState<ListMode>("mute")
  const [selectedRepos, setSelectedRepos] = useState<RepoInfo[]>([])
  const [curPage, setPage] = useState(1)
  const [hasMore, setHasMore] = useState(true)

  useEffect(() => {
    // Fetch installations on mount
    async function fetchInstallations() {
      try {
        const response = await axios.get('/api/installations')
        setIsLoggedIn(true)
        setInstallations(response.data)
      } catch (error: any) {
        if (error.response.status === 401) {
          setIsLoggedIn(false)
        }
        console.error('Failed to fetch installations', error)
      }
    }

    fetchInstallations()
  }, [])

  useEffect(() => {
    // Fetch settings when selected account changes
    if (selectedAccount) {
      async function fetchSettings() {
        try {
          const response = await axios.get(`/api/settings/${selectedAccount?.account}`)
          setSettings(response.data)
        } catch (error) {
          console.error('Failed to fetch settings', error)
        }
      }

      async function fetchRepos() {
        try {
          const response: AxiosResponse<RepoInfo[]> = await axios.get(`/api/repos/${selectedAccount?.id}`)
          setRepos(response.data)
        } catch (error) {
          console.error('Failed to fetch repos', error)
        }
      }

      fetchSettings()
      setRepos([])
      fetchRepos()
      setPage(1)
      setHasMore(true)
      setSelectedRepos([])
    }
  }, [selectedAccount, installations])

  const handleAccountChange = (event: React.ChangeEvent<HTMLSelectElement>) => {
    setSelectedAccount(installations.find((installation) => installation.account === event.target.value) || null)
  }

  const handleSelectRepo = (event: React.ChangeEvent<HTMLUListElement>) => {
    const repo = event.target.value
    if (!selectedRepos.includes(repo)) {
      const repos = [...selectedRepos, repo]
      setSelectedRepos(repos)
    }
  }

  const handleUnselectRepo = (repo: RepoInfo) => {
    const repos = selectedRepos.filter((r) => r !== repo)
    setSelectedRepos(repos)
  }

  const loadMoreRepos = async () => {
    const newPage = curPage + 1
    try {
      const response = await axios.get(`/api/repos/${selectedAccount?.id}?page=${newPage}`)
      if (response.data.length === 0) {
        setHasMore(false)
        return
      }
      setRepos([...repos, ...response.data])
      setPage(newPage)
      setHasMore(true)
    } catch (error) {
      console.error('Failed to fetch repos', error)
    }
  }

  const handleTestSettings = async () => {
    const repoNames = selectedRepos.map((r) => r.full_name)
    listMode === "allow"
      ? setSettings({ ...settings!, allow_repos: repoNames, mute_repos: [] })
      : setSettings({ ...settings!, mute_repos: repoNames, allow_repos: [] })

    try {
      await axios.post('/api/settings/test', settings)
      toast.success('Test successful')
    } catch (error) {
      console.error('Failed to test settings', error)
      toast.error('Test failed')
    }
  }

  const handleSaveSettings = async () => {
    const repoNames = selectedRepos.map((r) => r.full_name)
    listMode === "allow"
      ? setSettings({ ...settings!, allow_repos: repoNames, mute_repos: [] })
      : setSettings({ ...settings!, mute_repos: repoNames, allow_repos: [] })

    try {
      await axios.post(`/api/settings/${selectedAccount}`, settings)
      toast.success('Settings saved successfully')
    } catch (error) {
      console.error('Failed to save settings', error)
      toast.error('Failed to save settings')
    }
  }

  if (!isLoggedIn) {
    return (
      <div className={styles.container}>
        <Header />
        <main>
          <section>
            <p>Please log in through GitHub to continue.</p>
            <a href='/api/authorize' className={styles.loginButton}>
              Log in with GitHub
            </a>
          </section>
        </main>
        <Footer />
        <ToastContainer />
      </div>
    )
  }

  return (
    <div className={styles.container}>
      <Header />
      <main>
        <section>
          <label htmlFor='account-select'>Select Account:</label>
          <AccountSelect
            handleAccountChange={handleAccountChange}
            selectedAccount={selectedAccount!}
            installations={installations}
            updateAccountState={() => {
              try {
                axios.get('/api/installations').then((response) => {
                  setInstallations(response.data)
                })
              } catch (error) {
                console.error('Failed to fetch installations', error)
              }
            }}
          />
        </section>
        {selectedAccount && settings && (
          <section>
            <NotificationConfig settings={settings} setSettings={setSettings} />
            <RepoSelector
              repos={repos.filter((repo) => !selectedRepos.includes(repo))}
              onSelect={handleSelectRepo}
              loadMoreRepos={loadMoreRepos}
              hasMore={hasMore}
            />

            {/* TODO: can only select one mode at a time*/}
            <div className={styles.listModeContainer}>
              <button className={styles.toggleButton} onClick={() => setListMode(ListMode.Allow)}>
                Allow notifications from these repos only
              </button>
              <button className={styles.toggleButton} onClick={() => setListMode(ListMode.Mute)}>
                Allow notifications from these repos only
              </button>
            </div>

            {/* TODO: a button to display a RepoSelector popup for user to select repo */}
            {/* after select, the popup dismiss */}
            <button className={styles.selectRepoButton}>
              <GoArrowDown /> Select Repositories
            </button>

            <div className={styles.selectedReposContainer}>
              <p>Selected {selectedRepos.length} repositories</p>
              {selectedRepos.map((repo) => (
                <div key={repo.id} className={styles.repoItem}>
                  {repo.fork ? <GoRepoForked /> : <GoRepo />}
                  {repo.owner}/<strong>{repo.name}</strong>
                  <button className={styles.removeButton} onClick={() => handleUnselectRepo(repo)}>
                    <FaTimes />
                  </button>
                </div>
              ))}
            </div>

            <div className={styles.settingsContainer}>
              <div className={styles.settingItem}>
                <label
                  htmlFor='mute-star-lost'
                  data-tooltip-id='tooltip'
                  data-tooltip-content="Don't send notifications when lost stars"
                >
                  Mute Star Lost:
                </label>
                <input
                  type='checkbox'
                  id='mute-star-lost'
                  checked={settings.mute_lost_stars}
                  onChange={(e) =>
                    setSettings({
                      ...settings,
                      mute_lost_stars: e.target.checked,
                    })
                  }
                />
              </div>
            </div>

            <div className={styles.buttonGroup}>
              <button
                className={styles.testButton}
                data-tooltip-id='tooltip'
                data-tooltip-content='Send a test notification to verify settings'
                onClick={handleTestSettings}
              >
                Test Settings
              </button>
              <button className={styles.saveButton} onClick={handleSaveSettings}>
                Save Settings
              </button>
            </div>
          </section>
        )}
      </main>
      <Footer />
      <ToastContainer closeOnClick />
      <Tooltip id='tooltip' />
    </div>
  )
}

export default App
