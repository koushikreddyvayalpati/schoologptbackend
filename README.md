# SchoolGPT Backend

A production-ready Go backend for an AI-powered education platform.

## Features

- Secure REST API using Gin framework
- Firebase Authentication and Firestore integration
- OpenAI GPT integration with function calling
- Google Cloud Speech-to-Text and Text-to-Speech integration
- Role-Based Access Control (RBAC)

## Project Structure

```
schoolgpt/
├── cmd/
│   └── api/              # Application entrypoint
├── internal/             # Private application code
│   ├── auth/             # Authentication logic
│   ├── config/           # Configuration management
│   ├── handlers/         # HTTP request handlers
│   ├── middleware/       # HTTP middleware
│   ├── models/           # Data models
│   ├── services/         # Business logic
│   └── storage/          # Database interactions
├── pkg/                  # Public libraries
│   ├── gpt/              # GPT integration
│   └── voice/            # Voice processing
└── scripts/              # Utility scripts
```

## Setup

1. Clone the repository
2. Copy `.env.example` to `.env` and fill in your configuration values
3. Run `go mod tidy` to install dependencies
4. Start the server with `go run cmd/api/main.go`

## API Endpoints

- `POST /api/v1/ask` - Ask a question to GPT
- `GET /api/v1/attendance/:student_id/:date` - Get attendance for a student
- `POST /api/v1/voice/transcribe` - Transcribe audio to text
- `POST /api/v1/voice/synthesize` - Synthesize text to audio

## Environment Variables

See `.env.example` for required environment variables.

## License

MIT # schoologptbackend
