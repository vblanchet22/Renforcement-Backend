import { BrowserRouter, Routes, Route, Navigate, Outlet } from 'react-router-dom';
import { AuthProvider, useAuth } from './context/AuthContext';
import { ColocationProvider } from './context/ColocationContext';
import { AppLayout } from './components/layout';
import { Login, Register, Dashboard, Expenses, Balances } from './pages';

function ProtectedRoute() {
  const { isAuthenticated, isLoading } = useAuth();

  if (isLoading) {
    return (
      <div className="min-h-screen bg-slate-50 flex items-center justify-center">
        <div className="text-center">
          <div className="w-10 h-10 rounded-full border-[3px] border-primary border-t-transparent animate-spin mx-auto mb-4" />
          <p className="text-slate-400 text-sm">Chargement...</p>
        </div>
      </div>
    );
  }

  if (!isAuthenticated) {
    return <Navigate to="/login" replace />;
  }

  return (
    <ColocationProvider>
      <AppLayout>
        <Outlet />
      </AppLayout>
    </ColocationProvider>
  );
}

function PublicRoute() {
  const { isAuthenticated, isLoading } = useAuth();

  if (isLoading) {
    return (
      <div className="min-h-screen bg-slate-50 flex items-center justify-center">
        <div className="w-10 h-10 rounded-full border-[3px] border-primary border-t-transparent animate-spin" />
      </div>
    );
  }

  if (isAuthenticated) {
    return <Navigate to="/dashboard" replace />;
  }

  return <Outlet />;
}

function PlaceholderPage({ title }: { title: string }) {
  return (
    <div className="flex flex-col items-center justify-center min-h-[60vh]">
      <h1 className="font-display text-3xl text-slate-900 mb-2">{title}</h1>
      <p className="text-slate-500">Cette page est en cours de développement</p>
    </div>
  );
}

function App() {
  return (
    <BrowserRouter>
      <AuthProvider>
        <Routes>
          <Route element={<PublicRoute />}>
            <Route path="/login" element={<Login />} />
            <Route path="/register" element={<Register />} />
          </Route>

          <Route element={<ProtectedRoute />}>
            <Route path="/dashboard" element={<Dashboard />} />
            <Route path="/expenses" element={<Expenses />} />
            <Route path="/balances" element={<Balances />} />
            <Route path="/payments" element={<PlaceholderPage title="Remboursements" />} />
            <Route path="/decisions" element={<PlaceholderPage title="Décisions" />} />
            <Route path="/funds" element={<PlaceholderPage title="Cagnottes" />} />
            <Route path="/notifications" element={<PlaceholderPage title="Notifications" />} />
            <Route path="/settings" element={<PlaceholderPage title="Paramètres" />} />
          </Route>

          <Route path="/" element={<Navigate to="/dashboard" replace />} />
          <Route path="*" element={<Navigate to="/dashboard" replace />} />
        </Routes>
      </AuthProvider>
    </BrowserRouter>
  );
}

export default App;
