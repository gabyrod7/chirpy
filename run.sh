#!/bin/bash 

PORT="${PORT:-8080}" # Use PORT env var, defaulting to 8080
LOG_FILE="/tmp/chirpy-run.log" # Store server logs in /tmp
TIMEOUT=30
SLEEP_TIME=0.25

kill_process_on_port() { # Define helper to free the target port
  local pids="" # Hold matching process IDs

  if command -v lsof >/dev/null 2>&1; then # Prefer lsof if installed
    pids=$(lsof -tiTCP:"$PORT" -sTCP:LISTEN 2>/dev/null || true) # Find listener PIDs on PORT
  elif command -v fuser >/dev/null 2>&1; then # Otherwise try fuser
    pids=$(fuser "$PORT"/tcp 2>/dev/null || true) # Find TCP users of PORT
  elif command -v ss >/dev/null 2>&1; then # Otherwise try ss
    pids=$(ss -ltnp "sport = :$PORT" 2>/dev/null | grep -oP 'pid=\K[0-9]+' | sort -u || true) # Extract listener PIDs
  fi # End port lookup choices

  if [ -n "$pids" ]; then # If any listener was found
    echo "Port $PORT is in use by PID(s): $pids. Closing..." # Report processes to kill
    kill $pids 2>/dev/null || true # Ask processes to stop
    sleep $SLEEP_TIME # Give them time to exit
    kill -9 $pids 2>/dev/null || true # Force-kill any remaining processes
  else # If no listener was found
    echo "Could not find a process listening on port $PORT." # Report no match
  fi # End PID check
} # End kill_process_on_port

start_server() { # Define helper to start the app
  : > "$LOG_FILE" # Clear previous log output
  timeout $TIMEOUT ./out >"$LOG_FILE" 2>&1 & # Run server for up to TIMEOUT in background
  local pid=$! # Save background server PID

  sleep $SLEEP_TIME # Give server time to start or fail

  if ! kill -0 "$pid" 2>/dev/null; then # Check if server already exited
    wait "$pid" 2>/dev/null || true # Reap exited process

    if grep -q "bind: address already in use" "$LOG_FILE"; then # Detect port conflict
      return 98 # Signal address-in-use error
    fi # End port-conflict check

    cat "$LOG_FILE" >&2 # Print startup error logs
    return 1 # Signal generic startup failure
  fi # End early-exit check

  echo "Started server with PID $pid" # Report successful start
  return 0 # Signal success
} # End start_server

go build -o out # Build the Go app into ./out

start_server # Try to start the server
status=$? # Save start result

if [ "$status" -eq 98 ]; then # If the port was already in use
  echo "Address already in use. Closing existing listener and retrying..." # Report retry plan
  kill_process_on_port # Stop current listener on PORT
  start_server # Retry server start
elif [ "$status" -ne 0 ]; then # If another startup error occurred
  exit "$status" # Exit with the same error code
fi # End status handling
