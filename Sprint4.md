# Sprint 4 - RideX

## 1. Sprint Goal
Sprint 4 focused on finishing the rough edges left after Sprint 3, improving consistency between backend behavior and frontend UX, expanding automated verification for the new behavior, and updating the project documentation so the app is easier to run, test, and present.

## 2. Sprint 4 Work Completed

### Entire Team
- Reviewed the current Sprint 3 implementation and closed missing integration gaps.
- Added tests for newly completed Sprint 4 behavior.
- Updated the front-page `README.md` with setup, usage, requirements, and test instructions.
- Added this `Sprint4.md` deliverable with completed work, tests, and backend API documentation.

### Frontend Completed
- Removed the winter-only date restriction from ride request forms, published ride forms, and ride search filters so users can select any valid date.
- Refreshed `Publish a ride -> View Ride Requests` accept flow so it immediately updates:
  - `My Trips`
  - `Notifications`
  - driver-side ride request feed state
- Added a real `Chats` screen backed by live backend data, including conversation list, message history, and sending messages.
- Added chat support for image sharing and browser-based location sharing.
- Improved the home `Upcoming trip` summary to prioritize the latest confirmed trip instead of falling back to older request data when a trip already exists.
- Added a frontend unit test covering the ride-request accept refresh flow.

### Backend Completed
- Removed the backend winter-season validation and now accept any valid `YYYY-MM-DD` ride date.
- Added persistent chat storage and APIs for:
  - listing conversations
  - listing messages in a conversation
  - sending messages
- Extended chat messages to support text, image, and shared location message types.
- Automatically create rider-driver conversations from real accept/negotiate flows so chats are tied to actual ride interactions.
- Added notifications for the ride-request response flow:
  - rider receives a notification when a driver accepts the request
  - rider receives a notification when a driver starts negotiation
  - driver receives confirmation/negotiation notifications for the same actions
- Added notifications for new incoming chat messages.
- Extended backend handler tests to verify notifications are created for ride-request accept and to cover chat creation/message flows.

## 3. Frontend Unit and Cypress Tests

### Frontend Unit Test Files
- `frontend/src/App.test.jsx`
- `frontend/src/api/auth.test.js`

### Cypress Test Files
- `frontend/cypress/e2e/auth.cy.js`
- `frontend/cypress/e2e/sprint3_nav.cy.js`

### Frontend Test Commands
```bash
cd frontend
npm test
```

```bash
cd frontend
npx cypress run
```

## 4. Backend Unit Tests

### Backend Test Files
- `backend/internal/auth/password_test.go`
- `backend/internal/auth/validation_test.go`
- `backend/internal/auth/jwt_test.go`
- `backend/internal/handlers/chat_handler_test.go`
- `backend/internal/handlers/season_window_test.go`
- `backend/internal/handlers/trip_handler_test.go`

### Backend Test Command
```bash
cd backend
go test ./...
```

## 5. Backend API Documentation (Updated)

### Authentication APIs

#### `POST /api/auth/signup`
Purpose:
- Create a new email/password account.

#### `POST /api/auth/login`
Purpose:
- Authenticate a user and return a JWT plus user payload.

#### `POST /api/auth/verify-email`
Purpose:
- Verify an account using the emailed code.

#### `POST /api/auth/resend-verification`
Purpose:
- Send a new verification code.

#### `POST /api/auth/forgot-password`
Purpose:
- Send a password reset code.

#### `POST /api/auth/reset-password`
Purpose:
- Reset a password using the emailed code.

#### `GET /api/auth/oauth/google/start`
Purpose:
- Start Google OAuth.

#### `GET /api/auth/oauth/google/callback`
Purpose:
- Complete Google OAuth, create or link a user, then redirect back to the frontend.

#### `GET /api/auth/oauth/github/start`
Purpose:
- Start GitHub OAuth.

#### `GET /api/auth/oauth/github/callback`
Purpose:
- Complete GitHub OAuth, create or link a user, then redirect back to the frontend.

#### `GET /api/auth/me`
Purpose:
- Return the current authenticated user.

Headers:
```text
Authorization: Bearer <token>
```

#### `PUT /api/auth/me`
Purpose:
- Update the current user profile, including name, avatar, and interests.

Headers:
```text
Authorization: Bearer <token>
```

### Ride Request APIs

#### `POST /api/ride-requests`
Purpose:
- Create a rider request.

#### `GET /api/ride-requests`
Purpose:
- List the current user's own ride requests.

#### `PUT /api/ride-requests/:id`
Purpose:
- Update a previously created ride request.

#### `DELETE /api/ride-requests/:id`
Purpose:
- Cancel a ride request.

#### `GET /api/ride-requests/feed`
Purpose:
- List ride requests for drivers with filter support.

Headers:
```text
Authorization: Bearer <token>
```

#### `POST /api/ride-requests/:id/respond`
Purpose:
- Driver accepts or negotiates on a rider request.
- On `accept`, the backend creates a confirmed trip.
- Sprint 4 update: this flow now also creates notifications for both rider and driver.

Headers:
```text
Authorization: Bearer <token>
```

Request body:
```json
{
  "action": "accept",
  "publishedRideId": "optional-driver-ride-id",
  "message": "Driver is ready to confirm this route."
}
```

### Published Ride APIs

#### `POST /api/published-rides`
Purpose:
- Create a published ride listing as a driver.

#### `GET /api/published-rides`
Purpose:
- List the current user's own published rides.

#### `GET /api/published-rides/feed`
Purpose:
- List active published rides from other users with filters.

Headers:
```text
Authorization: Bearer <token>
```

#### `POST /api/published-rides/:id/respond`
Purpose:
- Rider accepts or negotiates on a published ride.
- On `accept`, the backend creates a ride request, confirms a trip, hides the ride from the feed, and creates notifications.

Headers:
```text
Authorization: Bearer <token>
```

Request body:
```json
{
  "action": "accept",
  "message": "Rider is ready to confirm this ride.",
  "passengers": 2
}
```

### Trips API

#### `GET /api/trips`
Purpose:
- List confirmed or cancelled trips for the authenticated user with enriched route, schedule, rider, and driver details.

Headers:
```text
Authorization: Bearer <token>
```

#### `POST /api/trips/:id/cancel`
Purpose:
- Cancel a confirmed trip.
- If linked to a published ride, the ride becomes active again in the discovery feed.
- Cancellation also creates notifications.

Headers:
```text
Authorization: Bearer <token>
```

### Notifications API

#### `GET /api/notifications`
Purpose:
- List notifications for the authenticated user, ordered newest first.

Headers:
```text
Authorization: Bearer <token>
```

### Chats API

#### `GET /api/chats`
Purpose:
- List real rider-driver chat conversations for the authenticated user.

Headers:
```text
Authorization: Bearer <token>
```

#### `GET /api/chats/:id/messages`
Purpose:
- List messages in a chat conversation the authenticated user belongs to.

Headers:
```text
Authorization: Bearer <token>
```

#### `POST /api/chats/:id/messages`
Purpose:
- Send a message to a rider-driver conversation.
- Sending a message also creates a notification for the other participant.
- Supports text, image, and shared location messages.

Headers:
```text
Authorization: Bearer <token>
```

Request body:
```json
{
  "body": "I can meet at the north entrance around 7:15."
}
```

Image message example:
```json
{
  "messageType": "image",
  "body": "Here is the pickup spot photo.",
  "imageData": "data:image/png;base64,..."
}
```

Location message example:
```json
{
  "messageType": "location",
  "body": "Meet me here.",
  "locationLabel": "Airport Terminal A",
  "locationLat": 28.4312,
  "locationLon": -81.3081
}
```

### Standard Error Shape
```json
{
  "error": "message"
}
```

## 6. Verification Commands

### Frontend
```bash
cd frontend
npm test
```

### Cypress
```bash
cd frontend
npx cypress run
```

### Backend
```bash
cd backend
go test ./...
```
