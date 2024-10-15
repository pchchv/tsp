# tsp

*tsp* is a simple, customizable status page generator that allows you to monitor the status of various services and display them on a clean, dark mode, responsive web page.

## Features

- Monitor HTTP endpoints, ping hosts, and check open ports
- Responsive design for both status page and history page
- Customizable service checks via YAML configuration
- Incident history tracking
- Automatic status updates at configurable intervals
- The generated HTML is only 5KB in size

## Configuration

1. Create a `.env` file in the project root and customize the variables.
2. Edit the `checks.yaml` file to add or modify the services you want to monitor
3. (Optional) Customize the `incidents.html` file to add any known incidents or maintenance schedules.
4. (Optional) Modify the `templateFile` and `historyTemplateFile` constant to customize the look and feel of your status pages.
