import styles from './App.module.css';
import React, {useEffect, useState} from "react";
import axios from "axios";
import NotificationConfig from "./NotificationConfig";
import RepoSelector from "./RepoSelector";
import {toast, ToastContainer} from 'react-toastify';
import {Tooltip} from 'react-tooltip';
import 'react-toastify/dist/ReactToastify.css';
import 'react-tooltip/dist/react-tooltip.css'

const App = () => {
    const [installations, setInstallations] = useState([]);
    const [repos, setRepos] = useState([]);
    const [selectedAccount, setSelectedAccount] = useState(null);
    const [settings, setSettings] = useState(null);
    const [listMode, setListMode] = useState('mute');
    const [selectedRepos, setSelectedRepos] = useState([]);
    const [curPage, setPage] = useState(1);
    const [isLoggedIn, setIsLoggedIn] = useState(true);

    const toggleListMode = () => {
        const mode = listMode === 'allow' ? 'mute' : 'allow';
        setListMode(mode);
        mode === 'allow' ?
            setSettings({...settings, allow_repos: selectedRepos, mute_repos: []})
            : setSettings({...settings, mute_repos: selectedRepos, allow_repos: []});
    };

    useEffect(() => {
        // Fetch installations on mount
        async function fetchInstallations() {
            try {
                const response = await axios.get('/api/installations');
                setIsLoggedIn(true);
                setInstallations(response.data);
            } catch (error) {
                if (error.response.status === 401) {
                    setIsLoggedIn(false);
                }
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
            const repos = [...selectedRepos, repo];
            setSelectedRepos(repos);
            listMode === 'allow' ?
                setSettings({...settings, allow_repos: repos, mute_repos: []})
                : setSettings({...settings, mute_repos: repos, allow_repos: []});
        }
    };

    const handleUnselectRepo = (repo) => {
        const repos = selectedRepos.filter(r => r !== repo);
        setSelectedRepos(repos);
        listMode === 'allow' ?
            setSettings({...settings, allow_repos: repos, mute_repos: []})
            : setSettings({...settings, mute_repos: repos, allow_repos: []});
    };

    const loadMoreRepos = async () => {
        const newPage = curPage + 1;
        const account = installations.find(installation => installation.account === selectedAccount);
        try {
            const response = await axios.get(`/api/repos/${account.id}?page=${newPage}`);
            if (response.data.length === 0) {
                return;
            }
            setRepos([...repos, ...response.data])
            setPage(newPage);
        } catch (error) {
            console.error('Failed to fetch repos', error);
        }
    };

    const handleTestSettings = async () => {
        try {
            await axios.post('/api/settings/test', settings);
            toast.success('Test successful');
        } catch (error) {
            console.error('Failed to test settings', error);
            toast.error('Test failed');
        }
    };

    const handleSaveSettings = async () => {
        try {
            await axios.post(`/api/settings/${selectedAccount}`, settings);
            toast.success('Settings saved successfully');
        } catch (error) {
            console.error('Failed to save settings', error);
            toast.error('Failed to save settings');
        }
    };

    const header = (
        <header>
            <h1 className={styles.title}>Star++ Configuration</h1>
        </header>
    );
    const footer = (
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
    );

    if (!isLoggedIn) {
        return (
            <div className={styles.container}>
                {header}
                <main>
                    <section>
                        <p>Please log in through GitHub to continue.</p>
                        <a href={'/api/authorize'} className={styles.loginButton}>
                            Log in with GitHub
                        </a>
                    </section>
                </main>
                {footer}
                <ToastContainer/>
            </div>
        );
    }

    return (
        <div className={styles.container}>
            {header}
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
                        <option onClick={() => toast.warning("Not implemented yet.")}>Add GitHub Account</option>
                    </select>
                </section>
                {selectedAccount && settings && (
                    <section>
                        <NotificationConfig settings={settings} setSettings={setSettings}/>
                        <RepoSelector
                            repos={repos.filter(repo => !selectedRepos.includes(repo))}
                            onSelect={handleSelectRepo}
                            loadMoreRepos={loadMoreRepos}
                        />

                        <h3>{listMode === 'allow' ? 'Allow notifications from these repos only' : 'Mute notifications from these repos'}</h3>
                        <button onClick={toggleListMode}>Change to {listMode === 'allow' ? 'Mute' : 'Allow'}</button>
                        <div className={styles.configSection}>
                            {selectedRepos.map(repo => (
                                <div key={repo}>
                                    {repo}
                                    <button onClick={() => handleUnselectRepo(repo)}>Remove</button>
                                </div>
                            ))}
                        </div>
                        <div className={styles.configSection}>
                            <label htmlFor="mute-star-lost"
                                   data-tooltip-id="tooltip"
                                   data-tooltip-content="Don't send notifications when lost stars">Mute Star Lost:
                            </label>
                            <input type="checkbox"
                                   id="mute-star-lost"
                                   checked={settings.mute_lost_stars}
                                   onChange={(e) => setSettings({...settings, mute_lost_stars: e.target.checked})}
                            />
                        </div>
                        <div className={styles.buttonGroup}>
                            <button data-tooltip-id="tooltip"
                                    data-tooltip-content="Send a test notification to verify settings"
                                    onClick={handleTestSettings}>Test Settings
                            </button>
                            <button onClick={handleSaveSettings}>Save Settings</button>
                        </div>
                    </section>
                )}
            </main>
            {footer}
            <ToastContainer closeOnClick/>
            <Tooltip id="tooltip"/>
        </div>
    );
};

export default App;
