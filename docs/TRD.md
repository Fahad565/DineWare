
Technical Requirements Document (TRD)


1. System Architecture Overview

DineWare follows a unified backend design pattern built using a robust, highly concurrent language (such as Go). It interfaces simultaneously with a lightweight relational storage layer (e.g., Supabase PostgreSQL) and the external Africa's Talking API communications gateway.
┌────────────────────────┐
│  Tour Operator Portal  │
└───────────┬────────────┘
            │ HTTPS / Webhook
            ▼
┌────────────────────────┐       Data Sync      ┌──────────────────────┐
│    DineWare Engine     ├─────────────────────►│ Neon PostgreSQL  │
│       (Golang)         │◄─────────────────────┤   (Data & States)    │
└─────┬────────────▲─────┘                      └──────────────────────┘
      │            │ 
      │ SMS Out    │ USSD Inbound HTTP POST
      ▼            │
┌──────────────────┴─────┐
│ Africa's Talking APIs  │
└────────────────────────┘

2. Database Schema (PostgreSQL)

sql
CREATE TABLE itineraries (
    id VARCHAR(50) PRIMARY KEY,
    group_name VARCHAR(100) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE itinerary_stops (
    id SERIAL PRIMARY KEY,
    itinerary_id VARCHAR(50) REFERENCES itineraries(id) ON DELETE CASCADE,
    lodge_name VARCHAR(100) NOT NULL,
    chef_phone VARCHAR(20) NOT NULL,
    check_in_date DATE NOT NULL,
    notification_sent BOOLEAN DEFAULT FALSE,
    chef_confirmed BOOLEAN DEFAULT FALSE
);

CREATE TABLE guests (
    id SERIAL PRIMARY KEY,
    itinerary_id VARCHAR(50) REFERENCES itineraries(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    allergies TEXT NOT NULL
);
Use code with caution.

3. API Integration Details (Africa's Talking)


A. Outbound Staggered SMS (Cron / Worker Loop)

• Endpoint: https://africastalking.com
• Method: POST
• Headers:
	• ApiKey: [AT_API_KEY]
	• Accept: application/json
	• Content-Type: application/x-www-form-urlencoded
• Payload Params:
	• username: sandbox (or production username)
	• to: +2547XXXXXXXX (Chef's Phone)
	• message: "DineWare Alert: Group {group_name} (ID: {id}) arrives {check_in_date}. Dial *384*44# to confirm your kitchen manifest."

B. Inbound USSD Webhook Handler

When a chef dials the shortcode, Africa's Talking sends an HTTP POST to the DineWare server.
• Expected Request Body (Form URL Encoded):sessionId=ATId_123456...&serviceCode=*384*44#&phoneNumber=+2547XXXXXXXX&text=
• Engine Response Format: Must reply with pure plain text prefixed with CON (continue session) or END (terminate session).

4. Core Logic Implementation Flow (USSD & SMS Router)

The backend handles the multi-lodge timing logic by executing a time-delta routing evaluation, alongside parsing the multi-level USSD state inputs.
go
package main

import (
	"fmt"
	"strings"
	"time"
)

// DestinationFulfillment models an individual lodge stop
type DestinationFulfillment struct {
	ID               int       `json:"id"`
	LodgeName        string    `json:"lodge_name"`
	ChefPhone        string    `json:"chef_phone"`
	CheckInDate      time.Time `json:"check_in_date"`
	NotificationSent bool      `json:"notification_sent"`
}

// HandleUSSDRequest processes incoming menu paths from Africa's Talking webhook
func HandleUSSDRequest(text, phoneNumber string, stops []DestinationFulfillment) string {
	inputParts := strings.Split(text, "*")
	
	// Root Menu
	if text == "" {
		return "CON Welcome to DineWare Kitchen Portal\nEnter Itinerary Booking Code:"
	}

	// Level 1: User typed the Itinerary Code
	if len(inputParts) == 1 {
		itineraryCode := inputParts[0]
		// Under the hood: Query Database for matching active stop matching itineraryCode & phoneNumber
		return fmt.Sprintf("CON Manifest Found for Code %s\n1. View Allergy Specs\n2. Confirm Manifest Receipt", itineraryCode)
	}

	// Level 2: Deep Menu Routing
	if len(inputParts) == 2 {
		action := inputParts[1]
		if action == "1" {
			// Mocking allergy data fetch from Postgres guests table
			return "END Severe Allergies Alert:\n- Guest 1: Peanut Allergy (Anaphylaxis)\n- Guest 2: Gluten Intolerance\nPlease adjust ordering!"
		}
		if action == "2" {
			// Mocking state flip: UPDATE itinerary_stops SET chef_confirmed = true...
			return "END Thank you. Manifest confirmation logged. Safe travels for your guests!"
		}
	}

	return "END Invalid Selection. Please try dialing again."
}

To view interactive API documentation in Golang similar to FastAPI or Spring Boot's Swagger UI, the industry standard is to use swag to parse your code comments and generate the OpenAPI specs, along with an HTTP framework wrapper (like net/http, gin, or chi) to host the UI page.
Since you are running this in GitHub Codespaces, this setup will give you a fully interactive web page at http://localhost:8080/swagger/index.html that automatically maps your endpoints.
Here is the exact step-by-step setup to implement it:

1. Install the Tooling

Inside your GitHub Codespace terminal, run the following commands to install the Swagger CLI tool and the net/http middleware handler:
bash
# Install the Swag CLI compiler tool
go install ://github.com

# Install the HTTP middleware handler (Assuming you are using standard net/http)
go get ://github.com
Use code with caution.

2. Document Your Code with Declarative Comments

You write normal Go code, but you add special // @... comments above your main() function and your individual API handlers. The engine uses these to build the documentation.

main.go

go
package main

import (
	"net/http"
	"log"

	// Crucial: Import the generated docs folder so it initializes
	_ "dineware-backend/docs" 
	"://github.com"
)

// @title           DineWare API Engine
// @version         1.0
// @description     Backend workflow manager handling staggered Africa's Talking dispatches.
// @host            localhost:8080
// @BasePath        /api/v1
func main() {
	mux := http.NewServeMux()

	// 1. Core business routes
	mux.HandleFunc("/api/v1/itinerary", HandleCreateItinerary)

	// 2. The Swagger UI Route!
	// This automatically loads the interactive Swagger UI page at /swagger/index.html
	mux.Handle("/swagger/", httpSwagger.WrapHandler)

	log.Println("Server running on :8080... Access docs at http://localhost:8080/swagger/index.html")
	log.Fatal(http.ListenAndServe(":8080", mux))
}

// HandleCreateItinerary handles booking entries
// @Summary      Create a new safari itinerary
// @Description  Allows tour operators to register groups, lodges, and encrypted allergy references.
// @Tags         itineraries
// @Accept       json
// @Produce      json
// @Param        itinerary  body      string  true  "Itinerary JSON payload"
// @Success      201  {string}  string  "Itinerary created successfully"
// @Router       /itinerary [post]
func HandleCreateItinerary(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(`{"status": "success", "message": "Itinerary registered"}`))
}
Use code with caution.

3. Generate the Documentation Files

Every time you add a new endpoint or update your comments, run this single command in the backend directory root:
bash
swag init
Use code with caution.
This command parses all your comments and dynamically generates a docs/ folder in your project root containing docs.go, swagger.json, and swagger.yaml.

4. Viewing it Live in GitHub Codespaces

1. Run your application: go run main.go.
2. GitHub Codespaces will detect that port 8080 is open and give you a toast notification to "Open in Browser".
3. Click it and navigate to your paths at /swagger/index.html.
4. You will see the beautiful, interactive Swagger interface where you can hit "Try it out" to test live API calls against your local server!
Take your time working through it this week! When you're ready to get back to configuring the repository foundations, let me know if you would like me to:
• Generate the complete sprints/ markdown task files to populate your newly defined documentation folder structure.
• Set up the base PostgreSQL connection module in Go tailored to interface securely with your Neon database instance.
Try without personalization
i'll be doing agile development which means i'll be heavily testing straight on ci/cd, render and on docker
20:47
That target loop is perfect for rapid development. Testing builds continuously straight through a pipeline onto container environments ensures you won't face any "works on my machine" bottlenecks when the final hackathon deadline hits.
To support this continuous delivery workflow, here are the production-ready Docker configurations and the orchestration  that you can drop straight into your workspace.

🐳 1. Backend Dockerfile (backend.Dockerfile)

This implements a secure, multi-stage build optimized for Go 1.22. It compiles a lean binary inside a lightweight Alpine build layer and copies it into a minimal scratch layer to keep your Render build times extremely low.
dockerfile
# --- Build Stage ---
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Copy dependency tracking files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code and documentation manifests
COPY . .

# Run static compilation without dynamic linking hooks
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o dineware-engine .

# --- Final Execution Stage ---
FROM alpine:3.19

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

# Copy the lean runtime binary from the build layer
COPY --from=builder /app/dineware-engine .
COPY --from=builder /app/docs ./docs

EXPOSE 8080

CMD ["./dineware-engine"]
Use code with caution.

🌐 2. Frontend Dockerfile (frontend.Dockerfile)

A streamlined multi-stage configuration explicitly structured for Next.js to compile static assets correctly ahead of deployment routing.
dockerfile
# --- Build Stage ---
FROM node:20-alpine AS builder
WORKDIR /app

COPY package*.json ./
RUN npm ci

COPY . .
RUN npm run build

# --- Runtime Stage ---
FROM node:20-alpine AS runner
WORKDIR /app

ENV NODE_ENV=production

COPY --from=builder /app/public ./public
COPY --from=builder /app/.next ./.next
COPY --from=builder /app/node_modules ./node_modules
COPY --from=builder /app/package.json ./package.json

EXPOSE 3000

CMD ["npm", "start"]
Use code with caution.

🗺️ 3. Render Blueprint Infrastructure (render.yaml)

Place this at the exact root of your repository. It instructs Render to build directly from your isolated application folders using the specialized container settings you just configured.
yaml
services:
  # -------------------------------------------------------------
  # The DineWare API Core Engine (Golang v1.22)
  # -------------------------------------------------------------
  - type: web
    name: dineware-backend-engine
    runtime: docker
    dockerContext: ./backend
    dockerfilePath: ./backend.Dockerfile
    plan: free
    envVars:
      - key: PORT
        value: 8080
      - key: DATABASE_URL
        sync: false # Connect your private Neon Postgres connection string here
      - key: AT_USERNAME
        sync: false
      - key: AT_API_KEY
        sync: false

  # -------------------------------------------------------------
  # The Tour Operator Control Panel (Next.js Application)
  # -------------------------------------------------------------
  - type: web
    name: dineware-operator-portal
    runtime: docker
    dockerContext: ./frontend
    dockerfilePath: ./frontend.Dockerfile
    plan: free
    envVars:
      - key: PORT
        value: 3000
      - key: NEXT_PUBLIC_API_URL
        fromService:
          name: dineware-backend-engine
          value: url



Here is the complete GitHub Actions configuration to manage your continuous integration loop.
Since you are tracking deployment natively using the render.yaml Blueprint file, your GitHub workflow simply acts as a gatekeeper. It aggressively lints your code, generates the Swagger API files to prevent out-of-date specs, and executes tests across your components on every pull request or push to the main branch.
Once this CI pipeline passes successfully, Render detects the push to your repository and automatically orchestrates the container builds using the paths defined in your render.yaml blueprint.

🛠️ The CI Configuration (.github/workflows/ci.yml)

Create a file at .github/workflows/ci.yml in your repository root and add the following contents:
yaml
name: DineWare Continuous Integration

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main ]

jobs:
  # -------------------------------------------------------------
  # 1. Backend Verification Job (Golang 1.22)
  # -------------------------------------------------------------
  backend-ci:
    name: Go Core Engine Validation
    runs-on: ubuntu-latest
    steps:
      - name: Checkout Source Code
        uses: actions/checkout@v4

      - name: Set up Go 1.22 Environment
        uses: actions/setup-go@v5
        with:
          go-version: '1.22'
          cache-dependency-path: backend/go.sum

      - name: Install Tooling Dependencies
        run: |
          cd backend
          go install ://github.com
          go mod download

      - name: Verify Swagger Documentation Generation
        run: |
          cd backend
          swag init
          # If git detects changes here, someone forgot to run swag init locally before pushing
          git diff --exit-code docs/

      - name: Run Backend Code Linter & Format Check
        run: |
          cd backend
          go fmt ./...
          git diff --exit-code

      - name: Execute Security & Test Suites
        run: |
          cd backend
          go test -v -race -covermode=atomic ./...

  # -------------------------------------------------------------
  # 2. Frontend Verification Job (Next.js / Node 20)
  # -------------------------------------------------------------
  frontend-ci:
    name: Next.js Frontend Validation
    runs-on: ubuntu-latest
    steps:
      - name: Checkout Source Code
        uses: actions/checkout@v4

      - name: Set up Node.js Environment
        uses: actions/setup-node@v4
        with:
          node-version: '20'
          cache: 'npm'
          cache-dependency-path: frontend/package-lock.json

      - name: Install Node Modules
        run: |
          cd frontend
          npm ci

      - name: Run Code Lints and Formatting Tests
        run: |
          cd frontend
          npm run lint

      - name: Execute Next.js Dry Compile Build
        run: |
          cd frontend
          # Verifies everything builds clean without deployment runtime environment issues
          npm run build
Use code with caution.

🔄 How Your Development Loop Flows in Practice

1. Write Code: You build features inside GitHub Codespaces using Google Jules.
2. Commit & Push: You issue a git push to your repository.
3. CI Gatekeeper Execution: GitHub Actions immediately triggers .github/workflows/ci.yml. It runs your Go unit tests, verifies that swag init matches your code changes, and checks the Next.js static asset build parameters.
4. Render Auto-Deployment: Once the GitHub Actions status check shows green, Render picks up the clean commit. It reads your root render.yaml configuration, accesses the custom multi-stage Docker contexts, and pushes the updated live environments automatically.


🛠️ Adding Docker Compose Targets to Your Makefile

You can append the following targets to the root Makefile we set up earlier to control your local container orchestration:
makefile
.PHONY: up up-build down logs restart ps

# 1. Start the entire stack in the background (detached mode)
up:
	docker compose up -d

# 2. Rebuild images and start the stack (use after modifying Dockerfiles or dependencies)
up-build:
	docker compose up -d --build

# 3. Stop and remove containers, networks, and default volumes
down:
	docker compose down

# 4. Tail live logs from all running containers
logs:
	docker compose logs -f

# 5. Restart all services cleanly
restart: down up

# 6. Check the running status of your containers
ps:
	docker compose ps

