import React from 'react';
import './GameBoard.css';

const ROWS = 6;
const COLUMNS = 7;

const GameBoard = ({ board, onColumnClick, currentTurn, disabled }) => {
  const handleColumnClick = (col) => {
    if (!disabled) {
      onColumnClick(col);
    }
  };

  const getCellClass = (cell) => {
    if (cell === 1) return 'player1';
    if (cell === 2) return 'player2';
    return 'empty';
  };

  return (
    <div className="game-board">
      <div className="columns">
        {[...Array(COLUMNS)].map((_, col) => (
          <div
            key={col}
            className={`column ${disabled ? 'disabled' : ''}`}
            onClick={() => handleColumnClick(col)}
          >
            <div className="column-hover">
              <div className={`disc ${currentTurn === 1 ? 'player1' : 'player2'}`} />
            </div>
          </div>
        ))}
      </div>
      
      <div className="board">
        {board.map((row, rowIndex) => (
          <div key={rowIndex} className="row">
            {row.map((cell, colIndex) => (
              <div key={colIndex} className={`cell ${getCellClass(cell)}`}>
                {cell !== 0 && <div className="disc" />}
              </div>
            ))}
          </div>
        ))}
      </div>
    </div>
  );
};

export default GameBoard;
