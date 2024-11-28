# GCP Permissions Checker

This project was originally a fork of [https://github.com/carlospolop/Bruteforce-GCP-Permissions](https://github.com/carlospolop/Bruteforce-GCP-Permissions).

This Go script checks your permissions over a specific Google Cloud Platform (GCP) project, folder, or organization by leveraging the IAM permissions reference and the Cloud Resource Manager API.

---

## Features

- Fetches a comprehensive list of GCP permissions dynamically from the Google Cloud IAM permissions reference.
- Supports checking permissions for a specified GCP project, folder, or organization.
- Parallel processing for faster results, with configurable thread and chunk size options.
- Verbose output to display permissions as they are found.

---

## Requirements

- Go 1.18 or later installed on your machine.
- A GCP service account credentials file (`credentials.json`) with sufficient permissions to access the Cloud Resource Manager API.
- Internet access to fetch permissions from the Google Cloud IAM reference.

---

## Installation

1. Clone the repository:
   ```bash
   git clone <repository_url>
   cd <repository_directory>
   ```

2. Build the executable:
   ```bash
   go build -o gcp_permissions gcp_permissions.go
   ```

---

## Usage

### Command-Line Flags

| Flag                | Description                                                                                  | Example                                    |
|---------------------|----------------------------------------------------------------------------------------------|--------------------------------------------|
| `-project`          | GCP project ID. Required if `-folder` or `-organization` is not provided.                   | `-project my-project-id`                   |
| `-folder`           | GCP folder ID. Mutually exclusive with `-project` and `-organization`.                      | `-folder 123456789012`                     |
| `-organization`     | GCP organization ID. Mutually exclusive with `-project` and `-folder`.                      | `-organization 123456789012`              |
| `-credentials`      | Path to the service account credentials JSON file. Required.                               | `-credentials ./path/to/credentials.json` |
| `-verbose`          | Enable verbose output to display found permissions as they are identified.                  | `-verbose`                                 |
| `-threads`          | Number of threads for parallel processing (default: 3).                                     | `-threads 5`                               |
| `-size`             | Size of permission chunks for parallel processing (default: 50).                            | `-size 100`                                |

### Examples

#### Check Permissions for a GCP Project
```bash
./gcp_permissions -project my-project-id -credentials ./credentials.json -verbose
```

#### Check Permissions for a GCP Folder with 5 Threads
```bash
./gcp_permissions -folder 123456789012 -credentials ./credentials.json -threads 5
```

#### Check Permissions for a GCP Organization with Custom Chunk Size
```bash
./gcp_permissions -organization 123456789012 -credentials ./credentials.json -size 100
```

---

## How It Works

1. The script fetches the list of all GCP permissions from the Google Cloud IAM permissions reference.
2. It divides the permissions into chunks for parallel testing.
3. Using the Cloud Resource Manager API, it checks which permissions the specified resource (project, folder, or organization) allows.
4. Results are displayed, showing the permissions you have.

---

## Notes

- Ensure the Cloud Resource Manager API is enabled in your GCP project before running the script.
- Be cautious with the number of threads (`-threads`) to avoid hitting API rate limits.
- Use the `-verbose` flag to monitor progress and see permissions as they are found.

---

