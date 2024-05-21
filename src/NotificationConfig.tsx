import axios from "axios";
import React, { useState } from "react";
import { FaBell, FaDiscord, FaPlug, FaTelegram } from "react-icons/fa";

import styles from "./NotificationConfig.module.css";
import { NotifySetting, Settings } from "./models";

const serviceIcons: { [key: string]: React.ReactElement } = {
  telegram: <FaTelegram />,
  discord_webhook: <FaDiscord />,
  discord_bot: <FaDiscord />,
  bark: <FaBell />,
  webhook: <FaPlug />,
};

interface ServiceDetail {
  chat_id: string | null;
  webhook_id: string | null;
  webhook_token: string | null;
  username: string | null;
  avatar_url: string | null;
  color: number | null;
  key: string | null;
  server: string | null;
  url: string | null;
  method: string | null;
  headers: string | null;
  body: string | null;
}

interface ConnectionToken {
  token: string;
  bot_url: string;
  bot_group_url: string | null;
}

const NotificationConfig: React.FC<{
  settings: Settings;
  setSettings: (settings: Settings) => void;
}> = ({ settings, setSettings }) => {
  const [service, setService] = useState<string | null>(null);
  const [serviceDetails, setServiceDetails] = useState<ServiceDetail | null>(null);
  const [connectionToken, setConnectionToken] = useState<ConnectionToken | null>(null);
  const [showMore, setShowMore] = useState(false);

  const handleAddService = () => {
    const newService = { service: service!, ...serviceDetails };
    setSettings({
      ...settings,
      notify_settings: [...settings.notify_settings, newService],
    });
    setService(null);
    setServiceDetails(null);
  };

  const handleRemoveService = (index: number) => {
    const newSettings = {
      ...settings,
      notify_settings: settings.notify_settings.filter((_, i) => i !== index),
    };
    setSettings(newSettings);
  };

  const handleEditService = (index: number, updatedService: NotifySetting) => {
    const newSettings = {
      ...settings,
      notify_settings: settings.notify_settings.map((s, i) => (i === index ? updatedService : s)),
    };
    setSettings(newSettings);
  };

  const handleConnect = async () => {
    try {
      const response = await axios.post(`/api/connect/${service}`);
      setConnectionToken(response.data);
    } catch (error) {
      console.error("Failed to connect Telegram", error);
    }
  };

  const handleConnectResult = async () => {
    try {
      const response = await axios.get(`/api/connect/${service}/${connectionToken?.token}`);
      setServiceDetails({ ...serviceDetails, ...response.data });
    } catch (error) {
      console.error("Failed to get connect result", error);
    }
  };

  const selectService = (e: React.ChangeEvent<HTMLSelectElement>) => {
    setService(e.target.value);
    setServiceDetails({
      chat_id: null,
      webhook_id: null,
      webhook_token: null,
      username: null,
      avatar_url: null,
      color: null,
      key: null,
      server: null,
      url: null,
      method: null,
      headers: null,
      body: null,
    });
    setConnectionToken(null);
    setShowMore(false);
  };

  const supportedServices = ["bark", "telegram", "discord_webhook", "discord_bot", "webhook"];

  return (
    <div className={styles.notificationConfig}>
      <h3>
        <label htmlFor="service-select">Add Notification Service</label>
      </h3>
      <select id="service-select" value={service ?? ""} onChange={selectService}>
        <option value="" disabled>
          Select a service
        </option>
        {supportedServices.map((s) => (
          <option key={s} value={s}>
            {s.toUpperCase()}
          </option>
        ))}
      </select>

      {/* TODO: 增加配置后调用 check 检查配置 */}
      {service && (
        <div className={styles.serviceConfig}>
          {service === "telegram" && (
            <div>
              <button onClick={handleConnect}>Connect to a Telegram Chat</button>
              {connectionToken && (
                <div>
                  <p>
                    <a href={connectionToken.bot_url} target="_blank" rel="noopener noreferrer">
                      Private Chat
                    </a>
                  </p>
                  <p>
                    <a href={connectionToken.bot_group_url!} target="_blank" rel="noopener noreferrer">
                      Group Chat
                    </a>
                  </p>
                  <button onClick={handleConnectResult}>获取连接结果</button>
                  <p>Connected with Chat ID: {serviceDetails?.chat_id}</p>
                </div>
              )}
            </div>
          )}
          {service === "discord_webhook" && (
            <div>
              <label>Webhook ID:</label>
              <input
                type="text"
                value={serviceDetails?.webhook_id || ""}
                onChange={(e) =>
                  setServiceDetails({
                    ...serviceDetails!,
                    webhook_id: e.target.value,
                  })
                }
              />
              <label>Webhook Token:</label>
              <input
                type="text"
                value={serviceDetails?.webhook_token || ""}
                onChange={(e) =>
                  setServiceDetails({
                    ...serviceDetails!,
                    webhook_token: e.target.value,
                  })
                }
              />
              {showMore ? (
                <div>
                  <label>Username:</label>
                  <input
                    type="text"
                    value={serviceDetails?.username || ""}
                    onChange={(e) =>
                      setServiceDetails({
                        ...serviceDetails!,
                        username: e.target.value,
                      })
                    }
                  />
                  <label>Avatar URL:</label>
                  <input
                    type="text"
                    value={serviceDetails?.avatar_url || ""}
                    onChange={(e) =>
                      setServiceDetails({
                        ...serviceDetails!,
                        avatar_url: e.target.value,
                      })
                    }
                  />
                  <label>Color:</label>
                  <input
                    type="number"
                    value={serviceDetails?.color || ""}
                    onChange={(e) =>
                      setServiceDetails({
                        ...serviceDetails!,
                        color: parseInt(e.target.value),
                      })
                    }
                  />
                </div>
              ) : (
                <button onClick={() => setShowMore(true)}>Show More</button>
              )}
            </div>
          )}
          {service === "discord_bot" && (
            <div>
              <p>TODO</p>
            </div>
          )}
          {service === "bark" && (
            <div>
              <label>Bark Key:</label>
              <input
                type="text"
                value={serviceDetails?.key || ""}
                onChange={(e) =>
                  setServiceDetails({
                    ...serviceDetails!,
                    key: e.target.value,
                  })
                }
              />
              <label>Bark Server:</label>
              <input
                type="text"
                value={serviceDetails?.server || "https://api.day.app/"}
                onChange={(e) =>
                  setServiceDetails({
                    ...serviceDetails!,
                    server: e.target.value,
                  })
                }
              />
            </div>
          )}
          {service === "webhook" && (
            <div>
              <label>Webhook URL:</label>
              <input
                type="text"
                value={serviceDetails?.url || ""}
                onChange={(e) =>
                  setServiceDetails({
                    ...serviceDetails!,
                    url: e.target.value,
                  })
                }
              />
              <label>Method:</label>
              <select
                value={serviceDetails?.method || "GET"}
                onChange={(e) =>
                  setServiceDetails({
                    ...serviceDetails!,
                    method: e.target.value,
                  })
                }
              >
                <option value="GET">GET</option>
                <option value="POST">POST</option>
                <option value="PUT">PUT</option>
              </select>
              {showMore ? (
                <div>
                  <label>Headers:</label>
                  <textarea
                    value={serviceDetails?.headers || ""}
                    onChange={(e) =>
                      setServiceDetails({
                        ...serviceDetails!,
                        headers: e.target.value,
                      })
                    }
                  />
                  <label>Body Template:</label>
                  <textarea
                    value={serviceDetails?.body || ""}
                    onChange={(e) =>
                      setServiceDetails({
                        ...serviceDetails!,
                        body: e.target.value,
                      })
                    }
                  />
                </div>
              ) : (
                <button onClick={() => setShowMore(true)}>Show More</button>
              )}
            </div>
          )}
          <button onClick={handleAddService}>Add Service</button>
        </div>
      )}

      <div className={styles.currentSettings}>
        <h3>Current Notification Settings</h3>
        <div className={styles.settingsList}>
          {settings.notify_settings.map((ns, index) => (
            <div key={index} className={styles.settingCard}>
              <div className={styles.cardHeader}>
                <span className={styles.serviceIcon}>{serviceIcons[ns.service]}</span>
                <strong>{ns.service.toUpperCase()}</strong>
              </div>
              <div className={styles.cardContent}>
                <pre>{JSON.stringify(ns, null, 2)}</pre>
              </div>
              <div className={styles.cardActions}>
                <button onClick={() => handleEditService(index, ns)}>Edit</button>
                <button onClick={() => handleRemoveService(index)} className={styles.removeButton}>
                  Remove
                </button>
              </div>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
};

export default NotificationConfig;
