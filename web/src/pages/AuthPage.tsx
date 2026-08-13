import { useState } from 'react'
import { Button } from '@/components/ui/button'
import api from '@/lib/api'
import type { AuthMode, AuthNotice, LoginResult, UserAccount } from '@/types'
import { apiErrorMessage, DEFAULT_ADMIN_EMAIL } from '@/types'

interface AuthPageProps {
  onAuthenticated: (user: UserAccount) => void
}

const copy: Record<AuthMode, { title: string; description: string }> = {
  login: { title: 'Welcome Back', description: 'Enter your credentials to access the portal.' },
  register: { title: 'Create Account', description: 'Create an account and verify your email to access the portal.' },
  verify: { title: 'Verify Email', description: 'Enter the verification token associated with your account.' },
}

function noticeForLogin(result: LoginResult): UserAccount {
  return {
    ID: 0,
    email: result.email,
    username: result.username,
    isVerified: true,
    isAdmin: Boolean(result.isAdmin),
  }
}

export function AuthPage({ onAuthenticated }: AuthPageProps) {
  const [mode, setMode] = useState<AuthMode>('login')
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [username, setUsername] = useState('')
  const [token, setToken] = useState('')
  const [notice, setNotice] = useState<AuthNotice | null>(null)
  const [isSubmitting, setIsSubmitting] = useState(false)

  const showError = (error: unknown) => setNotice({ text: apiErrorMessage(error) })

  const submitLogin = async (event: React.FormEvent) => {
    event.preventDefault()
    setIsSubmitting(true)
    setNotice(null)

    try {
      const result = await api.login(email, password)
      if (!result.success) throw new Error(result.message || '登录失败')
      onAuthenticated(noticeForLogin(result))
    } catch (error) {
      showError(error)
    } finally {
      setIsSubmitting(false)
    }
  }

  const submitRegister = async (event: React.FormEvent) => {
    event.preventDefault()
    setIsSubmitting(true)
    setNotice(null)

    try {
      const result = await api.register(email, password, username)
      if (!result.success) throw new Error(result.message || '注册失败')
      setNotice({ text: '注册成功！请继续完成邮箱验证。', developmentToken: result.verifyToken })
      setMode('verify')
    } catch (error) {
      showError(error)
    } finally {
      setIsSubmitting(false)
    }
  }

  const submitVerification = async (event: React.FormEvent) => {
    event.preventDefault()
    setIsSubmitting(true)
    setNotice(null)

    try {
      const result = await api.verifyEmail(email, token)
      if (!result.success) throw new Error(result.message || '验证失败')
      setNotice({ text: '邮箱验证成功，请使用账号登录。' })
      setMode('login')
    } catch (error) {
      showError(error)
    } finally {
      setIsSubmitting(false)
    }
  }

  const action = mode === 'login' ? submitLogin : mode === 'register' ? submitRegister : submitVerification

  return (
    <div className="min-h-screen bg-slate-950 text-slate-50 flex items-center justify-center p-4">
      <div className="w-full max-w-[400px] space-y-8">
        <div className="text-center">
          <h1 className="text-4xl font-black tracking-tighter text-indigo-500">MaiGoDX</h1>
          <p className="text-slate-400 mt-2">Next-Gen Arcade Game Server Portal</p>
        </div>

        <section className="bg-slate-900 border border-slate-800 shadow-2xl rounded-xl p-6">
          <h2 className="text-2xl font-bold text-white">{copy[mode].title}</h2>
          <p className="mt-2 text-sm text-slate-400">{copy[mode].description}</p>

          {notice && (
            <div className="mt-6 p-3 bg-indigo-500/10 border border-indigo-500/20 rounded-lg text-sm text-indigo-300 text-center">
              <p>{notice.text}</p>
              {notice.developmentToken && (
                <p className="mt-1 font-mono text-[10px] text-indigo-200/70 break-all">
                  Development token: {notice.developmentToken}
                </p>
              )}
            </div>
          )}

          <form onSubmit={action} className="space-y-4 mt-6">
            {mode === 'register' && (
              <label className="block space-y-2">
                <span className="text-sm font-medium text-slate-300">Username</span>
                <input
                  required
                  value={username}
                  onChange={(event) => setUsername(event.target.value)}
                  className="w-full h-10 px-3 bg-slate-800 border border-slate-700 rounded-md text-white focus:ring-2 focus:ring-indigo-500 outline-none"
                />
              </label>
            )}

            {mode !== 'verify' || !email ? (
              <label className="block space-y-2">
                <span className="text-sm font-medium text-slate-300">Email</span>
                <input
                  type="email"
                  required
                  value={email}
                  onChange={(event) => setEmail(event.target.value)}
                  placeholder={DEFAULT_ADMIN_EMAIL}
                  className="w-full h-10 px-3 bg-slate-800 border border-slate-700 rounded-md text-white focus:ring-2 focus:ring-indigo-500 outline-none"
                />
              </label>
            ) : null}

            {mode !== 'verify' && (
              <label className="block space-y-2">
                <span className="text-sm font-medium text-slate-300">Password</span>
                <input
                  type="password"
                  required
                  value={password}
                  onChange={(event) => setPassword(event.target.value)}
                  placeholder="••••••••"
                  className="w-full h-10 px-3 bg-slate-800 border border-slate-700 rounded-md text-white focus:ring-2 focus:ring-indigo-500 outline-none"
                />
              </label>
            )}

            {mode === 'verify' && (
              <label className="block space-y-2">
                <span className="text-sm font-medium text-slate-300">Verification token</span>
                <input
                  required
                  value={token}
                  onChange={(event) => setToken(event.target.value)}
                  placeholder="Enter verification token"
                  className="w-full h-10 px-3 bg-slate-800 border border-slate-700 rounded-md text-white focus:ring-2 focus:ring-indigo-500 outline-none"
                />
              </label>
            )}

            <Button isDisabled={isSubmitting} type="submit" className="w-full bg-indigo-600 hover:bg-indigo-500 text-white font-bold h-11">
              {isSubmitting ? 'Please wait…' : mode === 'login' ? 'Sign In' : mode === 'register' ? 'Sign Up' : 'Verify Email'}
            </Button>
          </form>

          <div className="mt-5 text-center text-xs text-slate-500">
            {mode === 'login' && (
              <>New here? <button type="button" onClick={() => setMode('register')} className="text-indigo-400 hover:underline">Create an account</button></>
            )}
            {mode === 'register' && (
              <>Already have an account? <button type="button" onClick={() => setMode('login')} className="text-indigo-400 hover:underline">Sign in</button></>
            )}
            {mode === 'verify' && (
              <button type="button" onClick={() => setMode('login')} className="text-indigo-400 hover:underline">Back to sign in</button>
            )}
          </div>
        </section>
      </div>
    </div>
  )
}
