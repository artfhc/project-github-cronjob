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
