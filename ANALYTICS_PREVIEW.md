# Analytics Dashboard Preview

## Overview

The Analytics dashboard provides real-time insights into game activity with hourly and daily granularity.

## Dashboard Layout

```
┌─────────────────────────────────────────────────────────────────┐
│                    Game Analytics                               │
│                                                                 │
│  [Hourly (24h)]  [Daily (30d)]  ← Toggle Buttons              │
└─────────────────────────────────────────────────────────────────┘

┌─────────────┬─────────────┬─────────────┬─────────────┐
│   Total     │   Games     │   Total     │    Avg      │
│   Games     │  Completed  │   Moves     │  Duration   │
│  Started    │             │             │             │
│             │             │             │             │
│    150      │     142     │    2,450    │   145.2s    │
└─────────────┴─────────────┴─────────────┴─────────────┘
           Summary Cards (with gradient colors)

┌─────────────────────────────────────────────────────────────────┐
│                         Peak Activity                           │
│                                                                 │
│         Oct 30, 6:00 PM - 25 games started                     │
└─────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────┐
│                      Data Table                                 │
├──────────────┬────────┬──────────┬────────┬────────────────────┤
│ Hour         │ Started│ Completed│ Moves  │ Avg Duration       │
├──────────────┼────────┼──────────┼────────┼────────────────────┤
│ Oct 30, 6 PM │   25   │    24    │  412   │  152.3s           │
│ Oct 30, 5 PM │   18   │    16    │  289   │  148.7s           │
│ Oct 30, 4 PM │   15   │    14    │  245   │  141.2s           │
│ Oct 30, 3 PM │   12   │    11    │  198   │  139.8s           │
│ Oct 30, 2 PM │   20   │    19    │  356   │  145.6s           │
│     ...      │   ...  │    ...   │  ...   │   ...             │
└──────────────┴────────┴──────────┴────────┴────────────────────┘

┌─────────────────────────────────────────────────────────────────┐
│                  Games Activity Chart                           │
│                                                                 │
│  25 │                   ▄▄▄                                    │
│  20 │         ▄▄▄       ███       ▄▄▄                          │
│  15 │   ▄▄▄   ███ ▄▄▄   ███ ▄▄▄   ███                          │
│  10 │   ███   ███ ███   ███ ███   ███   ▄▄▄                   │
│   5 │   ███   ███ ███   ███ ███   ███   ███   ▄▄▄             │
│   0 └─────────────────────────────────────────────────────────┤
│       2PM  3PM  4PM  5PM  6PM  7PM  8PM  9PM 10PM             │
└─────────────────────────────────────────────────────────────────┘
```

## View Modes

### Hourly View (Last 24 Hours)
- **Granularity**: Per hour
- **Time Range**: Last 24 hours
- **Use Case**: Monitor recent activity, identify peak hours
- **Chart**: Shows hourly game counts
- **Example**: "Oct 30, 2:00 PM - 6:00 PM"

### Daily View (Last 30 Days)
- **Granularity**: Per day
- **Time Range**: Last 30 days
- **Use Case**: Track trends, measure growth
- **Chart**: Shows daily game counts
- **Example**: "Oct 1 - Oct 30"

## Color Scheme

### Summary Cards (Gradient Backgrounds)
1. **Total Games Started**: Purple gradient (667eea → 764ba2)
2. **Games Completed**: Pink gradient (f093fb → f5576c)
3. **Total Moves**: Blue gradient (4facfe → 00f2fe)
4. **Avg Duration**: Green gradient (43e97b → 38f9d7)

### Chart Bars
- **Default**: Blue gradient (3498db → 5dade2)
- **Hover**: Darker blue (2980b9 → 3498db)
- **Peak Hour**: Highlighted automatically

## Features

### Auto-Refresh
- Updates every 30 seconds automatically
- Shows loading indicator during refresh
- No page reload required

### Interactive Elements
- **Bar Chart**: Hover to see exact values
- **Toggle Buttons**: Switch between views instantly
- **Responsive**: Works on mobile and desktop

### Data Display
- **Empty State**: Friendly message when no data available
- **Error Handling**: Clear error messages with retry option
- **Loading State**: Smooth loading animations

## Sample Data Scenarios

### Low Activity Period
```
┌─────────────┬─────────────┬─────────────┬─────────────┐
│      5      │      4      │     82      │    95.3s    │
│   Games     │  Completed  │   Moves     │  Duration   │
└─────────────┴─────────────┴─────────────┴─────────────┘
```

### High Activity Period
```
┌─────────────┬─────────────┬─────────────┬─────────────┐
│     245     │     238     │   4,123     │   152.8s    │
│   Games     │  Completed  │   Moves     │  Duration   │
└─────────────┴─────────────┴─────────────┴─────────────┘
```

### Peak Time Alert
```
╔═════════════════════════════════════════════════════════╗
║                    🏆 Peak Activity                     ║
║                                                         ║
║        Today, 6:00 PM - 45 games started!              ║
╚═════════════════════════════════════════════════════════╝
```

## API Endpoints Used

### Frontend Calls
```javascript
// Hourly data
GET /api/analytics/hourly?hours=24

// Daily data  
GET /api/analytics/daily?days=30
```

### Response Format
```json
{
  "hour": "2025-10-30T18:00:00Z",
  "games_started": 25,
  "games_completed": 24,
  "total_moves": 412,
  "avg_game_duration": 152.3
}
```

## Navigation

Access from main navigation bar:
```
[ Play ]  [ Leaderboard ]  [ Analytics ]
                              ↑
                         Click here
```

## Mobile View

On mobile devices (< 768px):
- Summary cards stack vertically
- Table scrolls horizontally
- Chart shows fewer bars (better visibility)
- Toggle buttons stack if needed
- All data remains accessible

## Real-World Example

### Typical Hourly View (Evening)
```
Hour              Games    Moves    Duration
5:00 PM - 6:00 PM   12      203     145s
6:00 PM - 7:00 PM   25      412     152s  ← Peak Hour
7:00 PM - 8:00 PM   18      298     148s
8:00 PM - 9:00 PM   15      245     141s
9:00 PM - 10:00 PM  8       134     139s
```

### Typical Daily View (Week)
```
Date        Games    Moves    Duration
Oct 24        45      782      148s
Oct 25        52      891      151s
Oct 26        38      645      142s
Oct 27        67     1,143     155s  ← Busiest Day
Oct 28        41      701      146s
Oct 29        35      598      144s
Oct 30        58      987      149s
```

## Benefits

1. **For Players**: See when others are most active
2. **For Admins**: Identify optimal maintenance windows
3. **For Marketing**: Plan campaigns around peak times
4. **For Development**: Optimize server resources

## Testing the Feature

1. **Start Services**:
   ```bash
   docker-compose up
   ```

2. **Play Games**:
   - Open http://localhost:3000
   - Play 5-10 games
   - Make various moves

3. **View Analytics**:
   - Click "Analytics" in navigation
   - Wait for data to load
   - Toggle between hourly/daily views

4. **Verify Real-time Updates**:
   - Play more games
   - Wait 30 seconds (auto-refresh)
   - See counts increment

5. **Check Kafka Pipeline**:
   - Open http://localhost:8090
   - View "game-events" topic
   - See events correlating with analytics

## Summary

The Analytics dashboard provides:
- ✅ Real-time game activity monitoring
- ✅ Hourly and daily views
- ✅ Beautiful visualizations
- ✅ Auto-refresh functionality
- ✅ Mobile-responsive design
- ✅ Peak activity identification
- ✅ Comprehensive data tables
- ✅ Interactive charts

Perfect for understanding player behavior and optimizing the gaming experience!
