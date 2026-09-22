@echo off
echo Starting Multiplayer Ludo Server...
set PORT=8080
set JWT_SECRET=supersecretjwtkey
set DB_DSN=root:@tcp(127.0.0.1:3306)/ludo?parseTime=true
ludo_server.exe
pause
