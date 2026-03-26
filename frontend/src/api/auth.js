const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080'

async function request(path, options = {}) {
  const response = await fetch(`${API_BASE_URL}${path}`, {
    headers: {
      'Content-Type': 'application/json',
      ...(options.headers || {}),
    },
    ...options,
  })

  const isJSON = response.headers.get('content-type')?.includes('application/json')
  const payload = isJSON ? await response.json() : {}

  if (!response.ok) {
    throw new Error(payload.error || 'Something went wrong')
  }

  return payload
}

export function signup(input) {
  return request('/api/auth/signup', {
    method: 'POST',
    body: JSON.stringify(input),
  })
}

export function login(input) {
  return request('/api/auth/login', {
    method: 'POST',
    body: JSON.stringify(input),
  })
}

export function verifyEmail(input) {
  return request('/api/auth/verify-email', {
    method: 'POST',
    body: JSON.stringify(input),
  })
}

export function resendVerification(input) {
  return request('/api/auth/resend-verification', {
    method: 'POST',
    body: JSON.stringify(input),
  })
}

export function forgotPassword(input) {
  return request('/api/auth/forgot-password', {
    method: 'POST',
    body: JSON.stringify(input),
  })
}

export function resetPassword(input) {
  return request('/api/auth/reset-password', {
    method: 'POST',
    body: JSON.stringify(input),
  })
}

export function me(token) {
  return request('/api/auth/me', {
    method: 'GET',
    headers: {
      Authorization: `Bearer ${token}`,
    },
  })
}

export function updateProfile(token, input) {
  return request('/api/auth/me', {
    method: 'PUT',
    headers: {
      Authorization: `Bearer ${token}`,
    },
    body: JSON.stringify(input),
  })
}

export function createRideRequest(token, input) {
  return request('/api/ride-requests', {
    method: 'POST',
    headers: {
      Authorization: `Bearer ${token}`,
    },
    body: JSON.stringify(input),
  })
}

export function listRideRequests(token) {
  return request('/api/ride-requests', {
    method: 'GET',
    headers: {
      Authorization: `Bearer ${token}`,
    },
  })
}

export function deleteRideRequest(token, requestID) {
  return request(`/api/ride-requests/${requestID}`, {
    method: 'DELETE',
    headers: {
      Authorization: `Bearer ${token}`,
    },
  })
}

export function updateRideRequest(token, requestID, input) {
  return request(`/api/ride-requests/${requestID}`, {
    method: 'PUT',
    headers: {
      Authorization: `Bearer ${token}`,
    },
    body: JSON.stringify(input),
  })
}

export function getPublicUserProfile(token, userID) {
  return request(`/api/users/${userID}/profile`, {
    method: 'GET',
    headers: {
      Authorization: `Bearer ${token}`,
    },
  })
}

export function createPublishedRide(token, input) {
  return request('/api/published-rides', {
    method: 'POST',
    headers: {
      Authorization: `Bearer ${token}`,
    },
    body: JSON.stringify(input),
  })
}

export function listPublishedRides(token) {
  return request('/api/published-rides', {
    method: 'GET',
    headers: {
      Authorization: `Bearer ${token}`,
    },
  })
}

export function listRideRequestFeed(token, filters = {}) {
  const params = new URLSearchParams()
  Object.entries(filters).forEach(([key, value]) => {
    if (value !== undefined && value !== null && value !== '') params.set(key, String(value))
  })
  const suffix = params.toString() ? `?${params.toString()}` : ''
  return request(`/api/ride-requests/feed${suffix}`, {
    method: 'GET',
    headers: {
      Authorization: `Bearer ${token}`,
    },
  })
}

export function respondToRideRequest(token, requestID, input) {
  return request(`/api/ride-requests/${requestID}/respond`, {
    method: 'POST',
    headers: {
      Authorization: `Bearer ${token}`,
    },
    body: JSON.stringify(input),
  })
}

export { API_BASE_URL }
