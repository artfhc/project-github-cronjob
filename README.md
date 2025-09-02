# Multi-CSV Download to B2

This repository sets up a GitHub Actions workflow that downloads multiple CSV files and uploads them to Backblaze B2. The workflow is configuration-driven, allowing you to easily manage multiple data sources.

## 🎯 What it does

1. **Configuration-Driven**: Uses `download-config.json` to define multiple CSV downloads
2. **Parallel Processing**: Downloads multiple files simultaneously using matrix strategy  
3. **CSV Processing**: Optional newline fixing and validation for corrupted CSV files
4. **B2 Upload**: Uploads processed files to Backblaze B2 with date-based organization
5. **Flexible Control**: Each download can have different processing rules

## 📋 Configuration

### Configuration File Structure

Create `download-config.json` in the repository root:

```json
{
  "downloads": [
    {
      "name": "symphony-composer",
      "csv_url_secret": "SYMPHONY_CSV_URL",
      "b2_bucket_secret": "SYMPHONY_B2_BUCKET", 
      "output_prefix": "composer",
      "description": "Symphony Composer DB (Official) - daily download",
      "fix_csv_newlines": true,
      "validate_csv": true
    },
    {
      "name": "ICDB",
      "csv_url_secret": "ICDB_DATA_CSV_URL",
      "b2_bucket_secret": "ICDB_DATA_B2_BUCKET",
      "output_prefix": "icdb", 
      "description": "https://icdb.solarwolf.xyz/ - daily download",
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

### Configuration Schema

#### Required Fields per Download
- `name`: Unique identifier for the download (used in job names)
- `csv_url_secret`: Name of GitHub secret containing the CSV download URL
- `b2_bucket_secret`: Name of GitHub secret containing the B2 bucket name
- `output_prefix`: Prefix used for file naming and B2 path structure

#### Optional Fields per Download  
- `description`: Human-readable description of the download
- `fix_csv_newlines`: Enable/disable CSV newline fixing for this download (default: false)
- `validate_csv`: Enable/disable CSV validation for this download (default: false)

#### Global Settings
- `schedule`: Cron expression (informational only)
- `retention_days`: Future use for artifact retention

## 🔧 Setup

### 1. Required GitHub Secrets

#### Shared Secrets (required once)
- `B2_KEY_ID`: Backblaze B2 Key ID
- `B2_APP_KEY`: Backblaze B2 Application Key

#### Per-Download Secrets
For each download configuration, you need:
- The URL secret (referenced in `csv_url_secret`)
- The bucket secret (referenced in `b2_bucket_secret`)

**Example for the configuration above:**
- `SYMPHONY_CSV_URL`
- `SYMPHONY_B2_BUCKET`
- `ICDB_DATA_CSV_URL`
- `ICDB_DATA_B2_BUCKET`

### 2. Create Configuration File
Create `download-config.json` in your repository root with your specific downloads.

### 3. Test Configuration
1. Go to Actions → "Multi-CSV Download to B2"
2. Click "Run workflow"
3. Optionally specify a different config file path
4. Monitor the parallel job execution

## 📊 Example Configurations

### Single Download (Backward Compatible)
```json
{
  "downloads": [
    {
      "name": "composer",
      "csv_url_secret": "CSV_URL",
      "b2_bucket_secret": "B2_BUCKET",
      "output_prefix": "composer",
      "description": "Legacy composer data download",
      "fix_csv_newlines": true,
      "validate_csv": true
    }
  ]
}
```

### Multiple Downloads with Different Processing
```json
{
  "downloads": [
    {
      "name": "clean-data",
      "csv_url_secret": "CLEAN_DATA_URL",
      "b2_bucket_secret": "CLEAN_DATA_BUCKET",
      "output_prefix": "clean",
      "fix_csv_newlines": true,
      "validate_csv": true
    },
    {
      "name": "raw-export", 
      "csv_url_secret": "RAW_EXPORT_URL",
      "b2_bucket_secret": "RAW_EXPORT_BUCKET",
      "output_prefix": "raw",
      "fix_csv_newlines": false,
      "validate_csv": false
    }
  ]
}
```

## 📁 File Organization

Files are uploaded to B2 with this structure:
```
bucket-name/
  output-prefix/
    YYYY/
      MM/
        output-prefix_YYYY-MM-DD_HH-MM-SS.csv
```

**Examples:**
- Symphony data: `symphony-bucket/composer/2025/01/composer_2025-01-15_14-05-30.csv`
- ICDB data: `icdb-bucket/icdb/2025/01/icdb_2025-01-15_14-05-30.csv`

## 🛠️ Workflow Features

### CSV Processing
The workflow includes a sophisticated CSV fixer (`fix_csv.py`) that:
- Fixes newlines within quoted CSV fields
- Validates CSV structure after processing
- Compares record counts to detect data loss
- Handles escaped quotes and edge cases properly

### Parallel Processing
- Each download runs in its own job simultaneously
- Uses GitHub Actions matrix strategy for efficiency
- Individual downloads can fail without affecting others (`fail-fast: false`)

### Error Handling
- Validates configuration file exists and is valid JSON
- Checks that all referenced secrets exist
- Validates downloaded files are not empty
- Comprehensive logging for debugging

## 🏃 Running the Workflow

### Scheduled Runs
The workflow runs daily at 14:05 UTC by default (configured in the workflow file).

### Manual Runs
1. Go to the Actions tab
2. Select "Multi-CSV Download to B2"
3. Click "Run workflow"
4. Optionally specify a custom config file path

### Custom Configuration Files
You can use different configuration files by specifying the path when manually triggering the workflow.

## 🔍 Troubleshooting

### Common Issues

1. **"Configuration file not found"**
   - Verify `download-config.json` exists in repository root
   - Check file path if using custom config

2. **"No downloads configured"**
   - Ensure `downloads` array is not empty
   - Verify JSON syntax is valid

3. **"Secret not found" errors**
   - Verify all secret names in config match GitHub secrets
   - Check both URL and bucket secrets exist for each download

4. **Matrix job failures**
   - Check individual job logs for specific download issues
   - Verify CSV URLs are accessible and return valid data

### Debugging
- Each job logs detailed information about downloads, processing, and uploads
- Use the Actions tab to view logs for each parallel job
- Failed jobs will show specific error messages

## 🔄 Migration from Legacy Setup

If you have an existing single-download setup, create this configuration to maintain compatibility:

```json
{
  "downloads": [
    {
      "name": "legacy-download",
      "csv_url_secret": "CSV_URL", 
      "b2_bucket_secret": "B2_BUCKET",
      "output_prefix": "composer",
      "fix_csv_newlines": true,
      "validate_csv": true
    }
  ]
}
```

This maintains compatibility with your existing secrets while enabling the new configuration system.

## ✨ Benefits

✅ **Multiple Downloads**: Handle multiple CSV sources in parallel  
✅ **Flexible Configuration**: Easy to add/remove downloads without workflow changes  
✅ **Separate Buckets**: Each download can target different B2 buckets  
✅ **Custom Prefixes**: Organize files with different prefixes per download  
✅ **Granular Control**: Each download can have different processing rules  
✅ **Shared Credentials**: Reuse B2 credentials across all downloads  
✅ **Backward Compatible**: Existing setups work with minimal config changes  
✅ **Robust Processing**: Advanced CSV fixing with validation and error detection
