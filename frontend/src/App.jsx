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

 

export default App
