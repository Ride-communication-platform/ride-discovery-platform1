# Sprint 3 - RideX

## 1. Sprint Goal
Sprint 3 focused on closing the main deferred workflow from Sprint 2 by **persisting a confirmed trip record** when a driver accepts a ride request, and exposing that confirmed trip data through a dedicated API endpoint. The sprint also expanded automated test coverage for the new trip workflow and updated the backend API documentation to reflect the new endpoint and response shape.

## 2. Sprint 3 Work Completed

### Entire Team
- Implemented a confirmed trip record that is created when a driver **accepts** a ride request.
- Added an authenticated `GET /api/trips` endpoint to list trips for the current user (as rider or driver).
- Updated the ride request respond flow to return a `trip` object when `action=accept`.
- Added new automated unit tests (frontend + backend) and verified all unit tests (including Sprint 2 tests) pass.
- Updated backend API documentation to include the new Trips API and the updated respond payload.

### Frontend Completed
- Added a frontend API client function `listTrips()` for retrieving confirmed trips from the backend.
- Added a frontend unit test validating the `GET /api/trips` request behavior.

### Backend Completed
- Added a persistent `trips` table (SQLite) with a unique constraint on `ride_request_id` to prevent duplicate confirmed trips.
- Added store functions for creating a confirmed trip, listing trips for the current user, and returning an **enriched trip view** (joined rider/driver names + route/schedule details).
- Added handler support for:
  - `GET /api/trips` (enriched trip view)
  - `POST /api/trips/:id/cancel` (cancels a confirmed trip and reactivates the published ride listing when applicable)
- Extended `POST /api/ride-requests/:id/respond` so `action=accept` creates and returns a `trip` record.
- Added `GET /api/published-rides/feed` (Find-a-ride) with filters and “exclude self”.
- Added `POST /api/published-rides/:id/respond` so riders can `accept` / `negotiate` on published rides:
  - On `accept`: confirms a trip, sets the published ride `status` to inactive (disappears for others).
- Added `notifications` persistence and `GET /api/notifications` so both rider and driver receive real notifications for accept/negotiate/cancel events.
- Added handler-level unit tests verifying trip creation, cancellation/reactivation, and published-ride accept behavior.

## 3. Remaining / Deferred Work
- Trip lifecycle beyond confirmation (e.g., completed/cancelled states, ratings, and trip completion flows) is not implemented yet.
- Live push updates through WebSockets/SSE are still not implemented.

## 4. Frontend Unit Tests

### Cypress Test (Simple)
- `frontend/cypress/e2e/auth.cy.js`
- `frontend/cypress/e2e/sprint3_nav.cy.js`

### Frontend Unit Test Files
- `frontend/src/App.test.jsx`
- `frontend/src/api/auth.test.js`

### Frontend Test Command
```bash
cd frontend
npm test
```

## 5. Backend Unit Tests

### Backend Test Files
- `backend/internal/auth/password_test.go`
- `backend/internal/auth/validation_test.go`
- `backend/internal/auth/jwt_test.go`
- `backend/internal/handlers/trip_handler_test.go`

### Backend Test Command
```bash
cd backend
go test ./...
```

## 6. Backend API Documentation (Updated)

### Trips API

#### `GET /api/trips`
Purpose:
- List trips for the authenticated user (as rider or driver) with enriched details (route + schedule + names).

Headers:
```text
Authorization: Bearer <token>
```

Response body:
```json
{
  "trips": [
    {
      "id": "trip-id",
      "status": "confirmed",
      "createdAt": "2026-04-13T15:04:05Z",
      "rideRequestId": "ride-request-id",
      "publishedRideId": "published-ride-id",
      "riderUserId": "rider-user-id",
      "riderName": "Rider Name",
      "driverUserId": "driver-user-id",
      "driverName": "Driver Name",
      "fromLabel": "Miami, FL",
      "toLabel": "Orlando, FL",
      "rideDate": "2026-04-15",
      "rideTime": "07:30",
      "passengers": 2,
      "routeMiles": 234,
      "routeDuration": "4h 24m",
      "pricePerSeat": 25,
      "estimatedTotal": 50
    }
  ]
}
```

Auth required:
- Yes

#### `POST /api/trips/:id/cancel`
Purpose:
- Cancel a confirmed trip (rider or driver can cancel).
- If linked to a published ride, the ride is set back to `active` so it reappears in Find-a-ride.

Headers:
```text
Authorization: Bearer <token>
```

Auth required:
- Yes

### Published Ride Discovery APIs

#### `GET /api/published-rides/feed`
Purpose:
- List published rides for riders to browse (excludes your own rides).
- Returns only `active` rides.

Headers:
```text
Authorization: Bearer <token>
```

Query params (optional):
- `date`, `route`, `seats`, `maxPrice`
- `rideType`, `vehicleType`, `luggageAllowed`, `flexibility`

Auth required:
- Yes

#### `POST /api/published-rides/:id/respond`
Purpose:
- Rider responds to a published ride with `accept` or `negotiate`.
- On `accept`, confirms a trip and sets the published ride to inactive so it disappears for others.

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

Auth required:
- Yes

### Notifications API

#### `GET /api/notifications`
Purpose:
- List notifications for the authenticated user (newest first).

Headers:
```text
Authorization: Bearer <token>
```

Auth required:
- Yes

### Ride Request Actions (Updated)

#### `POST /api/ride-requests/:id/respond`
Purpose:
- Record a driver response to a rider request.
- When `action=accept`, the backend now also creates a confirmed trip record and returns it as `trip`.

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

Response body (accept):
```json
{
  "message": "Ride request accepted.",
  "action": {
    "id": "action-id",
    "rideRequestId": "ride-request-id",
    "driverUserId": "driver-user-id",
    "publishedRideId": "",
    "action": "accept",
    "message": "Driver is ready to confirm this route.",
    "createdAt": "2026-04-13T15:04:05Z"
  },
  "trip": {
    "id": "trip-id",
    "rideRequestId": "ride-request-id",
    "riderUserId": "rider-user-id",
    "driverUserId": "driver-user-id",
    "publishedRideId": "",
    "status": "confirmed",
    "createdAt": "2026-04-13T15:04:05Z"
  }
}
```

Response body (negotiate):
```json
{
  "message": "Negotiation started.",
  "action": {
    "id": "action-id",
    "rideRequestId": "ride-request-id",
    "driverUserId": "driver-user-id",
    "publishedRideId": "",
    "action": "negotiate",
    "message": "Driver wants to discuss pricing and pickup details.",
    "createdAt": "2026-04-13T15:04:05Z"
  },
  "trip": null
}
```

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

### Cypress (Simple E2E)
```bash
cd frontend
npm run dev
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
- GitHub Repository: `https://github.com/Ride-communication-platform/ride-discovery-platform1/tree/main`
- Video Demo: `https://drive.google.com/drive/u/0/folders/1pbfFmdI8RoOfTg5x2qVKaMCLpl3IP9uo`

### Video checklist (what to show)
- Demonstrate new Sprint 3 features:
  - Find a ride → filter/search → Negotiate → Notifications
  - Accept a ride → ride disappears from feed → Upcoming trip updates → View all → My Trips
  - Cancel trip → ride reappears for all → Notifications updated
  - RideX logo click returns to Home
- Show unit test results (Sprint 2 + Sprint 3):
  - Frontend: `npm test`
  - Backend: `go test ./...`
  - Cypress: run `sprint3_nav.cy.js` (simple navigation test)

## 9. GitHub issues — work completed after Sprint 2 (12 items)

Use these as copy-paste GitHub issues (labels: `frontend` / `backend`, `sprint-3`). Each maps to work shipped in this repo after Sprint 2.

### Frontend (6)

#### FE-1: Find-a-ride — published rides discovery UI
- **Summary:** Add a **Find a ride** flow that lists **other users’** published rides with filters (date, route text, seats, max price per seat, ride type, vehicle type, luggage, flexibility) consistent with the rest of the app.
- **Acceptance criteria:** Home “Find a ride” opens the view; filters call the feed API; empty/loading states are clear.

#### FE-2: Find-a-ride — rider Accept / Negotiate on published rides
- **Summary:** Match the **driver feed card** pattern (route, meta, notes, actions). Wire **Accept** and **Negotiate** to the backend; on **Accept**, remove the card locally and refresh trips/notifications.
- **Acceptance criteria:** Buttons show loading state; errors surface from API; successful accept updates UI without full page reload.

#### FE-3: My Trips — list, details, and cancel
- **Summary:** **My Trips** nav opens a screen listing enriched trips (route, schedule, parties, miles/duration, estimate). **Cancel trip** calls the API and refreshes lists.
- **Acceptance criteria:** Confirmed trips show **Cancel**; cancelled trips display correct status; feed can refresh after cancel when on Find-a-ride.

#### FE-4: Notifications screen
- **Summary:** **Notifications** nav shows items from `GET /api/notifications` (title, body, timestamp). Placeholder empty state when none.
- **Acceptance criteria:** New notifications appear after accept/negotiate/cancel flows when backend creates them.

#### FE-5: Home — Upcoming trip uses real data + View all
- **Summary:** **Upcoming trip** shows the latest **confirmed** trip summary (not generic placeholder). **View all** navigates to **My Trips**.
- **Acceptance criteria:** With at least one confirmed trip, home reflects it; View all lands on My Trips.

#### FE-6: Brand / nav polish — RideX logo goes home (not a button)
- **Summary:** Logo is a **link-styled** home control (no button chrome); logged-in users return to **home** view; logged-out users reset to a clean login-focused auth state.
- **Acceptance criteria:** Keyboard focus ring only; no accidental full reload (`preventDefault` on click).

### Backend (6)

#### BE-1: Published rides feed API
- **Summary:** `GET /api/published-rides/feed` returns published rides **excluding the current user**, with query filters and **driver** summary fields (name, rating).
- **Acceptance criteria:** Auth required; filters work; only **active** listings appear (see BE-2).

#### BE-2: Published ride listing lifecycle (`status`)
- **Summary:** Add `published_rides.status` (`active` / `inactive`). Feed returns **active** only. On **rider accept** of a published ride, mark listing **inactive** so it disappears for others.
- **Acceptance criteria:** DB migration applies on existing SQLite; create path defaults to `active`.

#### BE-3: Rider respond to published ride
- **Summary:** `POST /api/published-rides/:id/respond` with `accept` | `negotiate`, optional `passengers` and `message`. On **accept**, create a **ride_request** row for the rider, then **confirmed trip** linking rider, driver, and `published_ride_id`.
- **Acceptance criteria:** Cannot respond to own ride; validation errors are consistent with other handlers.

#### BE-4: Trips API — enriched views + cancel + re-list ride
- **Summary:** `GET /api/trips` returns **joined** trip views (labels, schedule, passengers, route meta, names, price estimate). `POST /api/trips/:id/cancel` sets trip **cancelled** and sets linked published ride back to **active** when applicable.
- **Acceptance criteria:** Only participant can cancel; cancel is idempotent-safe (404 if not found / wrong state).

#### BE-5: Notifications API
- **Summary:** `notifications` table + `GET /api/notifications`. Create notifications on key events (e.g. accept/negotiate on published ride, trip cancel — align with implemented handlers).
- **Acceptance criteria:** Auth required; list is ordered newest-first.

#### BE-6: Tests — handlers for trips, published-ride respond, notifications plumbing
- **Summary:** Extend `go test` coverage for trip creation/listing, published-ride accept path (trip + notifications), and related handler behavior.
- **Acceptance criteria:** `go test ./...` passes in CI/local.

## 10. Team Contribution Summary
- Frontend Member 1: `Manikanta Srinivas Penumarthi` worked on Sprint 3 end-to-end UX for Find-a-ride, My Trips, notifications UI wiring, Cypress + Vitest tests, and overall frontend integration.
- Frontend Member 2: `Avighna Yarlagadda` worked on UI consistency, card layout alignment with the existing feed, navigation polish (RideX home link), and Find-a-ride filter UX refinement.
- Backend Member 1: `Srija Chowdary Chava` worked on Sprint 3 backend handlers for published-ride feed/respond, trips (enriched views + cancel), and notifications API integration.
- Backend Member 2: `Sai Krishna Reddy Kethireddy` worked on database migrations (trips, notifications, published_ride status), store-layer queries, and backend test coverage for trip lifecycle and ride visibility rules.

