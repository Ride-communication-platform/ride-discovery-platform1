import React from 'react'
import { fireEvent, render, screen, waitFor } from '@testing-library/react'
import App from './App'

vi.mock('@lottiefiles/dotlottie-react', () => ({
  DotLottieReact: () => <div data-testid="hero-lottie" />,
}))

vi.mock('react-leaflet', () => ({
  MapContainer: ({ children }) => <div data-testid="map-container">{children}</div>,
  TileLayer: () => null,
  Polyline: () => null,
  Marker: () => null,
  CircleMarker: () => null,
}))

describe('App auth screen', () => {
  beforeEach(() => {
    localStorage.clear()
    vi.restoreAllMocks()
    vi.stubGlobal('fetch', vi.fn())
  })

  it('renders login tab by default', () => {
    render(<App />)

    expect(screen.getByRole('tab', { name: /login/i })).toHaveAttribute('aria-selected', 'true')
    expect(screen.getByLabelText(/^email$/i)).toBeInTheDocument()
    expect(screen.getByLabelText(/^password$/i)).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /continue with google/i })).toBeInTheDocument()
  })

  it('switches to signup tab and shows full name field', () => {
    render(<App />)

    fireEvent.click(screen.getByRole('tab', { name: /sign up/i }))

    expect(screen.getByRole('tab', { name: /sign up/i })).toHaveAttribute('aria-selected', 'true')
    expect(screen.getByLabelText(/full name/i)).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /create free account/i })).toBeInTheDocument()
  })

  it('shows validation errors when login is submitted empty', async () => {
    render(<App />)

    fireEvent.click(screen.getByRole('button', { name: /^login$/i }))

    expect(await screen.findByText(/email is required/i)).toBeInTheDocument()
    expect(await screen.findByText(/password is required/i)).toBeInTheDocument()
  })

  it('refreshes trips and notifications after accepting a ride request', async () => {
    localStorage.setItem('ridex_token', 'test-token')

    let tripCalls = 0
    let notificationCalls = 0

    fetch.mockImplementation(async (url, options = {}) => {
      if (url.includes('/api/auth/me')) {
        return {
          ok: true,
          headers: { get: () => 'application/json' },
          json: async () => ({
            user: {
              id: 'driver-1',
              name: 'Driver Demo',
              email: 'driver@example.com',
              avatarData: '',
              interests: [],
              rating: 0,
              ratingCount: 0,
              tripsCompleted: 0,
              emailVerified: true,
              authProvider: 'password',
              createdAt: '2026-04-13T00:00:00Z',
            },
          }),
        }
      }

      if (url.endsWith('/api/ride-requests')) {
        return {
          ok: true,
          headers: { get: () => 'application/json' },
          json: async () => ({ requests: [] }),
        }
      }

      if (url.endsWith('/api/published-rides')) {
        return {
          ok: true,
          headers: { get: () => 'application/json' },
          json: async () => ({
            rides: [{ id: 'ride-1', fromLabel: 'Miami', toLabel: 'Orlando' }],
          }),
        }
      }

      if (url.includes('/api/ride-requests/feed')) {
        return {
          ok: true,
          headers: { get: () => 'application/json' },
          json: async () => ({
            requests: [
              {
                id: 'request-1',
                userId: 'rider-1',
                requesterName: 'Rider Demo',
                fromLabel: 'Miami',
                toLabel: 'Orlando',
                fromLat: 25.7617,
                fromLon: -80.1918,
                toLat: 28.5383,
                toLon: -81.3792,
                rideDate: '2026-01-15',
                rideTime: '07:30',
                passengers: 2,
                maxBudget: 50,
                routeMiles: 234,
                routeDuration: '4h 24m',
                notes: 'Need a morning ride.',
                rideType: 'shared',
                vehiclePreference: 'any',
                luggage: 'small',
              },
            ],
          }),
        }
      }

      if (url.endsWith('/api/trips')) {
        tripCalls += 1
        return {
          ok: true,
          headers: { get: () => 'application/json' },
          json: async () => ({
            trips: tripCalls > 1
              ? [{
                  id: 'trip-1',
                  fromLabel: 'Miami',
                  toLabel: 'Orlando',
                  rideDate: '2026-01-15',
                  rideTime: '07:30',
                  passengers: 2,
                  routeMiles: 234,
                  routeDuration: '4h 24m',
                  status: 'confirmed',
                  driverName: 'Driver Demo',
                  riderName: 'Rider Demo',
                }]
              : [],
          }),
        }
      }

      if (url.endsWith('/api/notifications')) {
        notificationCalls += 1
        return {
          ok: true,
          headers: { get: () => 'application/json' },
          json: async () => ({
            notifications: notificationCalls > 1
              ? [{ id: 'note-1', title: 'Ride request accepted', body: 'Trip confirmed.' }]
              : [],
          }),
        }
      }

      if (url.includes('/api/ride-requests/request-1/respond') && options.method === 'POST') {
        return {
          ok: true,
          headers: { get: () => 'application/json' },
          json: async () => ({ message: 'Ride request accepted.' }),
        }
      }

      return {
        ok: true,
        headers: { get: () => 'application/json' },
        json: async () => ({}),
      }
    })

    render(<App />)

    expect(await screen.findByText(/welcome back, driver demo/i)).toBeInTheDocument()

    fireEvent.click(screen.getByRole('heading', { name: /publish a ride/i }))
    fireEvent.click(screen.getByRole('button', { name: /view ride requests/i }))

    expect(await screen.findByText(/rider demo/i)).toBeInTheDocument()

    fireEvent.click(screen.getByRole('button', { name: /^accept$/i }))

    expect(await screen.findByText(/ride request accepted\./i)).toBeInTheDocument()

    await waitFor(() => {
      expect(tripCalls).toBeGreaterThan(1)
      expect(notificationCalls).toBeGreaterThan(1)
    })

    fireEvent.click(screen.getByRole('button', { name: /my trips/i }))

    expect(await screen.findByText((content) => content.includes('Miami') && content.includes('Orlando'))).toBeInTheDocument()

    fireEvent.click(screen.getByRole('button', { name: /notifications/i }))

    expect(await screen.findByText(/ride request accepted/i)).toBeInTheDocument()
  })
})
