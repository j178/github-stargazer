import React, {useState} from 'react';
import axios from 'axios';

const NotificationConfig = ({settings, setSettings}) => {
    const [service, setService] = useState('');
    const [serviceDetails, setServiceDetails] = useState({});
    const [connectionToken, setConnectionToken] = useState(null);

    const handleAddService = () => {
        const newService = {service: service, ...serviceDetails};
        setSettings({...settings, notify_settings: [...settings.notify_settings, newService]});
        setService('');
        setServiceDetails({});
    };

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
            console.log(response.data)
            setServiceDetails({...serviceDetails, ...response.data})
        } catch (error) {
            console.error('Failed to get connect result', error);
        }
    }

    return (
        <div>
            <label htmlFor="service-select">Add Notification Service:</label>
            <select id="service-select" value={service} onChange={(e) => {
                setService(e.target.value)
                setServiceDetails({})
            }}>
                <option value="" disabled selected>Select a service</option>
                {['bark', 'telegram', 'discord_webhook', 'discord_bot', 'webhook'].map((s) => (
                    <option key={s} value={s}>{s}</option>
                ))}
            </select>

            {/* TODO: 增加配置后调用 check 检查配置 */}
            {service && (
                <div>
                    {service === 'telegram' && (
                        <button onClick={handleConnect}>Connect to a Telegram Chat</button>
                    )}
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
                            {/* TODO: 增加可选配置：username, avatar_url, color */}
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
                        </div>
                    )}
                    {service === 'webhook' && (
                        <div>
                            <label>Webhook URL:</label>
                            <input type="text" value={serviceDetails.url || ''}
                                   onChange={(e) => setServiceDetails({...serviceDetails, url: e.target.value})}/>
                        </div>
                    )}
                    <button onClick={handleAddService}>Add Service</button>
                </div>
            )}

            <div>
                {/* TODO: 支持移除某项配置，将 Add a new service 按钮放在这个列表最下发 */}
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
