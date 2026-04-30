# RideX

RideX is a full-stack ride discovery and coordination app for shared travel. Riders can post ride requests or browse driver-published rides, and drivers can publish seats or respond to incoming requests. Confirmed matches become trips, both sides receive notifications for important events, and real chat conversations are available once a ride interaction starts.

## Tech Stack
- Frontend: React + Vite
- Backend: Go + net/http
- Database: SQLite
- Auth: JWT, bcrypt, Google OAuth, GitHub OAuth
- Tests: Vitest, Cypress, Go `testing`

## Core Features
- Email/password signup and login
- Email verification and password reset
- Google and GitHub sign-in entry points
- Profile editing with avatar crop and interests
- Ride request creation, editing, cancellation, and driver feed
- Published ride creation and rider discovery feed
- Confirmed trips list with trip cancellation
- Notifications for accept, negotiate, and cancel flows
- Real rider-driver chat threads backed by the backend database
- Route preview using OpenStreetMap search and OSRM routing

## Requirements
- Node.js 18+
- npm 9+
- Go 1.22+ recommended

## Project Structure
- `frontend/` React client
- `backend/` Go API
- `Sprint1.md`, `Sprint2.md`, `Sprint3.md`, `Sprint4.md` sprint summaries

## Backend Setup
```bash
cd backend
go mod download
go run ./cmd/server
```

The backend loads `backend/.env` automatically when present.

### Backend Environment Variables
- `PORT` default `8080`
- `JWT_SECRET` default `ridex-dev-secret-change-me`
- `DB_PATH` default `data/ridex.db`
- `FRONTEND_ORIGIN` default `http://localhost:5173`
- `SMTP_HOST`, `SMTP_PORT`, `SMTP_USER`, `SMTP_PASS`, `SMTP_FROM` for real email delivery
- `GOOGLE_CLIENT_ID`, `GOOGLE_CLIENT_SECRET`, `GOOGLE_REDIRECT_URI` for Google OAuth
- `GITHUB_CLIENT_ID`, `GITHUB_CLIENT_SECRET`, `GITHUB_REDIRECT_URI` for GitHub OAuth

Example `backend/.env`:
```bash
PORT=8080
JWT_SECRET=ridex-dev-secret-change-me
DB_PATH=data/ridex.db
FRONTEND_ORIGIN=http://localhost:5173
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USER=your-account@gmail.com
SMTP_PASS=your-app-password
SMTP_FROM=your-account@gmail.com
GOOGLE_CLIENT_ID=your-google-client-id.apps.googleusercontent.com
GOOGLE_CLIENT_SECRET=your-google-client-secret
GOOGLE_REDIRECT_URI=http://localhost:8080/api/auth/oauth/google/callback
GITHUB_CLIENT_ID=your-github-client-id
GITHUB_CLIENT_SECRET=your-github-client-secret
GITHUB_REDIRECT_URI=http://localhost:8080/api/auth/oauth/github/callback
```

## Frontend Setup
```bash
cd frontend
npm install
npm run dev
```

### Frontend Environment Variables
- `VITE_API_BASE_URL` default `http://localhost:8080`

## How To Use
1. Start the backend on port `8080`.
2. Start the frontend on port `5173`.
3. Create an account or sign in.
4. Use `Request a ride` to post a rider request.
5. Use `Publish a ride` to offer seats as a driver.
6. Use `Find a ride` to browse other users' published rides.
7. Accepting or negotiating on a ride creates a real chat conversation in `Chats`.
8. Accepting a ride or rider request creates a confirmed trip in `My Trips`.
9. Check `Notifications` for accept, negotiate, cancel, and new-message updates.

## Test Commands

### Frontend Unit Tests
```bash
cd frontend
npm test
```

### Cypress
```bash
cd frontend
npx cypress run
```

### Backend Unit Tests
```bash
cd backend
go test ./...
```

## Backend API Summary

### Authentication
- `POST /api/auth/signup`
- `POST /api/auth/login`
- `POST /api/auth/verify-email`
- `POST /api/auth/resend-verification`
- `POST /api/auth/forgot-password`
- `POST /api/auth/reset-password`
- `GET /api/auth/oauth/google/start`
- `GET /api/auth/oauth/google/callback`
- `GET /api/auth/oauth/github/start`
- `GET /api/auth/oauth/github/callback`
- `GET /api/auth/me`
- `PUT /api/auth/me`

### Riders and Drivers
- `GET /api/users/:id/profile`
- `POST /api/ride-requests`
- `GET /api/ride-requests`
- `PUT /api/ride-requests/:id`
- `DELETE /api/ride-requests/:id`
- `GET /api/ride-requests/feed`
- `POST /api/ride-requests/:id/respond`
- `POST /api/published-rides`
- `GET /api/published-rides`
- `GET /api/published-rides/feed`
- `POST /api/published-rides/:id/respond`

### Trips and Notifications
- `GET /api/trips`
- `POST /api/trips/:id/cancel`
- `GET /api/notifications`

### Standard Error Shape
```json
{
  "error": "message"
}
```

## Sprint 4 Notes
- Sprint 4 deliverables are documented in `Sprint4.md`.
- The current repo includes frontend, Cypress, and backend test files for the implemented flows.
