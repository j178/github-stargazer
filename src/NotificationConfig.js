import React, { useState } from 'react';
import axios from 'axios';

const NotificationConfig = ({ settings, setSettings }) => {
    const [service, setService] = useState('');
    const [serviceDetails, setServiceDetails] = useState({});
    const [telegramToken, setTelegramToken] = useState(null);

    const handleAddService = () => {
        const newService = { service, ...serviceDetails };
        setSettings({ ...settings, notify_settings: [...settings.notify_settings, newService] });
        setService('');
        setServiceDetails({});
    };

    const handleConnectTelegram = async () => {
        try {
            const response = await axios.post('/api/connect/telegram');
            setTelegramToken(response.data.token);
        } catch (error) {
            console.error('Failed to connect Telegram', error);
        }
    };

    return (
        <div>
            <label htmlFor="service-select">Add Notification Service:</label>
            <select id="service-select" value={service} onChange={(e) => setService(e.target.value)}>
                <option value="" disabled selected>Select a service</option>
                {['telegram', 'discord', 'bark', 'webhook'].map((s) => (
                    <option key={s} value={s}>{s}</option>
                ))}
            </select>

            {service && (
                <div>
                    {service === 'telegram' && !settings.notify_settings.find(ns => ns.service === 'telegram') && (
                        <button onClick={handleConnectTelegram}>Connect Telegram</button>
                    )}
                    {telegramToken && (
                        <p>Go to the bot and start the chat: <a href={`https://t.me/gh_stargazer_bot?start=${telegramToken}`} target="_blank" rel="noopener noreferrer">Start Chat</a></p>
                    )}
                    {service === 'discord' && (
                        <div>
                            <label>Webhook ID:</label>
                            <input type="text" value={serviceDetails.webhook_id || ''} onChange={(e) => setServiceDetails({ ...serviceDetails, webhook_id: e.target.value })} />
                            <label>Webhook Token:</label>
                            <input type="text" value={serviceDetails.webhook_token || ''} onChange={(e) => setServiceDetails({ ...serviceDetails, webhook_token: e.target.value })} />
                        </div>
                    )}
                    {service === 'bark' && (
                        <div>
                            <label>Key:</label>
                            <input type="text" value={serviceDetails.key || ''} onChange={(e) => setServiceDetails({ ...serviceDetails, key: e.target.value })} />
                        </div>
                    )}
                    {service === 'webhook' && (
                        <div>
                            <label>URL:</label>
                            <input type="text" value={serviceDetails.url || ''} onChange={(e) => setServiceDetails({ ...serviceDetails, url: e.target.value })} />
                        </div>
                    )}
                    <button onClick={handleAddService}>Add Service</button>
                </div>
            )}

            <div>
                <h3>Current Notification Settings</h3>
                <ul>
                    {settings.notify_settings.map((ns, index) => (
                        <li key={index}>{JSON.stringify(ns)}</li>
                    ))}
                </ul>
            </div>
        </div>
    );
};

export default NotificationConfig;
