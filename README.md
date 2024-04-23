# proxai 🐨

Welcome to the `proxai` repository! This project, created by [@harperreed](https://github.com/harperreed), provides a proxy server for the OpenAI API. 🚀

## Features ✨

- Forwards requests to the OpenAI API 📡
- Supports various command-line arguments for configuration 🛠️
- Provides a helpful `/help` endpoint for integration guidance 📚
- Displays a real-time status bar with request and token usage statistics 📊
- Uses cool emoji to make the experience more fun! 😎

## Usage:

Send requests to the proxy server endpoint, and it will forward them to the OpenAI API.

### Command-line Arguments:

*   `-port`: Port to listen on (default: 8080, current: {{.Port}})
*   `-address`: Address to listen on (default: localhost, current: {{.Address}})

### Example:

To start the proxy server on port 9000 and allow access from any IP address:

`./proxai -port=9000 -address=0.0.0.0`

or

`go run proxai.go -port=9000 -address=0.0.0.0`


## Integration 🔌

To use the proxy with your code, make the following changes:

### Langchain
```python
from langchain.chat_models.openai import ChatOpenAI

openai = ChatOpenAI(
    model_name="your-model-name",
    openai_api_key="your-api-key",
    openai_api_base="http://{{.Address}}:{{.Port}}/v1",
)
```

### OpenAI Python Client
```python
from openai import OpenAI

client = OpenAI(
    base_url="http://{{.Address}}:{{.Port}}/v1",
)
```

For more details, visit the `/help` endpoint of the running proxy server. 📖

## Building 🚀

1. Clone the repository:
   ```
   git clone https://github.com/harperreed/proxai.git
   ```

2. Navigate to the project directory:
   ```
   cd proxai
   ```

3. Build and run the proxy server:
   ```
   go run proxai.go -port=9000 -address=0.0.0.0
   ```

   You can customize the port and address using the command-line arguments.

4. Send requests to the proxy server endpoint, and it will forward them to the OpenAI API. 📬


## Contributing 🤝

Contributions are welcome! If you find any issues or have suggestions for improvements, please open an issue or submit a pull request. Let's make this project even better together! 💪

## License 📜

This project is licensed under the [MIT License](https://github.com/harperreed/proxai/blob/main/LICENSE). Feel free to use, modify, and distribute the code as per the terms of the license.

Happy proxying! 🎉
