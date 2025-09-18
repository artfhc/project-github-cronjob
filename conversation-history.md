# Conversation History Fetcher

This document explains how to use the conversation history fetcher tool and GitHub workflow to bulk export Slack conversation data.

## Overview

The conversation history fetcher provides two ways to export Slack messages:

1. **Command Line Tool** - Direct Go binary for local execution
2. **GitHub Workflow** - Automated cloud execution with artifact downloads

Both methods bypass the MCP protocol and directly call the Slack API for maximum performance when handling large data exports.

## Command Line Tool

### Location
```
cmd/fetch-all-history/main.go
```

### Building
```bash
go build -o ./build/fetch-all-history ./cmd/fetch-all-history/main.go
```

### Usage
```bash
./build/fetch-all-history [options]
```

### Command Line Options

| Option | Description | Default | Example |
|--------|-------------|---------|---------|
| `-start` | Start date (YYYY-MM-DD) | No limit | `-start=2023-12-01` |
| `-end` | End date (YYYY-MM-DD) | No limit | `-end=2023-12-07` |
| `-limit` | Max messages per channel | 100 | `-limit=500` |
| `-output` | Output format | console | `-output=json` |
| `-channels` | Channel filter | all | `-channels=public` |

### Channel Filter Options

- **`all`** - Fetch from all channel types (public, private, group DMs, DMs)
- **`public`** - Public channels only
- **`private`** - Private channels only
- **`channel1,channel2`** - Specific channels (comma-separated)

### Output Formats

#### Console (Default)
Human-readable format for terminal viewing:
```
[2023-12-01 14:30:15] general (U123456): Hello everyone!
[2023-12-01 14:31:22] random (U789012): How's everyone doing?
```

#### JSON
Structured data format for programmatic processing:
```json
[
  {
    "channel": "general",
    "channel_id": "C123456",
    "user": "U123456",
    "text": "Hello everyone!",
    "timestamp": "1701434215.123456",
    "date": "2023-12-01T14:30:15Z"
  }
]
```

#### CSV
Spreadsheet-compatible format:
```csv
Channel,ChannelID,User,Timestamp,Date,Text
general,C123456,U123456,1701434215.123456,2023-12-01 14:30:15,"Hello everyone!"
```

### Usage Examples

#### Fetch Last Week's Messages (All Channels)
```bash
./build/fetch-all-history \
  -start=2023-12-01 \
  -end=2023-12-07 \
  -output=json \
  -limit=200
```

#### Export Specific Channels to CSV
```bash
./build/fetch-all-history \
  -channels="general,random,dev-team" \
  -output=csv \
  -limit=1000 > conversations.csv
```

#### Fetch Only Public Channels
```bash
./build/fetch-all-history \
  -channels=public \
  -start=2023-11-01 \
  -output=json
```

#### Quick Console Preview
```bash
./build/fetch-all-history \
  -channels="general" \
  -limit=10
```

## GitHub Workflow

### Location
```
.github/workflows/fetch-all-conversations.yaml
```

### Accessing the Workflow

1. Navigate to your repository on GitHub
2. Click the **Actions** tab
3. Select **"Fetch All Conversations History"** from the workflow list
4. Click **"Run workflow"** button

### Workflow Parameters

| Parameter | Description | Options | Default |
|-----------|-------------|---------|---------|
| **Start Date** | Beginning of date range | YYYY-MM-DD format | (optional) |
| **End Date** | End of date range | YYYY-MM-DD format | (optional) |
| **Limit Per Channel** | Max messages per channel | Number | 100 |
| **Output Format** | Export format | console, json, csv | console |
| **Channels Filter** | Channel types to include | all, public, private | all |
| **Specific Channels** | Override with channel names | Comma-separated list | (optional) |

### Workflow Features

#### Manual Trigger Only
The workflow only runs when manually triggered to prevent accidental bulk data exports that could:
- Consume API rate limits
- Generate large artifacts
- Incur unnecessary costs

#### Smart Parameter Handling
- **Date ranges** are optional - omit for no time filtering
- **Specific channels** override the channels filter when provided
- **Output formats** determine whether artifacts are created

#### Artifact Management
- **JSON/CSV outputs** are automatically saved as downloadable artifacts
- **Console output** is displayed in the workflow logs only
- **Filenames** include timestamp and parameters for easy identification
- **Retention** is set to 30 days to manage storage costs

#### Example Filename
```
conversations_20231201_143022_from_2023-11-01_to_2023-11-30_channels_public.json
```

### Sample Workflow Executions

#### Export All Data for Analysis
```yaml
Start Date: 2023-01-01
End Date: 2023-12-31
Limit Per Channel: 1000
Output Format: json
Channels Filter: all
```

#### Quick Channel Audit
```yaml
Limit Per Channel: 50
Output Format: console
Channels Filter: public
```

#### Specific Project Export
```yaml
Start Date: 2023-10-01
End Date: 2023-12-01
Limit Per Channel: 500
Output Format: csv
Specific Channels: project-alpha,project-beta,project-gamma
```

## Environment Variables

Both the command line tool and GitHub workflow require Slack API credentials:

### Required (Choose One Authentication Method)

**Browser Token Method (Stealth Mode):**
- `SLACK_MCP_XOXC_TOKEN` - Browser cookie token
- `SLACK_MCP_XOXD_TOKEN` - Browser session token

**OAuth Token Method:**
- `SLACK_MCP_XOXP_TOKEN` - OAuth bot token

### GitHub Secrets Setup

For the GitHub workflow, add these as repository secrets:

1. Go to **Settings** → **Secrets and variables** → **Actions**
2. Add the required tokens as secrets
3. The workflow will automatically use them

## Best Practices

### Rate Limiting
- Use reasonable limits per channel (100-500 messages)
- For large exports, consider running during off-peak hours
- The tool automatically handles API rate limits with built-in delays

### Date Ranges
- **Narrow date ranges** for faster execution and smaller files
- **Omit dates** to fetch all available history
- **End dates** are inclusive (adds 24 hours to include full day)

### Channel Selection
- **Start with public channels** to test your setup
- **Use specific channel lists** for targeted exports
- **Be mindful of private channel permissions**

### Output Format Selection
- **Console** for quick previews and testing
- **JSON** for programmatic processing and analysis
- **CSV** for spreadsheet analysis and reporting

### Large Workspace Considerations
- **Test with small limits first** to estimate data volume
- **Use date ranges** to break large exports into chunks
- **Monitor artifact sizes** as GitHub has storage limits

## Troubleshooting

### Common Issues

#### "No Slack credentials provided"
- Ensure environment variables are set correctly
- For GitHub workflows, check that secrets are properly configured

#### "Failed to fetch history for channel X"
- Channel may be private and bot lacks access
- Channel may be archived or deleted
- Rate limiting may be in effect (tool will retry automatically)

#### Large artifact files
- Reduce the limit per channel
- Use narrower date ranges
- Consider using console output for very large datasets

#### Empty results
- Check date range format (must be YYYY-MM-DD)
- Verify channel names are correct
- Ensure the bot has proper permissions

### Getting Help

1. Check the workflow logs for detailed error messages
2. Verify your Slack token permissions
3. Test with a small limit and single channel first
4. Review the [Authentication Setup Guide](01-authentication-setup.md)

## Security Considerations

### Data Handling
- **Exported data contains sensitive information** - handle artifacts securely
- **Artifacts are stored temporarily** - download and delete promptly
- **Console output may contain sensitive data** - be cautious with logs

### Access Control
- **Limit repository access** to authorized users only
- **Use least-privilege tokens** that only access necessary channels
- **Regularly rotate API tokens** following security best practices

### Compliance
- **Check data retention policies** before exporting
- **Ensure compliance** with your organization's data handling requirements
- **Consider data protection regulations** (GDPR, CCPA, etc.)