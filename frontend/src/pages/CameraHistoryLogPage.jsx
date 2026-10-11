import React, { useCallback, useEffect, useState } from 'react';

const API_BASE = 'http://localhost:4000/api';
const PAGE_SIZE = 20;

// Keep the patient-specific measurement selection in one place for the
// client-chair calculation. The API expects the measurements as a list.
function calculateClientChair(bodyMeasurements = []) {
  return Array.isArray(bodyMeasurements) ? bodyMeasurements : [];
}

export default function CameraHistoryLogPage() {
  const [result, setResult] = useState({
    data: [],
    page: 1,
    limit: PAGE_SIZE,
    total: 0,
    totalPages: 0,
  });

  const [page, setPage] = useState(1);
  const [startTime, setStartTime] = useState('');
  const [endTime, setEndTime] = useState('');
  const [loading, setLoading] = useState(false);
  const [adjustingLogId, setAdjustingLogId] = useState(null);
  const [error, setError] = useState('');

  const loadLogs = useCallback(async (requestedPage = page) => {
    setLoading(true);
    setError('');

    try {
      const params = new URLSearchParams({
        page: String(requestedPage),
        limit: String(PAGE_SIZE),
      });

      if (startTime) {
        params.set('startTime', new Date(startTime).toISOString());
      }

      if (endTime) {
        params.set('endTime', new Date(endTime).toISOString());
      }

      const res = await fetch(`${API_BASE}/camera/logs?${params}`);

      const responseText = await res.text();

      let data;
      try {
        data = responseText ? JSON.parse(responseText) : {};
      } catch {
        data = {};
      }

      if (!res.ok) {
        throw new Error(
          data.error || responseText || 'Failed to load camera history'
        );
      }

      setResult(data);
      setPage(data.page);
    } catch (err) {
      setError(err.message || 'Failed to load camera history');
    } finally {
      setLoading(false);
    }
  }, [page, startTime, endTime]);

  useEffect(() => {
    loadLogs(1);
  }, []);

  const submitFilters = (event) => {
    event.preventDefault();
    loadLogs(1);
  };

  const create3dAdjust = async (entry) => {
    setAdjustingLogId(entry.id);
    setError('');

    try {
      const itemsRes = await fetch(`${API_BASE}/adjustitem`);
      const itemsText = await itemsRes.text();
      let adjustableItems;
      try {
        adjustableItems = itemsText ? JSON.parse(itemsText) : [];
      } catch {
        adjustableItems = [];
      }
      if (!itemsRes.ok) {
        throw new Error(itemsText || 'Failed to load adjustable items');
      }

      const loggedNames = new Set((entry.adjustmentItems || []).map((name) =>
        String(name).toLowerCase()
      ));
      const adjustmentItems = adjustableItems
        .filter((item) => loggedNames.has(String(item.adjustName).toLowerCase()))
        .map((item) => ({
          adjustName: item.adjustName,
          adjustId: item.adjustItemId,
          dataUsed: item.patientAdjust
            ? calculateClientChair(entry.bodyMeasurements)
            : (entry.bodyMeasurements || []),
        }));

      const res = await fetch(`${API_BASE}/3dadust`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          historyLogId: entry.id,
          status: false,
          adjustmentItems,
        }),
      });
      const responseText = await res.text();
      let data;
      try {
        data = responseText ? JSON.parse(responseText) : {};
      } catch {
        data = {};
      }

      if (!res.ok) {
        throw new Error(data.error || responseText || 'Failed to create 3D adjustment');
      }

      await loadLogs(page);
    } catch (err) {
      setError(err.message || 'Failed to create 3D adjustment');
    } finally {
      setAdjustingLogId(null);
    }
  };

  return (
    <section className="panel">
      <h2>Camera History</h2>

      {/* Filters */}
      <form
        onSubmit={submitFilters}
        className="row"
        style={{
          alignItems: 'end',
          marginBottom: 16,
        }}
      >
        <div className="field">
          <label htmlFor="history-start">Start date/time</label>
          <input
            id="history-start"
            type="datetime-local"
            value={startTime}
            onChange={(e) => setStartTime(e.target.value)}
          />
        </div>

        <div className="field">
          <label htmlFor="history-end">End date/time</label>
          <input
            id="history-end"
            type="datetime-local"
            value={endTime}
            onChange={(e) => setEndTime(e.target.value)}
          />
        </div>

        <button type="submit" disabled={loading}>
          {loading ? 'Loading…' : 'Apply filters'}
        </button>
      </form>

      {/* Error */}
      {error && (
        <div className="error" role="alert">
          {error}
        </div>
      )}

      {/* Loading */}
      {loading && <p>Loading camera history…</p>}

      {/* Empty */}
      {!loading && !error && result.data.length === 0 && (
        <p>No camera history found.</p>
      )}

      {/* Table */}
      {!loading && !error && result.data.length > 0 && (
        <div
          style={{
            width: '100%',
            overflowX: 'auto',
          }}
        >
          <table
            style={{
              width: '100%',
              borderCollapse: 'collapse',
              minWidth: 1020,
            }}
          >
            <thead>
              <tr>
                <th style={thStyle}>Time</th>
                <th style={thStyle}>Camera</th>
                <th style={thStyle}>DentistID</th>
                <th style={thStyle}>Recognition</th>
                <th style={thStyle}>Body Measurements</th>
                <th style={thStyle}>Adjustment Items</th>
                <th style={thStyle}>Adjustment Status</th>
                <th style={thStyle}>3D Adjustment</th>
              </tr>
            </thead>

            <tbody>
              {result.data.map((entry) => {
                const unknown = entry.dentistId === 'UNKNOWN';
                const adjustmentStatus = entry.adjustmentStatus || 'not_applicable';

                return (
                  <tr key={entry.id}>
                    {/* Time */}
                    <td style={tdStyle}>
                      {entry.created_at
                        ? new Date(entry.created_at).toLocaleString()
                        : 'Time unavailable'}
                    </td>

                    {/* Camera */}
                    <td style={tdStyle}>
                      {entry.cameraName || 'Unknown Camera'}
                    </td>

                    {/* DentistID */}
                    <td style={tdStyle}>
                      {unknown ? 'Unknown' : entry.dentistId}
                    </td>

                    {/* Recognition */}
                    <td style={tdStyle}>
                      <span
                        style={{
                          ...statusStyle,
                          ...(unknown
                            ? unknownStatusStyle
                            : recognizedStatusStyle),
                        }}
                      >
                        {unknown
                          ? 'Unknown Person'
                          : 'Recognized Dentist'}
                      </span>
                    </td>

                    {/* Body Measurements */}
                    <td style={tdStyle}>
                      {entry.bodyMeasurements?.length
                        ? entry.bodyMeasurements.map((measurement) => `${measurement.part}: ${measurement.value} ${measurement.unit}`).join(' · ')
                        : 'Not measured'}
                    </td>

                    {/* Adjustment Items */}
                    <td style={tdStyle}>
                      {entry.adjustmentItems?.length
                        ? entry.adjustmentItems.join(', ')
                        : unknown
                          ? 'None'
                          : 'Not determined yet'}
                    </td>

                    {/* Adjustment Status */}
                    <td style={tdStyle}>
                      {adjustmentStatus === 'not_applicable'
                        ? 'Not applicable'
                        : adjustmentStatus === 'success'
                          ? 'Success'
                          : 'Pending'}
                    </td>

                    <td style={tdStyle}>
                      <button
                        type="button"
                        onClick={() => create3dAdjust(entry)}
                        disabled={adjustmentStatus !== 'pending'}
                      >
                        {adjustingLogId === entry.id ? 'Creating…' : 'Create 3D adjustment'}
                      </button>
                    </td>
                  </tr>
                );
              })}
            </tbody>
          </table>
        </div>
      )}

      {/* Pagination */}
      {result.totalPages > 0 && (
        <div
          className="controls"
          aria-label="Camera history pagination"
          style={{
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            gap: 16,
            marginTop: 20,
          }}
        >
          <button
            onClick={() => loadLogs(page - 1)}
            disabled={loading || page <= 1}
          >
            Previous
          </button>

          <span>
            Page {page} of {result.totalPages} · {result.total} records
          </span>

          <button
            onClick={() => loadLogs(page + 1)}
            disabled={loading || page >= result.totalPages}
          >
            Next
          </button>
        </div>
      )}
    </section>
  );
}

/* Table styles */
const thStyle = {
  textAlign: 'left',
  padding: '12px 16px',
  borderBottom: '2px solid #ddd',
  whiteSpace: 'nowrap',
};

const tdStyle = {
  padding: '12px 16px',
  borderBottom: '1px solid #eee',
  verticalAlign: 'top',
};

const statusStyle = {
  display: 'inline-block',
  padding: '4px 8px',
  borderRadius: 6,
  fontSize: 13,
  fontWeight: 500,
};

const recognizedStatusStyle = {
  background: '#e8f5e9',
  color: '#2e7d32',
};

const unknownStatusStyle = {
  background: '#f5f5f5',
  color: '#666',
};
