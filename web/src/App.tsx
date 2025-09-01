import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import ParentDashboard from './pages/ParentDashboard'
import ChildView from './pages/ChildView'
import Login from './pages/Login'

function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<Navigate to="/login" replace />} />
        <Route path="/login" element={<Login />} />
        <Route path="/parent/:parentId" element={<ParentDashboard />} />
        <Route path="/child/:childId" element={<ChildView />} />
      </Routes>
    </BrowserRouter>
  )
}

export default App
