import { useEffect, useState } from 'react'
import { ThemeProvider, SidebarLayout, NotificationProvider, LoginCallbackPage, ProtectedRoute, SidebarHeader } from '@gofreego/tsutils'
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'


// Icons
import DashboardIcon from '@mui/icons-material/Dashboard'

import AccountTreeIcon from '@mui/icons-material/AccountTree' // For projects
import TimelineIcon from '@mui/icons-material/Timeline' // For events
import PeopleIcon from '@mui/icons-material/People' // For persons
import FlagIcon from '@mui/icons-material/Flag' // For feature flags

// Components
import { ProjectSelector } from './components/ProjectSelector'

// Pages
import { DashboardPage } from './pages/dashboard/DashboardPage'
import { ProjectsPage } from './pages/projects/ProjectsPage'
import { EventsPage } from './pages/events/EventsPage'
import { PersonsPage } from './pages/persons/PersonsPage'
import { FeatureFlagsPage } from './pages/feature_flags/FeatureFlagsPage'

import { authService, sessionManager } from './services'

const LOGIN_URL = import.meta.env.VITE_LOGIN_URL as string || 'http://localhost:8080/login'

function App() {
  const [isInitialized, setIsInitialized] = useState(false);

  useEffect(() => {
    authService.initializeAuth()
    setIsInitialized(true);
  }, [])

  const handleLoginFailed = () => {
    console.log("Login failed, redirecting to -> ", LOGIN_URL)
    window.location.href = LOGIN_URL
  }

  const menuItems = [
    {
      id: 'dashboard',
      label: 'Dashboard',
      path: '/openclick/dashboard',
      icon: <DashboardIcon />,
    },
    {
      id: 'projects',
      label: 'Projects',
      path: '/openclick/projects',
      icon: <AccountTreeIcon />,
    },
    {
      id: 'events',
      label: 'Events & Replay',
      path: '/openclick/events',
      icon: <TimelineIcon />,
    },
    {
      id: 'persons',
      label: 'Persons & Cohorts',
      path: '/openclick/persons',
      icon: <PeopleIcon />,
    },
    {
      id: 'feature-flags',
      label: 'Feature Flags',
      path: '/openclick/feature-flags',
      icon: <FlagIcon />,
    },
  ]

  if (!isInitialized) {
    return <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', height: '100vh' }}>
      <div>Loading...</div>
    </div>
  }

  return (
    <ThemeProvider>
      <NotificationProvider>
        <BrowserRouter>
          <Routes>
            <Route path="/openclick/login-callback" element={<LoginCallbackPage authService={authService} navigateTo="/openclick/dashboard" onLoginFailed={handleLoginFailed} />} />
            <Route
              path="/"
              element={
                <ProtectedRoute sessionManager={sessionManager} loginUrl={LOGIN_URL} callbackPath="/openclick/login-callback">
                  <SidebarLayout
                    menuItems={menuItems}
                    isRouter={true}
                    isBrowserRouter={false}
                    style={{ height: '100vh' }}
                    header={<SidebarHeader title="OpenClick" homePath="/" />}
                    footer={<ProjectSelector />}
                  />
                </ProtectedRoute>
              }
            >
              <Route index element={<Navigate to="/openclick/dashboard" replace />} />
              <Route path="openclick/dashboard" element={<DashboardPage />} />
              <Route path="openclick/projects" element={<ProjectsPage />} />
              <Route path="openclick/events" element={<EventsPage />} />
              <Route path="openclick/persons" element={<PersonsPage />} />
              <Route path="openclick/feature-flags" element={<FeatureFlagsPage />} />
              <Route path="*" element={<Navigate to="/openclick/dashboard" replace />} />
            </Route>
          </Routes>
        </BrowserRouter>
      </NotificationProvider>
    </ThemeProvider>
  )
}

export default App
