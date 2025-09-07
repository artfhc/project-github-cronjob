#!/bin/sh
set -eu

echo "=== DiscordChatExporter Container ==="
echo "DISCORD_TOKEN: [REDACTED]"
echo "GUILD_ID: ${GUILD_ID:-'not set'}"
echo "CHANNEL_ID: ${CHANNEL_ID:-'not set'}"
echo "SCOPE: ${SCOPE:-'guild'}"
echo "EXPORT_FORMAT: ${EXPORT_FORMAT:-'Json'}"
echo "ARCHIVE_URI: ${ARCHIVE_URI:-'not set'}"
echo "AFTER_TS: ${AFTER_TS:-'not set'}"
echo "BEFORE_TS: ${BEFORE_TS:-'not set'}"

# Configure rclone if RCLONE_CONFIG is provided
if [ -n "${RCLONE_CONFIG:-}" ]; then
    echo "Configuring rclone..."
    export HOME=/root
    export RCLONE_CONFIG_DIR=/root/.config/rclone
    mkdir -p /root/.config/rclone
    echo "$RCLONE_CONFIG" > /root/.config/rclone/rclone.conf
    
    echo "Rclone config file contents:"
    cat /root/.config/rclone/rclone.conf
    
    echo "Config file path check:"
    ls -la /root/.config/rclone/
    
    echo "Testing rclone configuration..."
    if ! rclone --config /root/.config/rclone/rclone.conf config show; then
        echo "Error: rclone config show failed - invalid configuration"
        exit 1
    fi
    
    if ! rclone --config /root/.config/rclone/rclone.conf listremotes; then
        echo "Error: rclone listremotes failed - no remotes configured"
        exit 1
    fi
    
    echo "✓ Rclone configuration validated successfully"
else
    echo "Warning: No RCLONE_CONFIG provided - uploads will fail"
fi

# Build DiscordChatExporter command based on scope
if [ "$SCOPE" = "guild" ]; then
    if [ -z "${GUILD_ID:-}" ]; then
        echo "Error: GUILD_ID required for guild scope"
        exit 1
    fi
    echo "Using exportguild command for guild: $GUILD_ID"
    DCE_CMD="./DiscordChatExporter.Cli exportguild"
    DCE_CMD="$DCE_CMD --token \"$DISCORD_TOKEN\""
    DCE_CMD="$DCE_CMD --format \"$EXPORT_FORMAT\""
    DCE_CMD="$DCE_CMD --output \"/output/\""
    DCE_CMD="$DCE_CMD --guild \"$GUILD_ID\""
    DCE_CMD="$DCE_CMD --include-threads"
elif [ "$SCOPE" = "channel" ]; then
    if [ -z "${CHANNEL_ID:-}" ]; then
        echo "Error: CHANNEL_ID required for channel scope"
        exit 1
    fi
    echo "Using export command for channel: $CHANNEL_ID"
    DCE_CMD="./DiscordChatExporter.Cli export"
    DCE_CMD="$DCE_CMD --token \"$DISCORD_TOKEN\""
    DCE_CMD="$DCE_CMD --format \"$EXPORT_FORMAT\""
    DCE_CMD="$DCE_CMD --output \"/output\""
    DCE_CMD="$DCE_CMD --channel \"$CHANNEL_ID\""
else
    echo "Error: Unknown scope: $SCOPE"
    exit 1
fi

# Add time range filters - convert ISO format to DCE expected format
if [ -n "${AFTER_TS:-}" ]; then
    # Convert from 2025-09-06T06:45:30Z to 2025-09-06 06:45:30
    AFTER_DCE=$(echo "$AFTER_TS" | sed 's/T/ /' | sed 's/Z$//')
    DCE_CMD="$DCE_CMD --after \"$AFTER_DCE\""
    echo "After timestamp (converted): $AFTER_DCE"
fi

if [ -n "${BEFORE_TS:-}" ]; then
    # Convert from 2025-09-06T06:45:30Z to 2025-09-06 06:45:30
    BEFORE_DCE=$(echo "$BEFORE_TS" | sed 's/T/ /' | sed 's/Z$//')
    DCE_CMD="$DCE_CMD --before \"$BEFORE_DCE\""
    echo "Before timestamp (converted): $BEFORE_DCE"
fi

# Check if DiscordChatExporter exists
cd /app
echo "Contents of /app:"
ls -la

if [ ! -f "DiscordChatExporter.Cli" ]; then
    echo "Error: DiscordChatExporter.Cli not found!"
    echo "Available files:"
    find . -name "*DiscordChatExporter*" -o -name "*Cli*"
    exit 1
fi

# Test executable
echo "Testing DiscordChatExporter.Cli..."
./DiscordChatExporter.Cli --version || echo "Warning: Could not get version"

echo "=== Checking available commands and options ==="
echo "Export command options:"
./DiscordChatExporter.Cli export --help 2>&1 | head -20

echo ""
echo "Exportguild command options:"
./DiscordChatExporter.Cli exportguild --help 2>&1 | head -20

# Run export
echo "Running: $DCE_CMD"
eval "$DCE_CMD"

# Upload to archive if configured
if [ -n "${ARCHIVE_URI:-}" ] && [ -n "${RCLONE_CONFIG:-}" ]; then
    # Create timestamped subdirectory
    TIMESTAMP=$(date -u +"%Y%m%d_%H%M%S")
    TIMESTAMPED_URI="${ARCHIVE_URI%/}/${TIMESTAMP}/"
    
    echo "Upload timestamp: $TIMESTAMP"
    echo "Uploading to: $TIMESTAMPED_URI"
    rclone --config /root/.config/rclone/rclone.conf copy /output "$TIMESTAMPED_URI" --progress
    echo "✓ Upload completed to timestamped directory: $TIMESTAMP"
else
    echo "Skipping upload - no archive URI or rclone config"
fi

echo "✓ Export process completed"