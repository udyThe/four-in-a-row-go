class WebSocketService {
  constructor() {
    this.ws = null;
    this.messageHandlers = {};
    this.isConnected = false;
    this.reconnectAttempts = 0;
    this.maxReconnectAttempts = 5;
    this.messageQueue = [];
  }

  connect(onOpen) {
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const wsUrl = process.env.REACT_APP_WS_URL || `${protocol}//${window.location.hostname}:8080/ws`;
    
    this.ws = new WebSocket(wsUrl);

    this.ws.onopen = () => {
      console.log('WebSocket connected');
      this.isConnected = true;
      this.reconnectAttempts = 0;
      
      // Flush any queued messages
      while (this.messageQueue.length > 0) {
        const { type, payload } = this.messageQueue.shift();
        this.send(type, payload);
      }
      
      if (onOpen) onOpen();
      this.startHeartbeat();
    };

    this.ws.onmessage = (event) => {
      // The server may batch multiple JSON messages in one frame, separated by newlines.
      const raw = typeof event.data === 'string' ? event.data : '';
      const chunks = raw.split('\n').filter((c) => c.trim().length > 0);

      for (const chunk of chunks) {
        try {
          const message = JSON.parse(chunk);
          console.log('Received message:', message);
          const handler = this.messageHandlers[message.type];
          if (handler) handler(message.payload);
        } catch (error) {
          console.error('Error parsing message chunk:', { error, chunkSample: chunk.slice(0, 200) });
        }
      }
    };

    this.ws.onerror = (error) => {
      console.error('WebSocket error:', error);
    };

    this.ws.onclose = () => {
      console.log('WebSocket disconnected');
      this.isConnected = false;
      this.stopHeartbeat();
      
      // Attempt reconnection
      if (this.reconnectAttempts < this.maxReconnectAttempts) {
        this.reconnectAttempts++;
        console.log(`Reconnecting... Attempt ${this.reconnectAttempts}`);
        setTimeout(() => this.connect(onOpen), 2000 * this.reconnectAttempts);
      }
    };
  }

  disconnect() {
    if (this.ws) {
      this.stopHeartbeat();
      this.ws.close();
      this.ws = null;
      this.isConnected = false;
    }
  }

  send(type, payload) {
    if (this.ws && this.ws.readyState === WebSocket.OPEN) {
      const message = JSON.stringify({ type, payload });
      this.ws.send(message);
      console.log('Sent message:', { type, payload });
    } else if (this.ws && this.ws.readyState === WebSocket.CONNECTING) {
      // Queue message to send once connection opens
      this.messageQueue.push({ type, payload });
      console.log('Queued message (connecting):', { type, payload });
    } else {
      console.warn('WebSocket not ready, state:', this.ws ? this.ws.readyState : 'null');
    }
  }

  on(messageType, handler) {
    this.messageHandlers[messageType] = handler;
  }

  off(messageType) {
    delete this.messageHandlers[messageType];
  }

  startHeartbeat() {
    // Clear any existing interval first
    this.stopHeartbeat();
    
    this.heartbeatInterval = setInterval(() => {
      if (this.ws && this.ws.readyState === WebSocket.OPEN) {
        this.send('heartbeat', {});
      }
    }, 10000); // Send heartbeat every 10 seconds
  }

  stopHeartbeat() {
    if (this.heartbeatInterval) {
      clearInterval(this.heartbeatInterval);
      this.heartbeatInterval = null;
    }
  }

  joinGame(username) {
    this.send('join', { username });
  }

  makeMove(column) {
    this.send('move', { column });
  }

  reconnect(playerId, gameId) {
    this.send('reconnect', { player_id: playerId, game_id: gameId });
  }
}

export default new WebSocketService();
