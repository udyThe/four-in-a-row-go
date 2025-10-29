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

  useEffect(() => {
    // Connect to WebSocket
    wsService.connect();

    // Set up message handlers
    wsService.on('player_info', handlePlayerInfo);
    wsService.on('waiting', handleWaiting);
    wsService.on('game_update', handleGameUpdate);
    wsService.on('error', handleError);
    wsService.on('reconnected', handleReconnected);

    // Check for existing game session
    const savedPlayerInfo = localStorage.getItem('playerInfo');
    if (savedPlayerInfo) {
      try {
        const info = JSON.parse(savedPlayerInfo);
        setPlayerInfo(info);
        setStatus('connecting');
        // Send reconnect - handleError will clear stale data if game not found
        wsService.reconnect(info.player_id, info.game_id);
      } catch (err) {
        localStorage.removeItem('playerInfo');
      }
    }

    return () => {
      wsService.disconnect();
    };
  }, []);

  const handlePlayerInfo = useCallback((payload) => {
    setPlayerInfo(payload);
    localStorage.setItem('playerInfo', JSON.stringify(payload));
  }, []);

  const handleWaiting = useCallback((payload) => {
    setStatus('waiting');
    setMessage(payload.message || 'Waiting for opponent...');
  }, []);

  const handleGameUpdate = useCallback((payload) => {
    setGameState(payload);
    
    if (payload.status === 'in_progress') {
      setStatus('playing');
      setMessage('');
    } else if (payload.status === 'finished') {
      setStatus('finished');
      handleGameEnd(payload);
    }
  }, []);

  const handleError = useCallback((payload) => {
    // If we get "Game not found" during reconnect attempt, clear stale data and go to lobby
    if (payload.message === 'Game not found') {
      localStorage.removeItem('playerInfo');
      setPlayerInfo(null);
      setGameState(null);
      setStatus('lobby');
      setError('Previous game expired. Please join a new game.');
      setTimeout(() => setError(''), 5000);
      return;
    }
    
    setError(payload.message || 'An error occurred');
    setTimeout(() => setError(''), 5000);
  }, []);

  const handleReconnected = useCallback((payload) => {
    console.log('Reconnected:', payload);
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
    localStorage.removeItem('playerInfo');
  };

  const handleJoinGame = (e) => {
    e.preventDefault();
    if (username.trim()) {
      wsService.joinGame(username.trim());
      setStatus('waiting');
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
          <div className="spinner"></div>
        </div>
      </div>
    );
  }

  return (
    <div className="game-container">
      <h1>4 in a Row</h1>
      
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
