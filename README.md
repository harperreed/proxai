# proxai 🐨

Welcome to the `proxai` repository! This project, created by [@harperreed](https://github.com/harperreed), provides a proxy server for the OpenAI API. 🚀

## Summary of Project 📜

`proxai` is a proxy server that simplifies interactions with the OpenAI API. Key features include forwarding requests, flexible configuration via command-line arguments, real-time status monitoring, and integration guidance via a `/help` endpoint.

## How to Use 🤖

To start using the proxy server, follow these steps:

1. **Run the Proxy Server:**

    Navigate to the directory where you've cloned the repository and run:

    ```sh
    ./proxai -port=9000 -address=0.0.0.0
    ```

    or if you prefer using Go:

    ```sh
    go run proxai.go -port=9000 -address=0.0.0.0
    ```

2. **Integration:**

   Use the proxy in your code with minor tweaks to your OpenAI client setup.

    ### Langchain

    ```python
    from langchain.chat_models.openai import ChatOpenAI

    openai = ChatOpenAI(
        model_name="your-model-name",
        openai_api_key="your-api-key",
        openai_api_base="http://<address>:<port>/v1",
    )
    ```

    ### OpenAI Python Client

    ```python
    from openai import OpenAI

    client = OpenAI(
        base_url="http://<address>:<port>/v1",
    )
    ```

3. **Command-line Arguments:**

    - `-port`: Port to listen on (default: 8080)
    - `-address`: Address to listen on (default: localhost)

    Start customizing the server by using these arguments as needed.

4. **Access Help:**

   Visit the `/help` endpoint to get more details and guidance while the proxy server is running.

## Tech Info ⚙️

This section provides technical details about the `proxai` repository, its file structure, and content.

### Directory/File Tree

```
proxai/
├── LICENSE
├── README.md
├── cache.go
├── go.mod
├── go.sum
├── handlers.go
├── logger.go
├── logs/
│   ├── costs.log
│   ├── prompts.log
│   ├── requests.log
│   └── responses.log
├── main.go
├── main_text.go
├── proxy.go
├── ui.go
└── utils.go
```

### Tech Stack

- **Programming Language:** Go
- **HTTP Client:** `net/http` package
- **Markdown Rendering:** `github.com/gomarkdown/markdown`
- **Logging:** `github.com/peterbourgon/diskv`
- **CLI:** `github.com/charmbracelet` suite for CLI and UI interactions

### File Summaries

- **`LICENSE`:** Contains the MIT License information.
- **`README.md`:** Documentation file (You're reading it!).
- **`cache.go`:** Implements caching mechanism using `diskv`.
- **`go.mod`:** Go module dependencies.
- **`go.sum`:** Checksums for Go module dependencies.
- **`handlers.go`:** HTTP handlers for proxy and help endpoints.
- **`logger.go`:** Logger implementation for request and response logs.
- **`logs/`:** Directory containing log files.
- **`main.go`:** Main entry point to start the proxy server.
- **`main_text.go`:** Test cases for the main components.
- **`proxy.go`:** Core proxy server implementation.
- **`ui.go`:** CLI UI for monitoring server status.
- **`utils.go`:** Utility functions.

## Building 🚀

1. **Clone the Repository:**

    ```sh
    git clone https://github.com/harperreed/proxai.git
    ```

2. **Navigate to the Project Directory:**

    ```sh
    cd proxai
    ```

3. **Build and Run the Proxy Server:**

    ```sh
    go run proxai.go -port=9000 -address=0.0.0.0
    ```

4. **Send Requests:**

    Point your requests to the proxy server endpoint, and it will handle the rest.

## Contributing 🤝

Contributions are welcome! If you find any issues or have suggestions for improvements, please open an issue or submit a pull request. Let's make this project even better together! 💪

## License 📜

This project is licensed under the [MIT License](https://github.com/harperreed/proxai/blob/main/LICENSE). Feel free to use, modify, and distribute the code as per the terms of the license.

Happy proxying! 🎉
