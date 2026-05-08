import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import Layout from './components/Layout'
import Dashboard from './pages/Dashboard'
import Diagnosis from './pages/Diagnosis'
import DiagnosisHistory from './pages/DiagnosisHistory'
import Assistant from './pages/Assistant'
import Users from './pages/Users'
import Clusters from './pages/Clusters'
import System from './pages/System'

function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<Layout />}>
          <Route index element={<Navigate to="/dashboard" replace />} />
          <Route path="dashboard" element={<Dashboard />} />
          <Route path="metastore" element={<div>元仓 (开发中)</div>} />
          <Route path="diagnosis/job" element={<Diagnosis />} />
          <Route path="diagnosis/history" element={<DiagnosisHistory />} />
          <Route path="assistant" element={<Assistant />} />
          <Route path="settings/users" element={<Users />} />
          <Route path="settings/clusters" element={<Clusters />} />
          <Route path="settings/system" element={<System />} />
        </Route>
      </Routes>
    </BrowserRouter>
  )
}

export default App