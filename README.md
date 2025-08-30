# Project GitHub Cronjob

This repository sets up a GitHub Actions workflow that runs daily. The job downloads a CSV file and uploads it to Backblaze B2.

## What it does
1. Runs once per day using GitHub Actions cron.
2. Downloads the CSV file from ${CSV_URL}
3. Saves it with a timestamped filename.
4. Uploads the file into a Backblaze B2 bucket.

## Setup

### 1. Add repository secrets
Go to: Settings > Secrets and variables > Actions  
Add these secrets:
- `B2_KEY_ID` - Backblaze B2 key ID
- `B2_APP_KEY` - Backblaze B2 application key
- `CSV_URL` - URL to download the CSV from
- `CSV_OUTPUT_PREFIX` - Prefix for output files
- `B2_BUCKET` - Target B2 bucket name

### 2. Workflow file
The workflow is defined in:
```
.github/workflows/csv_to_b2.yml
```

It contains a `cron` schedule:
```
cron: '5 14 * * *'
```
This means it runs every day at 14:05 UTC.

### 3. Run manually
You can also trigger the job manually from the Actions tab by selecting "Run workflow".

## File storage
Files are uploaded to Backblaze B2 in a date-based folder structure:
```
${B2_BUCKET}/YYYY/MM/${CSV_OUTPUT_PREFIX}_YYYY-MM-DD_%H-%M-%S.csv
```

## Managing the workflow
- To change the time: edit the cron expression in the workflow file.
- To pause: comment out or remove the schedule block.
- To debug: check the Actions tab logs for each run.
