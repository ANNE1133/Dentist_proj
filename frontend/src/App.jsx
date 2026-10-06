import React from 'react';
import Navbar from './components/Navbar';
import { Routes, Route, Navigate } from 'react-router-dom';
import CreateDentistPage from './pages/CreateDentistPage';
import CameraHistoryLogPage from './pages/CameraHistoryLogPage';
import LiveMonitorPage from './pages/LiveMonitorPage';

export default function App() {
  return (
    <div className="app">
        <Navbar />
        <main className="page-container">
            <Routes>
                <Route path="/" element={<Navigate to="/create" replace />} />

                <Route path="/create" element={<CreateDentistPage />} />

                <Route path="/live" element={<LiveMonitorPage />} />

                <Route path="/history" element={<CameraHistoryLogPage />} />
            </Routes>
        </main>
    </div>
  );
}
