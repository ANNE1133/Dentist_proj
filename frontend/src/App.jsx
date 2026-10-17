import { useEffect, useState } from 'react';
import { Navigate, Outlet, Route, Routes } from 'react-router-dom';
import Navbar from './components/Navbar';
import CameraHistoryLogPage from './pages/CameraHistoryLogPage';
import CreateDentistPage from './pages/CreateDentistPage';
import LiveMonitorPage from './pages/LiveMonitorPage';
import LoginPage from './pages/LoginPage';
import { checkSession, logout } from './services/authService';

function ProtectedRoute({ isAuthenticated }) {
  return isAuthenticated ? <Outlet /> : <Navigate to="/login" replace />;
}

function AppLayout({ onLogout, logoutError }) {
  return (
    <>
      <Navbar onLogout={onLogout} />
      <main className="page-container">
        {logoutError && <p className="app-error" role="alert">{logoutError}</p>}
        <Outlet />
      </main>
    </>
  );
}

export default function App() {
  const [isAuthenticated, setIsAuthenticated] = useState(false);
  const [isCheckingSession, setIsCheckingSession] = useState(true);
  const [logoutError, setLogoutError] = useState('');
  const [sessionError, setSessionError] = useState('');

  useEffect(() => {
    let isActive = true;

    checkSession()
      .then((authenticated) => {
        if (isActive) setIsAuthenticated(authenticated);
      })
      .catch(() => {
        if (isActive) {
          setIsAuthenticated(false);
          setSessionError(error.message || 'Unable to verify your login session.');
        }
      })
      .finally(() => {
        if (isActive) setIsCheckingSession(false);
      });

    return () => {
      isActive = false;
    };
  }, []);

  const handleLogout = async () => {
    setLogoutError('');
    try {
      await logout();
      setIsAuthenticated(false);
    } catch (error) {
      setLogoutError(error.message || 'Unable to log out. Please try again.');
    }
  };

  if (isCheckingSession) {
    return <div className="app-loading" role="status">Checking your session...</div>;
  }

  return (
    <Routes>
      <Route
        path="/login"
        element={
          isAuthenticated
            ? <Navigate to="/create" replace />
            : <LoginPage
                onLogin={() => setIsAuthenticated(true)}
                initialError={sessionError}
              />
        }
      />
      <Route element={<ProtectedRoute isAuthenticated={isAuthenticated} />}>
        <Route element={<AppLayout onLogout={handleLogout} logoutError={logoutError} />}>
          <Route path="/" element={<Navigate to="/create" replace />} />
          <Route path="/create" element={<CreateDentistPage />} />
          <Route path="/live" element={<LiveMonitorPage />} />
          <Route path="/history" element={<CameraHistoryLogPage />} />
          <Route path="*" element={<Navigate to="/create" replace />} />
        </Route>
      </Route>
      <Route path="*" element={<Navigate to={isAuthenticated ? '/create' : '/login'} replace />} />
    </Routes>
  );
}
