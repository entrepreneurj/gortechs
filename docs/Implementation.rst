

Client Perspective
1. Opens connection to server
2. Registers channel
3. If channel is present subscribes to channel and waits for data otherwise closes
4. If receives EOC message, disconnects.

Server Perspective
1. Receives connection from client
2. Tries to register client on channel
3. If channel exists, subscribes client and publishes data from channel, otherwise sends signal to close websocket
4. If event ends, unsubscribe client from channel and send EOC
