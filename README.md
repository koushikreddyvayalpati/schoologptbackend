# SchoolGPT Backend


> Production-ready AI education backend in Go — OpenAI function calling,
> voice interfaces, Firebase auth, and role-based access control.


## What Is This?


SchoolGPT is a backend API for an AI-powered education platform.
Students ask questions via text or voice. The backend routes them through
OpenAI's GPT with function calling to retrieve attendance, grades, and
course info — delivering intelligent, context-aware answers in real time.
## Architecture


┌─────────────┐    REST API    ┌──────────────────────────────┐
│  Student /  │ ────────────▶ │  Go Backend (Gin Framework)  │
│  Teacher UI │               │  ┌──────────┐ ┌───────────┐  │
└─────────────┘               │  │  Auth    │ │  Handlers │  │
                              │  │ Firebase │ │  Services │  │
                              │  └──────────┘ └───────────┘  │
                              └──────┬───────────────┬────────┘
                                     │               │
                              ┌──────▼──┐    ┌───────▼──────┐
                              │  OpenAI │    │ Google Cloud │
                              │   GPT   │    │ Speech API   │
                              └─────────┘    └──────────────┘

## Tech Stack


| Layer         | Technology                        |
|---------------|-----------------------------------|
| Language      | Go 1.21+                          |
| Framework     | Gin HTTP Framework                |
| AI / LLM      | OpenAI GPT-4 (Function Calling)   |
| Voice         | Google Cloud Speech-to-Text / TTS |
| Auth          | Firebase Authentication           |
| Database      | Google Firestore                  |
| DevOps        | Docker, docker-compose, Makefile  |
| Access        | Role-Based Access Control (RBAC)  |

## Key Features


- **GPT Function Calling** — AI selects the right data retrieval function
  based on the student's question (attendance, grades, schedule)
- **Voice Interface** — Students ask questions by voice; backend transcribes,
  processes, and responds in natural language
- **Firebase Auth + RBAC** — Students, teachers, and admins get different
  access levels enforced at middleware level
- **Production Structure** — Clean cmd/internal/pkg separation, Dockerfile,
  docker-compose, environment config, Makefile commands

## API Endpoints


| Method | Endpoint                         | Description                    |
|--------|----------------------------------|--------------------------------|
| POST   | /api/v1/ask                      | Ask GPT a question             |
| GET    | /api/v1/attendance/:student/:date| Get student attendance         |
| POST   | /api/v1/voice/transcribe         | Convert audio to text          |
| POST   | /api/v1/voice/synthesize         | Convert text to audio          |

## Quick Start


# 1. Clone
git clone https://github.com/koushikreddyvayalpati/schoologptbackend.git
cd schoologptbackend


# 2. Configure
cp env.example .env
# Fill in your OpenAI key, Firebase credentials, Google Cloud keys


# 3. Run with Docker
docker-compose up


# OR run locally
go run cmd/api/main.go

## Environment Variables


OPENAI_API_KEY=your_openai_key
FIREBASE_PROJECT_ID=your_project_id
GOOGLE_APPLICATION_CREDENTIALS=path/to/credentials.json
PORT=8080
ENVIRONMENT=development


See env.example for the complete list.
## Project Structure


schoolgpt/
├── cmd/api/          # Entrypoint — main.go
├── internal/
│   ├── auth/         # Firebase auth logic
│   ├── config/       # Environment & app config
│   ├── handlers/     # HTTP request handlers
│   ├── middleware/   # Auth, RBAC middleware
│   ├── models/       # Data models
│   ├── services/     # Core business logic
│   └── storage/      # Firestore interactions
├── pkg/
│   ├── gpt/          # OpenAI GPT + function calling
│   └── voice/        # Speech-to-Text / TTS
├── admin-portal/     # TypeScript admin UI
├── Dockerfile
├── docker-compose.yml
└── Makefile

## Built By
Koushik Reddy Vayalpati
MS Computer Science — University at Buffalo
AI & Machine Learning Specialization
linkedin.com/in/YOUR_LINKEDIN   |   github.com/koushikreddyvayalpati

