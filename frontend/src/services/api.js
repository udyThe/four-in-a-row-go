import axios from 'axios';

const API_BASE_URL = process.env.REACT_APP_API_URL || 'http://13.204.212.197:8080/api';

const api = axios.create({
  baseURL: API_BASE_URL,
  timeout: 10000,
});

export const getLeaderboard = async (limit = 10) => {
  const response = await api.get(`/leaderboard?limit=${limit}`);
  return response.data;
};

export const getUserStats = async (username) => {
  const response = await api.get(`/user/${username}`);
  return response.data;
};

export const getRecentGames = async (limit = 20) => {
  const response = await api.get(`/games/recent?limit=${limit}`);
  return response.data;
};

export const getUserGames = async (username, limit = 20) => {
  const response = await api.get(`/games/user/${username}?limit=${limit}`);
  return response.data;
};

export const checkHealth = async () => {
  const response = await api.get('/health');
  return response.data;
};

export default api;
