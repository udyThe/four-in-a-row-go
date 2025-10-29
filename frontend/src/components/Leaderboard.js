import React, { useState, useEffect } from 'react';
import { getLeaderboard } from '../services/api';
import './Leaderboard.css';

const Leaderboard = () => {
  const [leaderboard, setLeaderboard] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  useEffect(() => {
    fetchLeaderboard();
  }, []);

  const fetchLeaderboard = async () => {
    try {
      setLoading(true);
      const data = await getLeaderboard(10);
      setLeaderboard(data || []);
      setError('');
    } catch (err) {
      setError('Failed to load leaderboard');
    } finally {
      setLoading(false);
    }
  };

  if (loading) {
    return (
      <div className="leaderboard-container">
        <div className="spinner"></div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="leaderboard-container">
        <div className="error">{error}</div>
        <button onClick={fetchLeaderboard}>Retry</button>
      </div>
    );
  }

  return (
    <div className="leaderboard-container">
      <h1>🏆 Leaderboard</h1>
      
      {leaderboard.length === 0 ? (
        <div className="empty-state">
          <p>No games played yet. Be the first!</p>
        </div>
      ) : (
        <div className="leaderboard-table">
          <div className="table-header">
            <div className="rank">Rank</div>
            <div className="username">Player</div>
            <div className="stats">Wins</div>
            <div className="stats">Losses</div>
            <div className="stats">Draws</div>
            <div className="stats">Win Rate</div>
          </div>
          
          {leaderboard.map((player, index) => {
            const totalGames = player.games_won + player.games_lost + player.games_drawn;
            const winRate = totalGames > 0 ? ((player.games_won / totalGames) * 100).toFixed(1) : 0;
            
            return (
              <div key={player.id} className={`table-row ${index < 3 ? `rank-${index + 1}` : ''}`}>
                <div className="rank">
                  {index === 0 && '🥇'}
                  {index === 1 && '🥈'}
                  {index === 2 && '🥉'}
                  {index > 2 && `#${index + 1}`}
                </div>
                <div className="username">{player.username}</div>
                <div className="stats">{player.games_won}</div>
                <div className="stats">{player.games_lost}</div>
                <div className="stats">{player.games_drawn}</div>
                <div className="stats">{winRate}%</div>
              </div>
            );
          })}
        </div>
      )}
      
      <button onClick={fetchLeaderboard} className="refresh-btn">
        🔄 Refresh
      </button>
    </div>
  );
};

export default Leaderboard;
