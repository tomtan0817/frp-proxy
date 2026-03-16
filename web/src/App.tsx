import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { getUser, isAdmin } from './auth';
import AppLayout from './components/Layout';
import Login from './pages/Login';
import Register from './pages/Register';
import Domains from './pages/Domains';
import AdminUsers from './pages/AdminUsers';
import AdminDomains from './pages/AdminDomains';
import AdminInvites from './pages/AdminInvites';

function RequireAuth({ children }: { children: React.ReactNode }) {
  const user = getUser();
  if (!user) return <Navigate to="/login" replace />;
  return <>{children}</>;
}

function RequireAdmin({ children }: { children: React.ReactNode }) {
  if (!isAdmin()) return <Navigate to="/domains" replace />;
  return <>{children}</>;
}

function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/login" element={<Login />} />
        <Route path="/register" element={<Register />} />
        <Route
          path="/domains"
          element={
            <RequireAuth>
              <AppLayout>
                <Domains />
              </AppLayout>
            </RequireAuth>
          }
        />
        <Route
          path="/admin/users"
          element={
            <RequireAuth>
              <RequireAdmin>
                <AppLayout>
                  <AdminUsers />
                </AppLayout>
              </RequireAdmin>
            </RequireAuth>
          }
        />
        <Route
          path="/admin/domains"
          element={
            <RequireAuth>
              <RequireAdmin>
                <AppLayout>
                  <AdminDomains />
                </AppLayout>
              </RequireAdmin>
            </RequireAuth>
          }
        />
        <Route
          path="/admin/invites"
          element={
            <RequireAuth>
              <RequireAdmin>
                <AppLayout>
                  <AdminInvites />
                </AppLayout>
              </RequireAdmin>
            </RequireAuth>
          }
        />
        <Route path="/" element={<Navigate to="/domains" replace />} />
      </Routes>
    </BrowserRouter>
  );
}

export default App;
