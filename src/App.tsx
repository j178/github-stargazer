import axios from "axios";
import React, { useEffect, useState } from "react";
import { ToastContainer, toast } from "react-toastify";
import "react-toastify/dist/ReactToastify.css";
import { Tooltip } from "react-tooltip";
import "react-tooltip/dist/react-tooltip.css";

import styles from "./App.module.css";
import NotificationConfig from "./NotificationConfig";
import RepoSelector from "./RepoSelector";
import { Installation, RepoInfo, Settings } from "./models";

enum ListMode {
  Allow = "allow",
  Mute = "mute",
}

const AccountSelect: React.FC<{
  installations: Installation[];
  selectedAccount: Installation | null;
  handleAccountChange: (event: React.ChangeEvent<HTMLSelectElement>) => void;
  updateAccountState: () => void;
}> = ({ installations, selectedAccount, handleAccountChange, updateAccountState }) => {
  const [, setPopup] = useState<Window | null>(null);

  const handleAddAccount = () => {
    const newPopup = window.open(
      "https://github.com/apps/stars-notifier/installations/new",
      "popup",
      "width=600,height=600",
    )!;
    setPopup(newPopup);

    const checkPopup = setInterval(() => {
      if (newPopup.closed) {
        clearInterval(checkPopup);
        setPopup(null);
        // Call a function to update the account state after the user installs the app
        updateAccountState();
      }
    }, 1000);
  };

  const handleChange = (event: React.ChangeEvent<HTMLSelectElement>) => {
    const value = event.target.value;
    if (value === "add-account") {
      handleAddAccount();
    } else {
      handleAccountChange(event);
    }
  };

  return (
    <select id="account-select" onChange={handleChange} value={selectedAccount?.account}>
      <option value="" disabled>
        Select an account
      </option>
      {installations.map((installation) => (
        <option key={installation.id} value={installation.account}>
          {installation.account} ({installation.account_type})
        </option>
      ))}
      <option value="add-account">Add GitHub Account</option>
    </select>
  );
};
const Header: React.FC = () => {
  return (
    <header>
      <h1 className={styles.title}>Star++ Configuration</h1>
    </header>
  );
};

const Footer: React.FC = () => {
  return (
    <footer>
      <a href="https://github.com/apps/stars-notifier" target="_blank" rel="noopener noreferrer">
        Powered by <img src="/avatar.png" alt="Star++" className={styles.logo} />
      </a>
    </footer>
  );
};

const App: React.FC = () => {
  const [isLoggedIn, setIsLoggedIn] = useState(true);
  const [installations, setInstallations] = useState<Installation[]>([]);
  const [repos, setRepos] = useState<RepoInfo[]>([]);
  const [selectedAccount, setSelectedAccount] = useState<Installation | null>(null);
  const [settings, setSettings] = useState<Settings | null>(null);
  const [listMode, setListMode] = useState(ListMode.Mute);
  const [selectedRepos, setSelectedRepos] = useState<string[]>([]);
  const [curPage, setPage] = useState(1);
  const [hasMore, setHasMore] = useState(true);

  const toggleListMode = () => {
    const mode = listMode === ListMode.Allow ? ListMode.Mute : ListMode.Allow;
    setListMode(mode);
    mode === ListMode.Allow
      ? setSettings({
          ...settings!,
          allow_repos: selectedRepos,
          mute_repos: [],
        })
      : setSettings({
          ...settings!,
          mute_repos: selectedRepos,
          allow_repos: [],
        });
  };

  useEffect(() => {
    // Fetch installations on mount
    async function fetchInstallations() {
      try {
        const response = await axios.get("/api/installations");
        setIsLoggedIn(true);
        setInstallations(response.data);
      } catch (error: any) {
        if (error.response.status === 401) {
          setIsLoggedIn(false);
        }
        console.error("Failed to fetch installations", error);
      }
    }

    fetchInstallations();
  }, []);

  useEffect(() => {
    // Fetch settings when selected account changes
    if (selectedAccount) {
      async function fetchSettings() {
        try {
          const response = await axios.get(`/api/settings/${selectedAccount?.account}`);
          setSettings(response.data);
        } catch (error) {
          console.error("Failed to fetch settings", error);
        }
      }

      async function fetchRepos() {
        try {
          const response = await axios.get(`/api/repos/${selectedAccount?.id}`);
          setRepos(response.data);
        } catch (error) {
          console.error("Failed to fetch repos", error);
        }
      }

      fetchSettings();
      setRepos([]);
      fetchRepos();
      setPage(1);
      setHasMore(true);
      setSelectedRepos([]);
    }
  }, [selectedAccount, installations]);

  const handleAccountChange = (event: React.ChangeEvent<HTMLSelectElement>) => {
    setSelectedAccount(installations.find((installation) => installation.account === event.target.value) || null);
  };

  const handleSelectRepo = (event: React.ChangeEvent<HTMLSelectElement>) => {
    const repo = event.target.value;
    if (!selectedRepos.includes(repo)) {
      const repos = [...selectedRepos, repo];
      setSelectedRepos(repos);
      listMode === ListMode.Allow
        ? setSettings({
            ...settings!,
            allow_repos: repos,
            mute_repos: [],
          })
        : setSettings({
            ...settings!,
            mute_repos: repos,
            allow_repos: [],
          });
    }
  };

  const handleUnselectRepo = (repo: string) => {
    const repos = selectedRepos.filter((r) => r !== repo);
    setSelectedRepos(repos);
    listMode === ListMode.Allow
      ? setSettings({ ...settings!, allow_repos: repos, mute_repos: [] })
      : setSettings({ ...settings!, mute_repos: repos, allow_repos: [] });
  };

  const loadMoreRepos = async () => {
    const newPage = curPage + 1;
    try {
      const response = await axios.get(`/api/repos/${selectedAccount?.id}?page=${newPage}`);
      if (response.data.length === 0) {
        setHasMore(false);
        return;
      }
      setRepos([...repos, ...response.data]);
      setPage(newPage);
      setHasMore(true);
    } catch (error) {
      console.error("Failed to fetch repos", error);
    }
  };

  const handleTestSettings = async () => {
    try {
      await axios.post("/api/settings/test", settings);
      toast.success("Test successful");
    } catch (error) {
      console.error("Failed to test settings", error);
      toast.error("Test failed");
    }
  };

  const handleSaveSettings = async () => {
    try {
      await axios.post(`/api/settings/${selectedAccount}`, settings);
      toast.success("Settings saved successfully");
    } catch (error) {
      console.error("Failed to save settings", error);
      toast.error("Failed to save settings");
    }
  };

  if (!isLoggedIn) {
    return (
      <div className={styles.container}>
        <Header />
        <main>
          <section>
            <p>Please log in through GitHub to continue.</p>
            <a href="/api/authorize" className={styles.loginButton}>
              Log in with GitHub
            </a>
          </section>
        </main>
        <Footer />
        <ToastContainer />
      </div>
    );
  }

  return (
    <div className={styles.container}>
      <Header />
      <main>
        <section>
          <label htmlFor="account-select">Select Account:</label>
          <AccountSelect
            handleAccountChange={handleAccountChange}
            selectedAccount={selectedAccount!}
            installations={installations}
            updateAccountState={() => {
              try {
                axios.get("/api/installations").then((response) => {
                  setInstallations(response.data);
                });
              } catch (error) {
                console.error("Failed to fetch installations", error);
              }
            }}
          />
        </section>
        {selectedAccount && settings && (
          <section>
            <NotificationConfig settings={settings} setSettings={setSettings} />
            <RepoSelector
              repos={repos.filter((repo) => !selectedRepos.includes(repo.name))}
              onSelect={handleSelectRepo}
              loadMoreRepos={loadMoreRepos}
              hasMore={hasMore}
            />

            <div className={styles.listModeContainer}>
              <h3>
                {listMode === "allow"
                  ? "Allow notifications from these repos only"
                  : "Mute notifications from these repos"}
              </h3>
              <button className={styles.toggleButton} onClick={toggleListMode}>
                Change to {listMode === "allow" ? "Mute" : "Allow"}
              </button>
            </div>

            <div className={styles.selectedReposContainer}>
              {selectedRepos.map((repo) => (
                <div key={repo} className={styles.repoItem}>
                  {repo}
                  <button className={styles.removeButton} onClick={() => handleUnselectRepo(repo)}>
                    Remove
                  </button>
                </div>
              ))}
            </div>

            <div className={styles.settingsContainer}>
              <div className={styles.settingItem}>
                <label
                  htmlFor="mute-star-lost"
                  data-tooltip-id="tooltip"
                  data-tooltip-content="Don't send notifications when lost stars"
                >
                  Mute Star Lost:
                </label>
                <input
                  type="checkbox"
                  id="mute-star-lost"
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
                data-tooltip-id="tooltip"
                data-tooltip-content="Send a test notification to verify settings"
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
      <Tooltip id="tooltip" />
    </div>
  );
};

export default App;
