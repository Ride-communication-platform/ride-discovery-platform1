import React from 'react'
import { fireEvent, render, screen } from '@testing-library/react'
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
})
