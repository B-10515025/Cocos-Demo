# ThirteenPoker Server API Document

Use websocket(SSL) protocol to connect, use json format string for communication.
Maximum 4 people per room, the game will start 5 seconds after the room is full, 30 seconds to set the cards, 15 seconds to show result the cards.

NTUST server: 140.118.157.18:2345

If you have a untrusted certificate problem, please link to [this](https://140.118.157.18:2345) and make it trusted.

Server program: [github.com/B-10515025/Cocos-Demo/tree/ThirteenPoker/server](https://github.com/B-10515025/Cocos-Demo/tree/ThirteenPoker/server)

[Client2Server](#Client2Server)
[Server2Client](#Server2Client)

## Client2Server

### Join Lobby
- Enter the game lobby
```

{
    "Type": "join.lobby",   // Request type
    "Name": "nickname"      // Player's nickname
}

```

### Exit Lobby
- Leave from the lobby
```

{
    "Type": "exit.lobby"    // Request type
}

```

### Join Room
- Enter the game room
```

{
    "Type": "join.room",    // Request type
    "Name": "roomname"      // Room's name
}

```

### Exit Room
- Leave from the room
```

{
    "Type": "exit.room"   // Request type
}

```

### Set Card
- Set player's card sequence
```

Request:
{
    "Type": "set.card",     // Request type
    "Cards": [0, 1, 2, ...] // Integer array of length 13, element is player's card index, array index 0~2: front set, 3~7: middle set, 8~12: back set
}

```

## Server2Client

### Location Message
- Player location information
```

{
    "Type": "LocationMsg"   // Response type
    "Location": 0           // Player's location, 0: Logout, 1: Lobby, 2: Room 
    "Name": "roomname"      // Location name
}

```

### Room Message
- All room information
```

{
    "Type": "RoomMsg"   // Response type
    "Rooms":    [
                    {
                        "Name": "roomname", // Room's name
                        "Players":  [
                                        "nickname1",
                                        "nickname2",
                                        "nickname3",
                                        "nickname4"
                                    ]       // String array of length 4, the name of the players in this room, empty means space
                    }, ...
                ]       // All existing rooms
}

```

### Card Message
- Player card information
```

{
    "Type": "CardMsg"   // Response type
    "Cards":    [
                    [0, 1, 2, ...],
                    [-1, ...],
                    [-1, ...],
                    [-1, ...]
                ]       // All player cards, 4-by-13 2-dimensional integer array, element is card index, -1 is means unknown
}

```