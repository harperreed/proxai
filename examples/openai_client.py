from openai import OpenAI

client = OpenAI(
    base_url="http://localhost:8080/v1"
)


chat_completion = client.chat.completions.create(
    messages=[
        {
            "role": "user",
            "content": "Say this is a test",
        }
    ],
    model="gpt-3.5-turbo",
)

print(chat_completion.get("choices")[0].get("message").get("content"))
