import styles from './App.module.css';
import React, {useEffect, useState} from "react";
import axios from "axios";
import NotificationConfig from "./NotificationConfig";
import RepoSelector from "./RepoSelector";

const App = () => {
    const [installations, setInstallations] = useState([]);
    const [repos, setRepos] = useState([]);
    const [selectedAccount, setSelectedAccount] = useState(null);
    const [settings, setSettings] = useState(null);
    const [listMode, setListMode] = useState('allow');
    const [selectedRepos, setSelectedRepos] = useState([]);
    const [curPage, setPage] = useState(1);

    const toggleListMode = () => {
        setListMode(listMode === 'allow' ? 'mute' : 'allow');
        listMode === 'allow' ?
            setSettings({...settings, allow_repos: selectedRepos, mute_repos: []})
            : setSettings({...settings, mute_repos: selectedRepos, allow_repos: []});
    };

    useEffect(() => {
        // Fetch installations on mount
        async function fetchInstallations() {
            try {
                const response = await axios.get('/api/installations');
                setInstallations(response.data);
            } catch (error) {
                console.error('Failed to fetch installations', error);
            }
        }

        fetchInstallations();
    }, []);

    useEffect(() => {
        // Fetch settings when selected account changes
        if (selectedAccount) {
            async function fetchSettings() {
                try {
                    const response = await axios.get(`/api/settings/${selectedAccount}`);
                    setSettings(response.data);
                } catch (error) {
                    console.error('Failed to fetch settings', error);
                }
            }

            async function fetchRepos() {
                const account = installations.find(installation => installation.account === selectedAccount);
                try {
                    const response = await axios.get(`/api/repos/${account.id}`);
                    setRepos(response.data)
                } catch (error) {
                    console.error('Failed to fetch repos', error);
                }
            }

            fetchSettings();
            setRepos([]);
            fetchRepos();
            setPage(1);
        }
    }, [selectedAccount, installations]);

    const handleAccountChange = (event) => {
        setSelectedAccount(event.target.value);
        setSelectedRepos([]);
    };

    const handleSelectRepo = (event) => {
        const repo = event.target.value;
        if (!selectedRepos.includes(repo)) {
            setSelectedRepos([...selectedRepos, repo]);
        }
    };

    const handleUnselectRepo = (repo) => {
        setSelectedRepos(selectedRepos.filter(r => r !== repo));
    };

    const loadMoreRepos = async () => {
        const newPage = curPage + 1;
        const account = installations.find(installation => installation.account === selectedAccount);
        try {
            const response = await axios.get(`/api/repos/${account.id}?page=${newPage}`);
            setRepos([...repos, ...response.data])
            setPage(newPage);
        } catch (error) {
            console.error('Failed to fetch repos', error);
        }
    };

    const handleTestSettings = async () => {
        try {
            await axios.post('/api/settings/test', settings);
            alert('Test successful');
        } catch (error) {
            console.error('Failed to test settings', error);
            alert('Test failed');
        }
    };

    const handleSaveSettings = async () => {
        try {
            await axios.post(`/api/settings/${selectedAccount}`, settings);
            alert('Settings saved successfully');
        } catch (error) {
            console.error('Failed to save settings', error);
            alert('Failed to save settings');
        }
    };

    return (
        <div className={styles.container}>
            <header>
                <h1 className={styles.title}>Star++ Configuration</h1>
            </header>
            <main>
                <section>
                    <label htmlFor="account-select">Select Account:</label>
                    <select id="account-select" onChange={handleAccountChange} value={selectedAccount || ''}>
                        <option value="" disabled>Select an account</option>
                        {installations.map((installation) => (
                            <option key={installation.id} value={installation.account}>
                                {installation.account} ({installation.account_type})
                            </option>
                        ))}
                    </select>
                </section>
                {selectedAccount && settings && (
                    <section>
                        <NotificationConfig settings={settings} setSettings={setSettings}/>
                        <RepoSelector
                            repos={repos.filter(repo => !selectedRepos.includes(repo))}
                            onSelect={handleSelectRepo}
                            maxVisible={10}
                            loadMoreRepos={loadMoreRepos}
                        />

                        <button onClick={toggleListMode}>{listMode === 'allow' ? 'Mute List' : 'Allow List'}</button>
                        <div className={styles.configSection}>
                            {selectedRepos.map(repo => (
                                <div key={repo}>
                                    {repo}
                                    <button onClick={() => handleUnselectRepo(repo)}>Remove</button>
                                </div>
                            ))}
                        </div>
                        <div className={styles.configSection}>
                            <label>Mute Star Lost:</label>
                            <input
                                type="checkbox"
                                checked={settings.mute_lost_stars}
                                onChange={(e) => setSettings({...settings, mute_lost_stars: e.target.checked})}
                            />
                        </div>
                        <div className={styles.buttonGroup}>
                            <button onClick={handleTestSettings}>Test Settings</button>
                            <button onClick={handleSaveSettings}>Save Settings</button>
                        </div>
                    </section>
                )}
            </main>
            <footer>
                <a
                    href="https://github.com/apps/stars-notifier"
                    target="_blank"
                    rel="noopener noreferrer"
                >
                    Powered by{' '}
                    <img src="/avatar.png" alt="Star++" className={styles.logo}/>
                </a>
            </footer>
        </div>
    );
};

export default App;
