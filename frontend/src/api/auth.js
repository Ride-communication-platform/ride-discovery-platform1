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

export function me(token) {
  return request('/api/auth/me', {
    method: 'GET',
    headers: {
      Authorization: `Bearer ${token}`,
    },
  })
}

export { API_BASE_URL }
