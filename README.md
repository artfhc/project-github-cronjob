# Multi-CSV Download to B2

GitHub Actions workflow that downloads multiple CSV files and uploads them to Backblaze B2.

## What it does

1. Reads configuration from `download-config.json`
2. Downloads CSV files from configured URLs in parallel
3. Optionally processes CSV files (fixes newlines, validates structure)
4. Uploads files to Backblaze B2 with date-based organization

## Configuration

### `download-config.json`

Defines what CSV files to download and how to process them:

```json
{
  "downloads": [
    {
      "name": "symphony-composer",
      "csv_url_secret": "SYMPHONY_CSV_URL",
      "b2_bucket_secret": "SYMPHONY_B2_BUCKET",
      "output_prefix": "composer",
      "description": "Symphony Composer DB - daily download",
      "fix_csv_newlines": true,
      "validate_csv": true
    }
  ],
  "global_settings": {
    "schedule": "5 14 * * *",
    "retention_days": 30
  }
}
```

**Required per download:**
- `name`: Unique identifier
- `csv_url_secret`: GitHub secret name containing the download URL
- `b2_bucket_secret`: GitHub secret name containing the B2 bucket name
- `output_prefix`: File prefix for naming and B2 paths

**Optional per download:**
- `description`: Human-readable description
- `fix_csv_newlines`: Fix newlines in quoted CSV fields (default: false)
- `validate_csv`: Validate CSV structure after processing (default: false)

### `.github/workflows/csv_to_b2.yml`

GitHub Actions workflow that:
1. **load-config job**: Reads `download-config.json` and creates matrix strategy
2. **fetch-process-upload job**: Runs in parallel for each configured download
   - Downloads CSV from URL
   - Processes CSV if enabled (fixes newlines, validates structure) 
   - Uploads to B2 bucket with path: `bucket/prefix/YYYY/MM/filename.csv`

## Setup

### 1. GitHub Secrets

**Shared secrets:**
- `B2_KEY_ID`: Backblaze B2 Key ID
- `B2_APP_KEY`: Backblaze B2 Application Key

**Per-download secrets (example):**
- `SYMPHONY_CSV_URL`: URL to download CSV from
- `SYMPHONY_B2_BUCKET`: Target B2 bucket name

### 2. Run Workflow

**Scheduled:** Daily at 14:05 UTC (configured in workflow file)

**Manual:** Actions tab → "Multi-CSV Download to B2" → "Run workflow"
