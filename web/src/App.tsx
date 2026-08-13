import { useState, useEffect } from 'react'
import { 
  LayoutDashboard, 
  Users, 
  Settings, 
  LogOut, 
  Home, 
  Gamepad2, 
  TrendingUp,
  Menu,
  X,
  Award,
  BookOpen,
  Server,
  CreditCard,
  Sliders
} from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Avatar, AvatarFallback } from '@/components/ui/avatar'
import { Badge } from '@/components/ui/badge'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Separator } from '@/components/ui/separator'
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer } from 'recharts'

// --- Types ---
interface UserAccount {
  ID: number
  email: string
  username: string
  isVerified: boolean
  isAdmin: boolean
}

interface UserCard {
  ID: number
  accessCode: string
  cardName: string
}

interface SystemConfig {
  ID: number
  key: string
  value: string
  desc: string
}

interface Playlog {
  ID: number
  musicId: number
  level: number
  achievement: number
  score: number
  createDate: string
}

interface UserDetail {
  UserID: number
  UserName: string
  Rating: number
  MaxRating: number
  TotalPoint: number
}

interface SongComp {
  title: string
  level: string
  score: number
  rating: number
}

interface Stats {
  totalUsers: number
  totalPlays: number
  recentPlays: Playlog[]
  user: UserDetail
  ratingComposition: {
    bests: SongComp[]
    newBests: SongComp[]
  }
}

// --- Components ---

export default function App() {
  const [isLoggedIn, setIsLoggedIn] = useState(false)
  const [user, setUser] = useState<UserAccount | null>(null)
  const [tab, setTab] = useState<'home' | 'dashboard' | 'maimai' | 'setup' | 'admin' | 'settings'>('home')
  const [authView, setAuthView] = useState<'login' | 'register' | 'verify'>('login')
  const [sidebarOpen, setSidebarOpen] = useState(true)
  
  // Form States
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [username, setUsername] = useState('')
  const [token, setToken] = useState('')
  const [message, setMessage] = useState('')
  const [devToken, setDevToken] = useState('')

  // Card & Config States
  const [accessCode, setAccessCode] = useState('')
  const [cardName, setCardName] = useState('')
  const [userCards, setUserCards] = useState<UserCard[]>([])
  const [configs, setConfigs] = useState<SystemConfig[]>([])
  
  // Data States
  const [stats, setStats] = useState<Stats | null>(null)
  const [userList, setUserList] = useState<UserAccount[]>([])

  useEffect(() => {
    if (isLoggedIn && user) {
      fetchStats()
      fetchCards()
      if (user?.isAdmin) {
        fetchUsers()
        fetchConfigs()
      }
    }
  }, [isLoggedIn, user])

  const fetchStats = async () => {
    try {
      const res = await fetch('/api/stats')
      const data = await res.json()
      if (data.success) setStats(data)
    } catch (e) { console.error(e) }
  }

  const fetchUsers = async () => {
    try {
      const res = await fetch('/api/admin/users')
      const data = await res.json()
      if (data.success) setUserList(data.users || [])
    } catch (e) { console.error(e) }
  }

  const fetchCards = async () => {
    if (!user) return
    try {
      const res = await fetch(`/api/card/list?email=${user.email}`)
      const data = await res.json()
      if (data.success) setUserCards(data.cards || [])
    } catch (e) { console.error(e) }
  }

  const fetchConfigs = async () => {
    try {
      const res = await fetch('/api/admin/config/get')
      const data = await res.json()
      if (data.success) setConfigs(data.configs || [])
    } catch (e) { console.error(e) }
  }

  const handleBindCard = async (e: React.FormEvent) => {
    e.preventDefault()
    setMessage('')
    try {
      const res = await fetch('/api/card/bind', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ email: user?.email, accessCode, cardName }),
      })
      const data = await res.json()
      if (data.success) {
        setMessage('卡片绑定成功！')
        setAccessCode('')
        setCardName('')
        fetchCards()
      } else {
        setMessage(data.message || '卡片绑定失败')
      }
    } catch { setMessage('网络错误') }
  }

  const handleUpdateConfig = async (key: string, value: string) => {
    try {
      const res = await fetch('/api/admin/config/update', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ key, value }),
      })
      const data = await res.json()
      if (data.success) {
        fetchConfigs()
      }
    } catch { console.error('Failed to update config') }
  }

  const handleLogin = async (e: React.FormEvent) => {
    e.preventDefault()
    setMessage('')
    try {
      const res = await fetch('/api/auth/login', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ email, password }),
      })
      const data = await res.json()
      if (data.success) {
        setIsLoggedIn(true)
        const loggedUser: UserAccount = {
          ID: 0,
          email: data.email,
          username: data.username,
          isVerified: true,
          isAdmin: data.email === 'admin@maigodx.local'
        }
        setUser(loggedUser)
        setTab('dashboard')
      } else {
        setMessage(data.message || '登录失败')
      }
    } catch { setMessage('网络错误') }
  }

  const handleRegister = async (e: React.FormEvent) => {
    e.preventDefault()
    setMessage('')
    try {
      const res = await fetch('/api/auth/register', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ email, password, username }),
      })
      const data = await res.json()
      if (data.success) {
        setMessage('注册成功！请查收验证邮件。')
        if (data.verifyToken) setDevToken(data.verifyToken)
        setAuthView('verify')
      } else { setMessage(data.message || '注册失败') }
    } catch { setMessage('网络错误') }
  }

  const handleVerify = async (e: React.FormEvent) => {
    e.preventDefault()
    setMessage('')
    try {
      const res = await fetch('/api/auth/verify', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ email, token }),
      })
      const data = await res.json()
      if (data.success) {
        setMessage('验证成功！请登录。')
        setAuthView('login')
      } else { setMessage(data.message || '验证失败') }
    } catch { setMessage('网络错误') }
  }

  if (!isLoggedIn) {
    return (
      <div className="min-h-screen bg-slate-950 text-slate-50 flex items-center justify-center p-4">
        <div className="w-full max-w-[400px] space-y-8">
          <div className="text-center">
            <h1 className="text-4xl font-black tracking-tighter text-indigo-500">MaiGoDX</h1>
            <p className="text-slate-400 mt-2">Next-Gen Arcade Game Server Portal</p>
          </div>

          <Card className="bg-slate-900 border-slate-800 shadow-2xl">
            <CardHeader>
              <CardTitle className="text-2xl text-white">
                {authView === 'login' ? 'Welcome Back' : authView === 'register' ? 'Create Account' : 'Verify Email'}
              </CardTitle>
              <CardDescription className="text-slate-400">
                {authView === 'login' ? 'Enter your credentials to access the portal' : 'Join the community today'}
              </CardDescription>
            </CardHeader>
            <CardContent>
              {message && (
                <div className="mb-6 p-3 bg-indigo-500/10 border border-indigo-500/20 rounded-lg text-sm text-indigo-400 text-center">
                  {message}
                  {devToken && <div className="mt-1 font-mono text-[10px] opacity-50">Token: {devToken}</div>}
                </div>
              )}

              {authView === 'login' && (
                <form onSubmit={handleLogin} className="space-y-4">
                  <div className="space-y-2">
                    <label className="text-sm font-medium text-slate-300">Email</label>
                    <input type="email" required value={email} onChange={e => setEmail(e.target.value)} className="w-full h-10 px-3 bg-slate-800 border-slate-700 rounded-md text-white focus:ring-2 focus:ring-indigo-500 outline-none transition-all" placeholder="admin@maigodx.local" />
                  </div>
                  <div className="space-y-2">
                    <label className="text-sm font-medium text-slate-300">Password</label>
                    <input type="password" required value={password} onChange={e => setPassword(e.target.value)} className="w-full h-10 px-3 bg-slate-800 border-slate-700 rounded-md text-white focus:ring-2 focus:ring-indigo-500 outline-none transition-all" placeholder="••••••••" />
                  </div>
                  <Button type="submit" className="w-full bg-indigo-600 hover:bg-indigo-500 text-white font-bold h-11">Sign In</Button>
                  <p className="text-center text-xs text-slate-500">
                    New here? <button type="button" onClick={() => setAuthView('register')} className="text-indigo-400 hover:underline">Create an account</button>
                  </p>
                </form>
              )}

              {authView === 'register' && (
                <form onSubmit={handleRegister} className="space-y-4">
                  <div className="space-y-2">
                    <label className="text-sm font-medium text-slate-300">Username</label>
                    <input type="text" required value={username} onChange={e => setUsername(e.target.value)} className="w-full h-10 px-3 bg-slate-800 border-slate-700 rounded-md text-white focus:ring-2 focus:ring-indigo-500 outline-none transition-all" />
                  </div>
                  <div className="space-y-2">
                    <label className="text-sm font-medium text-slate-300">Email</label>
                    <input type="email" required value={email} onChange={e => setEmail(e.target.value)} className="w-full h-10 px-3 bg-slate-800 border-slate-700 rounded-md text-white focus:ring-2 focus:ring-indigo-500 outline-none transition-all" />
                  </div>
                  <div className="space-y-2">
                    <label className="text-sm font-medium text-slate-300">Password</label>
                    <input type="password" required value={password} onChange={e => setPassword(e.target.value)} className="w-full h-10 px-3 bg-slate-800 border-slate-700 rounded-md text-white focus:ring-2 focus:ring-indigo-500 outline-none transition-all" />
                  </div>
                  <Button type="submit" className="w-full bg-indigo-600 hover:bg-indigo-500 text-white font-bold h-11">Sign Up</Button>
                  <p className="text-center text-xs text-slate-500">
                    Already have an account? <button type="button" onClick={() => setAuthView('login')} className="text-indigo-400 hover:underline">Sign In</button>
                  </p>
                </form>
              )}

              {authView === 'verify' && (
                <form onSubmit={handleVerify} className="space-y-4">
                  <div className="space-y-2">
                    <label className="text-sm font-medium text-slate-300">Token</label>
                    <input type="text" required value={token} onChange={e => setToken(e.target.value)} className="w-full h-10 px-3 bg-slate-800 border-slate-700 rounded-md text-white focus:ring-2 focus:ring-indigo-500 outline-none transition-all" placeholder="Enter verification token" />
                  </div>
                  <Button type="submit" className="w-full bg-indigo-600 hover:bg-indigo-500 text-white font-bold h-11">Verify Email</Button>
                </form>
              )}
            </CardContent>
          </Card>
        </div>
      </div>
    )
  }

  const NavItem = ({ id, icon: Icon, label }: { id: typeof tab, icon: any, label: string }) => (
    <button
      onClick={() => setTab(id)}
      className={`flex items-center gap-3 px-4 py-3 rounded-lg transition-all ${
        tab === id ? 'bg-indigo-600 text-white shadow-lg shadow-indigo-500/20' : 'text-slate-400 hover:bg-slate-800 hover:text-white'
      }`}
    >
      <Icon size={20} />
      <span className={`font-medium ${!sidebarOpen && 'hidden'}`}>{label}</span>
    </button>
  )

  return (
    <div className="min-h-screen bg-slate-950 text-slate-50 flex overflow-hidden">
      {/* Sidebar */}
      <aside className={`${sidebarOpen ? 'w-64' : 'w-20'} bg-slate-900 border-r border-slate-800 flex flex-col transition-all duration-300 z-50`}>
        <div className="p-6 flex items-center justify-between">
          <div className={`font-black text-2xl text-indigo-500 tracking-tighter ${!sidebarOpen && 'hidden'}`}>MaiGoDX</div>
          <button onClick={() => setSidebarOpen(!sidebarOpen)} className="text-slate-400 hover:text-white">
            {sidebarOpen ? <X size={20} /> : <Menu size={20} />}
          </button>
        </div>
        
        <nav className="flex-1 px-3 space-y-2">
          <NavItem id="home" icon={Home} label="Home" />
          <NavItem id="dashboard" icon={LayoutDashboard} label="Dashboard" />
          <NavItem id="maimai" icon={Gamepad2} label="maimai DX" />
          <NavItem id="setup" icon={Server} label="Setup Guide" />
          {user?.isAdmin && <NavItem id="admin" icon={Users} label="Admin Panel" />}
          <NavItem id="settings" icon={Settings} label="Settings" />
        </nav>

        <div className="p-4 border-t border-slate-800">
          <button onClick={() => setIsLoggedIn(false)} className="flex items-center gap-3 px-4 py-3 w-full text-rose-400 hover:bg-rose-500/10 rounded-lg transition-all">
            <LogOut size={20} />
            <span className={`font-medium ${!sidebarOpen && 'hidden'}`}>Logout</span>
          </button>
        </div>
      </aside>

      {/* Main Content */}
      <main className="flex-1 overflow-y-auto relative">
        <header className="h-16 border-b border-slate-800 bg-slate-950/50 backdrop-blur-xl flex items-center justify-between px-8 sticky top-0 z-40">
          <h2 className="text-lg font-bold capitalize">{tab}</h2>
          <div className="flex items-center gap-4">
            <div className="text-right hidden sm:block">
              <div className="text-sm font-bold">{user?.username}</div>
              <div className="text-[10px] text-slate-500">{user?.email}</div>
            </div>
            <Avatar className="h-9 w-9 border border-slate-700">
              <AvatarFallback className="bg-indigo-600 text-white font-bold">{user?.username ? user.username[0].toUpperCase() : 'U'}</AvatarFallback>
            </Avatar>
          </div>
        </header>

        <div className="p-8">
          {tab === 'home' && (
            <div className="max-w-4xl space-y-8">
              <div className="space-y-2">
                <h1 className="text-4xl font-black tracking-tight">Welcome to MaiGoDX Portal</h1>
                <p className="text-slate-400 text-lg">A modern, high-performance arcade game server management system inspired by AquaDX.</p>
              </div>
              <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                <Card className="bg-slate-900 border-slate-800">
                  <CardHeader><CardTitle className="text-white">Supported Games</CardTitle></CardHeader>
                  <CardContent className="space-y-4">
                    <div className="flex items-center justify-between p-3 bg-slate-800 rounded-lg">
                      <span className="font-bold text-white">maimai DX</span>
                      <Badge className="bg-emerald-500">Active</Badge>
                    </div>
                    <div className="flex items-center justify-between p-3 bg-slate-800 rounded-lg opacity-50">
                      <span className="font-bold text-white">CHUNITHM</span>
                      <Badge variant="outline" className="text-slate-400 border-slate-700">Soon</Badge>
                    </div>
                  </CardContent>
                </Card>
                <Card className="bg-slate-900 border-slate-800">
                  <CardHeader><CardTitle className="text-white">Quick Stats</CardTitle></CardHeader>
                  <CardContent className="grid grid-cols-2 gap-4 text-center">
                    <div className="p-4 bg-indigo-500/10 rounded-xl border border-indigo-500/20">
                      <div className="text-2xl font-black text-indigo-400">{stats?.totalUsers || 0}</div>
                      <div className="text-[10px] uppercase tracking-wider text-slate-500">Users</div>
                    </div>
                    <div className="p-4 bg-emerald-500/10 rounded-xl border border-emerald-500/20">
                      <div className="text-2xl font-black text-emerald-400">{stats?.totalPlays || 0}</div>
                      <div className="text-[10px] uppercase tracking-wider text-slate-500">Total Plays</div>
                    </div>
                  </CardContent>
                </Card>
              </div>
            </div>
          )}

          {tab === 'dashboard' && (
            <div className="space-y-8">
              <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
                <Card className="bg-slate-900 border-slate-800">
                  <CardHeader className="pb-2"><CardDescription>Global Rating</CardDescription><CardTitle className="text-3xl text-white">{stats?.user?.Rating || 15420}</CardTitle></CardHeader>
                  <CardContent><div className="text-xs text-emerald-400 flex items-center gap-1"><TrendingUp size={12} /> Max: {stats?.user?.MaxRating || 16000}</div></CardContent>
                </Card>
                <Card className="bg-slate-900 border-slate-800">
                  <CardHeader className="pb-2"><CardDescription>Play Count</CardDescription><CardTitle className="text-3xl text-white">{stats?.totalPlays || 1248}</CardTitle></CardHeader>
                  <CardContent><div className="text-xs text-slate-500">Total recorded plays</div></CardContent>
                </Card>
                <Card className="bg-slate-900 border-slate-800">
                  <CardHeader className="pb-2"><CardDescription>Server Status</CardDescription><CardTitle className="text-3xl text-emerald-400">ONLINE</CardTitle></CardHeader>
                  <CardContent><div className="text-xs text-slate-500">All services operational</div></CardContent>
                </Card>
              </div>

              {/* Rating Composition Section */}
              <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
                <Card className="bg-slate-900 border-slate-800">
                  <CardHeader><CardTitle className="text-white flex items-center gap-2"><Award className="text-indigo-400" size={20} /> Best Bests (Top Rating)</CardTitle></CardHeader>
                  <CardContent className="space-y-3">
                    {stats?.ratingComposition?.bests?.map((song, idx) => (
                      <div key={idx} className="flex items-center justify-between p-3 bg-slate-800/60 rounded-lg">
                        <div>
                          <div className="font-bold text-white text-sm">{song.title}</div>
                          <div className="text-xs text-slate-400">Level: {song.level} | Score: {song.score}</div>
                        </div>
                        <Badge className="bg-indigo-600 text-white font-mono">+{song.rating}</Badge>
                      </div>
                    )) || <div className="text-sm text-slate-500">No composition data</div>}
                  </CardContent>
                </Card>

                <Card className="bg-slate-900 border-slate-800">
                  <CardHeader><CardTitle className="text-white flex items-center gap-2"><Award className="text-emerald-400" size={20} /> New Bests</CardTitle></CardHeader>
                  <CardContent className="space-y-3">
                    {stats?.ratingComposition?.newBests?.map((song, idx) => (
                      <div key={idx} className="flex items-center justify-between p-3 bg-slate-800/60 rounded-lg">
                        <div>
                          <div className="font-bold text-white text-sm">{song.title}</div>
                          <div className="text-xs text-slate-400">Level: {song.level} | Score: {song.score}</div>
                        </div>
                        <Badge className="bg-emerald-600 text-white font-mono">+{song.rating}</Badge>
                      </div>
                    )) || <div className="text-sm text-slate-500">No new best data</div>}
                  </CardContent>
                </Card>
              </div>

              <Card className="bg-slate-900 border-slate-800">
                <CardHeader><CardTitle className="text-white">Rating Trend</CardTitle></CardHeader>
                <CardContent className="h-[300px]">
                  <ResponsiveContainer width="100%" height="100%">
                    <LineChart data={[
                      { name: 'Mon', rating: 15100 }, { name: 'Tue', rating: 15150 }, { name: 'Wed', rating: 15120 },
                      { name: 'Thu', rating: 15200 }, { name: 'Fri', rating: 15350 }, { name: 'Sat', rating: 15400 }, { name: 'Sun', rating: 15420 },
                    ]}>
                      <CartesianGrid strokeDasharray="3 3" stroke="#1e293b" />
                      <XAxis dataKey="name" stroke="#64748b" fontSize={12} />
                      <YAxis stroke="#64748b" fontSize={12} domain={['dataMin - 100', 'dataMax + 100']} />
                      <Tooltip contentStyle={{ backgroundColor: '#0f172a', border: '1px solid #1e293b' }} />
                      <Line type="monotone" dataKey="rating" stroke="#6366f1" strokeWidth={3} dot={{ r: 4, fill: '#6366f1' }} />
                    </LineChart>
                  </ResponsiveContainer>
                </CardContent>
              </Card>
            </div>
          )}

          {tab === 'maimai' && (
            <div className="space-y-6">
              <Tabs defaultSelectedKey="recent" className="w-full">
                <TabsList className="bg-slate-900 border border-slate-800">
                  <TabsTrigger id="recent" className="data-[selected]:bg-indigo-600">Recent Plays</TabsTrigger>
                  <TabsTrigger id="stats" className="data-[selected]:bg-indigo-600">Statistics</TabsTrigger>
                </TabsList>
                <TabsContent id="recent" className="mt-6">
                  <Card className="bg-slate-900 border-slate-800 overflow-hidden">
                    <Table>
                      <TableHeader className="bg-slate-800/50">
                        <TableRow>
                          <TableHead className="text-slate-300">Music</TableHead>
                          <TableHead className="text-slate-300">Level</TableHead>
                          <TableHead className="text-slate-300 text-right">Achievement</TableHead>
                          <TableHead className="text-slate-300 text-right">Score</TableHead>
                          <TableHead className="text-slate-300 text-right">Date</TableHead>
                        </TableRow>
                      </TableHeader>
                      <TableBody>
                        {stats?.recentPlays && stats.recentPlays.length > 0 ? stats.recentPlays.map((p) => (
                          <TableRow key={p.ID} className="border-slate-800 hover:bg-slate-800/30">
                            <TableCell className="font-bold text-white">Song ID: {p.musicId}</TableCell>
                            <TableCell><Badge variant="outline" className="border-indigo-500/50 text-indigo-400">LV.{p.level}</Badge></TableCell>
                            <TableCell className="text-right font-mono text-emerald-400">{(p.achievement / 10000).toFixed(4)}%</TableCell>
                            <TableCell className="text-right font-mono">{p.score}</TableCell>
                            <TableCell className="text-right text-slate-500 text-xs">{p.createDate}</TableCell>
                          </TableRow>
                        )) : (
                          <TableRow><TableCell colSpan={5} className="text-center py-12 text-slate-500">No play history found.</TableCell></TableRow>
                        )}
                      </TableBody>
                    </Table>
                  </Card>
                </TabsContent>
                <TabsContent id="stats">
                  <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-4 gap-4">
                    {['SSS+', 'SSS', 'SS', 'S'].map(rank => (
                      <Card key={rank} className="bg-slate-900 border-slate-800 text-center p-6">
                        <div className="text-2xl font-black text-indigo-400 mb-1">{rank}</div>
                        <div className="text-sm text-slate-500 font-bold">124 songs</div>
                      </Card>
                    ))}
                  </div>
                </TabsContent>
              </Tabs>
            </div>
          )}

          {tab === 'setup' && (
            <div className="max-w-3xl space-y-6">
              <Card className="bg-slate-900 border-slate-800">
                <CardHeader>
                  <CardTitle className="text-white flex items-center gap-2"><BookOpen className="text-indigo-400" /> Connection & Setup Guide</CardTitle>
                  <CardDescription className="text-slate-400">Follow these instructions to connect your game client to MaiGoDX.</CardDescription>
                </CardHeader>
                <CardContent className="space-y-4 text-sm text-slate-300 leading-relaxed">
                  <div className="p-4 bg-slate-800/60 rounded-lg space-y-2">
                    <h3 className="font-bold text-white">1. Configure Hosts / DNS</h3>
                    <p>Redirect ALL.Net domain requests to your server IP address (e.g. `127.0.0.1 a net.am-all.net`).</p>
                  </div>
                  <div className="p-4 bg-slate-800/60 rounded-lg space-y-2">
                    <h3 className="font-bold text-white">2. Boot Game Client</h3>
                    <p>Ensure your game boots to the title screen and successfully authorizes via the local server instance.</p>
                  </div>
                  <div className="p-4 bg-slate-800/60 rounded-lg space-y-2">
                    <h3 className="font-bold text-white">3. Access Portal</h3>
                    <p>Log in with your registered email and password to track scores, view rating composition, and manage game profiles.</p>
                  </div>
                </CardContent>
              </Card>
            </div>
          )}

          {tab === 'admin' && user?.isAdmin && (
            <div className="space-y-8">
              <div className="space-y-6">
                <div className="flex justify-between items-center">
                  <h3 className="text-xl font-bold flex items-center gap-2"><Users /> User Management</h3>
                  <Button onClick={fetchUsers} size="sm" variant="outline" className="border-slate-700">Refresh List</Button>
                </div>
                <Card className="bg-slate-900 border-slate-800 overflow-hidden">
                  <Table>
                    <TableHeader className="bg-slate-800/50">
                      <TableRow>
                        <TableHead className="text-slate-300">User</TableHead>
                        <TableHead className="text-slate-300">Status</TableHead>
                        <TableHead className="text-slate-300">Role</TableHead>
                        <TableHead className="text-slate-300 text-right">Actions</TableHead>
                      </TableRow>
                    </TableHeader>
                    <TableBody>
                      {userList.map((u) => (
                        <TableRow key={u.ID} className="border-slate-800 hover:bg-slate-800/30">
                          <TableCell>
                            <div className="flex items-center gap-3">
                              <Avatar className="h-8 w-8"><AvatarFallback>{u.username ? u.username[0] : 'U'}</AvatarFallback></Avatar>
                              <div>
                                <div className="font-bold text-white text-sm">{u.username}</div>
                                <div className="text-[10px] text-slate-500">{u.email}</div>
                              </div>
                            </div>
                          </TableCell>
                          <TableCell>{u.isVerified ? <Badge className="bg-emerald-500/10 text-emerald-500 border-emerald-500/20">Verified</Badge> : <Badge variant="outline" className="text-slate-500 border-slate-800">Pending</Badge>}</TableCell>
                          <TableCell>{u.isAdmin ? <Badge className="bg-amber-500/10 text-amber-500 border-amber-500/20">Admin</Badge> : <Badge variant="outline" className="text-slate-500 border-slate-800">User</Badge>}</TableCell>
                          <TableCell className="text-right"><Button variant="ghost" size="sm" className="text-indigo-400 hover:text-indigo-300 hover:bg-indigo-500/10">Edit</Button></TableCell>
                        </TableRow>
                      ))}
                    </TableBody>
                  </Table>
                </Card>
              </div>

              {/* Admin System Configurations Panel */}
              <div className="space-y-6">
                <div className="flex justify-between items-center">
                  <h3 className="text-xl font-bold flex items-center gap-2"><Sliders /> Server Configuration Management</h3>
                  <Button onClick={fetchConfigs} size="sm" variant="outline" className="border-slate-700">Refresh Configs</Button>
                </div>
                <Card className="bg-slate-900 border-slate-800">
                  <CardHeader><CardTitle className="text-white text-sm">Global System Settings & Flags</CardTitle></CardHeader>
                  <CardContent className="space-y-4">
                    {configs.map((cfg) => (
                      <div key={cfg.ID} className="flex items-center justify-between p-4 bg-slate-800/50 rounded-lg">
                        <div>
                          <div className="font-mono font-bold text-indigo-400 text-sm">{cfg.key}</div>
                          <div className="text-xs text-slate-400 mt-1">{cfg.desc}</div>
                        </div>
                        <div className="flex items-center gap-3">
                          <input 
                            type="text" 
                            defaultValue={cfg.value} 
                            onBlur={(e) => handleUpdateConfig(cfg.key, e.target.value)}
                            className="w-36 h-9 px-3 bg-slate-900 border border-slate-700 rounded-md text-sm text-white focus:outline-none focus:ring-2 focus:ring-indigo-500"
                          />
                        </div>
                      </div>
                    ))}
                  </CardContent>
                </Card>
              </div>
            </div>
          )}

          {tab === 'settings' && (
            <div className="max-w-2xl space-y-8">
              <Card className="bg-slate-900 border-slate-800">
                <CardHeader><CardTitle className="text-white">Profile Settings</CardTitle></CardHeader>
                <CardContent className="space-y-6">
                  <div className="flex items-center gap-6">
                    <Avatar className="h-20 w-20 border-2 border-slate-800">
                      <AvatarFallback className="bg-indigo-600 text-2xl font-black">{user?.username ? user.username[0].toUpperCase() : 'U'}</AvatarFallback>
                    </Avatar>
                    <Button variant="outline" className="border-slate-700">Change Avatar</Button>
                  </div>
                  <Separator className="bg-slate-800" />
                  <div className="grid grid-cols-2 gap-4">
                    <div className="space-y-2">
                      <label className="text-xs font-bold text-slate-500 uppercase tracking-wider">Username</label>
                      <input type="text" disabled value={user?.username || ''} className="w-full h-10 px-3 bg-slate-800/50 border-slate-700 rounded-md text-slate-400 cursor-not-allowed" />
                    </div>
                    <div className="space-y-2">
                      <label className="text-xs font-bold text-slate-500 uppercase tracking-wider">Email</label>
                      <input type="email" disabled value={user?.email || ''} className="w-full h-10 px-3 bg-slate-800/50 border-slate-700 rounded-md text-slate-400 cursor-not-allowed" />
                    </div>
                  </div>
                </CardContent>
              </Card>

              {/* Aime Card Binding Section */}
              <Card className="bg-slate-900 border-slate-800">
                <CardHeader>
                  <CardTitle className="text-white flex items-center gap-2"><CreditCard className="text-indigo-400" /> Aime Card Binding</CardTitle>
                  <CardDescription className="text-slate-400">Link your arcade Aime card access code to your account.</CardDescription>
                </CardHeader>
                <CardContent className="space-y-6">
                  <form onSubmit={handleBindCard} className="space-y-4">
                    <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                      <div className="space-y-2">
                        <label className="text-xs font-medium text-slate-300">Access Code (20 digits)</label>
                        <input type="text" required value={accessCode} onChange={e => setAccessCode(e.target.value)} placeholder="01234567890123456789" className="w-full h-10 px-3 bg-slate-800 border border-slate-700 rounded-md text-sm text-white focus:outline-none focus:ring-2 focus:ring-indigo-500" />
                      </div>
                      <div className="space-y-2">
                        <label className="text-xs font-medium text-slate-300">Card Alias / Nickname</label>
                        <input type="text" value={cardName} onChange={e => setCardName(e.target.value)} placeholder="My Main Card" className="w-full h-10 px-3 bg-slate-800 border border-slate-700 rounded-md text-sm text-white focus:outline-none focus:ring-2 focus:ring-indigo-500" />
                      </div>
                    </div>
                    <Button type="submit" className="bg-indigo-600 hover:bg-indigo-500 text-white font-bold">Bind Card</Button>
                  </form>

                  <Separator className="bg-slate-800" />

                  <div className="space-y-3">
                    <h4 className="font-bold text-sm text-white">Your Bound Cards</h4>
                    <div className="space-y-2">
                      {userCards.length > 0 ? userCards.map(card => (
                        <div key={card.ID} className="flex items-center justify-between p-3 bg-slate-800/60 rounded-lg">
                          <div>
                            <div className="font-bold text-white text-sm">{card.cardName}</div>
                            <div className="font-mono text-xs text-indigo-400">{card.accessCode}</div>
                          </div>
                          <Badge className="bg-emerald-500/10 text-emerald-500 border-emerald-500/20">Bound</Badge>
                        </div>
                      )) : (
                        <div className="text-sm text-slate-500">No cards bound yet.</div>
                      )}
                    </div>
                  </div>
                </CardContent>
              </Card>
            </div>
          )}
        </div>
      </main>
    </div>
  )
}
