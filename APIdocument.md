# Minesweeper Server API Document

Use websocket(SSL) protocol to connect, use json format string for communication.
Server up to 64 connections, up to 64 players per game, up to 16 players alive at the same time.
If all the squares are clicked, or all players die and the current game player is full, the next game will start.
The game start time takes about 5 seconds, and each request takes about 20 milliseconds.

NTUST server: 140.118.157.18:1234

If you have a untrusted certificate problem, please link to [this](https://140.118.157.18:1234) and make it trusted.

Server program: [github.com/B-10515025/Cocos-Demo/tree/Minesweeper/server](https://github.com/B-10515025/Cocos-Demo/tree/Minesweeper/server)

[Get](#Get)  
[Action](#Action)  
[System](#System)  

## Get

### Game Board
- Get current game information
```

Request:
{
    "Type": "get.board"     // Request type
}

Response:
{
    "GID": 0,               // Game index
    "Height": 64,           // Number of game rows
    "Width": 64,            // Number of game columns
    "Playing": true,        // Game has started
    "Client": [-1, -1, ...] // Integer Array with (Height * Width) elements, -1 value is means unknown, other value is PlayerID * 10 + X, X is the number of mines around, if X is 9 it means the flag has been set up 
}

```

### Players
- Get current players information
```

Request:
{
    "Type": "get.players"   // Request type
}

Response:
[
    {
        "Alive": true,      // The player is Alive
        "Name": "nickname", // Player's nickname
        "Score": 0,         // Player's score
    }, ...
]   // All players in current game

```

### History
- Get player records for all games
```

Request:
{
    "Type": "get.history"   // Request type
}

Response:
[
    {
        "GID": 0,       // Game index
        "Players":  [
                        {
                            "Alive": true,      // The player is Alive
                            "Name": "nickname", // Player's nickname
                            "Score": 0,         // Player's score
                        }, ...
                    ]   // All players in that game
    }, ...
]   // All player records in server

```


## Action

### Join Game
- Join the current game
```

Request:
{
    "Type": "action.join"   // Request type
    "Name": "nickname"      // Player's nickname
}

Response:
{
    "Code": 0,  // Status code, 0: successed, 1: name already used, 2: current players is full, 3: game players is full, 4: already in the game
    "Pid": 0    // If successed, this is your player index
}

```

### Click
- Click on the current game
```

Request:
{
    "Type": "action.click" // Request type
    "Index": 0          // Element index of game array
    "Flag": false       // false: normal click, true: set up flag
}

Response:
{
    "Code": 0,  // Status code, 0: successed, 1: has been clicked, 2: player is dead, 3: have not join game, 4: game has not started
    "Score": 0  // -1: you died, 0: click failed, other: score obtained
}
```

## System

### New Game
- Force open the next game
```

Request:
{
    "Type": "system.nextgame"   // Request type
    "Name": "password"          // System command password
}

Response:
{
    "Code": 0,  // Status code, 0: successed, 1: failed
    "Msg": 0    // System message
}

```