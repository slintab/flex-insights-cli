# Flex Insights CLI
CLI tool for exporting reports from Twilio Flex Insights.

## Setup
Set the following environment variables:
```
FLEX_INSIGHTS_USER=<YOUR_FLEX_INSIGHTS_USERNAME>
FLEX_INSIGHTS_PASSWORD=<YOUR_FLEX_INSIGHTS_PASSWORD>
```

Alternatively, you can also supply your username and password as command-line arguments when running the tool.

## Usage
1. `cd bin/`
2. Ensure the CLI utility is executable: `chmod +x flex-insights-cli`
3. Run the CLI utility: `./flex-insights-cli`

## Features
### Supported commands
- `export` - for exporting reports from Flex Insights. Note:  For more details, please see the [Flex Insights documentation](https://www.twilio.com/docs/flex/developer/insights/api/export-data).
    - Supported flags:
        - `-o, --objectid` object ID of the report (required)
        - `-w, --workspace` workspace ID of the report (required
        - `-f, --output` name of the file for saving the report (required)
        - `-u, --user` your Flex Insights username (optional if `FLEX_INSIGHTS_USER` environmental variable is set)
        - `-p, --password` your Flex Insights password (optional if `FLEX_INSIGHTS_PASSWORD` environmental variable is set)









