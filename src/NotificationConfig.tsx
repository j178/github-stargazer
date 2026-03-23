import axios from 'axios'
import {
  type Dispatch,
  type FC,
  type ReactElement,
  type ReactNode,
  type SetStateAction,
  useCallback,
  useEffect,
  useState,
} from 'react'
import { FaBell, FaDiscord, FaPlug, FaTelegram } from 'react-icons/fa'
import { FiCheckCircle, FiEdit2, FiExternalLink, FiLink, FiPlus, FiRefreshCw, FiTrash2, FiX } from 'react-icons/fi'

import { type NotificationService, type NotifySetting, notificationServices, type Settings } from './models'

import styles from './NotificationConfig.module.css'

const serviceIcons: Record<NotificationService, ReactElement> = {
  telegram: <FaTelegram />,
  discord_webhook: <FaDiscord />,
  discord_bot: <FaDiscord />,
  bark: <FaBell />,
  webhook: <FaPlug />,
}

const serviceMeta: Record<
  NotificationService,
  {
    label: string
    description: string
    quickConnect?: string
  }
> = {
  telegram: {
    label: 'Telegram',
    description: 'Send star activity to a Telegram chat.',
    quickConnect: 'Link a Telegram chat with the default bot, or enter the chat ID manually.',
  },
  discord_webhook: {
    label: 'Discord Webhook',
    description: 'Post rich embeds through a Discord webhook.',
  },
  discord_bot: {
    label: 'Discord Bot',
    description: 'Send updates with a linked Discord bot.',
    quickConnect: 'Invite the bot, then run `/connect <token>` in the channel you want to use.',
  },
  bark: {
    label: 'Bark',
    description: 'Push lightweight alerts to Bark.',
  },
  webhook: {
    label: 'Generic Webhook',
    description: 'Send star updates to any HTTP endpoint.',
  },
}

interface ConnectionToken {
  token: string
  bot_url?: string
  bot_group_url?: string | null
  expire?: number
}

type ConnectionState = 'idle' | 'pending' | 'connected' | 'error'

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

const maskValue = (value?: string) => {
  if (!value) {
    return 'Not set'
  }

  if (value.length <= 10) {
    return value
  }

  return `${value.slice(0, 4)}•••${value.slice(-4)}`
}

const createDraft = (service: NotificationService): NotifySetting => {
  switch (service) {
    case 'telegram':
      return {
        service,
        chat_id: '',
        token: '',
        telegram_username: '',
      }
    case 'discord_webhook':
      return {
        service,
        webhook_id: '',
        webhook_token: '',
        username: '',
        avatar_url: '',
        color: 'fd9a00',
      }
    case 'discord_bot':
      return {
        service,
        channel_id: '',
        token: '',
        username: '',
        avatar_url: '',
        color: 'fd9a00',
      }
    case 'bark':
      return {
        service,
        key: '',
        server: 'https://api.day.app/',
      }
    case 'webhook':
      return {
        service,
        url: '',
        method: 'POST',
        headers: '',
        body: '',
      }
  }
}

const getRequiredFieldMessage = (draft: NotifySetting) => {
  switch (draft.service) {
    case 'telegram':
      return draft.chat_id?.trim() ? null : 'Chat ID is required for Telegram notifications.'
    case 'discord_webhook':
      return draft.webhook_id?.trim() && draft.webhook_token?.trim()
        ? null
        : 'Webhook ID and webhook token are required.'
    case 'discord_bot':
      return draft.channel_id?.trim() ? null : 'Channel ID is required for Discord bot delivery.'
    case 'bark':
      return draft.key?.trim() ? null : 'Bark key is required.'
    case 'webhook':
      return draft.url?.trim() ? null : 'Webhook URL is required.'
  }
}

const serializeDraft = (draft: NotifySetting): NotifySetting => {
  const next: NotifySetting = { service: draft.service }
  const include = (key: string) => {
    const value = draft[key]?.trim()
    if (value) {
      next[key] = value
    }
  }

  switch (draft.service) {
    case 'telegram':
      include('chat_id')
      include('token')
      include('telegram_username')
      break
    case 'discord_webhook':
      include('webhook_id')
      include('webhook_token')
      include('username')
      include('avatar_url')
      include('color')
      break
    case 'discord_bot':
      include('channel_id')
      include('token')
      include('username')
      include('avatar_url')
      include('color')
      include('guild_id')
      break
    case 'bark':
      include('key')
      include('server')
      break
    case 'webhook':
      include('url')
      include('method')
      include('headers')
      include('body')
      break
  }

  return next
}

const getConnectPlatform = (service: NotificationService) => {
  switch (service) {
    case 'telegram':
      return 'telegram'
    case 'discord_bot':
      return 'discord'
    default:
      return null
  }
}

const getSettingDetails = (setting: NotifySetting) => {
  switch (setting.service) {
    case 'telegram':
      return [
        { label: 'Chat ID', value: setting.chat_id ?? 'Not set' },
        { label: 'Username', value: setting.telegram_username ?? 'Default bot connection' },
        { label: 'Bot token', value: setting.token ? 'Custom token' : 'Default bot' },
      ]
    case 'discord_webhook':
      return [
        { label: 'Webhook ID', value: maskValue(setting.webhook_id) },
        { label: 'Webhook token', value: maskValue(setting.webhook_token) },
        { label: 'Display name', value: setting.username ?? 'Star++' },
      ]
    case 'discord_bot':
      return [
        { label: 'Channel ID', value: setting.channel_id ?? 'Not set' },
        { label: 'Bot token', value: setting.token ? 'Custom token' : 'Default bot' },
        { label: 'Display name', value: setting.username ?? 'Star++' },
      ]
    case 'bark':
      return [
        { label: 'Key', value: maskValue(setting.key) },
        { label: 'Server', value: setting.server ?? 'https://api.day.app/' },
      ]
    case 'webhook':
      return [
        { label: 'URL', value: setting.url ?? 'Not set' },
        { label: 'Method', value: setting.method ?? 'POST' },
        {
          label: 'Headers',
          value: setting.headers?.trim() ? 'Custom headers configured' : 'No custom headers',
        },
      ]
  }
}

const Field: FC<{
  label: string
  hint?: string
  wide?: boolean
  children: ReactNode
}> = ({ label, hint, wide, children }) => {
  return (
    <div className={wide ? `${styles.field} ${styles.fieldWide}` : styles.field}>
      <div className={styles.fieldMeta}>
        <span className={styles.fieldLabel}>{label}</span>
        {hint ? <span className={styles.fieldHint}>{hint}</span> : null}
      </div>
      {children}
    </div>
  )
}

const getSettingKey = (setting: NotifySetting) =>
  Object.entries(setting)
    .sort(([leftKey], [rightKey]) => leftKey.localeCompare(rightKey))
    .map(([key, value]) => `${key}:${value ?? ''}`)
    .join('|')

const NotificationConfig: FC<{
  settings: Settings
  setSettings: Dispatch<SetStateAction<Settings>>
}> = ({ settings, setSettings }) => {
  const [draft, setDraft] = useState<NotifySetting | null>(null)
  const [editingIndex, setEditingIndex] = useState<number | null>(null)
  const [connectionToken, setConnectionToken] = useState<ConnectionToken | null>(null)
  const [connectionState, setConnectionState] = useState<ConnectionState>('idle')
  const [connectionMessage, setConnectionMessage] = useState('')

  const resetConnectionState = () => {
    setConnectionToken(null)
    setConnectionState('idle')
    setConnectionMessage('')
  }

  const beginDraft = (service: NotificationService) => {
    setDraft(createDraft(service))
    setEditingIndex(null)
    resetConnectionState()
  }

  const cancelDraft = () => {
    setDraft(null)
    setEditingIndex(null)
    resetConnectionState()
  }

  const updateDraftField = (key: string, value: string) => {
    setDraft((current) => (current ? { ...current, [key]: value } : current))
  }

  const handleRemoveService = (index: number) => {
    setSettings((current) => ({
      ...current,
      notify_settings: current.notify_settings.filter((_, currentIndex) => currentIndex !== index),
    }))

    if (editingIndex === index) {
      cancelDraft()
    }
  }

  const handleEditService = (index: number) => {
    const service = settings.notify_settings[index]
    setDraft({
      ...createDraft(service.service),
      ...service,
    })
    setEditingIndex(index)
    resetConnectionState()
  }

  const handleApplyDraft = () => {
    if (!draft) {
      return
    }

    const validationMessage = getRequiredFieldMessage(draft)
    if (validationMessage) {
      setConnectionState('error')
      setConnectionMessage(validationMessage)
      return
    }

    const serialized = serializeDraft(draft)
    setSettings((current) => {
      if (editingIndex === null) {
        return {
          ...current,
          notify_settings: [...current.notify_settings, serialized],
        }
      }

      return {
        ...current,
        notify_settings: current.notify_settings.map((setting, index) =>
          index === editingIndex ? serialized : setting
        ),
      }
    })

    cancelDraft()
  }

  const handleStartConnection = async () => {
    if (!draft) {
      return
    }

    const platform = getConnectPlatform(draft.service)
    if (!platform) {
      return
    }

    try {
      const response = await axios.post(`/api/connect/${platform}`)
      setConnectionToken(response.data)
      setConnectionState('pending')
      setConnectionMessage(
        platform === 'discord'
          ? 'Invite the bot, then run `/connect <token>` in the target channel.'
          : 'Open one of the Telegram links and send the start command to finish linking.'
      )
    } catch (error) {
      setConnectionState('error')
      setConnectionMessage(getErrorMessage(error, 'Failed to generate a connection token'))
    }
  }

  const checkConnectionResult = useCallback(
    async (silent = false) => {
      if (!draft || !connectionToken) {
        return
      }

      const platform = getConnectPlatform(draft.service)
      if (!platform) {
        return
      }

      try {
        const response = await axios.get(`/api/connect/${platform}/${connectionToken.token}`)
        const result = response.data as Record<string, unknown>
        const hasConnectedValue = result.chat_id != null || result.channel_id != null

        if (!hasConnectedValue) {
          if (!silent) {
            setConnectionState('pending')
            setConnectionMessage('Still waiting for the bot to confirm the connection.')
          }
          return
        }

        setDraft((current) => {
          if (!current) {
            return current
          }

          const next = { ...current }
          for (const [key, value] of Object.entries(result)) {
            if (value != null) {
              next[key] = String(value)
            }
          }
          return next
        })
        setConnectionState('connected')
        setConnectionMessage(platform === 'discord' ? 'Discord channel linked.' : 'Telegram chat linked.')
      } catch (error) {
        if (axios.isAxiosError(error) && error.response?.status === 404) {
          if (!silent) {
            setConnectionState('pending')
            setConnectionMessage('Connection is still pending. Try again in a moment.')
          }
          return
        }

        if (!silent) {
          setConnectionState('error')
          setConnectionMessage(getErrorMessage(error, 'Failed to fetch connection result'))
        }
      }
    },
    [connectionToken, draft]
  )

  useEffect(() => {
    if (!connectionToken || !draft || connectionState === 'connected') {
      return
    }

    const platform = getConnectPlatform(draft.service)
    if (!platform) {
      return
    }

    const initialPoll = window.setTimeout(() => {
      void checkConnectionResult(true)
    }, 0)
    const timer = window.setInterval(() => {
      void checkConnectionResult(true)
    }, 2500)

    return () => {
      window.clearTimeout(initialPoll)
      window.clearInterval(timer)
    }
  }, [checkConnectionResult, connectionState, connectionToken, draft])

  const validationMessage = draft ? getRequiredFieldMessage(draft) : null

  return (
    <div className={styles.notificationConfig}>
      <div className={styles.sectionHeader}>
        <div>
          <p className={styles.sectionEyebrow}>Notification channels</p>
          <h2 className={styles.sectionTitle}>Choose notification channels</h2>
          <p className={styles.sectionText}>Add up to 10 destinations for this account.</p>
        </div>
        <span className={styles.limitBadge}>{settings.notify_settings.length}/10 configured</span>
      </div>

      <div className={styles.serviceGrid}>
        {notificationServices.map((service) => {
          const meta = serviceMeta[service]
          const isActive = draft?.service === service

          return (
            <button
              className={isActive ? `${styles.serviceCard} ${styles.serviceCardActive}` : styles.serviceCard}
              key={service}
              onClick={() => beginDraft(service)}
              type='button'
            >
              <span className={styles.serviceIcon}>{serviceIcons[service]}</span>
              <span className={styles.serviceCopy}>
                <span className={styles.serviceCardLabel}>{meta.label}</span>
                <span className={styles.serviceCardDescription}>{meta.description}</span>
              </span>
            </button>
          )
        })}
      </div>

      {draft ? (
        <div className={styles.editorCard}>
          <div className={styles.editorHeader}>
            <div>
              <h3 className={styles.editorTitle}>
                {editingIndex === null ? 'Add' : 'Edit'} {serviceMeta[draft.service].label}
              </h3>
              <p className={styles.editorText}>{serviceMeta[draft.service].description}</p>
            </div>
            <button className={styles.ghostButton} onClick={cancelDraft} type='button'>
              <FiX />
              Cancel
            </button>
          </div>

          {serviceMeta[draft.service].quickConnect ? (
            <div className={styles.connectCallout}>
              <div>
                <h4 className={styles.connectTitle}>Quick connect</h4>
                <p className={styles.connectText}>{serviceMeta[draft.service].quickConnect}</p>
              </div>
              {getConnectPlatform(draft.service) ? (
                <button className={styles.secondaryButton} onClick={() => void handleStartConnection()} type='button'>
                  <FiLink />
                  Generate token
                </button>
              ) : null}
            </div>
          ) : null}

          {connectionToken ? (
            <div className={styles.tokenCard}>
              <div className={styles.tokenRow}>
                <span className={styles.tokenLabel}>Connect token</span>
                <code className={styles.tokenValue}>{connectionToken.token}</code>
              </div>
              {connectionToken.bot_url ? (
                <div className={styles.linkRow}>
                  <a href={connectionToken.bot_url} rel='noopener noreferrer' target='_blank'>
                    <FiExternalLink />
                    Open bot
                  </a>
                  {connectionToken.bot_group_url ? (
                    <a href={connectionToken.bot_group_url} rel='noopener noreferrer' target='_blank'>
                      <FiExternalLink />
                      Open group flow
                    </a>
                  ) : null}
                </div>
              ) : null}
              <div className={styles.inlineActions}>
                <button className={styles.ghostButton} onClick={() => void checkConnectionResult(false)} type='button'>
                  <FiRefreshCw />
                  Check now
                </button>
                <span
                  className={
                    connectionState === 'connected'
                      ? `${styles.connectionHint} ${styles.connectionHintSuccess}`
                      : connectionState === 'error'
                        ? `${styles.connectionHint} ${styles.connectionHintError}`
                        : styles.connectionHint
                  }
                >
                  {connectionState === 'connected' ? <FiCheckCircle /> : null}
                  {connectionMessage}
                </span>
              </div>
            </div>
          ) : null}

          <div className={styles.fieldGrid}>
            {draft.service === 'telegram' ? (
              <>
                <Field hint='Required' label='Chat ID'>
                  <input
                    className={styles.input}
                    onChange={(event) => updateDraftField('chat_id', event.target.value)}
                    placeholder='379650434'
                    type='text'
                    value={draft.chat_id ?? ''}
                  />
                </Field>
                <Field hint='Optional custom bot token' label='Bot Token'>
                  <input
                    className={styles.input}
                    onChange={(event) => updateDraftField('token', event.target.value)}
                    placeholder='Leave empty to use the default bot'
                    type='text'
                    value={draft.token ?? ''}
                  />
                </Field>
                {draft.telegram_username ? (
                  <Field hint='Filled after quick connect' label='Telegram Username'>
                    <input className={styles.input} readOnly type='text' value={draft.telegram_username} />
                  </Field>
                ) : null}
              </>
            ) : null}

            {draft.service === 'discord_webhook' ? (
              <>
                <Field hint='Required' label='Webhook ID'>
                  <input
                    className={styles.input}
                    onChange={(event) => updateDraftField('webhook_id', event.target.value)}
                    placeholder='Webhook ID'
                    type='text'
                    value={draft.webhook_id ?? ''}
                  />
                </Field>
                <Field hint='Required' label='Webhook Token'>
                  <input
                    className={styles.input}
                    onChange={(event) => updateDraftField('webhook_token', event.target.value)}
                    placeholder='Webhook token'
                    type='text'
                    value={draft.webhook_token ?? ''}
                  />
                </Field>
                <Field label='Display Name'>
                  <input
                    className={styles.input}
                    onChange={(event) => updateDraftField('username', event.target.value)}
                    placeholder='Star++'
                    type='text'
                    value={draft.username ?? ''}
                  />
                </Field>
                <Field label='Avatar URL'>
                  <input
                    className={styles.input}
                    onChange={(event) => updateDraftField('avatar_url', event.target.value)}
                    placeholder='https://...'
                    type='text'
                    value={draft.avatar_url ?? ''}
                  />
                </Field>
                <Field hint='Hex value without #' label='Accent Color'>
                  <input
                    className={styles.input}
                    onChange={(event) => updateDraftField('color', event.target.value)}
                    placeholder='fd9a00'
                    type='text'
                    value={draft.color ?? ''}
                  />
                </Field>
              </>
            ) : null}

            {draft.service === 'discord_bot' ? (
              <>
                <Field hint='Required' label='Channel ID'>
                  <input
                    className={styles.input}
                    onChange={(event) => updateDraftField('channel_id', event.target.value)}
                    placeholder='Channel ID'
                    type='text'
                    value={draft.channel_id ?? ''}
                  />
                </Field>
                <Field hint='Optional custom bot token' label='Bot Token'>
                  <input
                    className={styles.input}
                    onChange={(event) => updateDraftField('token', event.target.value)}
                    placeholder='Leave empty to use the default bot'
                    type='text'
                    value={draft.token ?? ''}
                  />
                </Field>
                <Field label='Display Name'>
                  <input
                    className={styles.input}
                    onChange={(event) => updateDraftField('username', event.target.value)}
                    placeholder='Star++'
                    type='text'
                    value={draft.username ?? ''}
                  />
                </Field>
                <Field label='Avatar URL'>
                  <input
                    className={styles.input}
                    onChange={(event) => updateDraftField('avatar_url', event.target.value)}
                    placeholder='https://...'
                    type='text'
                    value={draft.avatar_url ?? ''}
                  />
                </Field>
                <Field hint='Hex value without #' label='Accent Color'>
                  <input
                    className={styles.input}
                    onChange={(event) => updateDraftField('color', event.target.value)}
                    placeholder='fd9a00'
                    type='text'
                    value={draft.color ?? ''}
                  />
                </Field>
              </>
            ) : null}

            {draft.service === 'bark' ? (
              <>
                <Field hint='Required' label='Key'>
                  <input
                    className={styles.input}
                    onChange={(event) => updateDraftField('key', event.target.value)}
                    placeholder='Your Bark key'
                    type='text'
                    value={draft.key ?? ''}
                  />
                </Field>
                <Field label='Server'>
                  <input
                    className={styles.input}
                    onChange={(event) => updateDraftField('server', event.target.value)}
                    placeholder='https://api.day.app/'
                    type='text'
                    value={draft.server ?? ''}
                  />
                </Field>
              </>
            ) : null}

            {draft.service === 'webhook' ? (
              <>
                <Field hint='Required' label='Webhook URL' wide>
                  <input
                    className={styles.input}
                    onChange={(event) => updateDraftField('url', event.target.value)}
                    placeholder='https://example.com/webhooks/stars'
                    type='text'
                    value={draft.url ?? ''}
                  />
                </Field>
                <Field label='Method'>
                  <select
                    className={styles.select}
                    onChange={(event) => updateDraftField('method', event.target.value)}
                    value={draft.method ?? 'POST'}
                  >
                    <option value='GET'>GET</option>
                    <option value='POST'>POST</option>
                    <option value='PUT'>PUT</option>
                  </select>
                </Field>
                <Field hint='Semicolon-separated key:value pairs' label='Headers' wide>
                  <textarea
                    className={styles.textarea}
                    onChange={(event) => updateDraftField('headers', event.target.value)}
                    placeholder='Authorization: Bearer token; Content-Type: application/json'
                    value={draft.headers ?? ''}
                  />
                </Field>
                <Field hint='Use {{.Title}} and {{.Message}} placeholders' label='Body Template' wide>
                  <textarea
                    className={styles.textarea}
                    onChange={(event) => updateDraftField('body', event.target.value)}
                    placeholder='{"title":"{{.Title}}","message":"{{.Message}}"}'
                    value={draft.body ?? ''}
                  />
                </Field>
              </>
            ) : null}
          </div>

          <div className={styles.editorFooter}>
            <div className={styles.validationMessage}>{validationMessage ?? 'Required fields are complete.'}</div>
            <div className={styles.inlineActions}>
              <button
                className={styles.primaryButton}
                disabled={Boolean(validationMessage)}
                onClick={handleApplyDraft}
                type='button'
              >
                {editingIndex === null ? <FiPlus /> : <FiEdit2 />}
                {editingIndex === null ? 'Add channel' : 'Update channel'}
              </button>
              <button className={styles.secondaryButton} onClick={cancelDraft} type='button'>
                <FiX />
                Cancel
              </button>
            </div>
          </div>
        </div>
      ) : (
        <div className={styles.emptyGuide}>
          <div className={styles.emptyGuideHeader}>
            <span className={styles.emptyGuideBadge}>Start here</span>
            <div aria-hidden='true' className={styles.emptyGuideIcons}>
              {notificationServices.map((service) => (
                <span className={styles.emptyGuideIcon} key={service}>
                  {serviceIcons[service]}
                </span>
              ))}
            </div>
          </div>
          <div className={styles.emptyGuideSteps}>
            <div className={`${styles.emptyGuideStep} ${styles.emptyGuideStepActive}`}>
              <span className={styles.emptyGuideStepNumber}>1</span>
              <span className={styles.emptyGuideStepLabel}>Pick channel</span>
            </div>
            <div className={styles.emptyGuideStep}>
              <span className={styles.emptyGuideStepNumber}>2</span>
              <span className={styles.emptyGuideStepLabel}>Fill details</span>
            </div>
            <div className={styles.emptyGuideStep}>
              <span className={styles.emptyGuideStepNumber}>3</span>
              <span className={styles.emptyGuideStepLabel}>Add channel</span>
            </div>
          </div>
        </div>
      )}

      <div className={styles.currentSettings}>
        <div className={styles.sectionHeader}>
          <div>
            <h3 className={styles.subsectionTitle}>Configured channels</h3>
            <p className={styles.sectionText}>Existing destinations can be edited or removed at any time.</p>
          </div>
        </div>

        {settings.notify_settings.length === 0 ? (
          <div className={styles.emptyState}>No notification channels configured yet.</div>
        ) : (
          <div className={styles.settingsGrid}>
            {settings.notify_settings.map((setting, index) => (
              <div className={styles.settingCard} key={getSettingKey(setting)}>
                <div className={styles.settingCardHeader}>
                  <div className={styles.settingCardTitle}>
                    <span className={styles.serviceIcon}>{serviceIcons[setting.service]}</span>
                    <div>
                      <strong>{serviceMeta[setting.service].label}</strong>
                      <p>{serviceMeta[setting.service].description}</p>
                    </div>
                  </div>
                  <span className={styles.settingIndex}>#{index + 1}</span>
                </div>
                <dl className={styles.detailList}>
                  {getSettingDetails(setting).map((detail) => (
                    <div className={styles.detailRow} key={`${setting.service}-${detail.label}`}>
                      <dt>{detail.label}</dt>
                      <dd>{detail.value}</dd>
                    </div>
                  ))}
                </dl>
                <div className={styles.cardActions}>
                  <button className={styles.ghostButton} onClick={() => handleEditService(index)} type='button'>
                    <FiEdit2 />
                    Edit
                  </button>
                  <button className={styles.dangerButton} onClick={() => handleRemoveService(index)} type='button'>
                    <FiTrash2 />
                    Remove
                  </button>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  )
}

export default NotificationConfig
