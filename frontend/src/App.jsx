import { useEffect, useMemo, useState } from 'react'
import { login, me, signup } from './api/auth'
import { DotLottieReact } from '@lottiefiles/dotlottie-react'

const TOKEN_KEY = 'ridex_token'

const emptyLogin = { email: '', password: '' }
const emptySignup = { name: '', email: '', password: '' }
const emailPattern = /^[^\s@]+@[^\s@]+\.[^\s@]+$/

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

function App() {
  const [activeTab, setActiveTab] = useState('login')
  const [loginForm, setLoginForm] = useState(emptyLogin)
  const [signupForm, setSignupForm] = useState(emptySignup)
  const [errors, setErrors] = useState({})
  const [loading, setLoading] = useState(false)
  const [booting, setBooting] = useState(true)
  const [banner, setBanner] = useState('')
  const [user, setUser] = useState(null)

  useEffect(() => {
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

  const switchTab = (tab) => {
    setErrors({})
    setBanner('')
    setActiveTab(tab)
  }

  const validate = () => {
    const next = {}

    if (activeTab === 'signup') {
      if (!signupForm.name.trim()) next.name = 'Full name is required'
      if (!signupForm.email.trim()) next.email = 'Email is required'
      else if (!emailPattern.test(signupForm.email.trim())) next.email = 'Enter a valid email address'
      if (!signupForm.password) next.password = 'Password is required'
      if (signupForm.password && signupForm.password.length < 6) next.password = 'Password must be at least 6 characters'
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
      else if (activeTab === 'signup' && value.length < 6) message = 'Password must be at least 6 characters'
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
        await signup(signupForm)
        setSignupForm(emptySignup)
        setErrors({})
        setActiveTab('login')
        setBanner('Account created. Login now.')
      } else {
        const res = await login(loginForm)
        localStorage.setItem(TOKEN_KEY, res.token)
        setUser(res.user)
      }
    } catch (err) {
      setErrors((prev) => ({ ...prev, form: err.message }))
    } finally {
      setLoading(false)
    }
  }

  const handleLogout = () => {
    localStorage.removeItem(TOKEN_KEY)
    setUser(null)
    setLoginForm(emptyLogin)
    setErrors({})
    setBanner('')
    setActiveTab('login')
  }

  if (booting) {
    return <div className="loading-screen">Restoring session...</div>
  }

  if (user) {
    return (
      <main className="app-shell">
        <section className="session-card">
          <h1 className="logo">
            <span className="logo-dark">Ride</span>
            <span className="logo-accent">X</span>
          </h1>
          <p className="session-title">You are logged in</p>
          <p className="session-subtitle">{user.name} · {user.email}</p>
          <button className="primary-btn" onClick={handleLogout}>Logout</button>
        </section>
      </main>
    )
  }

  return (
    <main className="app-shell">
      <section className="auth-frame">
        <div className="auth-left">
          <h1 className="logo">
            <span className="logo-dark">Ride</span>
            <span className="logo-accent">X</span>
          </h1>
          <h2 className="hero-heading">
            Turn Empty Seats into Earnings.
          </h2>
          <p className="hero-copy">
            Post available seats or book your next trip with live pricing, verified profiles, and fast confirmation.
          </p>
          <ul className="trust-list" aria-label="RideX trust signals">
            <li>Verified rider profiles</li>
            <li>Secure live price negotiation</li>
          </ul>
          <div className="hero-visual" aria-hidden="true">
            <DotLottieReact
              src="/animations/man-waiting-car.lottie"
              loop
              autoplay
              className="hero-lottie"
            />
          </div>
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
                  placeholder={activeTab === 'signup' ? 'At least 6 characters' : 'Enter your password'}
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
            </div>
          </form>
        </div>
      </section>
    </main>
  )
}

export default App
