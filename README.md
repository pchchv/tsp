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

## Usage

1. Run the script:
   ```sh
   go run main.go
   ```

2. The script will generate 3 files:
   - `index.html`: The main status page
   - `history.html`: The status history page
   - `history.json`: The status history and timestamp data

3. To keep the status page continuously updated, you can run the script in the background:
   - On Unix-like systems (Linux, macOS):
   Build:
     ```
     go build -ldflags="-w -s" .
     ```
   Now just run the go app as service.
   - On Windows, you can use the Task Scheduler to run the exe file at startup.

4. Serve the generated HTML files using HTTP server at specific PORT.

## Using Docker

In order to run the script using Docker:

   ```
    docker build -t tsp .
    docker run -ti --rm --name tsp -v "$PWD":/usr/src/myapp -w /usr/src/myapp tsp
   ```