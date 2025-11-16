# GitHub Cronjob Workflows

Automated data collection workflows using GitHub Actions for CSV downloads, Discord exports, and Slack conversation archival to Backblaze B2 storage.

## Table of Contents

- [Workflows](#workflows)
  - [CSV Download Workflow](#csv-download-workflow)
  - [Discord Export Workflow](#discord-export-workflow)
  - [Slack Conversation Workflow](#slack-conversation-workflow)
- [Setup & Configuration](#setup--configuration)
- [GitHub Secrets Reference](#github-secrets-reference)
- [Usage](#usage)

## Workflows

### CSV Download Workflow

Downloads CSV files from configured URLs and uploads them to Backblaze B2 with automatic processing.

**Workflow File:** `.github/workflows/csv_to_b2.yml`

**Features:**
- Parallel downloads from multiple configured URLs
- CSV validation and newline fixing
- Organized B2 uploads: `bucket/prefix/YYYY/MM/filename.csv`

**Configuration:** `download-config.json`

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

---

### Discord Export Workflow

Exports Discord server chats using DiscordChatExporter and archives them to Backblaze B2.

**Workflow File:** `.github/workflows/discord_export.yml`

**Features:**
- Docker-based DiscordChatExporter
- Guild and channel scope support
- Flexible time range options
- Timestamped B2 directories: `bucket/YYYYMMDD_HHMMSS/`

**Schedule:** Daily at 09:00 UTC

**Configuration:** `discord-export-config.yml`

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

**Time Range Options:**
- `everything` - Full history
- `today` - Today's messages
- `yesterday` - Previous day
- `last_24_hours` - Last 24 hours
- `last_7_days` - Last week
- `last_30_days` - Last month

---

### Slack Conversation Workflow

Fetches Slack conversation history from all channels and uploads to Backblaze B2 using a Go-based application.

**Workflow File:** `.github/workflows/fetch_slack_conversations.yaml`

**Features:**
- Efficient Go-based Slack API client
- Multiple authentication methods (XOXP or XOXC+XOXD tokens)
- All channel types: public, private, DMs, group messages
- Configurable date ranges and message limits
- Smart file naming with timestamps and date ranges

**Schedule:** Weekly on Saturday at 2:00 AM UTC (manual trigger available)

**Configuration:** `slack-conversation-config.yaml`

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

## Setup & Configuration

### 1. Configure Workflows

Create or edit the configuration files for each workflow you want to use:
- `download-config.json` - CSV download settings
- `discord-export-config.yml` - Discord export settings
- `slack-conversation-config.yaml` - Slack conversation settings

### 2. Set Up GitHub Secrets

Add the required secrets to your GitHub repository:
**Settings → Secrets and variables → Actions → New repository secret**

See [GitHub Secrets Reference](#github-secrets-reference) for the complete list.

### 3. Enable Workflows

Workflows are located in `.github/workflows/` and will run automatically based on their schedules, or can be triggered manually via GitHub Actions.

## GitHub Secrets Reference

### B2 Storage (Common)

Required for all workflows that upload to Backblaze B2:

| Secret | Description |
|--------|-------------|
| `B2_KEY_ID` | Backblaze B2 Application Key ID |
| `B2_APP_KEY` | Backblaze B2 Application Key |

### CSV Download Workflow

| Secret | Description |
|--------|-------------|
| `SYMPHONY_CSV_URL` | CSV file download URL |
| `SYMPHONY_B2_BUCKET` | B2 bucket name for CSV storage |

### Discord Export Workflow

| Secret | Description |
|--------|-------------|
| `DISCORD_TOKEN` | Discord bot token |
| `GUILD_ID` | Discord server/guild ID |
| `ARCHIVE_URI` | B2 destination path (e.g., `b2:bucket/path`) |
| `RCLONE_CONFIG` | Rclone configuration for B2 access |

### Slack Conversation Workflow

| Secret | Description |
|--------|-------------|
| `SLACK_MCP_XOXC_TOKEN` | Slack session cookie token (session-based auth) |
| `SLACK_MCP_XOXD_TOKEN` | Slack session token (session-based auth) |
| `SLACK_MCP_XOXP_TOKEN` | Slack OAuth token (alternative auth method) |
| `B2_SLACK_KEY_ID` | B2 Application Key ID for Slack bucket |
| `B2_SLACK_APP_KEY` | B2 Application Key for Slack bucket |
| `B2_SLACK_BUCKET_NAME` | B2 bucket name for Slack data |
| `B2_SLACK_ENDPOINT` | B2 S3-compatible endpoint (e.g., `https://s3.us-west-004.backblazeb2.com`) |

## Usage

### Manual Workflow Triggers

To manually trigger a workflow:

1. Go to **Actions** tab in your GitHub repository
2. Select the workflow you want to run
3. Click **Run workflow**
4. Configure parameters (if available)
5. Click **Run workflow** button

### Monitoring

- View workflow runs in the **Actions** tab
- Check logs for detailed execution information
- Failed workflows will send notifications based on repository settings

### Scheduled Runs

- **CSV Download:** As configured in workflow file
- **Discord Export:** Daily at 09:00 UTC
- **Slack Conversations:** Weekly on Saturday at 2:00 AM UTC
