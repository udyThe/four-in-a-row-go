import React, { useState, useEffect } from 'react';
import './Analytics.css';

function Analytics() {
  const [hourlyData, setHourlyData] = useState([]);
  const [dailyData, setDailyData] = useState([]);
  const [view, setView] = useState('hourly');
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  useEffect(() => {
    fetchAnalytics();
    const interval = setInterval(fetchAnalytics, 30000); // Refresh every 30 seconds
    return () => clearInterval(interval);
  }, [view]);

  const fetchAnalytics = async () => {
    try {
      setLoading(true);
      const apiUrl = process.env.REACT_APP_API_URL || 'http://localhost:8080/api';
      
      if (view === 'hourly') {
        const response = await fetch(`${apiUrl}/analytics/hourly?hours=24`);
        if (!response.ok) throw new Error('Failed to fetch hourly analytics');
        const data = await response.json();
        setHourlyData(data || []);
      } else {
        const response = await fetch(`${apiUrl}/analytics/daily?days=30`);
        if (!response.ok) throw new Error('Failed to fetch daily analytics');
        const data = await response.json();
        setDailyData(data || []);
      }
      setError(null);
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  const formatDate = (dateString) => {
    const date = new Date(dateString);
    if (view === 'hourly') {
      return date.toLocaleString('en-US', { 
        month: 'short', 
        day: 'numeric', 
        hour: '2-digit',
        minute: '2-digit'
      });
    }
    return date.toLocaleDateString('en-US', { 
      month: 'short', 
      day: 'numeric',
      year: 'numeric'
    });
  };

  const calculateTotal = (data, field) => {
    return data.reduce((sum, item) => sum + (item[field] || 0), 0);
  };

  const calculateAverage = (data, field) => {
    if (data.length === 0) return 0;
    const total = calculateTotal(data, field);
    return (total / data.length).toFixed(1);
  };

  const findPeak = (data, field) => {
    if (data.length === 0) return null;
    return data.reduce((max, item) => 
      (item[field] || 0) > (max[field] || 0) ? item : max
    );
  };

  const currentData = view === 'hourly' ? hourlyData : dailyData;
  const peak = findPeak(currentData, 'games_started');

  return (
    <div className="analytics-container">
      <div className="analytics-header">
        <h1>Game Analytics</h1>
        <div className="view-toggle">
          <button 
            className={view === 'hourly' ? 'active' : ''}
            onClick={() => setView('hourly')}
          >
            Hourly (24h)
          </button>
          <button 
            className={view === 'daily' ? 'active' : ''}
            onClick={() => setView('daily')}
          >
            Daily (30d)
          </button>
        </div>
      </div>

      {loading && <div className="loading">Loading analytics...</div>}
      {error && <div className="error">Error: {error}</div>}

      {!loading && !error && currentData.length === 0 && (
        <div className="no-data">
          <p>No analytics data available yet.</p>
          <p>Play some games to generate analytics!</p>
        </div>
      )}

      {!loading && !error && currentData.length > 0 && (
        <>
          <div className="stats-summary">
            <div className="stat-card">
              <div className="stat-value">{calculateTotal(currentData, 'games_started')}</div>
              <div className="stat-label">Total Games Started</div>
            </div>
            <div className="stat-card">
              <div className="stat-value">{calculateTotal(currentData, 'games_completed')}</div>
              <div className="stat-label">Games Completed</div>
            </div>
            <div className="stat-card">
              <div className="stat-value">{calculateTotal(currentData, 'total_moves')}</div>
              <div className="stat-label">Total Moves</div>
            </div>
            <div className="stat-card">
              <div className="stat-value">{calculateAverage(currentData, 'avg_game_duration')}s</div>
              <div className="stat-label">Avg Game Duration</div>
            </div>
          </div>

          {peak && (
            <div className="peak-time">
              <h3>Peak Activity</h3>
              <p>
                {formatDate(view === 'hourly' ? peak.hour : peak.date)} - {peak.games_started} games started
              </p>
            </div>
          )}

          <div className="analytics-table">
            <table>
              <thead>
                <tr>
                  <th>{view === 'hourly' ? 'Hour' : 'Date'}</th>
                  <th>Games Started</th>
                  <th>Games Completed</th>
                  <th>Total Moves</th>
                  <th>Avg Duration</th>
                </tr>
              </thead>
              <tbody>
                {currentData.map((item, index) => (
                  <tr key={index}>
                    <td>{formatDate(view === 'hourly' ? item.hour : item.date)}</td>
                    <td>{item.games_started || 0}</td>
                    <td>{item.games_completed || 0}</td>
                    <td>{item.total_moves || 0}</td>
                    <td>{item.avg_game_duration ? item.avg_game_duration.toFixed(1) + 's' : '-'}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>

          <div className="chart-container">
            <h3>Games Activity Chart</h3>
            <div className="bar-chart">
              {currentData.slice(0, 20).reverse().map((item, index) => {
                const maxValue = Math.max(...currentData.map(d => d.games_started || 0));
                const height = maxValue > 0 ? ((item.games_started || 0) / maxValue * 100) : 0;
                
                return (
                  <div key={index} className="bar-wrapper">
                    <div 
                      className="bar" 
                      style={{ height: `${height}%` }}
                      title={`${item.games_started || 0} games`}
                    >
                      <span className="bar-value">{item.games_started || 0}</span>
                    </div>
                    <div className="bar-label">
                      {formatDate(view === 'hourly' ? item.hour : item.date).split(',')[0]}
                    </div>
                  </div>
                );
              })}
            </div>
          </div>
        </>
      )}
    </div>
  );
}

export default Analytics;
