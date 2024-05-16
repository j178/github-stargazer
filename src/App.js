import styles from './App.module.css';
import React, {useEffect, useState} from "react";
import axios from "axios";
import NotificationConfig from "./NotificationConfig";

const App = () => {
  const [installations, setInstallations] = useState([]);
  const [selectedAccount, setSelectedAccount] = useState(null);
  const [settings, setSettings] = useState(null);

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

      fetchSettings();
    }
  }, [selectedAccount]);

  const handleAccountChange = (event) => {
    setSelectedAccount(event.target.value);
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
      <div>
        <header>
          <h1 className={styles.title}>Star++ Configuration</h1>
        </header>
        <main>
          <section>
            <label htmlFor="account-select">Select Account:</label>
            <select id="account-select" onChange={handleAccountChange}>
              <option value="" disabled selected>Select an account</option>
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
                <div>
                  {/* TODO: 获取 repo 列表，手动选择增加或删除 */}
                  <label>Allow Repos:</label>
                  <textarea
                      value={JSON.stringify(settings.allow_repos, null, 2)}
                      onChange={(e) => setSettings({...settings, allow_repos: JSON.parse(e.target.value)})}
                  />
                </div>
                <div>
                  <label>Mute Repos:</label>
                  <textarea
                      value={JSON.stringify(settings.mute_repos, null, 2)}
                      onChange={(e) => setSettings({...settings, mute_repos: JSON.parse(e.target.value)})}
                  />
                </div>
                <div>
                  <label>Mute Star Lost:</label>
                    <input
                        type="checkbox"
                        checked={settings.mute_lost_stars}
                        onChange={(e) => setSettings({...settings, mute_lost_stars: e.target.checked})}
                    />
                </div>
                <button onClick={handleTestSettings}>Test Settings</button>
                <button onClick={handleSaveSettings}>Save Settings</button>
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
