import React, { useState, useEffect, useCallback } from 'react';
import GameBoard from './GameBoard';
import wsService from '../services/websocket';
import './Game.css';

const Game = () => {
  const [username, setUsername] = useState('');
  const [gameState, setGameState] = useState(null);
  const [playerInfo, setPlayerInfo] = useState(null);
  const [status, setStatus] = useState('lobby'); // lobby, waiting, playing, finished
  const [error, setError] = useState('');
  const [message, setMessage] = useState('');
  const [turnTimeLeft, setTurnTimeLeft] = useState(30);
  const [totalTimeLeft, setTotalTimeLeft] = useState(60);

  useEffect(() => {
    // Connect to WebSocket
    wsService.connect(() => {
      // After WebSocket connects, check for existing session
      const savedSession = localStorage.getItem('gameSession');
      if (savedSession) {
        try {
          const session = JSON.parse(savedSession);
          const sessionAge = Date.now() - session.timestamp;
          
          // Only reconnect if session is less than 30 seconds old
          if (sessionAge < 30000 && session.sessionToken) {
            console.log('Attempting to reconnect with saved session...');
            wsService.send('reconnect', { session_token: session.sessionToken });
          } else {
            console.log('Session expired or invalid, clearing...');
            localStorage.removeItem('gameSession');
          }
        } catch (e) {
          console.error('Error parsing saved session:', e);
          localStorage.removeItem('gameSession');
        }
      }
    });

    // Set up message handlers
    wsService.on('player_info', handlePlayerInfo);
    wsService.on('waiting', handleWaiting);
    wsService.on('game_update', handleGameUpdate);
    wsService.on('error', handleError);
    wsService.on('reconnected', handleReconnected);

    return () => {
      wsService.disconnect();
    };
  }, []);

  // Timer effect - updates turn and inactivity timers every second
  useEffect(() => {
    if (status !== 'playing' || !gameState) {
      return;
    }

    const interval = setInterval(() => {
      if (gameState.turn_started_at) {
        const turnStarted = new Date(gameState.turn_started_at);
        const turnElapsed = Math.floor((Date.now() - turnStarted) / 1000);
        const turnRemaining = Math.max(0, gameState.turn_timeout_sec - turnElapsed);
        setTurnTimeLeft(turnRemaining);
      }

      if (gameState.last_move_at) {
        const lastMove = new Date(gameState.last_move_at);
        const totalElapsed = Math.floor((Date.now() - lastMove) / 1000);
        const totalRemaining = Math.max(0, 60 - totalElapsed);
        setTotalTimeLeft(totalRemaining);
      }
    }, 1000);

    return () => clearInterval(interval);
  }, [status, gameState]);

  const handlePlayerInfo = useCallback((payload) => {
    setPlayerInfo(payload);
    
    // Save session token for reconnect (with timestamp)
    if (payload.session_token) {
      const session = {
        sessionToken: payload.session_token,
        playerID: payload.player_id,
        gameID: payload.game_id,
        username: payload.username,
        timestamp: Date.now()
      };
      localStorage.setItem('gameSession', JSON.stringify(session));
      console.log('Session saved for reconnect:', session);
    }
  }, []);

  const handleWaiting = useCallback((payload) => {
    setStatus('waiting');
    setMessage(payload.message || 'Waiting for opponent...');
  }, []);

  const handleGameUpdate = useCallback((payload) => {
    setGameState(payload);
    
    // Reset timers when game updates
    if (payload.turn_timeout_sec) {
      setTurnTimeLeft(payload.turn_timeout_sec);
    }
    setTotalTimeLeft(60);
    
    if (payload.status === 'in_progress') {
      setStatus('playing');
      setMessage('');
    } else if (payload.status === 'finished') {
      setStatus('finished');
      handleGameEnd(payload);
    }
  }, []);

  const handleError = useCallback((payload) => {
    const errorMsg = payload.message || 'An error occurred';
    
    // If reconnect fails, clear session and return to lobby
    if (errorMsg.includes('Reconnect failed') || 
        errorMsg.includes('Game not found') || 
        errorMsg.includes('session not found') ||
        errorMsg.includes('reconnect window expired')) {
      localStorage.removeItem('playerInfo');
      localStorage.removeItem('gameSession');
      setPlayerInfo(null);
      setGameState(null);
      setStatus('lobby');
      setError('Previous game expired. Please join a new game.');
      setTimeout(() => setError(''), 5000);
      return;
    }
    
    setError(errorMsg);
    setTimeout(() => setError(''), 5000);
  }, []);

  const handleReconnected = useCallback((payload) => {
    console.log('Reconnected:', payload);
    
    // Update player info with reconnected data
    setPlayerInfo({
      player_id: payload.player_id,
      game_id: payload.game_id,
      username: payload.username,
      session_token: payload.session_token
    });
    
    // Update session timestamp
    const session = {
      sessionToken: payload.session_token,
      playerID: payload.player_id,
      gameID: payload.game_id,
      username: payload.username,
      timestamp: Date.now()
    };
    localStorage.setItem('gameSession', JSON.stringify(session));
    
    setStatus('playing');
    setMessage('Reconnected to game!');
    setTimeout(() => setMessage(''), 3000);
  }, []);

  const handleGameEnd = (game) => {
    let resultMessage = '';
    
    if (game.result === 'draw') {
      resultMessage = "It's a draw!";
    } else if (game.winner) {
      const isWinner = playerInfo && game.winner.id === playerInfo.player_id;
      resultMessage = isWinner ? 'You won!' : `${game.winner.username} won!`;
    } else {
      resultMessage = 'Game ended';
    }
    
    setMessage(resultMessage);
    
    // Clear session data since game is over
    localStorage.removeItem('playerInfo');
    localStorage.removeItem('gameSession');
    console.log('Game ended, session cleared');
  };

  const handleJoinGame = (e) => {
    e.preventDefault();
    if (username.trim()) {
      wsService.joinGame(username.trim());
      setStatus('waiting');
    }
  };

  const handleManualReconnect = (e) => {
    e.preventDefault();
    const sessionToken = prompt('Enter your Session Token:');
    if (sessionToken && sessionToken.trim()) {
      console.log('Attempting manual reconnect with token:', sessionToken);
      wsService.send('reconnect', { session_token: sessionToken.trim() });
      setStatus('waiting');
      setMessage('Reconnecting...');
    }
  };

  const handleColumnClick = (column) => {
    if (status !== 'playing' || !gameState || !playerInfo) {
      return;
    }

    // Check if it's player's turn
    const currentPlayer = gameState.current_turn === 1 ? gameState.player1 : gameState.player2;
    if (currentPlayer && currentPlayer.id === playerInfo.player_id) {
      wsService.makeMove(column);
    }
  };

  const handlePlayAgain = () => {
    setStatus('lobby');
    setGameState(null);
    setPlayerInfo(null);
    setMessage('');
    setError('');
    localStorage.removeItem('playerInfo');
  };

  const renderTurnInfo = () => {
    if (!gameState || !playerInfo) return null;

    const currentPlayer = gameState.current_turn === 1 ? gameState.player1 : gameState.player2;
    const isMyTurn = currentPlayer && currentPlayer.id === playerInfo.player_id;
    
    return (
      <div className={`turn-info ${isMyTurn ? 'my-turn' : ''}`}>
        {isMyTurn ? 'Your Turn' : `${currentPlayer?.username}'s Turn`}
      </div>
    );
  };

  const renderPlayers = () => {
    if (!gameState) return null;

    return (
      <div className="players-info">
        <div className="player player1">
          <div className="player-disc"></div>
          <span>{gameState.player1?.username || 'Player 1'}</span>
          {gameState.player1?.is_bot && <span className="bot-badge">BOT</span>}
        </div>
        <div className="vs">VS</div>
        <div className="player player2">
          <div className="player-disc"></div>
          <span>{gameState.player2?.username || 'Waiting...'}</span>
          {gameState.player2?.is_bot && <span className="bot-badge">BOT</span>}
        </div>
        
        {/* Timer display */}
        <div className="timer-container">
          <div className={`timer turn-timer ${turnTimeLeft <= 10 ? 'warning' : ''}`}>
            <span className="timer-label">Turn:</span>
            <span className="timer-value">{turnTimeLeft}s</span>
          </div>
          <div className={`timer inactivity-timer ${totalTimeLeft <= 20 ? 'warning' : ''}`}>
            <span className="timer-label">Activity:</span>
            <span className="timer-value">{totalTimeLeft}s</span>
          </div>
        </div>
      </div>
    );
  };

  if (status === 'lobby') {
    return (
      <div className="game-container">
        <div className="lobby">
          <h1>4 in a Row</h1>
          <p className="subtitle">Connect four discs to win!</p>
          <form onSubmit={handleJoinGame}>
            <input
              type="text"
              placeholder="Enter your username"
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              maxLength={20}
              required
            />
            <button type="submit">Join Game</button>
          </form>
          <div className="divider">OR</div>
          <button className="reconnect-btn" onClick={handleManualReconnect}>
            Reconnect to Existing Game
          </button>
          {error && <div className="error">{error}</div>}
        </div>
      </div>
    );
  }

  if (status === 'waiting') {
    return (
      <div className="game-container">
        <div className="waiting">
          <h2>{message}</h2>
          <p>A bot will join if no player is found in 10 seconds...</p>
          
          {playerInfo && playerInfo.session_token && (
            <div className="session-info">
              <p className="session-label">Your Session Token (for reconnect from other devices):</p>
              <div className="session-token-box">
                <code>{playerInfo.session_token}</code>
                <button 
                  className="copy-btn"
                  onClick={() => {
                    navigator.clipboard.writeText(playerInfo.session_token);
                    setMessage('Session token copied!');
                    setTimeout(() => setMessage('Waiting for opponent...'), 2000);
                  }}
                >
                  Copy
                </button>
              </div>
              <p className="session-hint">💡 Save this to rejoin from another browser/device within 30s</p>
            </div>
          )}
          
          <div className="spinner"></div>
        </div>
      </div>
    );
  }

  return (
    <div className="game-container">
      <h1>4 in a Row</h1>
      
      {playerInfo && playerInfo.session_token && status === 'playing' && (
        <div className="session-info-compact">
          <span className="session-label-small">Session Token:</span>
          <code className="session-token-small">{playerInfo.session_token.substring(0, 8)}...</code>
          <button 
            className="copy-btn-small"
            onClick={() => {
              navigator.clipboard.writeText(playerInfo.session_token);
              setMessage('Session token copied!');
              setTimeout(() => setMessage(''), 2000);
            }}
            title="Copy full session token"
          >
            📋
          </button>
        </div>
      )}
      
      {renderPlayers()}
      {status === 'playing' && renderTurnInfo()}
      
      {gameState && (() => {
        const isMyTurn = playerInfo && (
          (gameState.current_turn === 1 && gameState.player1?.id === playerInfo.player_id) ||
          (gameState.current_turn === 2 && gameState.player2?.id === playerInfo.player_id)
        );
        const isDisabled = status !== 'playing' || !isMyTurn;
        
        return (
          <GameBoard
            board={gameState.board}
            onColumnClick={handleColumnClick}
            currentTurn={gameState.current_turn}
            disabled={isDisabled}
          />
        );
      })()}
      
      {message && <div className="message">{message}</div>}
      {error && <div className="error">{error}</div>}
      
      {status === 'finished' && (
        <div className="game-over">
          <button onClick={handlePlayAgain} className="play-again-btn">
            Play Again
          </button>
        </div>
      )}
    </div>
  );
};

export default Game;
