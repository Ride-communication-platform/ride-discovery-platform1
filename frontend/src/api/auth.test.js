import { describe, expect, it, vi, beforeEach } from 'vitest'
import { listChatMessages, listChats, listNotifications, listTrips, sendChatMessage } from './auth'

describe('api/auth listTrips', () => {
  beforeEach(() => {
    vi.restoreAllMocks()
    vi.stubGlobal('fetch', vi.fn())
  })

  it('calls /api/trips with bearer token and returns payload', async () => {
    fetch.mockResolvedValue({
      ok: true,
      headers: { get: () => 'application/json' },
      json: async () => ({ trips: [{ id: 't1' }] }),
    })

    const res = await listTrips('abc123')

    expect(fetch).toHaveBeenCalledTimes(1)
    const [url, options] = fetch.mock.calls[0]
    expect(url).toMatch(/\/api\/trips$/)
    expect(options.method).toBe('GET')
    expect(options.headers.Authorization).toBe('Bearer abc123')
    expect(res.trips[0].id).toBe('t1')
  })
})

describe('api/auth listNotifications', () => {
  beforeEach(() => {
    vi.restoreAllMocks()
    vi.stubGlobal('fetch', vi.fn())
  })

  it('calls /api/notifications with bearer token and returns payload', async () => {
    fetch.mockResolvedValue({
      ok: true,
      headers: { get: () => 'application/json' },
      json: async () => ({ notifications: [{ id: 'n1' }] }),
    })

    const res = await listNotifications('abc123')

    expect(fetch).toHaveBeenCalledTimes(1)
    const [url, options] = fetch.mock.calls[0]
    expect(url).toMatch(/\/api\/notifications$/)
    expect(options.method).toBe('GET')
    expect(options.headers.Authorization).toBe('Bearer abc123')
    expect(res.notifications[0].id).toBe('n1')
  })
})

describe('api/auth chats', () => {
  beforeEach(() => {
    vi.restoreAllMocks()
    vi.stubGlobal('fetch', vi.fn())
  })

  it('calls /api/chats with bearer token and returns payload', async () => {
    fetch.mockResolvedValue({
      ok: true,
      headers: { get: () => 'application/json' },
      json: async () => ({ conversations: [{ id: 'c1' }] }),
    })

    const res = await listChats('abc123')

    expect(fetch).toHaveBeenCalledTimes(1)
    const [url, options] = fetch.mock.calls[0]
    expect(url).toMatch(/\/api\/chats$/)
    expect(options.method).toBe('GET')
    expect(options.headers.Authorization).toBe('Bearer abc123')
    expect(res.conversations[0].id).toBe('c1')
  })

  it('posts a chat message to the conversation endpoint', async () => {
    fetch.mockResolvedValue({
      ok: true,
      headers: { get: () => 'application/json' },
      json: async () => ({ chatMessage: { id: 'm1' } }),
    })

    const res = await sendChatMessage('abc123', 'chat-1', { body: 'hello' })

    expect(fetch).toHaveBeenCalledTimes(1)
    const [url, options] = fetch.mock.calls[0]
    expect(url).toMatch(/\/api\/chats\/chat-1\/messages$/)
    expect(options.method).toBe('POST')
    expect(options.headers.Authorization).toBe('Bearer abc123')
    expect(options.body).toBe(JSON.stringify({ body: 'hello' }))
    expect(res.chatMessage.id).toBe('m1')
  })

  it('gets chat messages from the conversation endpoint', async () => {
    fetch.mockResolvedValue({
      ok: true,
      headers: { get: () => 'application/json' },
      json: async () => ({ messages: [{ id: 'm1' }] }),
    })

    const res = await listChatMessages('abc123', 'chat-1')

    expect(fetch).toHaveBeenCalledTimes(1)
    const [url, options] = fetch.mock.calls[0]
    expect(url).toMatch(/\/api\/chats\/chat-1\/messages$/)
    expect(options.method).toBe('GET')
    expect(options.headers.Authorization).toBe('Bearer abc123')
    expect(res.messages[0].id).toBe('m1')
  })
})
