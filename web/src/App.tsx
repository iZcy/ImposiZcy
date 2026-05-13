import { BrowserRouter, Routes, Route } from 'react-router-dom'
import { AppProvider } from './context/AppContext'
import { Layout, DashboardLogin, ProtectedDashboard, clearToken } from '@kzcy/dashboard'
import ErrorBoundary from './components/ErrorBoundary'
import Dashboard from './pages/Dashboard'
import Templates from './pages/Templates'
import TemplateEditor from './pages/TemplateEditor'
import Images from './pages/Images'
import RenderJobs from './pages/RenderJobs'
import Kafka from './pages/Kafka'
import Roles from './pages/Roles'
import Settings from './pages/Settings'

const navItems = [
  { to: '/', label: 'Home', icon: '🏠' },
  { to: '/templates', label: 'Templates', icon: '📄' },
  { to: '/render-jobs', label: 'Render Jobs', icon: '🖼️' },
  { to: '/images', label: 'Images', icon: '📁' },
  { to: '/kafka', label: 'Kafka', icon: '📡' },
  { to: '/roles', label: 'Roles', icon: '🔐' },
  { to: '/settings', label: 'Settings', icon: '🔧' },
]

const TOKEN_KEY = 'imposizcy_token'

function App() {
  const handleLogout = () => {
    clearToken(TOKEN_KEY)
    window.location.href = '/login'
  }

  return (
    <ErrorBoundary>
      <AppProvider>
        <BrowserRouter>
          <Routes>
            <Route path="/login" element={
              <DashboardLogin
                serviceName="ImposiZcy"
                serviceIcon="🖨️"
                tokenKey={TOKEN_KEY}
                authEndpoint="/api/v1/auth/login"
                showDefaults={true}
                defaultUsername="admin"
                defaultPassword="admin123"
              />
            } />
            <Route path="/*" element={
              <ProtectedDashboard tokenKey={TOKEN_KEY}>
                <Layout
                  serviceName="ImposiZcy"
                  serviceIcon="🖨️"
                  navItems={navItems}
                  onLogout={handleLogout}
                >
                  <Routes>
                    <Route index element={<Dashboard />} />
                    <Route path="templates" element={<Templates />} />
                    <Route path="templates/:id/editor" element={<TemplateEditor />} />
                    <Route path="render-jobs" element={<RenderJobs />} />
                    <Route path="images" element={<Images />} />
                    <Route path="kafka" element={<Kafka />} />
                    <Route path="roles" element={<Roles />} />
                    <Route path="settings" element={<Settings />} />
                  </Routes>
                </Layout>
              </ProtectedDashboard>
            } />
          </Routes>
        </BrowserRouter>
      </AppProvider>
    </ErrorBoundary>
  )
}

export default App
