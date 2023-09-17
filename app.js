'use strict';

const express = require('express');
const { createServer } = require('http');
const { join } = require('path');
const { Server } = require('socket.io');
require('dotenv').config();

const app = express();
const server = createServer(app);
const io = new Server(server);

app.get('/', (req, res) => {
    res.sendFile(join(__dirname, 'index.html'));
});

const users = {}; // To store users and their information

io.on('connection', (socket) => {

    // Handle user disconnect
    socket.on('disconnect', () => {
        const user = users[socket.id];
        if (user) {
            delete users[socket.id];
            io.emit('user_left', user);
            updateOnlineUsers();
        }
    });

    // Handle setting a nickname
    socket.on('set_user', (nickname) => {
        users[socket.id] = { id: socket.id, name: nickname, online: true };
        io.emit('user_joined', users[socket.id]);
        updateOnlineUsers();
    });

    // Handle chat messages
    socket.on('chatroom', (message) => {
        const sender = users[socket.id];
        if (sender) {
            // Don't send the message to the sender
            socket.broadcast.emit('chatroom', { sender, message });
            socket.emit('chatroom', { sender, message });
        }
    });

    // Update online users list
    function updateOnlineUsers() {
        io.emit('online_users', Object.values(users));
    }

});

server.listen(process.env.PORT, () => {
    console.log(`Server is running at http://${process.env.HOST}:${process.env.PORT}`);
});
