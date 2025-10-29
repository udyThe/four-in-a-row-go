import React from 'react';
import { BrowserRouter as Router, Routes, Route, Link } from 'react-router-dom';
import Game from './components/Game';
import Leaderboard from './components/Leaderboard';
import './App.css';

function App() {
  return (
    <Router>
      <div className="App">
        <nav className="navbar">
          <div className="nav-content">
            <Link to="/" className="nav-link">Play</Link>
            <Link to="/leaderboard" className="nav-link">Leaderboard</Link>
          </div>
        </nav>

        <Routes>
          <Route path="/" element={<Game />} />
          <Route path="/leaderboard" element={<Leaderboard />} />
        </Routes>

        <footer className="footer">
          <p>© 2025 4 in a Row | Built with React & Go</p>
        </footer>
      </div>
    </Router>
  );
}

export default App;
