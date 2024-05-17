import React, {useState} from 'react';
import axios from 'axios';
import styles from './NotificationConfig.module.css';
import {FaTelegram, FaDiscord, FaBell, FaPlug} from "react-icons/fa";

const serviceIcons = {
    telegram: <FaTelegram />,
    discord_webhook: <FaDiscord />,
    discord_bot: <FaDiscord />,
    bark: <FaBell />,
    webhook: <FaPlug />
};

const NotificationConfig = ({settings, setSettings}) => {
    const [service, setService] = useState('');
    const [serviceDetails, setServiceDetails] = useState({});
    const [connectionToken, setConnectionToken] = useState(null);
    const [showMore, setShowMore] = useState(false);

    const handleAddService = () => {
        const newService = {service: service, ...serviceDetails};
        setSettings({...settings, notify_settings: [...settings.notify_settings, newService]});
        setService('');
        setServiceDetails({});
    };

    const handleRemoveService = (index) => {
        const newSettings = {...settings, notify_settings: settings.notify_settings.filter((_, i) => i !== index)};
        setSettings(newSettings);
    }

    const handleConnect = async () => {
        try {
            const response = await axios.post(`/api/connect/${service}`);
            setConnectionToken(response.data);
        } catch (error) {
            console.error('Failed to connect Telegram', error);
        }
    };

    const handleConnectResult = async () => {
        try {
            const response = await axios.get(`/api/connect/${service}/${connectionToken.token}`);
            setServiceDetails({...serviceDetails, ...response.data})
        } catch (error) {
            console.error('Failed to get connect result', error);
        }
    }

    const selectService = (e) => {
        setService(e.target.value);
        setServiceDetails({});
        setConnectionToken(null);
        setShowMore(false);
    }

    return (
        <div className={styles.notificationConfig}>
            <label htmlFor="service-select">Add Notification Service:</label>
            <select id="service-select" value={service} onChange={selectService}>
                <option value="" disabled>Select a service</option>
                {['bark', 'telegram', 'discord_webhook', 'discord_bot', 'webhook'].map((s) => (
                    <option key={s} value={s}>{s.toUpperCase()}</option>
                ))}
            </select>

            {/* TODO: 增加配置后调用 check 检查配置 */}
            {service && (
                <div className={styles.serviceConfig}>
                    {service === 'telegram' && (
                        <div>
                            <button onClick={handleConnect}>Connect to a Telegram Chat</button>
                            {connectionToken && (
                                <div>
                                    <p><a href={connectionToken.bot_url} target="_blank" rel="noopener noreferrer">Private
                                        Chat</a></p>
                                    <p><a href={connectionToken.bot_group_url} target="_blank" rel="noopener noreferrer">Group
                                        Chat</a></p>
                                    <button onClick={handleConnectResult}>获取连接结果</button>
                                    <p>Connected with Chat ID: {serviceDetails.chat_id}</p>
                                </div>
                            )}
                        </div>
                    )}
                    {service === 'discord_webhook' && (
                        <div>
                            <label>Webhook ID:</label>
                            <input type="text" value={serviceDetails.webhook_id || ''}
                                   onChange={(e) => setServiceDetails({
                                       ...serviceDetails,
                                       webhook_id: e.target.value
                                   })}/>
                            <label>Webhook Token:</label>
                            <input type="text" value={serviceDetails.webhook_token || ''}
                                   onChange={(e) => setServiceDetails({
                                       ...serviceDetails,
                                       webhook_token: e.target.value
                                   })}/>
                            {
                                showMore? (
                                    <div>
                                        <label>Username:</label>
                                        <input type="text" value={serviceDetails.username || ''}
                                               onChange={(e) => e.target.value && setServiceDetails({...serviceDetails, username: e.target.value})}/>
                                        <label>Avatar URL:</label>
                                        <input type="text" value={serviceDetails.avatar_url || ''}
                                               onChange={(e) => e.target.value && setServiceDetails({...serviceDetails, avatar_url: e.target.value})}/>
                                        <label>Color:</label>
                                        <input type="number" value={serviceDetails.color || ''}
                                               onChange={(e) => e.target.value && setServiceDetails({...serviceDetails, color: e.target.value})}/>
                                    </div>
                                ):(<button onClick={() => setShowMore(true)}>Show More</button>)
                            }
                        </div>
                    )}
                    {service === 'discord_bot' && (
                        <div>
                            <p>TODO</p>
                        </div>
                    )}
                    {service === 'bark' && (
                        <div>
                            <label>Bark Key:</label>
                            <input type="text" value={serviceDetails.key || ''}
                                   onChange={(e) => setServiceDetails({...serviceDetails, key: e.target.value})}/>
                            <label>Bark Server:</label>
                            <input type="text" value={serviceDetails.server || 'https://api.day.app/'}
                                   onChange={(e) => setServiceDetails({...serviceDetails, server: e.target.value})}/>
                        </div>
                    )}
                    {service === 'webhook' && (
                        <div>
                            <label>Webhook URL:</label>
                            <input type="text" value={serviceDetails.url || ''}
                                   onChange={(e) => setServiceDetails({...serviceDetails, url: e.target.value})}/>
                            <label>Method:</label>
                            <select value={serviceDetails.method || 'GET'}
                                    onChange={(e) => setServiceDetails({...serviceDetails, method: e.target.value})}>
                                <option value="GET">GET</option>
                                <option value="POST">POST</option>
                                <option value="PUT">PUT</option>
                            </select>
                            {
                                showMore? (
                                    <div>
                                        <label>Headers:</label>
                                        <textarea value={serviceDetails.headers || ''}
                                                  onChange={(e) => setServiceDetails({...serviceDetails, headers: e.target.value})}/>
                                        <label>Body Template:</label>
                                        <textarea value={serviceDetails.body || ''}
                                                  onChange={(e) => setServiceDetails({...serviceDetails, body: e.target.value})}/>
                                    </div>
                                ):(<button onClick={() => setShowMore(true)}>Show More</button>)
                            }
                        </div>
                    )}
                    <button onClick={handleAddService}>Add Service</button>
                </div>
            )}

            <div className={styles.currentSettings}>
                {/* TODO: 将 Add a new service 按钮放在这个列表最下发 */}
                <h3>Current Notification Settings</h3>
                <ul>
                    {settings.notify_settings.map((ns, index) => (
                        <div key={index} className={styles.settingItem}>
                            <li className={styles.serviceDetail}>
                                <span className={styles.serviceIcon}>{serviceIcons[ns.service]}</span>
                            </li>
                            <div className={styles.serviceDetailText}>
                                <strong>{ns.service.toUpperCase()}</strong>
                                <pre>{JSON.stringify(ns, null, 2)}</pre>
                            </div>
                            <button onClick={() => handleRemoveService(index)} className={styles.removeButton}>Remove</button>
                        </div>
                    ))}
                </ul>
            </div>
        </div>
    );
};

export default NotificationConfig;
