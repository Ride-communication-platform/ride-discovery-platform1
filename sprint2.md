# Sprint 2 - RideX

## 1. Sprint Goal
Sprint 2 focused on extending the Sprint 1 authentication MVP into a more integrated RideX experience. The main goals were to complete remaining auth improvements, integrate frontend and backend end to end, add real profile and ride workflows, document the backend API, and introduce automated tests for both frontend and backend.

## 2. Sprint 2 Work Completed

### Entire Team
- Integrated the React frontend with the Go backend across authentication, profile, ride request, and published ride flows.
- Carried Sprint 1 auth work forward into a larger product flow with profile, rider, and driver interactions.
- Verified core application behavior through frontend and backend tests.

### Frontend Completed
- Added Google and GitHub social authentication entry points in the auth UI.
- Added email verification, resend verification, forgot password, and reset password overlays.
- Added authenticated home, profile, request ride, and publish ride experiences.
- Added avatar upload and crop support with persistent profile updates.
- Added worldwide location autocomplete, route preview, and real map integration.
- Added ride request management: create, edit, cancel, and rider-facing request history.
- Added driver-facing request feed with request detail modal and public rider profile modal.
- Added frontend automated tests using Vitest and Cypress.

### Backend Completed
- Added SMTP-based email verification and password reset flows.
- Added Google OAuth and GitHub OAuth start/callback flows.
- Added protected profile update support for name, avatar, and interests.
- Added public rider profile endpoint for ride-related frontend modals.
- Added persistent ride request create, list, update, delete, and feed APIs.
- Added published ride create/list APIs.
- Added driver response actions for ride requests such as accept and negotiate.
- Added backend unit tests for auth helper logic.

## 3. Remaining / Deferred Work
- Accepting a ride request stores a real backend response, but it does not yet create a final confirmed trip record.
- Live push updates through WebSockets or SSE are not implemented yet.

## 4. Frontend Tests

### Cypress Test
- `frontend/cypress/e2e/auth.cy.js`
- Test 1: switches from Login to Sign Up and confirms the signup fields are visible.
- Test 2: submits the login form empty and confirms validation errors appear.

### Frontend Unit Tests
- `frontend/src/App.test.jsx`
- Test 1: renders the Login tab by default.
- Test 2: switches to the Sign Up tab and shows the full name field.
- Test 3: shows validation errors when login is submitted empty.

### Frontend Test Commands
```bash
cd frontend
npm test
```

```bash
cd frontend
npm run dev
```

```bash
cd frontend
npx cypress run --spec cypress/e2e/auth.cy.js
```

### Frontend Test Results
- `npm test` passed.
- Cypress tests passed.

## 5. Backend Unit Tests

### Backend Test Files
- `backend/internal/auth/password_test.go`
- `backend/internal/auth/validation_test.go`
- `backend/internal/auth/jwt_test.go`

### Backend Test Coverage Added
- Password hashing produces a non-plain-text hash.
- Password comparison succeeds for the correct password and fails for the wrong password.
- Email validation accepts valid email strings and rejects invalid ones.
- Password strength validation accepts strong passwords and rejects weak ones.
- JWT generation returns a valid signed token.
- JWT verification returns the expected claims for a valid token.
- JWT verification rejects invalid tokens.

### Backend Test Command
```bash
cd backend
go test ./...
```

### Backend Test Results
- `go test ./...` passed.

## 6. Backend API Documentation

### Authentication APIs

#### `POST /api/auth/signup`
Purpose:
- Create a new email/password user account.

Request body:
```json
{
  "name": "Alex Rider",
  "email": "alex@example.com",
  "password": "secret123"
}
```

Behavior:
- Validates name, email, and password.
- Rejects duplicate emails.
- Hashes password with bcrypt.
- Creates a verification code and sends a verification email through SMTP.

Auth required:
- No

#### `POST /api/auth/login`
Purpose:
- Authenticate a user and return a JWT plus user payload.

Request body:
```json
{
  "email": "alex@example.com",
  "password": "secret123"
}
```

Behavior:
- Validates credentials.
- Rejects unverified email/password users until email is verified.
- Returns JWT token and user object on success.

Auth required:
- No

#### `POST /api/auth/verify-email`
Purpose:
- Verify a user account using the emailed verification code.

Request body:
```json
{
  "email": "alex@example.com",
  "code": "123456"
}
```

Auth required:
- No

#### `POST /api/auth/resend-verification`
Purpose:
- Generate and send a new verification code.

Request body:
```json
{
  "email": "alex@example.com"
}
```

Auth required:
- No

#### `POST /api/auth/forgot-password`
Purpose:
- Generate a password reset code and send it by email.

Request body:
```json
{
  "email": "alex@example.com"
}
```

Auth required:
- No

#### `POST /api/auth/reset-password`
Purpose:
- Reset the user password using the emailed reset code.

Request body:
```json
{
  "email": "alex@example.com",
  "code": "123456",
  "newPassword": "newsecret123"
}
```

Auth required:
- No

#### `GET /api/auth/oauth/google/start`
Purpose:
- Start the Google OAuth sign-in flow.

Auth required:
- No

#### `GET /api/auth/oauth/google/callback`
Purpose:
- Complete Google OAuth, create/link the RideX user, issue JWT, and redirect to the frontend.

Auth required:
- No

#### `GET /api/auth/oauth/github/start`
Purpose:
- Start the GitHub OAuth sign-in flow.

Auth required:
- No

#### `GET /api/auth/oauth/github/callback`
Purpose:
- Complete GitHub OAuth, create/link the RideX user, issue JWT, and redirect to the frontend.

Auth required:
- No

#### `GET /api/auth/me`
Purpose:
- Return the currently authenticated user.

Headers:
```text
Authorization: Bearer <token>
```

Auth required:
- Yes

#### `PUT /api/auth/me`
Purpose:
- Update the authenticated user’s profile.

Headers:
```text
Authorization: Bearer <token>
```

Request body:
```json
{
  "name": "Updated User",
  "avatarData": "data:image/png;base64,...",
  "interests": ["Orlando", "Road Trips"]
}
```

Behavior:
- Persists profile edits in the database.
- Used by the profile page for name, avatar, and interests.

Auth required:
- Yes

### Public Profile API

#### `GET /api/users/:id/profile`
Purpose:
- Return a safe public rider profile for use in ride detail and rider profile modals.

Headers:
```text
Authorization: Bearer <token>
```

Behavior:
- Exposes non-sensitive fields only.
- Does not expose private email or password data.

Auth required:
- Yes

### Ride Request APIs

#### `POST /api/ride-requests`
Purpose:
- Create a new rider request.

Headers:
```text
Authorization: Bearer <token>
```

Request body includes:
- `fromLabel`
- `fromLat`
- `fromLon`
- `toLabel`
- `toLat`
- `toLon`
- `rideDate`
- `rideTime`
- `flexibility`
- `passengers`
- `luggage`
- `maxBudget`
- `rideType`
- `vehiclePreference`
- `minimumRating`
- `verifiedDriversOnly`
- `notes`
- `routeMiles`
- `routeDuration`
- `priceEstimate`

Auth required:
- Yes

#### `GET /api/ride-requests`
Purpose:
- List the authenticated user’s own ride requests.

Auth required:
- Yes

#### `GET /api/ride-requests/:id`
Purpose:
- Return one ride request by ID for authenticated flows.

Auth required:
- Yes

#### `PUT /api/ride-requests/:id`
Purpose:
- Update one of the authenticated user’s ride requests.

Auth required:
- Yes

Behavior:
- Owner-only update.

#### `DELETE /api/ride-requests/:id`
Purpose:
- Cancel and delete one of the authenticated user’s ride requests.

Auth required:
- Yes

Behavior:
- Owner-only delete.

#### `GET /api/ride-requests/feed`
Purpose:
- Return ride requests for drivers to browse.

Supported filters:
- `date`
- `route`
- `passengers`
- `budget`

Behavior:
- Excludes the current user’s own requests from the feed.

Auth required:
- Yes

#### `POST /api/ride-requests/:id/respond`
Purpose:
- Record a driver response to a rider request.

Headers:
```text
Authorization: Bearer <token>
```

Request body:
```json
{
  "action": "accept",
  "publishedRideId": "optional-ride-id",
  "message": "Driver is ready to confirm this route."
}
```

Behavior:
- Stores real backend response actions such as `accept` and `negotiate`.

Auth required:
- Yes

### Published Ride APIs

#### `POST /api/published-rides`
Purpose:
- Create a new published driver ride.

Headers:
```text
Authorization: Bearer <token>
```

Request body includes:
- `fromLabel`
- `fromLat`
- `fromLon`
- `toLabel`
- `toLat`
- `toLon`
- `rideDate`
- `rideTime`
- `flexibility`
- `availableSeats`
- `totalSeats`
- `pricePerSeat`
- `vehicleType`
- `luggageAllowed`
- `rideType`
- `vehicleInfo`
- `notes`
- `routeMiles`
- `routeDuration`
- `earningsEstimate`

Auth required:
- Yes

#### `GET /api/published-rides`
Purpose:
- Return the authenticated driver’s published rides.

Auth required:
- Yes

### Standard Error Shape
```json
{
  "error": "message"
}
```

## 7. Commands Used for Verification

### Frontend
```bash
cd frontend
npm test
```

```bash
cd frontend
npx cypress open
```

### Backend
```bash
cd backend
go test ./...
```

## 8. Submission Links
- GitHub Repository: `https://github.com/Ride-communication-platform/ride-discovery-platform1`
- Frontend Video: `https://drive.google.com/drive/folders/1zTkvuGi01OlpNWRdS8vGDeibSBp5_tcc`
- Backend Video: `https://drive.google.com/drive/folders/1zTkvuGi01OlpNWRdS8vGDeibSBp5_tcc`
- Combined Demo Video: `https://drive.google.com/drive/folders/1zTkvuGi01OlpNWRdS8vGDeibSBp5_tcc`

## 9. Team Contribution Summary
- Frontend Member 1: `Manikanta Srinivas Penumarthi` worked on authentication flows, social login integration, session handling, frontend tests, and integrated ride-related product flows.
- Frontend Member 2: `Avighna Yarlagadda` worked on UI system, RideX theme consistency, responsive layout, profile/request/publish page polish, and interaction refinements.
- Backend Member 1: `Srija Chowdary Chava` worked on backend auth handlers, protected endpoints, public profile support, and ride request/published ride API flow integration.
- Backend Member 2: `Sai Krishna Reddy Kethireddy` worked on persistence, database-backed auth and ride data, JWT/bcrypt infrastructure, email verification/reset support, and backend test coverage.
