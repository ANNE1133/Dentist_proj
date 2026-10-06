import React, { useEffect, useState } from 'react';

const API_BASE = "http://localhost:4000/api";

export default function CameraHistoryLogPage() {
  const [logs, setLogs] = useState([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);
  const [form, setForm] = useState({
    dentistId: '',
    bodyComposition: '',
    adjustmentItems: '',
    adjustmentStatus: 'pending',
  });
  const [submitting, setSubmitting] = useState(false);

  const loadLogs = async () => {
    setLoading(true);
    setError(null);
    try {
      const res = await fetch(`${API_BASE}/camerahistorylog`);
      if (!res.ok) throw new Error('Failed to load logs');
      const data = await res.json();
      setLogs(data);
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  const handleCreateLog = async (e) => {
    e.preventDefault();
    if (!form.dentistId) {
      setError('Dentist ID is required.');
      return;
    }

    setSubmitting(true);
    setError(null);

    try {
      const payload = {
        dentistId: form.dentistId,
        bodyComposition: form.bodyComposition || 'unknown',
        adjustmentItems: form.adjustmentItems
          .split(',')
          .map((item) => item.trim())
          .filter(Boolean),
        adjustmentStatus: form.adjustmentStatus || 'pending',
      };

      const res = await fetch(`${API_BASE}/camerahistorylog`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(payload),
      });

      const data = await res.json().catch(() => ({}));
      if (!res.ok) {
        throw new Error(data.error || 'Failed to create log');
      }

      setForm({
        dentistId: '',
        bodyComposition: '',
        adjustmentItems: '',
        adjustmentStatus: 'pending',
      });
      await loadLogs();
    } catch (err) {
      setError(err.message || String(err));
    } finally {
      setSubmitting(false);
    }
  };

  useEffect(() => {
    loadLogs();
  }, []);

  return (
    <section className="panel">
      <h2>Camera History Log</h2>

      <form onSubmit={handleCreateLog} style={{ marginBottom: 16 }}>
        <div className="row">
          <div className="field">
            <label>Dentist ID</label>
            <input
              value={form.dentistId}
              onChange={(e) => setForm({ ...form, dentistId: e.target.value })}
              required
            />
          </div>

          <div className="field">
            <label>Body Composition</label>
            <input
              value={form.bodyComposition}
              onChange={(e) => setForm({ ...form, bodyComposition: e.target.value })}
              placeholder="e.g. normal"
            />
          </div>
        </div>

        <div className="row">
          <div className="field">
            <label>Adjustment Items</label>
            <input
              value={form.adjustmentItems}
              onChange={(e) => setForm({ ...form, adjustmentItems: e.target.value })}
              placeholder="Bracket, Crown"
            />
          </div>

          <div className="field">
            <label>Status</label>
            <select
              value={form.adjustmentStatus}
              onChange={(e) => setForm({ ...form, adjustmentStatus: e.target.value })}
            >
              <option value="pending">Pending</option>
              <option value="done">Done</option>
              <option value="failed">Failed</option>
            </select>
          </div>
        </div>

        <button type="submit" disabled={submitting} style={{ marginTop: 8 }}>
          {submitting ? 'Saving...' : 'Create Log'}
        </button>
      </form>

      <div className="controls">
        <button onClick={loadLogs} disabled={loading}>{loading ? 'Refreshing...' : 'Refresh'}</button>
      </div>
      {error && <div className="error">{error}</div>}
      {logs.length === 0 ? (
        <p>No logs yet.</p>
      ) : (
        <ul className="log-list">
          {logs.map((entry) => (
            <li key={entry.id}>
              <div>{entry.created_at} · {entry.dentistId || 'unknown'}</div>
              <div>Body: {entry.bodyComposition || 'unknown'} · Status: {entry.adjustmentStatus || 'pending'}</div>
              <div>Items: {entry.adjustmentItems?.join(', ') || 'none'}</div>
            </li>
          ))}
        </ul>
      )}
    </section>
  );
}
