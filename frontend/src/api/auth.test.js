import { describe, expect, it, vi, beforeEach } from 'vitest'
import { listNotifications, listTrips } from './auth'

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

