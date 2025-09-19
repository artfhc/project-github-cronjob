# GitHub Cronjob Workflows

Automated data collection workflows using GitHub Actions.

## CSV Download Workflow

Downloads CSV files and uploads them to Backblaze B2.

### Configuration: `download-config.json`

```json
{
  "downloads": [
    {
      "name": "symphony-composer",
      "csv_url_secret": "SYMPHONY_CSV_URL",
      "b2_bucket_secret": "SYMPHONY_B2_BUCKET",
      "output_prefix": "composer",
      "fix_csv_newlines": true,
      "validate_csv": true
    }
  ]
}
```

### Workflow: `.github/workflows/csv_to_b2.yml`

- Reads configuration from `download-config.json`
- Downloads CSV files from configured URLs in parallel
- Processes CSV files (fixes newlines, validates structure)
- Uploads to B2 with path: `bucket/prefix/YYYY/MM/filename.csv`

## Discord Export Workflow

Exports Discord chats and uploads them to Backblaze B2.

### Configuration: `discord-export-config.yml`

```yaml
exports:
  - name: "guild-export"
    guild_id_secret: "GUILD_ID"
    discord_token_secret: "DISCORD_TOKEN"
    scope: "guild"
    archive_uri_secret: "ARCHIVE_URI"
    rclone_config_secret: "RCLONE_CONFIG"
    export_format: "Json"
    enabled: true
    time_range: "last_24_hours"

global_settings:
  docker_image_tag: "dce-job:latest"
  fail_fast: false
```

### Workflow: `.github/workflows/discord_export.yml`

- Uses DiscordChatExporter in Docker container
- Supports guild and channel scopes
- Time range options: everything, today, yesterday, last_24_hours, last_7_days, last_30_days
- Uploads to B2 with timestamped directories: `bucket/YYYYMMDD_HHMMSS/`
- Scheduled: Daily at 09:00 UTC

## Slack Conversation Workflow

Fetches Slack conversation history and uploads to Backblaze B2.

### Configuration: `slack-conversation-config.yaml`

```yaml
storage:
  type: "b2"
  b2:
    bucket_name: "${B2_SLACK_BUCKET_NAME}"
    application_key_id: "${B2_SLACK_KEY_ID}"
    application_key: "${B2_SLACK_APP_KEY}"
    path_prefix: "slack-conversations/"
    endpoint: "${B2_SLACK_ENDPOINT}"

slack:
  xoxc_token: "${SLACK_MCP_XOXC_TOKEN}"
  xoxd_token: "${SLACK_MCP_XOXD_TOKEN}"
  xoxp_token: "${SLACK_MCP_XOXP_TOKEN}"

defaults:
  limit_per_channel: 0      # 0 = unlimited messages
  output_format: "json"     # console, json, csv
  channels_filter: "all"    # all, public, private
  past_days: 7             # Fetch last N days (0 = all history)

file_naming:
  include_timestamp: true
  include_date_range: true
  include_channels: true
  prefix: "conversations"
```

### Workflow: `.github/workflows/fetch_slack_conversations.yaml`

- Built with Go application for efficient Slack API usage
- Supports multiple authentication methods (XOXP or XOXC+XOXD tokens)
- Fetches from all channels (public, private, DMs, group messages)
- Configurable date ranges and message limits
- Automatic B2 upload with organized file naming
- Scheduled: Weekly on Saturday at 2:00 AM UTC
- Manual trigger available with custom parameters

## GitHub Secrets

**B2 Storage:**
- `B2_KEY_ID`: Backblaze B2 Key ID
- `B2_APP_KEY`: Backblaze B2 Application Key

**CSV Downloads:**
- `SYMPHONY_CSV_URL`: CSV download URL
- `SYMPHONY_B2_BUCKET`: B2 bucket name

**Discord Exports:**
- `DISCORD_TOKEN`: Discord bot token
- `GUILD_ID`: Discord server ID
- `ARCHIVE_URI`: B2 destination (e.g. `b2:bucket/path`)
- `RCLONE_CONFIG`: Rclone configuration for B2

**Slack Conversations:**
- `SLACK_MCP_XOXC_TOKEN`: Slack session cookie token (session-based auth)
- `SLACK_MCP_XOXD_TOKEN`: Slack session token (session-based auth)
- `SLACK_MCP_XOXP_TOKEN`: Slack OAuth token (alternative to session-based)
- `B2_SLACK_KEY_ID`: B2 Application Key ID for Slack bucket
- `B2_SLACK_APP_KEY`: B2 Application Key for Slack bucket
- `B2_SLACK_BUCKET_NAME`: B2 bucket name for Slack data
- `B2_SLACK_ENDPOINT`: B2 S3-compatible endpoint (e.g. `https://s3.us-west-004.backblazeb2.com`)
