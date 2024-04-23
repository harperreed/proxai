Proxai - OpenAI API Proxy
===================

This is a proxy server for the OpenAI API.

## Usage:

Send requests to the proxy server endpoint, and it will forward them to the OpenAI API.

### Command-line Arguments:

*   `-port`: Port to listen on (default: {{.Port}})
*   `-address`: Address to listen on (default: {{.Address}})

### Example:

To start the proxy server on port 9000 and allow access from any IP address:

`./proxai -port=9000 -address=0.0.0.0`

or

`go run proxai.go -port=9000 -address=0.0.0.0`


## Code Changes

To use the proxy with python code, you need to make the following changes:

### Langchain

```
from langchain.chat_models.openai import ChatOpenAI

openai = ChatOpenAI(
    model_name="your-model-name",
    openai_api_key="your-api-key",
    openai_api_base="http://{{.Address}}:{{.Port}}/v1",

)
```

### OpenAI Python Client

```
from openai import OpenAI

client = OpenAI(
    # Or use the `OPENAI_BASE_URL` env var
    base_url="http://{{.Address}}:{{.Port}}/v1",
)
```
