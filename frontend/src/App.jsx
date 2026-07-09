import { useEffect, useMemo, useRef, useState } from 'react'
import { MapContainer, Marker, Polyline, TileLayer } from 'react-leaflet'
import { divIcon } from 'leaflet'
import { createPublishedRide, createRideRequest, deleteRideRequest, forgotPassword, getPublicUserProfile, listPublishedRides, listRideRequestFeed, listRideRequests, login, me, resendVerification, resetPassword, respondToRideRequest, signup, updateProfile, updateRideRequest, verifyEmail } from './api/auth'
import { DotLottieReact } from '@lottiefiles/dotlottie-react'
import 'leaflet/dist/leaflet.css'

const TOKEN_KEY = 'ridex_token'
const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080'

const emptyLogin = { email: '', password: '' }
const emptySignup = { name: '', email: '', password: '' }
const emptyVerify = { email: '', code: '' }
const emptyReset = { email: '', code: '', newPassword: '' }
const emptyRequest = {
  from: '',
  to: '',
  date: '',
  time: '',
  flexibility: 'exact',
  passengers: 1,
  luggage: 'none',
  maxBudget: '',
  rideType: 'shared',
  vehiclePreference: 'any',
  minimumRating: '',
  verifiedOnly: false,
  notes: '',
}
const emptyPublishRide = {
  from: '',
  to: '',
  date: '',
  time: '',
  flexibility: 'exact',
  availableSeats: 1,
  totalSeats: 1,
  pricePerSeat: '',
  vehicleType: 'any',
  luggageAllowed: 'small',
  rideType: 'shared',
  vehicleInfo: '',
  notes: '',
}
const emailPattern = /^[^\s@]+@[^\s@]+\.[^\s@]+$/
const AVATAR_CROP_SIZE = 220
const NOMINATIM_URL = 'https://nominatim.openstreetmap.org/search'
const OSRM_URL = 'https://router.project-osrm.org/route/v1/driving'

const pickupMarkerIcon = divIcon({
  className: 'map-marker-icon map-marker-icon-start',
  html: '<span class="map-marker-pin" aria-hidden="true">📍</span>',
  iconSize: [34, 34],
  iconAnchor: [17, 30],
})

const dropoffMarkerIcon = divIcon({
  className: 'map-marker-icon map-marker-icon-end',
  html: '<span class="map-marker-pin" aria-hidden="true">📍</span>',
  iconSize: [34, 34],
  iconAnchor: [17, 30],
})

function RideMarkIcon() {
  return <img src="/ridex-mark.svg" alt="" />
}

function NavIcon({ path }) {
  return (
    <svg viewBox="0 0 24 24" aria-hidden="true">
      <path d={path} fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round" />
    </svg>
  )
}

function GoogleIcon() {
  return (
    <img src="/icons/google.svg" alt="" />
  )
}

function GitHubIcon() {
  return (
    <img src="/icons/github.svg" alt="" />
  )
}

function getInitials(name) {
  return name
    .trim()
    .split(/\s+/)
    .slice(0, 2)
    .map((part) => part[0]?.toUpperCase())
    .join('') || 'RX'
}

function formatMemberSince(dateValue) {
  if (!dateValue) return 'Member since recently'
  const date = new Date(dateValue)
  if (Number.isNaN(date.getTime())) return 'Member since recently'
  return `Member since ${date.toLocaleDateString('en-US')}`
}

function clamp(value, min, max) {
  return Math.min(Math.max(value, min), max)
}

function getCoverScale(width, height, viewportSize) {
  if (!width || !height) return 1
  return Math.max(viewportSize / width, viewportSize / height)
}

function clampCropOffset(offset, naturalWidth, naturalHeight, zoom, viewportSize) {
  const baseScale = getCoverScale(naturalWidth, naturalHeight, viewportSize)
  const displayWidth = naturalWidth * baseScale * zoom
  const displayHeight = naturalHeight * baseScale * zoom
  const maxOffsetX = Math.max(0, (displayWidth - viewportSize) / 2)
  const maxOffsetY = Math.max(0, (displayHeight - viewportSize) / 2)

  return {
    x: clamp(offset.x, -maxOffsetX, maxOffsetX),
    y: clamp(offset.y, -maxOffsetY, maxOffsetY),
  }
}

function getPasswordStrength(password) {
  if (!password) return { score: 0, label: '' }
  let score = 0
  if (password.length >= 6) score += 1
  if (password.length >= 8) score += 1
  if (/[A-Z]/.test(password) || /[0-9]/.test(password)) score += 1
  if (/[^A-Za-z0-9]/.test(password)) score += 1

  if (score <= 1) return { score: 1, label: 'Weak' }
  if (score <= 3) return { score: 2, label: 'Medium' }
  return { score: 3, label: 'Strong' }
}

function normalizeLocation(value) {
  return value.trim().toLowerCase()
}

function formatRouteDuration(totalSeconds) {
  if (!totalSeconds) return '—'
  const roundedMinutes = Math.round(totalSeconds / 60)
  const hours = Math.floor(roundedMinutes / 60)
  const minutes = roundedMinutes % 60
  return `${hours > 0 ? `${hours}h ` : ''}${minutes}m`
}

function estimatePriceRange(distanceMiles, rideType) {
  const lowMultiplier = rideType === 'private' ? 0.24 : 0.16
  const highMultiplier = rideType === 'private' ? 0.34 : 0.23
  return `$${Math.max(10, Math.round(distanceMiles * lowMultiplier))} - $${Math.max(16, Math.round(distanceMiles * highMultiplier))}`
}

function buildRoutePath(coordinates, width = 360, height = 148) {
  if (!coordinates?.length) return ''
  const lons = coordinates.map(([lon]) => lon)
  const lats = coordinates.map(([, lat]) => lat)
  const minLon = Math.min(...lons)
  const maxLon = Math.max(...lons)
  const minLat = Math.min(...lats)
  const maxLat = Math.max(...lats)
  const lonRange = maxLon - minLon || 1
  const latRange = maxLat - minLat || 1
  const padding = 18

  return coordinates
    .map(([lon, lat], index) => {
      const x = padding + ((lon - minLon) / lonRange) * (width - padding * 2)
      const y = height - padding - ((lat - minLat) / latRange) * (height - padding * 2)
      return `${index === 0 ? 'M' : 'L'} ${x.toFixed(1)} ${y.toFixed(1)}`
    })
    .join(' ')
}

async function fetchLocationSuggestions(query, signal) {
  const params = new URLSearchParams({
    q: query,
    format: 'jsonv2',
    addressdetails: '1',
    limit: '5',
  })
  const response = await fetch(`${NOMINATIM_URL}?${params.toString()}`, {
    signal,
    headers: {
      Accept: 'application/json',
    },
  })
  if (!response.ok) throw new Error('Could not load location suggestions')
  const results = await response.json()
  return results.map((item) => ({
    id: item.place_id,
    label: item.display_name,
    lat: Number(item.lat),
    lon: Number(item.lon),
  }))
}

async function fetchRouteData(fromLocation, toLocation, rideType, signal) {
  const response = await fetch(
    `${OSRM_URL}/${fromLocation.lon},${fromLocation.lat};${toLocation.lon},${toLocation.lat}?overview=full&geometries=geojson`,
    { signal, headers: { Accept: 'application/json' } },
  )
  if (!response.ok) throw new Error('Could not load route preview')
  const data = await response.json()
  const route = data.routes?.[0]
  if (!route) throw new Error('No route found for this trip')
  const miles = Math.round(route.distance / 1609.34)
  return {
    miles,
    duration: formatRouteDuration(route.duration),
    estimate: estimatePriceRange(miles, rideType),
    path: buildRoutePath(route.geometry?.coordinates || []),
    coordinates: (route.geometry?.coordinates || []).map(([lon, lat]) => [lat, lon]),
  }
}

function App() {
  const [activeTab, setActiveTab] = useState('login')
  const [loginForm, setLoginForm] = useState(emptyLogin)
  const [signupForm, setSignupForm] = useState(emptySignup)
  const [errors, setErrors] = useState({})
  const [loading, setLoading] = useState(false)
  const [verifyLoading, setVerifyLoading] = useState(false)
  const [booting, setBooting] = useState(true)
  const [banner, setBanner] = useState('')
  const [user, setUser] = useState(null)
  const [verifyForm, setVerifyForm] = useState(emptyVerify)
  const [resetForm, setResetForm] = useState(emptyReset)
  const [activeOverlay, setActiveOverlay] = useState('')
  const [overlayNotice, setOverlayNotice] = useState('')
  const [profileForm, setProfileForm] = useState({ name: '', interests: [] })
  const [profileEditing, setProfileEditing] = useState(false)
  const [profileNotice, setProfileNotice] = useState('')
  const [profileSaving, setProfileSaving] = useState(false)
  const [avatarUploading, setAvatarUploading] = useState(false)
  const [interestInput, setInterestInput] = useState('')
  const [avatarDraftSrc, setAvatarDraftSrc] = useState('')
  const [avatarDraftZoom, setAvatarDraftZoom] = useState(1)
  const [avatarDraftOffset, setAvatarDraftOffset] = useState({ x: 0, y: 0 })
  const [avatarDraftNatural, setAvatarDraftNatural] = useState({ width: 0, height: 0 })
  const [activeView, setActiveView] = useState('home')
  const [requestMode, setRequestMode] = useState('create')
  const [requestForm, setRequestForm] = useState(emptyRequest)
  const [requestErrors, setRequestErrors] = useState({})
  const [requestNotice, setRequestNotice] = useState('')
  const [requestSaving, setRequestSaving] = useState(false)
  const [editingRequestId, setEditingRequestId] = useState('')
  const [activeLocationField, setActiveLocationField] = useState('')
  const [requestSuggestions, setRequestSuggestions] = useState({ from: [], to: [] })
  const [locationLoading, setLocationLoading] = useState({ from: false, to: false })
  const [selectedLocations, setSelectedLocations] = useState({ from: null, to: null })
  const [routePreview, setRoutePreview] = useState({ loading: false, error: '', data: null })
  const [rideRequests, setRideRequests] = useState([])
  const [publishedRides, setPublishedRides] = useState([])
  const [publishMode, setPublishMode] = useState('create')
  const [publishForm, setPublishForm] = useState(emptyPublishRide)
  const [publishErrors, setPublishErrors] = useState({})
  const [publishNotice, setPublishNotice] = useState('')
  const [publishSaving, setPublishSaving] = useState(false)
  const [activePublishLocationField, setActivePublishLocationField] = useState('')
  const [publishSuggestions, setPublishSuggestions] = useState({ from: [], to: [] })
  const [publishLocationLoading, setPublishLocationLoading] = useState({ from: false, to: false })
  const [selectedPublishLocations, setSelectedPublishLocations] = useState({ from: null, to: null })
  const [publishRoutePreview, setPublishRoutePreview] = useState({ loading: false, error: '', data: null })
  const [rideRequestFeed, setRideRequestFeed] = useState([])
  const [rideFeedFilters, setRideFeedFilters] = useState({ date: '', route: '', passengers: '', budget: '' })
  const [rideFeedLoading, setRideFeedLoading] = useState(false)
  const [rideFeedNotice, setRideFeedNotice] = useState('')
  const [expandedFeedRequestID, setExpandedFeedRequestID] = useState('')
  const [requestActionLoading, setRequestActionLoading] = useState('')
  const [activeFeedRequest, setActiveFeedRequest] = useState(null)
  const [activeFeedRoutePreview, setActiveFeedRoutePreview] = useState({ loading: false, error: '', data: null })
  const [requesterProfile, setRequesterProfile] = useState(null)
  const [requesterProfileLoading, setRequesterProfileLoading] = useState(false)
  const [requesterProfileError, setRequesterProfileError] = useState('')
  const [cancelRequestLoading, setCancelRequestLoading] = useState('')
  const [editRequestLoading, setEditRequestLoading] = useState('')
  const [homeNotice, setHomeNotice] = useState('')
  const dragStateRef = useRef(null)

  useEffect(() => {
    const params = new URLSearchParams(window.location.search)
    const redirectedToken = params.get('token')
    const redirectedError = params.get('auth_error')

    if (redirectedToken) {
      localStorage.setItem(TOKEN_KEY, redirectedToken)
      window.history.replaceState({}, '', window.location.pathname)
    } else if (redirectedError) {
      setErrors((prev) => ({ ...prev, form: redirectedError }))
      window.history.replaceState({}, '', window.location.pathname)
    }

    const token = localStorage.getItem(TOKEN_KEY)
    if (!token) {
      setBooting(false)
      return
    }

    me(token)
      .then((res) => {
        setUser(res.user)
      })
      .catch(() => {
        localStorage.removeItem(TOKEN_KEY)
      })
      .finally(() => {
        setBooting(false)
      })
  }, [])

  const currentForm = useMemo(() => (activeTab === 'login' ? loginForm : signupForm), [activeTab, loginForm, signupForm])
  const passwordStrength = useMemo(() => getPasswordStrength(signupForm.password), [signupForm.password])
  const routeSummary = routePreview.data
  const fromSuggestions = requestSuggestions.from
  const toSuggestions = requestSuggestions.to
  const publishRouteSummary = publishRoutePreview.data
  const publishFromSuggestions = publishSuggestions.from
  const publishToSuggestions = publishSuggestions.to
  const estimatedEarnings = useMemo(() => {
    const seats = Number(publishForm.availableSeats || 0)
    const price = Number(publishForm.pricePerSeat || 0)
    return seats > 0 && price > 0 ? (seats * price).toFixed(0) : ''
  }, [publishForm.availableSeats, publishForm.pricePerSeat])

  useEffect(() => {
    if (!user) return
    setProfileForm({
      name: user.name || '',
      interests: user.interests || [],
    })
    setProfileEditing(false)
    setProfileNotice('')
    setInterestInput('')
  }, [user])

  useEffect(() => {
    if (!user) {
      setRideRequests([])
      setPublishedRides([])
      setHomeNotice('')
      return
    }
    const token = localStorage.getItem(TOKEN_KEY)
    if (!token) return
    listRideRequests(token)
      .then((res) => setRideRequests(res.requests || []))
      .catch(() => {})
    listPublishedRides(token)
      .then((res) => setPublishedRides(res.rides || []))
      .catch(() => {})
  }, [user])

  const handleCancelRideRequest = async (requestID) => {
    const token = localStorage.getItem(TOKEN_KEY)
    if (!token || !requestID) return

    setCancelRequestLoading(requestID)
    setHomeNotice('')
    try {
      const response = await deleteRideRequest(token, requestID)
      setRideRequests((prev) => prev.filter((request) => request.id !== requestID))
      if (editingRequestId === requestID) {
        resetRequestComposer()
      }
      setHomeNotice(response.message || 'Ride request cancelled.')
      setRequestNotice(response.message || 'Ride request cancelled.')
    } catch (error) {
      setHomeNotice(error.message)
      setRequestNotice(error.message)
    } finally {
      setCancelRequestLoading('')
    }
  }

  const handleEditRideRequest = (request) => {
    setEditRequestLoading(request.id)
    loadRequestIntoForm(request)
    setEditRequestLoading('')
  }

  useEffect(() => {
    const token = localStorage.getItem(TOKEN_KEY)
    if (!user || !token) return
    setRideFeedLoading(true)
    setRideFeedNotice('')
    listRideRequestFeed(token, rideFeedFilters)
      .then((res) => {
        const requests = Array.isArray(res.requests) ? res.requests : []
        setRideRequestFeed(requests.filter((request) => request.userId !== user.id))
      })
      .catch((error) => setRideFeedNotice(error.message))
      .finally(() => setRideFeedLoading(false))
  }, [user, rideFeedFilters])

  const switchTab = (tab) => {
    setErrors({})
    setBanner('')
    setOverlayNotice('')
    setActiveOverlay('')
    setActiveTab(tab)
  }

  const validate = () => {
    const next = {}

    if (activeTab === 'signup') {
      if (!signupForm.name.trim()) next.name = 'Full name is required'
      if (!signupForm.email.trim()) next.email = 'Email is required'
      else if (!emailPattern.test(signupForm.email.trim())) next.email = 'Enter a valid email address'
      if (!signupForm.password) next.password = 'Password is required'
      if (signupForm.password && signupForm.password.length < 8) next.password = 'Password must be at least 8 characters'
      if (signupForm.password && !/[A-Za-z]/.test(signupForm.password)) next.password = 'Password must include at least one letter'
      if (signupForm.password && !/[0-9]/.test(signupForm.password)) next.password = 'Password must include at least one number'
    }

    if (activeTab === 'login') {
      if (!loginForm.email.trim()) next.email = 'Email is required'
      else if (!emailPattern.test(loginForm.email.trim())) next.email = 'Enter a valid email address'
      if (!loginForm.password) next.password = 'Password is required'
    }

    setErrors(next)
    return Object.keys(next).length === 0
  }

  const validateField = (field, value) => {
    let message = ''

    if (field === 'name' && activeTab === 'signup' && !value.trim()) {
      message = 'Full name is required'
    }

    if (field === 'email') {
      if (!value.trim()) message = 'Email is required'
      else if (!emailPattern.test(value.trim())) message = 'Enter a valid email address'
    }

    if (field === 'password') {
      if (!value) message = 'Password is required'
      else if (activeTab === 'signup' && value.length < 8) message = 'Password must be at least 8 characters'
      else if (activeTab === 'signup' && !/[A-Za-z]/.test(value)) message = 'Password must include at least one letter'
      else if (activeTab === 'signup' && !/[0-9]/.test(value)) message = 'Password must include at least one number'
    }

    setErrors((prev) => {
      const next = { ...prev }
      if (message) next[field] = message
      else delete next[field]
      return next
    })
  }

  const handleSubmit = async (e) => {
    e.preventDefault()
    setBanner('')

    if (!validate()) return

    setLoading(true)
    try {
      if (activeTab === 'signup') {
        const res = await signup(signupForm)
        setSignupForm(emptySignup)
        setErrors({})
        setActiveTab('login')
        setVerifyForm({ email: signupForm.email, code: '' })
        setActiveOverlay('verify')
        setOverlayNotice('')
        setBanner(res.message)
      } else {
        const res = await login(loginForm)
        localStorage.setItem(TOKEN_KEY, res.token)
        setUser(res.user)
      }
    } catch (err) {
      if (err.message.includes('Email not verified')) {
        setVerifyForm((prev) => ({ ...prev, email: loginForm.email }))
        setActiveOverlay('verify')
      }
      setErrors((prev) => ({ ...prev, form: err.message }))
    } finally {
      setLoading(false)
    }
  }

  const handleVerify = async (e) => {
    e.preventDefault()
    setErrors((prev) => {
      const next = { ...prev }
      delete next.verify
      return next
    })

    if (!verifyForm.email.trim() || !verifyForm.code.trim()) {
      setErrors((prev) => ({ ...prev, verify: 'Email and verification code are required' }))
      return
    }
    if (!emailPattern.test(verifyForm.email.trim())) {
      setErrors((prev) => ({ ...prev, verify: 'Enter a valid email address' }))
      return
    }

    setVerifyLoading(true)
    try {
      const res = await verifyEmail(verifyForm)
      setBanner(res.message)
      setErrors({})
      setVerifyForm(emptyVerify)
      setOverlayNotice('')
      setActiveOverlay('')
      setActiveTab('login')
    } catch (err) {
      setErrors((prev) => ({ ...prev, verify: err.message }))
    } finally {
      setVerifyLoading(false)
    }
  }

  const handleResendVerification = async () => {
    if (!verifyForm.email.trim() || !emailPattern.test(verifyForm.email.trim())) {
      setErrors((prev) => ({ ...prev, verify: 'Enter a valid email before requesting a new code' }))
      return
    }

    setVerifyLoading(true)
    try {
      const res = await resendVerification({ email: verifyForm.email })
      setOverlayNotice(res.message)
      setErrors({})
    } catch (err) {
      setErrors((prev) => ({ ...prev, verify: err.message }))
    } finally {
      setVerifyLoading(false)
    }
  }

  const handleForgotPassword = async () => {
    if (!resetForm.email.trim() || !emailPattern.test(resetForm.email.trim())) {
      setErrors((prev) => ({ ...prev, reset: 'Enter a valid email address' }))
      return
    }

    setVerifyLoading(true)
    try {
      const res = await forgotPassword({ email: resetForm.email })
      setOverlayNotice(res.message)
      setErrors((prev) => {
        const next = { ...prev }
        delete next.reset
        return next
      })
      setActiveOverlay('reset')
    } catch (err) {
      setErrors((prev) => ({ ...prev, reset: err.message }))
    } finally {
      setVerifyLoading(false)
    }
  }

  const handleResetPassword = async () => {
    if (!resetForm.email.trim() || !resetForm.code.trim() || !resetForm.newPassword.trim()) {
      setErrors((prev) => ({ ...prev, reset: 'Email, reset code, and new password are required' }))
      return
    }
    if (!emailPattern.test(resetForm.email.trim())) {
      setErrors((prev) => ({ ...prev, reset: 'Enter a valid email address' }))
      return
    }

    setVerifyLoading(true)
    try {
      const res = await resetPassword(resetForm)
      setBanner(res.message)
      setErrors((prev) => {
        const next = { ...prev }
        delete next.reset
        return next
      })
      setResetForm(emptyReset)
      setOverlayNotice('')
      setActiveOverlay('')
    } catch (err) {
      setErrors((prev) => ({ ...prev, reset: err.message }))
    } finally {
      setVerifyLoading(false)
    }
  }

  const handleSocialClick = (provider) => {
    if (provider === 'Google') {
      window.location.href = `${API_BASE_URL}/api/auth/oauth/google/start`
      return
    }
    if (provider === 'GitHub') {
      window.location.href = `${API_BASE_URL}/api/auth/oauth/github/start`
      return
    }
    setErrors((prev) => ({
      ...prev,
      form: `${provider} sign-in requires OAuth client credentials and redirect setup. The UI entry point is added, but the provider is not configured in this environment.`,
    }))
  }

  const handleLogout = () => {
    localStorage.removeItem(TOKEN_KEY)
    setUser(null)
    setLoginForm(emptyLogin)
    setErrors({})
    setBanner('')
    setActiveTab('login')
    setActiveView('home')
  }

  const handleProfileSave = async () => {
    const token = localStorage.getItem(TOKEN_KEY)
    const name = profileForm.name.trim()

    if (!name) {
      setProfileNotice('Name is required.')
      return
    }
    if (!token) {
      handleLogout()
      return
    }

    setProfileSaving(true)
    try {
      const interests = profileForm.interests
      const res = await updateProfile(token, { name, interests })
      setUser(res.user)
      setProfileForm({ name: res.user.name || '', interests: res.user.interests || [] })
      setProfileEditing(false)
      setProfileNotice(res.message)
    } catch (err) {
      if (err.message.toLowerCase().includes('invalid token')) {
        handleLogout()
        return
      }
      setProfileNotice(err.message)
    } finally {
      setProfileSaving(false)
    }
  }

  const handleAvatarUpload = async (e) => {
    const file = e.target.files?.[0]
    e.target.value = ''
    if (!file) return
    if (!file.type.startsWith('image/')) {
      setProfileNotice('Please choose an image file.')
      return
    }
    if (file.size > 1_500_000) {
      setProfileNotice('Image must be smaller than 1.5 MB.')
      return
    }

    const reader = new FileReader()
    reader.onload = () => {
      const draftSrc = reader.result?.toString() || ''
      if (!draftSrc) {
        setProfileNotice('Could not read the selected image.')
        return
      }
      setAvatarDraftSrc(draftSrc)
      setAvatarDraftZoom(1)
      setAvatarDraftOffset({ x: 0, y: 0 })
      setAvatarDraftNatural({ width: 0, height: 0 })
      setActiveOverlay('avatarCrop')
    }
    reader.readAsDataURL(file)
  }

  const handleAvatarImageLoad = (e) => {
    setAvatarDraftNatural({
      width: e.currentTarget.naturalWidth,
      height: e.currentTarget.naturalHeight,
    })
    setAvatarDraftOffset({ x: 0, y: 0 })
  }

  const handleAvatarDragStart = (e) => {
    if (!avatarDraftNatural.width || !avatarDraftNatural.height) return
    dragStateRef.current = {
      pointerId: e.pointerId,
      startX: e.clientX,
      startY: e.clientY,
      startOffset: avatarDraftOffset,
    }
    e.currentTarget.setPointerCapture(e.pointerId)
  }

  const handleAvatarDragMove = (e) => {
    const dragState = dragStateRef.current
    if (!dragState || dragState.pointerId !== e.pointerId) return

    setAvatarDraftOffset(
      clampCropOffset(
        {
          x: dragState.startOffset.x + (e.clientX - dragState.startX),
          y: dragState.startOffset.y + (e.clientY - dragState.startY),
        },
        avatarDraftNatural.width,
        avatarDraftNatural.height,
        avatarDraftZoom,
        AVATAR_CROP_SIZE,
      ),
    )
  }

  const handleAvatarDragEnd = (e) => {
    if (dragStateRef.current?.pointerId !== e.pointerId) return
    dragStateRef.current = null
    e.currentTarget.releasePointerCapture(e.pointerId)
  }

  const handleAvatarZoomChange = (e) => {
    const nextZoom = Number(e.target.value)
    setAvatarDraftZoom(nextZoom)
    setAvatarDraftOffset((prev) =>
      clampCropOffset(prev, avatarDraftNatural.width, avatarDraftNatural.height, nextZoom, AVATAR_CROP_SIZE),
    )
  }

  const closeAvatarCrop = () => {
    setActiveOverlay('')
    setAvatarDraftSrc('')
    setAvatarDraftZoom(1)
    setAvatarDraftOffset({ x: 0, y: 0 })
    setAvatarDraftNatural({ width: 0, height: 0 })
    dragStateRef.current = null
  }

  const handleApplyAvatarCrop = async () => {
    const token = localStorage.getItem(TOKEN_KEY)
    if (!token) {
      handleLogout()
      return
    }
    if (!avatarDraftSrc || !avatarDraftNatural.width || !avatarDraftNatural.height) {
      setProfileNotice('Could not prepare the selected image.')
      return
    }

    const image = new Image()
    image.src = avatarDraftSrc
    await new Promise((resolve, reject) => {
      image.onload = resolve
      image.onerror = reject
    })

    const baseScale = getCoverScale(avatarDraftNatural.width, avatarDraftNatural.height, AVATAR_CROP_SIZE)
    const scaledWidth = avatarDraftNatural.width * baseScale * avatarDraftZoom
    const scaledHeight = avatarDraftNatural.height * baseScale * avatarDraftZoom
    const outputSize = 320
    const canvas = document.createElement('canvas')
    canvas.width = outputSize
    canvas.height = outputSize
    const ctx = canvas.getContext('2d')

    if (!ctx) {
      setProfileNotice('Could not prepare the cropped image.')
      return
    }

    const ratio = outputSize / AVATAR_CROP_SIZE
    const dx = (outputSize - scaledWidth * ratio) / 2 + avatarDraftOffset.x * ratio
    const dy = (outputSize - scaledHeight * ratio) / 2 + avatarDraftOffset.y * ratio

    ctx.drawImage(image, dx, dy, scaledWidth * ratio, scaledHeight * ratio)
    const avatarData = canvas.toDataURL('image/jpeg', 0.92)

    setAvatarUploading(true)
    try {
      const res = await updateProfile(token, {
        name: profileForm.name.trim() || user.name || '',
        avatarData,
        interests: profileForm.interests,
      })
      setUser(res.user)
      setProfileForm((prev) => ({ ...prev, interests: res.user.interests || prev.interests }))
      setProfileNotice('Avatar updated successfully.')
      closeAvatarCrop()
    } catch (err) {
      if (err.message.toLowerCase().includes('invalid token')) {
        handleLogout()
        return
      }
      setProfileNotice(err.message)
    } finally {
      setAvatarUploading(false)
    }
  }

  const handleInterestKeyDown = (e) => {
    if (e.key !== 'Enter') return
    e.preventDefault()
    if (!profileEditing) return
    const nextValue = interestInput.trim()
    if (!nextValue) return
    setProfileForm((prev) => {
      if (prev.interests.some((interest) => interest.toLowerCase() === nextValue.toLowerCase())) {
        return prev
      }
      if (prev.interests.length >= 10) return prev
      return { ...prev, interests: [...prev.interests, nextValue] }
    })
    setInterestInput('')
  }

  const handleRemoveInterest = (interestToRemove) => {
    if (!profileEditing) return
    setProfileForm((prev) => ({
      ...prev,
      interests: prev.interests.filter((interest) => interest !== interestToRemove),
    }))
  }

  const handleRequestFieldChange = (field, value) => {
    setRequestForm((prev) => ({ ...prev, [field]: value }))
    if (field === 'from' || field === 'to') {
      setSelectedLocations((prev) => ({ ...prev, [field]: null }))
      setRoutePreview((prev) => ({ ...prev, data: null, error: '' }))
    }
    setRequestErrors((prev) => {
      const next = { ...prev }
      delete next[field]
      return next
    })
    setRequestNotice('')
  }

  const resetRequestComposer = () => {
    setRequestForm(emptyRequest)
    setEditingRequestId('')
    setSelectedLocations({ from: null, to: null })
    setRequestSuggestions({ from: [], to: [] })
    setActiveLocationField('')
    setRoutePreview({ loading: false, error: '', data: null })
    setRequestErrors({})
  }

  const loadRequestIntoForm = (request) => {
    setRequestForm({
      from: request.fromLabel || '',
      to: request.toLabel || '',
      date: request.rideDate || '',
      time: request.rideTime || '',
      flexibility: request.flexibility || 'exact',
      passengers: Number(request.passengers || 1),
      luggage: request.luggage || 'none',
      maxBudget: request.maxBudget ? String(request.maxBudget) : '',
      rideType: request.rideType || 'shared',
      vehiclePreference: request.vehiclePreference || 'any',
      minimumRating: request.minimumRating ? String(request.minimumRating) : '',
      verifiedOnly: Boolean(request.verifiedDriversOnly),
      notes: request.notes || '',
    })
    setSelectedLocations({
      from: {
        id: `${request.id}-from`,
        label: request.fromLabel,
        lat: Number(request.fromLat),
        lon: Number(request.fromLon),
      },
      to: {
        id: `${request.id}-to`,
        label: request.toLabel,
        lat: Number(request.toLat),
        lon: Number(request.toLon),
      },
    })
    setEditingRequestId(request.id)
    setRequestSuggestions({ from: [], to: [] })
    setActiveLocationField('')
    setRequestErrors({})
    setRequestNotice('')
    setRequestMode('create')
  }

  const handlePublishFieldChange = (field, value) => {
    setPublishForm((prev) => ({ ...prev, [field]: value }))
    if (field === 'from' || field === 'to') {
      setSelectedPublishLocations((prev) => ({ ...prev, [field]: null }))
      setPublishRoutePreview({ loading: false, error: '', data: null })
    }
    setPublishErrors((prev) => {
      const next = { ...prev }
      delete next[field]
      return next
    })
    setPublishNotice('')
  }

  const selectPublishLocationSuggestion = (field, value) => {
    setPublishForm((prev) => ({ ...prev, [field]: value.label }))
    setSelectedPublishLocations((prev) => ({ ...prev, [field]: value }))
    setPublishSuggestions((prev) => ({ ...prev, [field]: [] }))
    setActivePublishLocationField('')
    setPublishErrors((prev) => {
      const next = { ...prev }
      delete next[field]
      return next
    })
  }

  const selectLocationSuggestion = (field, value) => {
    setRequestForm((prev) => ({ ...prev, [field]: value.label }))
    setSelectedLocations((prev) => ({ ...prev, [field]: value }))
    setRequestSuggestions((prev) => ({ ...prev, [field]: [] }))
    setRequestErrors((prev) => {
      const next = { ...prev }
      delete next[field]
      return next
    })
    setActiveLocationField('')
  }

  const validateRequest = () => {
    const next = {}
    if (!requestForm.from.trim()) next.from = 'Pickup location is required'
    if (!requestForm.to.trim()) next.to = 'Destination is required'
    if (requestForm.from.trim() && requestForm.to.trim() && normalizeLocation(requestForm.from) === normalizeLocation(requestForm.to)) {
      next.to = 'Destination must be different from pickup'
    }
    if (!requestForm.date) next.date = 'Date is required'
    if (!requestForm.time) next.time = 'Time is required'
    if (requestForm.maxBudget && Number(requestForm.maxBudget) <= 0) next.maxBudget = 'Budget must be greater than 0'
    if (requestForm.notes.length > 220) next.notes = 'Notes should be under 220 characters'
    setRequestErrors(next)
    return Object.keys(next).length === 0
  }

  const handleRequestSubmit = async (e) => {
    e.preventDefault()
    if (!validateRequest()) return
    const token = localStorage.getItem(TOKEN_KEY)
    if (!token) {
      handleLogout()
      return
    }
    if (!selectedLocations.from || !selectedLocations.to) {
      setRequestErrors((prev) => ({
        ...prev,
        from: prev.from || 'Select a pickup location from suggestions',
        to: prev.to || 'Select a destination from suggestions',
      }))
      return
    }

    setRequestSaving(true)
    try {
      const payload = {
        fromLabel: requestForm.from,
        fromLat: selectedLocations.from.lat,
        fromLon: selectedLocations.from.lon,
        toLabel: requestForm.to,
        toLat: selectedLocations.to.lat,
        toLon: selectedLocations.to.lon,
        rideDate: requestForm.date,
        rideTime: requestForm.time,
        flexibility: requestForm.flexibility,
        passengers: requestForm.passengers,
        luggage: requestForm.luggage,
        maxBudget: requestForm.maxBudget ? Number(requestForm.maxBudget) : 0,
        rideType: requestForm.rideType,
        vehiclePreference: requestForm.vehiclePreference,
        minimumRating: requestForm.minimumRating ? Number(requestForm.minimumRating) : 0,
        verifiedDriversOnly: requestForm.verifiedOnly,
        notes: requestForm.notes,
        routeMiles: routeSummary?.miles || 0,
        routeDuration: routeSummary?.duration || '',
        priceEstimate: routeSummary?.estimate || '',
      }
      const res = editingRequestId
        ? await updateRideRequest(token, editingRequestId, payload)
        : await createRideRequest(token, payload)
      setRideRequests((prev) => (
        editingRequestId
          ? prev.map((request) => (request.id === editingRequestId ? res.request : request))
          : [res.request, ...prev]
      ))
      setRequestNotice(res.message)
      resetRequestComposer()
      setRequestMode('view')
    } catch (error) {
      setRequestNotice('')
      setRequestErrors((prev) => ({ ...prev, form: error.message }))
    } finally {
      setRequestSaving(false)
    }
  }

  const validatePublishRide = () => {
    const next = {}
    if (!publishForm.from.trim()) next.from = 'Pickup location is required'
    if (!publishForm.to.trim()) next.to = 'Drop-off location is required'
    if (!publishForm.date) next.date = 'Date is required'
    if (!publishForm.time) next.time = 'Time is required'
    if (!publishForm.pricePerSeat || Number(publishForm.pricePerSeat) <= 0) next.pricePerSeat = 'Price per seat must be greater than 0'
    if (publishForm.availableSeats < 1) next.availableSeats = 'Available seats must be at least 1'
    if (publishForm.totalSeats < 1) next.totalSeats = 'Total seats must be at least 1'
    if (Number(publishForm.availableSeats) > Number(publishForm.totalSeats)) next.totalSeats = 'Total seats must be greater than or equal to available seats'
    setPublishErrors(next)
    return Object.keys(next).length === 0
  }

  const handlePublishRideSubmit = async (e) => {
    e.preventDefault()
    if (!validatePublishRide()) return
    const token = localStorage.getItem(TOKEN_KEY)
    if (!token) {
      handleLogout()
      return
    }
    if (!selectedPublishLocations.from || !selectedPublishLocations.to) {
      setPublishErrors((prev) => ({
        ...prev,
        from: prev.from || 'Select a pickup location from suggestions',
        to: prev.to || 'Select a drop-off location from suggestions',
      }))
      return
    }

    setPublishSaving(true)
    try {
      const res = await createPublishedRide(token, {
        fromLabel: publishForm.from,
        fromLat: selectedPublishLocations.from.lat,
        fromLon: selectedPublishLocations.from.lon,
        toLabel: publishForm.to,
        toLat: selectedPublishLocations.to.lat,
        toLon: selectedPublishLocations.to.lon,
        rideDate: publishForm.date,
        rideTime: publishForm.time,
        flexibility: publishForm.flexibility,
        availableSeats: Number(publishForm.availableSeats),
        totalSeats: Number(publishForm.totalSeats),
        pricePerSeat: Number(publishForm.pricePerSeat),
        vehicleType: publishForm.vehicleType,
        luggageAllowed: publishForm.luggageAllowed,
        rideType: publishForm.rideType,
        vehicleInfo: publishForm.vehicleInfo,
        notes: publishForm.notes,
        routeMiles: publishRouteSummary?.miles || 0,
        routeDuration: publishRouteSummary?.duration || '',
        earningsEstimate: Number(estimatedEarnings || 0),
      })
      setPublishedRides((prev) => [res.ride, ...prev])
      setPublishNotice(res.message)
      setPublishForm(emptyPublishRide)
      setSelectedPublishLocations({ from: null, to: null })
      setPublishSuggestions({ from: [], to: [] })
      setPublishRoutePreview({ loading: false, error: '', data: null })
      setPublishErrors({})
    } catch (error) {
      setPublishNotice('')
      setPublishErrors((prev) => ({ ...prev, form: error.message }))
    } finally {
      setPublishSaving(false)
    }
  }

  const handleRideRequestAction = async (requestID, action) => {
    const token = localStorage.getItem(TOKEN_KEY)
    if (!token) {
      handleLogout()
      return
    }
    setRequestActionLoading(`${requestID}:${action}`)
    setRideFeedNotice('')
    try {
      const res = await respondToRideRequest(token, requestID, {
        action,
        publishedRideId: publishedRides[0]?.id || '',
        message: action === 'accept' ? 'Driver is ready to confirm this route.' : 'Driver wants to discuss pricing and pickup details.',
      })
      setRideFeedNotice(res.message)
    } catch (error) {
      setRideFeedNotice(error.message)
    } finally {
      setRequestActionLoading('')
    }
  }

  const openFeedRequestDetails = (request) => {
    setActiveFeedRequest(request)
  }

  const closeFeedRequestDetails = () => {
    setActiveFeedRequest(null)
    setActiveFeedRoutePreview({ loading: false, error: '', data: null })
  }

  const openRequesterProfile = async (userID) => {
    const token = localStorage.getItem(TOKEN_KEY)
    if (!token || !userID) return

    setRequesterProfile(null)
    setRequesterProfileError('')
    setRequesterProfileLoading(true)
    try {
      const res = await getPublicUserProfile(token, userID)
      setRequesterProfile(res.user || null)
    } catch (error) {
      setRequesterProfileError(error.message)
    } finally {
      setRequesterProfileLoading(false)
    }
  }

  const closeRequesterProfile = () => {
    setRequesterProfile(null)
    setRequesterProfileError('')
    setRequesterProfileLoading(false)
  }

  useEffect(() => {
    if (!activeFeedRequest) {
      setActiveFeedRoutePreview({ loading: false, error: '', data: null })
      return undefined
    }

    const fromLocation = {
      lat: Number(activeFeedRequest.fromLat),
      lon: Number(activeFeedRequest.fromLon),
    }
    const toLocation = {
      lat: Number(activeFeedRequest.toLat),
      lon: Number(activeFeedRequest.toLon),
    }

    if ([fromLocation.lat, fromLocation.lon, toLocation.lat, toLocation.lon].some((value) => Number.isNaN(value))) {
      setActiveFeedRoutePreview({ loading: false, error: 'Route preview unavailable for this request.', data: null })
      return undefined
    }

    const controller = new AbortController()
    setActiveFeedRoutePreview({ loading: true, error: '', data: null })

    fetchRouteData(fromLocation, toLocation, activeFeedRequest.rideType || 'shared', controller.signal)
      .then((data) => setActiveFeedRoutePreview({ loading: false, error: '', data }))
      .catch((error) => {
        if (error.name !== 'AbortError') {
          setActiveFeedRoutePreview({ loading: false, error: error.message || 'Could not load route preview.', data: null })
        }
      })

    return () => controller.abort()
  }, [activeFeedRequest])

  useEffect(() => {
    const field = activeLocationField
    if (!field) return undefined
    const query = requestForm[field]
    if (!query.trim() || query.trim().length < 2) {
      setRequestSuggestions((prev) => ({ ...prev, [field]: [] }))
      setLocationLoading((prev) => ({ ...prev, [field]: false }))
      return undefined
    }

    const controller = new AbortController()
    const timeoutId = window.setTimeout(async () => {
      setLocationLoading((prev) => ({ ...prev, [field]: true }))
      try {
        const suggestions = await fetchLocationSuggestions(query, controller.signal)
        setRequestSuggestions((prev) => ({ ...prev, [field]: suggestions }))
      } catch (error) {
        if (error.name !== 'AbortError') {
          setRequestSuggestions((prev) => ({ ...prev, [field]: [] }))
        }
      } finally {
        setLocationLoading((prev) => ({ ...prev, [field]: false }))
      }
    }, 280)

    return () => {
      controller.abort()
      window.clearTimeout(timeoutId)
    }
  }, [activeLocationField, requestForm])

  useEffect(() => {
    const field = activePublishLocationField
    if (!field) return undefined
    const query = publishForm[field]
    if (!query.trim() || query.trim().length < 2) {
      setPublishSuggestions((prev) => ({ ...prev, [field]: [] }))
      setPublishLocationLoading((prev) => ({ ...prev, [field]: false }))
      return undefined
    }

    const controller = new AbortController()
    const timeoutId = window.setTimeout(async () => {
      setPublishLocationLoading((prev) => ({ ...prev, [field]: true }))
      try {
        const suggestions = await fetchLocationSuggestions(query, controller.signal)
        setPublishSuggestions((prev) => ({ ...prev, [field]: suggestions }))
      } catch (error) {
        if (error.name !== 'AbortError') {
          setPublishSuggestions((prev) => ({ ...prev, [field]: [] }))
        }
      } finally {
        setPublishLocationLoading((prev) => ({ ...prev, [field]: false }))
      }
    }, 280)

    return () => {
      controller.abort()
      window.clearTimeout(timeoutId)
    }
  }, [activePublishLocationField, publishForm])

  useEffect(() => {
    if (!selectedLocations.from || !selectedLocations.to) {
      setRoutePreview((prev) => ({ ...prev, loading: false, error: '', data: null }))
      return undefined
    }

    const controller = new AbortController()
    setRoutePreview((prev) => ({ ...prev, loading: true, error: '', data: null }))

    fetchRouteData(selectedLocations.from, selectedLocations.to, requestForm.rideType, controller.signal)
      .then((data) => {
        setRoutePreview({ loading: false, error: '', data })
      })
      .catch((error) => {
        if (error.name !== 'AbortError') {
          setRoutePreview({ loading: false, error: error.message, data: null })
        }
      })

    return () => controller.abort()
  }, [selectedLocations.from, selectedLocations.to, requestForm.rideType])

  useEffect(() => {
    if (!selectedPublishLocations.from || !selectedPublishLocations.to) {
      setPublishRoutePreview({ loading: false, error: '', data: null })
      return undefined
    }

    const controller = new AbortController()
    setPublishRoutePreview({ loading: true, error: '', data: null })
    fetchRouteData(selectedPublishLocations.from, selectedPublishLocations.to, publishForm.rideType, controller.signal)
      .then((data) => setPublishRoutePreview({ loading: false, error: '', data }))
      .catch((error) => {
        if (error.name !== 'AbortError') {
          setPublishRoutePreview({ loading: false, error: error.message, data: null })
        }
      })

    return () => controller.abort()
  }, [selectedPublishLocations.from, selectedPublishLocations.to, publishForm.rideType])

  if (booting) {
    return <div className="loading-screen">Restoring session...</div>
  }

  if (user) {
    const profileName = profileForm.name || user.name || 'RideX member'
    const hasRating = user.ratingCount > 0
    const cropBaseScale = getCoverScale(avatarDraftNatural.width, avatarDraftNatural.height, AVATAR_CROP_SIZE)
    const cropDisplayWidth = avatarDraftNatural.width * cropBaseScale * avatarDraftZoom
    const cropDisplayHeight = avatarDraftNatural.height * cropBaseScale * avatarDraftZoom

    return (
      <main className="profile-shell">
        <section className="profile-nav">
          <div className="profile-brand">
            <div className="profile-brand-mark">
              <RideMarkIcon />
            </div>
            <h1 className="profile-brand-text">
              <span className="logo-dark">Ride</span>
              <span className="logo-accent">X</span>
            </h1>
          </div>

          <nav className="profile-nav-links" aria-label="Primary navigation">
            <button className={`profile-nav-link ${activeView === 'home' || activeView === 'request' || activeView === 'publish' ? 'is-active' : ''}`} type="button" onClick={() => setActiveView('home')}>
              <NavIcon path="M12 8.5a3.5 3.5 0 1 0 0 7a3.5 3.5 0 0 0 0-7Zm0-5v2.2m0 12.6V20.5m8.5-8.5h-2.2M5.7 12H3.5" />
              <span>Explore</span>
            </button>
            <button className="profile-nav-link" type="button" onClick={() => setActiveView('home')}>
              <NavIcon path="M5 9.5h14M7.5 9.5V8A2.5 2.5 0 0 1 10 5.5h4A2.5 2.5 0 0 1 16.5 8v1.5M6 18.5h12a1.5 1.5 0 0 0 1.5-1.5v-6A1.5 1.5 0 0 0 18 9.5H6A1.5 1.5 0 0 0 4.5 11v6A1.5 1.5 0 0 0 6 18.5Z" />
              <span>My Trips</span>
            </button>
            <button className="profile-nav-link" type="button" onClick={() => setActiveView('home')}>
              <NavIcon path="M6 16.5V7.5A2.5 2.5 0 0 1 8.5 5h7A2.5 2.5 0 0 1 18 7.5v6A2.5 2.5 0 0 1 15.5 16H9l-3 3v-2.5Z" />
              <span>Chats</span>
            </button>
            <button className="profile-nav-link profile-nav-link-notification" type="button" onClick={() => setActiveView('home')}>
              <NavIcon path="M12 5.5a3 3 0 0 0-3 3v1.2c0 .6-.2 1.2-.6 1.7L7.2 13a1 1 0 0 0 .8 1.6h8a1 1 0 0 0 .8-1.6l-1.2-1.6a2.8 2.8 0 0 1-.6-1.7V8.5a3 3 0 0 0-3-3Zm0 13.5a2 2 0 0 1-1.9-1.4h3.8A2 2 0 0 1 12 19Z" />
              <span>Notifications</span>
            </button>
            <button className={`profile-nav-link ${activeView === 'profile' ? 'is-active' : ''}`} type="button" onClick={() => setActiveView('profile')}>
              <NavIcon path="M12 12a3.5 3.5 0 1 0 0-7a3.5 3.5 0 0 0 0 7Zm-6 7a6 6 0 0 1 12 0" />
              <span>Profile</span>
            </button>
          </nav>
        </section>

        {activeView === 'home' && (
          <section className="home-frame">
            <div className="home-hero">
              <h2>Welcome back, {profileName}</h2>
              <p>Where are you headed next?</p>
            </div>

            <div className="home-actions-grid">
              <article className="home-action-card is-featured">
                <div className="home-action-icon">⌕</div>
                <h3>Find a ride</h3>
                <p>Browse available rides near you.</p>
                <span className="home-action-arrow">→</span>
              </article>

              <article className="home-action-card" role="button" tabIndex={0} onClick={() => setActiveView('publish')} onKeyDown={(e) => e.key === 'Enter' && setActiveView('publish')}>
                <div className="home-action-icon">＋</div>
                <h3>Publish a ride</h3>
                <p>Offer seats on your next trip.</p>
                <span className="home-action-arrow">→</span>
              </article>

              <article className="home-action-card" role="button" tabIndex={0} onClick={() => setActiveView('request')} onKeyDown={(e) => e.key === 'Enter' && setActiveView('request')}>
                <div className="home-action-icon">↗</div>
                <h3>Request a ride</h3>
                <p>Post a request and let drivers find you.</p>
                <span className="home-action-arrow">→</span>
              </article>
            </div>

            <div className="home-panels-grid">
              <section className="home-panel">
                <div className="home-panel-header">
                  <h3>Upcoming trip</h3>
                  <button type="button">View all</button>
                </div>
                <div className="home-empty-card">
                  <strong>{rideRequests.length ? `${rideRequests[0].fromLabel} → ${rideRequests[0].toLabel}` : 'No upcoming trips yet'}</strong>
                  <p>
                    {rideRequests.length
                      ? `${rideRequests[0].rideDate} at ${rideRequests[0].rideTime} • ${rideRequests[0].priceEstimate || 'Price estimate pending'}`
                      : 'Your completed trips count will start at 0 and grow here as you finish rides.'}
                  </p>
                  <div className="home-stats-row">
                    <span>{rideRequests.length ? `${rideRequests.length} request${rideRequests.length > 1 ? 's' : ''} posted` : `${user.tripsCompleted} trips completed`}</span>
                    <span>{hasRating ? `${user.rating.toFixed(1)} rating` : 'Not rated yet'}</span>
                  </div>
                  {rideRequests.length ? (
                    <div className="home-card-actions">
                      <button
                        className="profile-ghost-btn home-card-btn"
                        type="button"
                        onClick={() => handleCancelRideRequest(rideRequests[0].id)}
                        disabled={cancelRequestLoading === rideRequests[0].id}
                      >
                        {cancelRequestLoading === rideRequests[0].id ? 'Cancelling...' : 'Cancel request'}
                      </button>
                    </div>
                  ) : null}
                  {homeNotice ? <p className="request-status home-status">{homeNotice}</p> : null}
                </div>
              </section>

              <section className="home-panel">
                <div className="home-panel-header">
                  <h3>Active negotiations</h3>
                  <button type="button">View all</button>
                </div>
                <div className="home-empty-card home-empty-card-negotiations">
                  <strong>No active negotiations</strong>
                  <p>When live pricing conversations start, they will appear here in real time.</p>
                  <div className="home-stats-row">
                    <span>{profileForm.interests.length} saved interests</span>
                    <span>{formatMemberSince(user.createdAt)}</span>
                  </div>
                </div>
              </section>
            </div>
          </section>
        )}

        {activeView === 'publish' && (
          <section className="request-frame publish-frame">
            <div className="request-page-header">
              <div>
                <h2>Publish a Ride</h2>
                <p>Create a ride or respond to live rider requests.</p>
              </div>
              <button className="request-cancel-top" type="button" onClick={() => setActiveView('home')}>Back</button>
            </div>

            <div className="publish-switch" role="tablist" aria-label="Publish ride modes">
              <button type="button" className={`publish-switch-btn ${publishMode === 'create' ? 'is-active' : ''}`} onClick={() => setPublishMode('create')}>
                Create New Ride
              </button>
              <button type="button" className={`publish-switch-btn ${publishMode === 'requests' ? 'is-active' : ''}`} onClick={() => setPublishMode('requests')}>
                View Ride Requests
              </button>
            </div>

            {publishMode === 'create' && (
              <div className="request-grid">
                <form className="request-form-card" onSubmit={handlePublishRideSubmit}>
                  <section className="request-block">
                    <div className="request-block-header">
                      <h3>Route</h3>
                      <span>Driver setup</span>
                    </div>
                    <div className="request-route-grid">
                      <label className="request-field">
                        <span>Pickup</span>
                        <div className={`request-input-wrap ${publishErrors.from ? 'has-error' : ''}`}>
                          <span className="request-input-icon">📍</span>
                          <input
                            type="text"
                            placeholder="Enter pickup location"
                            value={publishForm.from}
                            onFocus={() => setActivePublishLocationField('from')}
                            onBlur={() => window.setTimeout(() => setActivePublishLocationField((current) => (current === 'from' ? '' : current)), 120)}
                            onChange={(e) => handlePublishFieldChange('from', e.target.value)}
                          />
                        </div>
                        {activePublishLocationField === 'from' && (
                          <div className="request-suggestions">
                            {publishLocationLoading.from && <div className="request-suggestion-state">Searching locations...</div>}
                            {publishFromSuggestions.map((location) => (
                              <button key={location.id} type="button" onMouseDown={() => selectPublishLocationSuggestion('from', location)}>
                                {location.label}
                              </button>
                            ))}
                            {!publishLocationLoading.from && !publishFromSuggestions.length && <div className="request-suggestion-state">No matching places found.</div>}
                          </div>
                        )}
                        {publishErrors.from && <small>{publishErrors.from}</small>}
                      </label>

                      <label className="request-field">
                        <span>Drop-off</span>
                        <div className={`request-input-wrap ${publishErrors.to ? 'has-error' : ''}`}>
                          <span className="request-input-icon">📍</span>
                          <input
                            type="text"
                            placeholder="Enter drop-off location"
                            value={publishForm.to}
                            onFocus={() => setActivePublishLocationField('to')}
                            onBlur={() => window.setTimeout(() => setActivePublishLocationField((current) => (current === 'to' ? '' : current)), 120)}
                            onChange={(e) => handlePublishFieldChange('to', e.target.value)}
                          />
                        </div>
                        {activePublishLocationField === 'to' && (
                          <div className="request-suggestions">
                            {publishLocationLoading.to && <div className="request-suggestion-state">Searching locations...</div>}
                            {publishToSuggestions.map((location) => (
                              <button key={location.id} type="button" onMouseDown={() => selectPublishLocationSuggestion('to', location)}>
                                {location.label}
                              </button>
                            ))}
                            {!publishLocationLoading.to && !publishToSuggestions.length && <div className="request-suggestion-state">No matching places found.</div>}
                          </div>
                        )}
                        {publishErrors.to && <small>{publishErrors.to}</small>}
                      </label>
                    </div>
                  </section>

                  <section className="request-block">
                    <div className="request-inline-grid">
                      <label className="request-field">
                        <span>Date</span>
                        <div className={`request-input-wrap ${publishErrors.date ? 'has-error' : ''}`}>
                          <input type="date" value={publishForm.date} onChange={(e) => handlePublishFieldChange('date', e.target.value)} />
                        </div>
                        {publishErrors.date && <small>{publishErrors.date}</small>}
                      </label>
                      <label className="request-field">
                        <span>Time</span>
                        <div className={`request-input-wrap ${publishErrors.time ? 'has-error' : ''}`}>
                          <input type="time" value={publishForm.time} onChange={(e) => handlePublishFieldChange('time', e.target.value)} />
                        </div>
                        {publishErrors.time && <small>{publishErrors.time}</small>}
                      </label>
                    </div>
                    <div className="request-choice-group">
                      <span className="request-choice-label">Departure preference</span>
                      <div className="request-pill-row">
                        {[
                          ['exact', 'Exact departure'],
                          ['15', '±15 minutes'],
                          ['30', '±30 minutes'],
                          ['flexible', 'Flexible'],
                        ].map(([value, label]) => (
                          <button key={value} type="button" className={`request-pill ${publishForm.flexibility === value ? 'is-active' : ''}`} onClick={() => handlePublishFieldChange('flexibility', value)}>
                            {label}
                          </button>
                        ))}
                      </div>
                    </div>
                  </section>

                  <section className="request-block">
                    <div className="request-inline-grid">
                      <div className="request-counter-card">
                        <span className="request-choice-label">Available seats</span>
                        <div className="request-counter-row">
                          <button type="button" onClick={() => handlePublishFieldChange('availableSeats', Math.max(1, Number(publishForm.availableSeats) - 1))}>−</button>
                          <div>
                            <strong>{publishForm.availableSeats}</strong>
                            <span>Seats open</span>
                          </div>
                          <button type="button" onClick={() => handlePublishFieldChange('availableSeats', Math.min(8, Number(publishForm.availableSeats) + 1))}>+</button>
                        </div>
                        {publishErrors.availableSeats && <small>{publishErrors.availableSeats}</small>}
                      </div>
                      <label className="request-field">
                        <span>Total available seats</span>
                        <div className={`request-input-wrap ${publishErrors.totalSeats ? 'has-error' : ''}`}>
                          <input type="number" min="1" max="8" value={publishForm.totalSeats} onChange={(e) => handlePublishFieldChange('totalSeats', e.target.value)} />
                        </div>
                        {publishErrors.totalSeats && <small>{publishErrors.totalSeats}</small>}
                      </label>
                    </div>
                  </section>

                  <section className="request-block request-split-block">
                    <label className="request-field">
                      <span>Price per seat</span>
                      <div className={`request-input-wrap ${publishErrors.pricePerSeat ? 'has-error' : ''}`}>
                        <span className="request-input-icon">$</span>
                        <input type="number" min="1" placeholder="Enter price per seat" value={publishForm.pricePerSeat} onChange={(e) => handlePublishFieldChange('pricePerSeat', e.target.value)} />
                      </div>
                      {publishErrors.pricePerSeat && <small>{publishErrors.pricePerSeat}</small>}
                    </label>
                    <label className="request-field">
                      <span>Vehicle info</span>
                      <div className="request-input-wrap">
                        <input type="text" placeholder="Blue Honda Civic, 2021" value={publishForm.vehicleInfo} onChange={(e) => handlePublishFieldChange('vehicleInfo', e.target.value)} />
                      </div>
                    </label>
                  </section>

                  <section className="request-block request-split-block">
                    <div className="request-choice-group">
                      <span className="request-choice-label">Vehicle type</span>
                      <div className="request-pill-row">
                        {[
                          ['any', 'Any'],
                          ['sedan', 'Sedan'],
                          ['suv', 'SUV'],
                        ].map(([value, label]) => (
                          <button key={value} type="button" className={`request-pill ${publishForm.vehicleType === value ? 'is-active' : ''}`} onClick={() => handlePublishFieldChange('vehicleType', value)}>
                            {label}
                          </button>
                        ))}
                      </div>
                    </div>
                    <div className="request-choice-group">
                      <span className="request-choice-label">Luggage allowed</span>
                      <div className="request-pill-row">
                        {[
                          ['none', '🧳 None'],
                          ['small', '🎒 Small bag'],
                          ['large', '🧳 Large luggage'],
                        ].map(([value, label]) => (
                          <button key={value} type="button" className={`request-pill ${publishForm.luggageAllowed === value ? 'is-active' : ''}`} onClick={() => handlePublishFieldChange('luggageAllowed', value)}>
                            {label}
                          </button>
                        ))}
                      </div>
                    </div>
                  </section>

                  <section className="request-block">
                    <div className="request-choice-group">
                      <span className="request-choice-label">Ride type</span>
                      <div className="request-pill-row">
                        {[
                          ['shared', 'Shared'],
                          ['private', 'Private'],
                        ].map(([value, label]) => (
                          <button key={value} type="button" className={`request-pill ${publishForm.rideType === value ? 'is-active' : ''}`} onClick={() => handlePublishFieldChange('rideType', value)}>
                            {label}
                          </button>
                        ))}
                      </div>
                    </div>
                  </section>

                  <section className="request-block">
                    <label className="request-field">
                      <span>Notes for passengers</span>
                      <textarea
                        placeholder="Pickup at the north entrance. Two small backpacks fit comfortably."
                        value={publishForm.notes}
                        onChange={(e) => handlePublishFieldChange('notes', e.target.value)}
                      />
                    </label>
                  </section>

                  {publishErrors.form && <p className="request-status request-status-error">{publishErrors.form}</p>}
                  {publishNotice && <p className="request-status">{publishNotice}</p>}

                  <div className="request-actions">
                    <button className="profile-ghost-btn" type="button" onClick={() => setActiveView('home')}>Cancel</button>
                    <button className="profile-primary-btn request-submit-btn" type="submit" disabled={publishSaving}>
                      {publishSaving ? 'Publishing...' : 'Publish ride'}
                    </button>
                  </div>
                </form>

                <aside className="request-sidebar">
                  <section className="request-preview-card">
                    <div className="request-preview-top">
                      <span className="request-preview-label">Route preview</span>
                      <strong>{publishRouteSummary ? `$${estimatedEarnings || 0} est.` : 'Estimate pending'}</strong>
                    </div>
                    {publishRoutePreview.loading ? (
                      <>
                        <h3>Loading route preview</h3>
                        <p>Pulling live route data from map services.</p>
                      </>
                    ) : publishRoutePreview.error ? (
                      <>
                        <h3>Route unavailable</h3>
                        <p>{publishRoutePreview.error}</p>
                      </>
                    ) : publishForm.from && publishForm.to ? (
                      <>
                        <h3>{publishForm.from} → {publishForm.to}</h3>
                        <div className="request-metrics">
                          <span>{publishRouteSummary?.miles ?? '—'} miles</span>
                          <span>{publishRouteSummary?.duration ?? '—'}</span>
                          <span>${estimatedEarnings || 0} est. earnings</span>
                        </div>
                      </>
                    ) : (
                      <>
                        <h3>Add your route</h3>
                        <p>Enter pickup and drop-off to preview the ride and estimated earnings.</p>
                      </>
                    )}
                    <div className="request-map-preview">
                      {publishRouteSummary?.coordinates?.length ? (
                        <MapContainer
                          key={`${publishForm.from}-${publishForm.to}`}
                          center={publishRouteSummary.coordinates[Math.floor(publishRouteSummary.coordinates.length / 2)]}
                          zoom={6}
                          scrollWheelZoom={false}
                          className="request-map-live"
                          whenCreated={(map) => map.attributionControl.setPrefix(false)}
                        >
                          <TileLayer
                            attribution='&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors'
                            url="https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png"
                          />
                          <Polyline positions={publishRouteSummary.coordinates} pathOptions={{ color: '#ffb25f', weight: 5 }} />
                          <CircleMarker center={[selectedPublishLocations.from.lat, selectedPublishLocations.from.lon]} radius={8} pathOptions={{ color: '#ffb25f', fillColor: '#ffb25f', fillOpacity: 1 }} />
                          <CircleMarker center={[selectedPublishLocations.to.lat, selectedPublishLocations.to.lon]} radius={8} pathOptions={{ color: '#f08b2d', fillColor: '#f08b2d', fillOpacity: 1 }} />
                        </MapContainer>
                      ) : (
                        <div className="request-map-line"></div>
                      )}
                      <div className="request-map-dot dot-start"></div>
                      <div className="request-map-dot dot-end"></div>
                    </div>
                  </section>

                  <section className="request-sidebar-card">
                    <h3>Earnings preview</h3>
                    <div className="request-suggestion-list">
                      <span>💺 {publishForm.availableSeats || 0} seats × ${publishForm.pricePerSeat || 0} per seat</span>
                      <span>💰 Estimated earnings: ${estimatedEarnings || 0}</span>
                      <span>🚗 {publishForm.vehicleType === 'any' ? 'Flexible vehicle type' : publishForm.vehicleType}</span>
                    </div>
                  </section>
                </aside>
              </div>
            )}

            {publishMode === 'requests' && (
              <div className="publish-feed">
                <div className="publish-feed-filters">
                  <label className="request-field">
                    <span>Date</span>
                    <div className="request-input-wrap">
                      <input type="date" value={rideFeedFilters.date} onChange={(e) => setRideFeedFilters((prev) => ({ ...prev, date: e.target.value }))} />
                    </div>
                  </label>
                  <label className="request-field">
                    <span>Route</span>
                    <div className="request-input-wrap">
                      <input type="text" placeholder="Miami, Orlando, Tampa..." value={rideFeedFilters.route} onChange={(e) => setRideFeedFilters((prev) => ({ ...prev, route: e.target.value }))} />
                    </div>
                  </label>
                  <label className="request-field">
                    <span>Passengers</span>
                    <div className="request-input-wrap">
                      <input type="number" min="1" value={rideFeedFilters.passengers} onChange={(e) => setRideFeedFilters((prev) => ({ ...prev, passengers: e.target.value }))} />
                    </div>
                  </label>
                  <label className="request-field">
                    <span>Budget</span>
                    <div className="request-input-wrap">
                      <span className="request-input-icon">$</span>
                      <input type="number" min="0" value={rideFeedFilters.budget} onChange={(e) => setRideFeedFilters((prev) => ({ ...prev, budget: e.target.value }))} />
                    </div>
                  </label>
                </div>

                {rideFeedNotice && <p className="request-status">{rideFeedNotice}</p>}

                <div className="publish-feed-list">
                  {rideFeedLoading ? (
                    <div className="home-empty-card"><strong>Loading ride requests...</strong><p>Pulling live requests from the backend.</p></div>
                  ) : rideRequestFeed.length ? (
                    rideRequestFeed.map((request) => (
                      <article
                        key={request.id}
                        className="publish-request-card publish-request-card-clickable"
                        role="button"
                        tabIndex={0}
                        onClick={() => openFeedRequestDetails(request)}
                        onKeyDown={(e) => {
                          if (e.key === 'Enter' || e.key === ' ') {
                            e.preventDefault()
                            openFeedRequestDetails(request)
                          }
                        }}
                      >
                        <div className="publish-request-top">
                          <div>
                            <h3>{request.fromLabel} → {request.toLabel}</h3>
                            <p>{request.rideDate} • {request.rideTime} • {request.passengers} passenger{request.passengers > 1 ? 's' : ''}</p>
                          </div>
                          <strong>{request.maxBudget ? `$${request.maxBudget}` : request.priceEstimate || 'Budget open'}</strong>
                        </div>
                        <div className="request-metrics">
                          <button
                            className="requester-link-btn"
                            type="button"
                            onClick={(e) => {
                              e.stopPropagation()
                              openRequesterProfile(request.userId)
                            }}
                          >
                            {request.requesterName}
                          </button>
                          <span>{request.routeMiles || '—'} miles</span>
                          <span>{request.routeDuration || 'Duration pending'}</span>
                        </div>
                        <p className="publish-request-notes">{request.notes || 'No notes from rider.'}</p>
                        {expandedFeedRequestID === request.id && (
                          <div className="publish-request-details">
                            <span>Ride type: {request.rideType}</span>
                            <span>Vehicle pref: {request.vehiclePreference || 'Any'}</span>
                            <span>Luggage: {request.luggage}</span>
                            <span>Budget: {request.maxBudget ? `$${request.maxBudget}` : 'Open'}</span>
                          </div>
                        )}
                        <div className="publish-request-actions">
                          <button className="profile-primary-btn" type="button" onClick={(e) => {
                            e.stopPropagation()
                            handleRideRequestAction(request.id, 'accept')
                          }} disabled={requestActionLoading === `${request.id}:accept`}>
                            {requestActionLoading === `${request.id}:accept` ? 'Accepting...' : 'Accept'}
                          </button>
                          <button className="profile-ghost-btn" type="button" onClick={(e) => {
                            e.stopPropagation()
                            handleRideRequestAction(request.id, 'negotiate')
                          }} disabled={requestActionLoading === `${request.id}:negotiate`}>
                            {requestActionLoading === `${request.id}:negotiate` ? 'Sending...' : 'Negotiate'}
                          </button>
                          <button className="profile-ghost-btn" type="button" onClick={(e) => {
                            e.stopPropagation()
                            openFeedRequestDetails(request)
                          }}>
                            View details
                          </button>
                        </div>
                      </article>
                    ))
                  ) : (
                    <div className="home-empty-card">
                      <strong>No ride requests match these filters</strong>
                      <p>Adjust the route, date, passenger count, or budget filters to see more live requests.</p>
                    </div>
                  )}
                </div>
              </div>
            )}
          </section>
        )}

        {activeView === 'request' && (
          <section className="request-frame">
            <div className="request-page-header">
              <div>
                <h2>Request a Ride</h2>
                <p>Find drivers heading your way.</p>
              </div>
              <button className="request-cancel-top" type="button" onClick={() => setActiveView('home')}>Back</button>
            </div>

            <div className="publish-switch" role="tablist" aria-label="Ride request modes">
              <button type="button" className={`publish-switch-btn ${requestMode === 'create' ? 'is-active' : ''}`} onClick={() => setRequestMode('create')}>
                {editingRequestId ? 'Edit Request' : 'Create New Request'}
              </button>
              <button type="button" className={`publish-switch-btn ${requestMode === 'view' ? 'is-active' : ''}`} onClick={() => setRequestMode('view')}>
                View Your Requests
              </button>
            </div>

            {requestMode === 'create' ? (
            <div className="request-grid">
              <form className="request-form-card" onSubmit={handleRequestSubmit}>
                <section className="request-block">
                  <div className="request-block-header">
                    <h3>Route</h3>
                    <span>Top priority</span>
                  </div>
                  <div className="request-route-grid">
                    <label className="request-field">
                      <span>From</span>
                      <div className={`request-input-wrap ${requestErrors.from ? 'has-error' : ''}`}>
                        <span className="request-input-icon">📍</span>
                        <input
                          type="text"
                          placeholder="Enter pickup location"
                          value={requestForm.from}
                          onFocus={() => setActiveLocationField('from')}
                          onBlur={() => window.setTimeout(() => setActiveLocationField((current) => (current === 'from' ? '' : current)), 120)}
                          onChange={(e) => handleRequestFieldChange('from', e.target.value)}
                        />
                      </div>
                      {activeLocationField === 'from' && (
                        <div className="request-suggestions">
                          {locationLoading.from && <div className="request-suggestion-state">Searching locations...</div>}
                          {fromSuggestions.map((location) => (
                            <button key={location.id} type="button" onMouseDown={() => selectLocationSuggestion('from', location)}>
                              {location.label}
                            </button>
                          ))}
                          {!locationLoading.from && !fromSuggestions.length && <div className="request-suggestion-state">No matching places found.</div>}
                        </div>
                      )}
                      {requestErrors.from && <small>{requestErrors.from}</small>}
                    </label>

                    <label className="request-field">
                      <span>To</span>
                      <div className={`request-input-wrap ${requestErrors.to ? 'has-error' : ''}`}>
                        <span className="request-input-icon">📍</span>
                        <input
                          type="text"
                          placeholder="Enter destination"
                          value={requestForm.to}
                          onFocus={() => setActiveLocationField('to')}
                          onBlur={() => window.setTimeout(() => setActiveLocationField((current) => (current === 'to' ? '' : current)), 120)}
                          onChange={(e) => handleRequestFieldChange('to', e.target.value)}
                        />
                      </div>
                      {activeLocationField === 'to' && (
                        <div className="request-suggestions">
                          {locationLoading.to && <div className="request-suggestion-state">Searching locations...</div>}
                          {toSuggestions.map((location) => (
                            <button key={location.id} type="button" onMouseDown={() => selectLocationSuggestion('to', location)}>
                              {location.label}
                            </button>
                          ))}
                          {!locationLoading.to && !toSuggestions.length && <div className="request-suggestion-state">No matching places found.</div>}
                        </div>
                      )}
                      {requestErrors.to && <small>{requestErrors.to}</small>}
                    </label>
                  </div>
                </section>

                <section className="request-block">
                  <div className="request-inline-grid">
                    <label className="request-field">
                      <span>Date</span>
                      <div className={`request-input-wrap ${requestErrors.date ? 'has-error' : ''}`}>
                        <input type="date" value={requestForm.date} onChange={(e) => handleRequestFieldChange('date', e.target.value)} />
                      </div>
                      {requestErrors.date && <small>{requestErrors.date}</small>}
                    </label>

                    <label className="request-field">
                      <span>Time</span>
                      <div className={`request-input-wrap ${requestErrors.time ? 'has-error' : ''}`}>
                        <input type="time" value={requestForm.time} onChange={(e) => handleRequestFieldChange('time', e.target.value)} />
                      </div>
                      {requestErrors.time && <small>{requestErrors.time}</small>}
                    </label>
                  </div>

                  <div className="request-choice-group">
                    <span className="request-choice-label">Pickup flexibility</span>
                    <div className="request-pill-row">
                      {[
                        ['exact', 'Exact time'],
                        ['15', '±15 minutes'],
                        ['30', '±30 minutes'],
                        ['flexible', 'Flexible'],
                      ].map(([value, label]) => (
                        <button
                          key={value}
                          type="button"
                          className={`request-pill ${requestForm.flexibility === value ? 'is-active' : ''}`}
                          onClick={() => handleRequestFieldChange('flexibility', value)}
                        >
                          {label}
                        </button>
                      ))}
                    </div>
                  </div>
                </section>

                <section className="request-block">
                  <div className="request-inline-grid request-inline-grid-wide">
                    <div className="request-counter-card">
                      <span className="request-choice-label">Passengers</span>
                      <div className="request-counter-row">
                        <button type="button" onClick={() => handleRequestFieldChange('passengers', Math.max(1, requestForm.passengers - 1))}>−</button>
                        <div>
                          <strong>{requestForm.passengers}</strong>
                          <span>👥 {requestForm.passengers === 1 ? 'Passenger' : 'Passengers'}</span>
                        </div>
                        <button type="button" onClick={() => handleRequestFieldChange('passengers', Math.min(6, requestForm.passengers + 1))}>+</button>
                      </div>
                    </div>

                    <div className="request-choice-group">
                      <span className="request-choice-label">Luggage</span>
                      <div className="request-pill-row">
                        {[
                          ['none', 'None'],
                          ['small', 'Small bag'],
                          ['large', 'Large luggage'],
                        ].map(([value, label]) => (
                          <button
                            key={value}
                            type="button"
                            className={`request-pill ${requestForm.luggage === value ? 'is-active' : ''}`}
                            onClick={() => handleRequestFieldChange('luggage', value)}
                          >
                            {value === 'none' ? '🧳 No luggage' : value === 'small' ? '🎒 Small bag' : '🧳 Large luggage'}
                          </button>
                        ))}
                      </div>
                    </div>
                  </div>
                </section>

                <section className="request-block">
                  <label className="request-field">
                    <span>Max budget (optional)</span>
                    <div className={`request-input-wrap ${requestErrors.maxBudget ? 'has-error' : ''}`}>
                      <span className="request-input-icon">$</span>
                      <input
                        type="number"
                        min="0"
                        placeholder="Enter max budget"
                        value={requestForm.maxBudget}
                        onChange={(e) => handleRequestFieldChange('maxBudget', e.target.value)}
                      />
                    </div>
                    <small className="request-hint">Drivers may offer lower prices.</small>
                    {requestErrors.maxBudget && <small>{requestErrors.maxBudget}</small>}
                  </label>
                </section>

                <section className="request-block request-split-block">
                  <div className="request-choice-group">
                    <span className="request-choice-label">Ride type</span>
                    <div className="request-pill-row">
                      {[
                        ['shared', 'Shared ride'],
                        ['private', 'Private ride'],
                      ].map(([value, label]) => (
                        <button
                          key={value}
                          type="button"
                          className={`request-pill ${requestForm.rideType === value ? 'is-active' : ''}`}
                          onClick={() => handleRequestFieldChange('rideType', value)}
                        >
                          {label}
                        </button>
                      ))}
                    </div>
                  </div>

                  <div className="request-choice-group">
                    <span className="request-choice-label">Vehicle preference</span>
                    <div className="request-pill-row">
                      {[
                        ['any', 'Any'],
                        ['sedan', 'Sedan'],
                        ['suv', 'SUV'],
                      ].map(([value, label]) => (
                        <button
                          key={value}
                          type="button"
                          className={`request-pill ${requestForm.vehiclePreference === value ? 'is-active' : ''}`}
                          onClick={() => handleRequestFieldChange('vehiclePreference', value)}
                        >
                          {label}
                        </button>
                      ))}
                    </div>
                  </div>
                </section>

                <section className="request-block request-split-block">
                  <div className="request-choice-group">
                    <span className="request-choice-label">Minimum driver rating</span>
                    <div className="request-pill-row">
                      {[
                        ['', 'Any rating'],
                        ['4.0', '4.0+'],
                        ['4.5', '4.5+'],
                        ['5.0', '5.0'],
                      ].map(([value, label]) => (
                        <button
                          key={label}
                          type="button"
                          className={`request-pill ${requestForm.minimumRating === value ? 'is-active' : ''}`}
                          onClick={() => handleRequestFieldChange('minimumRating', value)}
                        >
                          {value ? `★ ${label}` : label}
                        </button>
                      ))}
                    </div>
                  </div>

                  <label className="request-toggle">
                    <span>Verified drivers only</span>
                    <input
                      type="checkbox"
                      checked={requestForm.verifiedOnly}
                      onChange={(e) => handleRequestFieldChange('verifiedOnly', e.target.checked)}
                    />
                    <strong aria-hidden="true"></strong>
                  </label>
                </section>

                <section className="request-block">
                  <label className="request-field">
                    <span>Notes for driver (optional)</span>
                    <textarea
                      placeholder="Small suitcase, flexible pickup point."
                      value={requestForm.notes}
                      onChange={(e) => handleRequestFieldChange('notes', e.target.value)}
                    />
                    {requestErrors.notes ? <small>{requestErrors.notes}</small> : <small className="request-hint">Write additional details for the driver.</small>}
                  </label>
                </section>

                {requestErrors.form && <p className="request-status request-status-error">{requestErrors.form}</p>}
                {requestNotice && <p className="request-status">{requestNotice}</p>}

                <div className="request-actions">
                  <button
                    className="profile-ghost-btn"
                    type="button"
                    onClick={() => {
                      if (editingRequestId) {
                        resetRequestComposer()
                      } else {
                        setActiveView('home')
                      }
                    }}
                  >
                    {editingRequestId ? 'Discard changes' : 'Cancel'}
                  </button>
                  <button className="profile-primary-btn request-submit-btn" type="submit" disabled={requestSaving}>
                    {requestSaving ? (editingRequestId ? 'Saving...' : 'Posting...') : (editingRequestId ? 'Save changes' : 'Post Ride Request')}
                  </button>
                </div>
              </form>

              <aside className="request-sidebar">
                <section className="request-preview-card">
                  <div className="request-preview-top">
                    <span className="request-preview-label">Route preview</span>
                    <strong>{routeSummary ? routeSummary.estimate : 'Estimate pending'}</strong>
                  </div>
                {routePreview.loading ? (
                    <>
                      <h3>Loading route preview</h3>
                      <p>Pulling live route data from map services.</p>
                    </>
                  ) : routePreview.error ? (
                    <>
                      <h3>Route unavailable</h3>
                      <p>{routePreview.error}</p>
                    </>
                  ) : requestForm.from && requestForm.to ? (
                    <>
                      <h3>{requestForm.from || 'Pickup'} → {requestForm.to || 'Destination'}</h3>
                      <div className="request-metrics">
                        <span>{routeSummary?.miles ?? '—'} miles</span>
                        <span>{routeSummary?.duration ?? '—'}</span>
                        <span>{routeSummary?.estimate ?? 'Estimate pending'}</span>
                      </div>
                    </>
                  ) : (
                    <>
                      <h3>Add your route</h3>
                      <p>Enter pickup and destination to see the trip summary, estimate, and matching hints.</p>
                    </>
                  )}
                  <div className="request-map-preview">
                    {routeSummary?.coordinates?.length ? (
                      <MapContainer
                        key={`${requestForm.from}-${requestForm.to}`}
                        center={routeSummary.coordinates[Math.floor(routeSummary.coordinates.length / 2)]}
                        zoom={6}
                        scrollWheelZoom={false}
                        className="request-map-live"
                        whenCreated={(map) => {
                          map.attributionControl.setPrefix(false)
                        }}
                      >
                        <TileLayer
                          attribution='&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors'
                          url="https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png"
                        />
                        <Polyline positions={routeSummary.coordinates} pathOptions={{ color: '#ffb25f', weight: 5 }} />
                        <CircleMarker center={[selectedLocations.from.lat, selectedLocations.from.lon]} radius={8} pathOptions={{ color: '#ffb25f', fillColor: '#ffb25f', fillOpacity: 1 }} />
                        <CircleMarker center={[selectedLocations.to.lat, selectedLocations.to.lon]} radius={8} pathOptions={{ color: '#f08b2d', fillColor: '#f08b2d', fillOpacity: 1 }} />
                      </MapContainer>
                    ) : (
                      <div className="request-map-line"></div>
                    )}
                    <div className="request-map-dot dot-start"></div>
                    <div className="request-map-dot dot-end"></div>
                  </div>
                </section>

                  <section className="request-sidebar-card">
                    <h3>Smart suggestions</h3>
                    <div className="request-suggestion-list">
                    <span>⏱ Best match window: {requestForm.flexibility === 'exact' ? 'exact departure' : 'wider driver match set'}</span>
                    <span>💸 {requestForm.rideType === 'private' ? 'Private rides trend 18% higher' : 'Shared rides usually price lower'}</span>
                    <span>🛡 {requestForm.verifiedOnly ? 'Only verified drivers will see this request' : 'Turn on verification filter for safer matches'}</span>
                  </div>
                </section>

                <section className="request-sidebar-card">
                  <h3>Passenger details</h3>
                  <div className="request-passenger-summary">
                    <div className="request-passenger-icons" aria-hidden="true">
                      {Array.from({ length: Math.min(requestForm.passengers, 4) }).map((_, index) => (
                        <span key={index}>●</span>
                      ))}
                    </div>
                    <p>👥 {requestForm.passengers} passenger{requestForm.passengers > 1 ? 's' : ''} • {requestForm.luggage === 'none' ? '🧳 No luggage' : requestForm.luggage === 'small' ? '🎒 Small bag' : '🧳 Large luggage'}</p>
                    </div>
                  </section>
                </aside>
              </div>
            ) : (
              <div className="publish-feed request-manage-feed">
                {requestNotice && <p className="request-status">{requestNotice}</p>}
                <div className="publish-feed-list">
                  {rideRequests.length ? (
                    rideRequests.map((request) => (
                      <article key={request.id} className="publish-request-card">
                        <div className="publish-request-top">
                          <div>
                            <h3>{request.fromLabel} → {request.toLabel}</h3>
                            <p>{request.rideDate} • {request.rideTime} • {request.passengers} passenger{request.passengers > 1 ? 's' : ''}</p>
                          </div>
                          <strong>{request.maxBudget ? `$${request.maxBudget}` : request.priceEstimate || 'Budget open'}</strong>
                        </div>
                        <div className="request-metrics">
                          <span>{request.routeMiles || '—'} miles</span>
                          <span>{request.routeDuration || 'Duration pending'}</span>
                          <span>{request.luggage || 'No luggage preference'}</span>
                        </div>
                        <p className="publish-request-notes">{request.notes || 'No notes for drivers.'}</p>
                        <div className="publish-request-details">
                          <span>Ride type: {request.rideType}</span>
                          <span>Vehicle pref: {request.vehiclePreference || 'Any'}</span>
                          <span>Flexibility: {request.flexibility}</span>
                          <span>{request.verifiedDriversOnly ? 'Verified drivers only' : 'All drivers'}</span>
                        </div>
                        <div className="publish-request-actions">
                          <button
                            className="profile-primary-btn"
                            type="button"
                            onClick={() => handleEditRideRequest(request)}
                            disabled={editRequestLoading === request.id}
                          >
                            {editRequestLoading === request.id ? 'Loading...' : 'Edit'}
                          </button>
                          <button
                            className="profile-ghost-btn"
                            type="button"
                            onClick={() => handleCancelRideRequest(request.id)}
                            disabled={cancelRequestLoading === request.id}
                          >
                            {cancelRequestLoading === request.id ? 'Cancelling...' : 'Cancel'}
                          </button>
                        </div>
                      </article>
                    ))
                  ) : (
                    <div className="home-empty-card">
                      <strong>No ride requests yet</strong>
                      <p>Create a request to start receiving offers from drivers heading your way.</p>
                    </div>
                  )}
                </div>
              </div>
            )}
          </section>
        )}

        {activeView === 'profile' && (
          <section className="profile-frame">
          <div className="profile-summary-card">
            <div className="profile-avatar-wrap">
              <div
                className="profile-avatar"
                style={user.avatarData ? { backgroundImage: `url(${user.avatarData})` } : undefined}
                aria-label="Profile avatar"
              >
                {!user.avatarData && <span>{getInitials(profileName)}</span>}
              </div>
              <label className="profile-upload-btn">
                {avatarUploading ? '...' : 'Upload'}
                <input type="file" accept="image/*" onChange={handleAvatarUpload} disabled={avatarUploading} />
              </label>
            </div>

            <div className="profile-summary-copy">
              <h2>{profileName}</h2>
              <div className="profile-rating-row">
                {hasRating ? (
                  <>
                    <span className="profile-stars" aria-hidden="true">★★★★★</span>
                    <strong>{user.rating.toFixed(1)}</strong>
                    <span>({user.ratingCount})</span>
                  </>
                ) : (
                  <span className="profile-rating-empty">Not rated yet</span>
                )}
              </div>
              <div className="profile-meta-row">
                <span>{user.tripsCompleted} trips completed</span>
                <span>{formatMemberSince(user.createdAt)}</span>
              </div>
              <p>{user.email}</p>
            </div>
          </div>

          <div className="profile-form-card">
            <label className="profile-field">
              <span>Name</span>
              <input
                type="text"
                value={profileForm.name}
                onChange={(e) => setProfileForm((prev) => ({ ...prev, name: e.target.value }))}
                disabled={!profileEditing}
              />
            </label>

            <label className="profile-field">
              <span>Interests</span>
              <div className={`profile-tags-input ${!profileEditing ? 'is-disabled' : ''}`}>
                <div className="profile-tags">
                  {profileForm.interests.map((interest) => (
                    <button
                      key={interest}
                      type="button"
                      className="profile-tag"
                      onClick={() => handleRemoveInterest(interest)}
                      disabled={!profileEditing}
                    >
                      <span>{interest}</span>
                      <strong>×</strong>
                    </button>
                  ))}
                </div>
                <input
                  type="text"
                  value={interestInput}
                  placeholder="Type interest and press Enter"
                  onChange={(e) => setInterestInput(e.target.value)}
                  onKeyDown={handleInterestKeyDown}
                  disabled={!profileEditing}
                />
              </div>
            </label>

            {profileNotice && <p className={`profile-status ${profileNotice.toLowerCase().includes('success') || profileNotice.toLowerCase().includes('updated') ? 'is-success' : 'is-error'}`}>{profileNotice}</p>}

            <div className="profile-actions">
              <button
                className="profile-ghost-btn"
                type="button"
                onClick={() => {
                  setProfileEditing(true)
                  setProfileNotice('')
                }}
              >
                Edit profile
              </button>
              <button className="profile-primary-btn" type="button" onClick={handleProfileSave} disabled={profileSaving}>
                {profileSaving ? 'Saving...' : 'Save changes'}
              </button>
              <button className="profile-ghost-btn profile-logout-btn" type="button" onClick={handleLogout}>
                Logout
              </button>
            </div>
          </div>

          {activeOverlay === 'avatarCrop' && (
            <div className="overlay-shell">
              <div className="overlay-card profile-crop-card">
                <button className="overlay-close" type="button" onClick={closeAvatarCrop}>Close</button>
                <h4 className="verify-title">Crop your photo</h4>
                <p className="verify-copy">Drag the image to reposition it and use zoom before saving.</p>
                <div
                  className="avatar-crop-stage"
                  onPointerDown={handleAvatarDragStart}
                  onPointerMove={handleAvatarDragMove}
                  onPointerUp={handleAvatarDragEnd}
                  onPointerCancel={handleAvatarDragEnd}
                >
                  {avatarDraftSrc && (
                    <img
                      src={avatarDraftSrc}
                      alt=""
                      className="avatar-crop-image"
                      onLoad={handleAvatarImageLoad}
                      draggable="false"
                      style={{
                        width: `${cropDisplayWidth}px`,
                        height: `${cropDisplayHeight}px`,
                        left: `calc(50% + ${avatarDraftOffset.x}px)`,
                        top: `calc(50% + ${avatarDraftOffset.y}px)`,
                      }}
                    />
                  )}
                  <div className="avatar-crop-frame" aria-hidden="true"></div>
                </div>
                <label className="crop-slider">
                  <span>Zoom</span>
                  <input type="range" min="1" max="2.4" step="0.01" value={avatarDraftZoom} onChange={handleAvatarZoomChange} />
                </label>
                <div className="verify-actions">
                  <button className="secondary-btn" type="button" onClick={closeAvatarCrop} disabled={avatarUploading}>
                    Cancel
                  </button>
                  <button className="primary-btn" type="button" onClick={handleApplyAvatarCrop} disabled={avatarUploading}>
                    {avatarUploading ? 'Saving...' : 'Apply photo'}
                  </button>
                </div>
              </div>
            </div>
          )}
          </section>
        )}

        {activeFeedRequest && (
          <div className="overlay-shell">
            <div className="overlay-card ride-detail-modal">
              <div className="ride-detail-grid">
                <div className="ride-detail-copy">
                  <div className="ride-detail-topbar">
                    <span className="request-preview-label">Ride request</span>
                    <button className="overlay-close ride-detail-close" type="button" onClick={closeFeedRequestDetails}>Close</button>
                  </div>
                  <div className="ride-detail-header">
                    <div>
                      <h3>{activeFeedRequest.fromLabel} → {activeFeedRequest.toLabel}</h3>
                      <p className="ride-detail-subtitle">
                        {activeFeedRequest.rideDate} • {activeFeedRequest.rideTime} • {activeFeedRequest.passengers} passenger{activeFeedRequest.passengers > 1 ? 's' : ''}
                      </p>
                    </div>
                    <div className="ride-detail-price">
                      <strong>{activeFeedRequest.maxBudget ? `$${activeFeedRequest.maxBudget}` : activeFeedRequest.priceEstimate || 'Budget open'}</strong>
                      <span>rider budget</span>
                    </div>
                  </div>

                  <div className="ride-detail-stats-grid">
                    <div className="ride-detail-stat">
                      <span>Distance</span>
                      <strong>{activeFeedRequest.routeMiles || '—'} miles</strong>
                    </div>
                    <div className="ride-detail-stat">
                      <span>Duration</span>
                      <strong>{activeFeedRequest.routeDuration || 'Pending'}</strong>
                    </div>
                    <div className="ride-detail-stat">
                      <span>Ride type</span>
                      <strong>{activeFeedRequest.rideType}</strong>
                    </div>
                    <div className="ride-detail-stat">
                      <span>Vehicle</span>
                      <strong>{activeFeedRequest.vehiclePreference || 'Any'}</strong>
                    </div>
                  </div>

                  <div className="ride-detail-section">
                    <h4>Rider</h4>
                    <button className="requester-profile-card" type="button" onClick={() => openRequesterProfile(activeFeedRequest.userId)}>
                      <div className="requester-profile-avatar">
                        {activeFeedRequest.requesterName ? getInitials(activeFeedRequest.requesterName) : 'RX'}
                      </div>
                      <div>
                        <strong>{activeFeedRequest.requesterName}</strong>
                        <span>
                          {activeFeedRequest.requesterRatingCount > 0
                            ? `${activeFeedRequest.requesterRating.toFixed(1)} rating • ${activeFeedRequest.requesterRatingCount} reviews`
                            : 'Not rated yet'}
                        </span>
                      </div>
                    </button>
                  </div>

                  <div className="ride-detail-section">
                    <h4>Additional details</h4>
                    <p>{activeFeedRequest.notes || 'No notes from rider.'}</p>
                    <div className="publish-request-details ride-detail-tags">
                      <span>Flexibility: {activeFeedRequest.flexibility}</span>
                      <span>Minimum driver rating: {activeFeedRequest.minimumRating ? `${activeFeedRequest.minimumRating}+` : 'Any'}</span>
                      <span>{activeFeedRequest.verifiedDriversOnly ? 'Verified drivers only' : 'All drivers welcome'}</span>
                      <span>Luggage: {activeFeedRequest.luggage}</span>
                    </div>
                  </div>

                  <div className="publish-request-actions">
                    <button className="profile-primary-btn" type="button" onClick={() => handleRideRequestAction(activeFeedRequest.id, 'accept')} disabled={requestActionLoading === `${activeFeedRequest.id}:accept`}>
                      {requestActionLoading === `${activeFeedRequest.id}:accept` ? 'Accepting...' : 'Accept'}
                    </button>
                    <button className="profile-ghost-btn" type="button" onClick={() => handleRideRequestAction(activeFeedRequest.id, 'negotiate')} disabled={requestActionLoading === `${activeFeedRequest.id}:negotiate`}>
                      {requestActionLoading === `${activeFeedRequest.id}:negotiate` ? 'Sending...' : 'Negotiate'}
                    </button>
                  </div>
                </div>

                <div className="ride-detail-map-panel">
                  <span className="request-preview-label">Map preview</span>
                  <div className="ride-detail-map">
                    {(() => {
                      const fromLat = Number(activeFeedRequest.fromLat)
                      const fromLon = Number(activeFeedRequest.fromLon)
                      const toLat = Number(activeFeedRequest.toLat)
                      const toLon = Number(activeFeedRequest.toLon)
                      const routeCoordinates = activeFeedRoutePreview.data?.coordinates?.length
                        ? activeFeedRoutePreview.data.coordinates
                        : [
                            [fromLat, fromLon],
                            [toLat, toLon],
                          ]

                      return (
                        <>
                          <MapContainer
                            key={`detail-${activeFeedRequest.id}-${routeCoordinates.length}`}
                            bounds={routeCoordinates}
                            boundsOptions={{ padding: [28, 28] }}
                            scrollWheelZoom
                            dragging
                            touchZoom
                            doubleClickZoom
                            className="request-map-live"
                            whenCreated={(map) => map.attributionControl.setPrefix(false)}
                          >
                            <TileLayer
                              attribution='&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors'
                              url="https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png"
                            />
                            <Polyline
                              positions={routeCoordinates}
                              pathOptions={{ color: '#ffb25f', weight: 5, lineCap: 'round' }}
                            />
                            <Marker position={[fromLat, fromLon]} icon={pickupMarkerIcon} />
                            <Marker position={[toLat, toLon]} icon={dropoffMarkerIcon} />
                          </MapContainer>
                          {activeFeedRoutePreview.loading ? (
                            <div className="ride-detail-map-state">Loading real driving route...</div>
                          ) : null}
                          {activeFeedRoutePreview.error ? (
                            <div className="ride-detail-map-state ride-detail-map-state-error">{activeFeedRoutePreview.error}</div>
                          ) : null}
                        </>
                      )
                    })()}
                  </div>
                  <div className="ride-detail-map-meta">
                    <span>Pickup pinned</span>
                    <span>Drop-off pinned</span>
                    <span>Pan and zoom enabled</span>
                  </div>
                </div>
              </div>
            </div>
          </div>
        )}

        {(requesterProfileLoading || requesterProfile || requesterProfileError) && (
          <div className="overlay-shell">
            <div className="overlay-card rider-profile-modal">
              <button className="overlay-close" type="button" onClick={closeRequesterProfile}>Close</button>
              <span className="request-preview-label">Rider profile</span>
              {requesterProfileLoading ? (
                <>
                  <h3>Loading rider profile</h3>
                  <p className="ride-detail-subtitle">Pulling live profile details from the backend.</p>
                </>
              ) : requesterProfileError ? (
                <>
                  <h3>Profile unavailable</h3>
                  <p className="ride-detail-subtitle">{requesterProfileError}</p>
                </>
              ) : requesterProfile ? (
                <>
                  <div className="rider-profile-header">
                    <div className="rider-profile-avatar" style={requesterProfile.avatarData ? { backgroundImage: `url(${requesterProfile.avatarData})` } : undefined}>
                      {!requesterProfile.avatarData && <span>{getInitials(requesterProfile.name || 'RideX')}</span>}
                    </div>
                    <div className="rider-profile-copy">
                      <h3>{requesterProfile.name}</h3>
                      <p>
                        {requesterProfile.ratingCount > 0
                          ? `${requesterProfile.rating.toFixed(1)} rating from ${requesterProfile.ratingCount} reviews`
                          : 'Not rated yet'}
                      </p>
                      <div className="publish-request-details">
                        <span>{requesterProfile.tripsCompleted} trips completed</span>
                        <span>{formatMemberSince(requesterProfile.createdAt)}</span>
                      </div>
                    </div>
                  </div>

                  <div className="ride-detail-section">
                    <h4>Interests</h4>
                    {requesterProfile.interests?.length ? (
                      <div className="publish-request-details">
                        {requesterProfile.interests.map((interest) => (
                          <span key={interest}>{interest}</span>
                        ))}
                      </div>
                    ) : (
                      <p>No saved interests yet.</p>
                    )}
                  </div>
                </>
              ) : null}
            </div>
          </div>
        )}
      </main>
    )
  }

  return (
    <main className="app-shell">
      <section className="auth-frame">
        <div className="auth-left">
          <div className="auth-brand">
            <div className="auth-brand-mark">
              <RideMarkIcon />
            </div>
            <h1 className="logo">
              <span className="logo-dark">Ride</span>
              <span className="logo-accent">X</span>
            </h1>
          </div>
          <h2 className="hero-heading">
            Turn Empty Seats into Earnings.
          </h2>
          <p className="hero-copy">
            Post available seats or book your next trip with live pricing, verified profiles, and fast confirmation.
          </p>
          <ul className="trust-list" aria-label="RideX trust signals">
            <li>Verified identity checks before every match</li>
            <li>Protected payments and secure in-app pricing</li>
            <li>Real-time route pricing with live updates</li>
          </ul>
          <div className="trust-stats" aria-label="RideX platform stats">
            <div className="trust-stat">
              <strong>120K+</strong>
              <span>members</span>
            </div>
            <div className="trust-stat">
              <strong>4.9/5</strong>
              <span>average rating</span>
            </div>
            <div className="trust-stat">
              <strong>18K+</strong>
              <span>verified drivers</span>
            </div>
          </div>
          <div className="hero-visual" aria-hidden="true">
            <DotLottieReact
              src="/animations/man-waiting-car.lottie"
              loop
              autoplay
              className="hero-lottie"
            />
            <div className="hero-flow-chip chip-top">Driver matched</div>
            <div className="hero-flow-chip chip-middle">Route confirmed</div>
            <div className="hero-flow-chip chip-bottom">Live pricing</div>
          </div>
          <blockquote className="testimonial-card">
            “RideX feels safer than a classifieds listing and faster than traditional carpool planning.”
            <span>Priya M., weekly commuter</span>
          </blockquote>
        </div>

        <div className="auth-right">
          <div className="tabs" role="tablist" aria-label="Authentication tabs">
            <button
              className={`tab-btn ${activeTab === 'login' ? 'active' : ''}`}
              role="tab"
              aria-selected={activeTab === 'login'}
              onClick={() => switchTab('login')}
              type="button"
            >
              Login
            </button>
            <button
              className={`tab-btn ${activeTab === 'signup' ? 'active' : ''}`}
              role="tab"
              aria-selected={activeTab === 'signup'}
              onClick={() => switchTab('signup')}
              type="button"
            >
              Sign Up
            </button>
          </div>

          <form className="form-panel" onSubmit={handleSubmit}>
            <div className={`form-content ${activeTab === 'signup' ? 'slide-left' : ''}`}>
              <h3 className="form-title">{activeTab === 'login' ? 'Welcome back' : 'Create your account'}</h3>
              <p className="form-subtitle">{activeTab === 'login' ? 'Login to continue to RideX.' : 'Start with your personal account details.'}</p>

              {banner && <p className="form-banner">{banner}</p>}
              {errors.form && <p className="form-error">{errors.form}</p>}

              {activeTab === 'signup' && (
                <label className="field">
                  Full name
                  <input
                    className={signupForm.name && !errors.name ? 'is-valid' : ''}
                    type="text"
                    placeholder="Jane Rider"
                    value={signupForm.name}
                    onChange={(e) => {
                      const value = e.target.value
                      setSignupForm((p) => ({ ...p, name: value }))
                      validateField('name', value)
                    }}
                    onBlur={(e) => validateField('name', e.target.value)}
                  />
                  {errors.name && <span className="field-error">{errors.name}</span>}
                </label>
              )}

              <label className="field">
                Email
                <input
                  className={currentForm.email && !errors.email ? 'is-valid' : ''}
                  type="email"
                  placeholder="you@example.com"
                  value={currentForm.email}
                  onChange={(e) => {
                    const value = e.target.value
                    if (activeTab === 'login') setLoginForm((p) => ({ ...p, email: value }))
                    else setSignupForm((p) => ({ ...p, email: value }))
                    validateField('email', value)
                  }}
                  onBlur={(e) => validateField('email', e.target.value)}
                />
                {errors.email && <span className="field-error">{errors.email}</span>}
              </label>

              <label className="field">
                Password
                <input
                  className={currentForm.password && !errors.password ? 'is-valid' : ''}
                  type="password"
                  placeholder={activeTab === 'signup' ? 'At least 8 characters with letters and numbers' : 'Enter your password'}
                  value={currentForm.password}
                  onChange={(e) => {
                    const value = e.target.value
                    if (activeTab === 'login') setLoginForm((p) => ({ ...p, password: value }))
                    else setSignupForm((p) => ({ ...p, password: value }))
                    validateField('password', value)
                  }}
                  onBlur={(e) => validateField('password', e.target.value)}
                />
                {errors.password && <span className="field-error">{errors.password}</span>}
              </label>

              {activeTab === 'login' && (
                <button className="text-link text-link-inline" type="button" onClick={() => {
                  setActiveOverlay((prev) => (prev === 'reset' ? '' : 'reset'))
                  setOverlayNotice('')
                  setResetForm((prev) => ({ ...prev, email: loginForm.email || prev.email }))
                }}>
                  {activeOverlay === 'reset' ? 'Hide password reset' : 'Forgot password?'}
                </button>
              )}

              {activeTab === 'signup' && signupForm.password && (
                <div className="password-strength" aria-live="polite">
                  <div className={`strength-track strength-${passwordStrength.score}`}>
                    <span></span>
                  </div>
                  <p className="strength-label">Password strength: {passwordStrength.label}</p>
                </div>
              )}

              <button className="primary-btn" disabled={loading} type="submit">
                {loading ? 'Please wait...' : activeTab === 'login' ? 'Login' : 'Create Free Account'}
              </button>

              <div className="social-auth">
                <div className="social-divider">
                  <span></span>
                  <p className="social-label">or</p>
                  <span></span>
                </div>
                <div className="social-grid">
                  <button className="social-btn" type="button" onClick={() => handleSocialClick('Google')}>
                    <span className="social-btn-inner">
                      <span className="social-icon social-icon-google"><GoogleIcon /></span>
                      <span>Continue with Google</span>
                    </span>
                  </button>
                  <button className="social-btn" type="button" onClick={() => handleSocialClick('GitHub')}>
                    <span className="social-btn-inner">
                      <span className="social-icon social-icon-github"><GitHubIcon /></span>
                      <span>Continue with GitHub</span>
                    </span>
                  </button>
                </div>
              </div>
            </div>

            {activeOverlay === 'verify' && (
              <div className="overlay-shell">
                <div className="overlay-card">
                  <button className="overlay-close" type="button" onClick={() => setActiveOverlay('')}>Close</button>
                  <h4 className="verify-title">Verify your email</h4>
                  <p className="verify-copy">Enter the verification code that was sent to your inbox. You can resend a fresh code if needed.</p>
                  {overlayNotice && <p className="form-banner">{overlayNotice}</p>}
                  <div className="verify-form">
                    <input
                      type="email"
                      placeholder="Email address"
                      value={verifyForm.email}
                      onChange={(e) => setVerifyForm((prev) => ({ ...prev, email: e.target.value }))}
                    />
                    <input
                      type="text"
                      placeholder="6-digit code"
                      value={verifyForm.code}
                      onChange={(e) => setVerifyForm((prev) => ({ ...prev, code: e.target.value }))}
                    />
                    {errors.verify && <p className="form-error">{errors.verify}</p>}
                    <div className="verify-actions">
                      <button className="primary-btn" disabled={verifyLoading} type="button" onClick={handleVerify}>
                        {verifyLoading ? 'Verifying...' : 'Verify Email'}
                      </button>
                      <button className="secondary-btn" disabled={verifyLoading} type="button" onClick={handleResendVerification}>
                        Resend Code
                      </button>
                    </div>
                  </div>
                </div>
              </div>
            )}

            {activeOverlay === 'reset' && (
              <div className="overlay-shell">
                <div className="overlay-card">
                  <button className="overlay-close" type="button" onClick={() => setActiveOverlay('')}>Close</button>
                  <h4 className="verify-title">Reset your password</h4>
                  <p className="verify-copy">Request a reset code by email, then enter the code with your new password here.</p>
                  {overlayNotice && <p className="form-banner">{overlayNotice}</p>}
                  <div className="verify-form">
                    <input
                      type="email"
                      placeholder="Email address"
                      value={resetForm.email}
                      onChange={(e) => setResetForm((prev) => ({ ...prev, email: e.target.value }))}
                    />
                    <div className="verify-actions">
                      <button className="secondary-btn" disabled={verifyLoading} type="button" onClick={handleForgotPassword}>
                        {verifyLoading ? 'Sending...' : 'Send Reset Code'}
                      </button>
                    </div>
                    <input
                      type="text"
                      placeholder="Reset code"
                      value={resetForm.code}
                      onChange={(e) => setResetForm((prev) => ({ ...prev, code: e.target.value }))}
                    />
                    <input
                      type="password"
                      placeholder="New password"
                      value={resetForm.newPassword}
                      onChange={(e) => setResetForm((prev) => ({ ...prev, newPassword: e.target.value }))}
                    />
                    {errors.reset && <p className="form-error">{errors.reset}</p>}
                    <div className="verify-actions">
                      <button className="primary-btn" disabled={verifyLoading} type="button" onClick={handleResetPassword}>
                        {verifyLoading ? 'Resetting...' : 'Reset Password'}
                      </button>
                    </div>
                  </div>
                </div>
              </div>
            )}
          </form>
        </div>
      </section>
    </main>
  )
}

export default App
