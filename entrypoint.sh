#!/bin/sh
set -eu

echo "=== DiscordChatExporter Container ==="
echo "DISCORD_TOKEN: ${DISCORD_TOKEN:0:10}..."
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
    mkdir -p ~/.config/rclone
    echo "$RCLONE_CONFIG" > ~/.config/rclone/rclone.conf
else
    echo "Warning: No RCLONE_CONFIG provided - uploads will fail"
fi

# Build DiscordChatExporter command
DCE_CMD="dotnet DiscordChatExporter.Cli.dll export"
DCE_CMD="$DCE_CMD --token \"$DISCORD_TOKEN\""
DCE_CMD="$DCE_CMD --format \"$EXPORT_FORMAT\""
DCE_CMD="$DCE_CMD --output \"/output\""

# Add guild or channel scope
if [ "$SCOPE" = "guild" ]; then
    if [ -z "${GUILD_ID:-}" ]; then
        echo "Error: GUILD_ID required for guild scope"
        exit 1
    fi
    DCE_CMD="$DCE_CMD --guild \"$GUILD_ID\""
elif [ "$SCOPE" = "channel" ]; then
    if [ -z "${CHANNEL_ID:-}" ]; then
        echo "Error: CHANNEL_ID required for channel scope"
        exit 1
    fi
    DCE_CMD="$DCE_CMD --channel \"$CHANNEL_ID\""
else
    echo "Error: Unknown scope: $SCOPE"
    exit 1
fi

# Add time range filters
if [ -n "${AFTER_TS:-}" ]; then
    DCE_CMD="$DCE_CMD --after \"$AFTER_TS\""
fi

if [ -n "${BEFORE_TS:-}" ]; then
    DCE_CMD="$DCE_CMD --before \"$BEFORE_TS\""
fi

# Check if DiscordChatExporter exists
cd /app
echo "Contents of /app:"
ls -la

if [ ! -f "DiscordChatExporter.Cli.dll" ]; then
    echo "Error: DiscordChatExporter.Cli.dll not found!"
    echo "Available files:"
    find . -name "*DiscordChatExporter*" -o -name "*Cli*"
    exit 1
fi

# Test dotnet runtime
echo "Testing .NET runtime..."
dotnet --version || echo "Warning: dotnet command not available"

# Run export
echo "Running: $DCE_CMD"
eval "$DCE_CMD"

# Upload to archive if configured
if [ -n "${ARCHIVE_URI:-}" ] && [ -n "${RCLONE_CONFIG:-}" ]; then
    echo "Uploading to: $ARCHIVE_URI"
    rclone copy /output "$ARCHIVE_URI" --progress
    echo "✓ Upload completed"
else
    echo "Skipping upload - no archive URI or rclone config"
fi

echo "✓ Export process completed"