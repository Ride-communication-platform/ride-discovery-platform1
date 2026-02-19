🚀 User Stories
👤 New User — Account Creation

As a new user, I open RideX for the first time and want a simple way to join.
I enter my basic details and create my account in one flow.
I expect clear guidance if anything is missing or invalid.
I want to feel confident that my account was created successfully.

🔑 Returning User — Login

As a returning user, I want to sign in quickly with my email and password.
I expect login to work smoothly without unnecessary steps.
If my credentials are wrong, I want a clear and respectful error.
I should immediately know what to fix and try again.

🔄 Session Persistence

As a user, I want the app to remember me after I log in.
When I refresh or reopen the app, I should stay signed in.
If my session is expired, the app should handle it safely.
I should be asked to log in again without confusion.

🚪 Logout

As a logged-in user, I want to log out whenever I choose.
My session should be cleared right away on this device.
No private account details should remain visible afterward.
I should return to a clean authentication screen.

📝 Form Experience

As a user, I want forms that feel clear and supportive.
Fields should validate in real time as I type.
Helpful messages should explain what is wrong and why.
I should be able to complete the form without guessing.

🔒 Security & Trust

As a user, I want to trust how RideX handles my credentials.
My password should never be stored in plain text.
The system should protect access using secure authentication tokens.
My account should feel safe every time I sign in.

👤 Consistent Profile Data

As a user, I want my profile data to be consistent after login.
The app should receive a clean user payload from the backend.
I should see the correct account identity every session.
The same behavior should work across refreshes and devices.

📧 Duplicate Email Prevention

As a user, I want account creation to prevent duplicate emails.
If an email is already used, I should be told clearly.
I should understand that I need to log in instead.
The message should be friendly and actionable.

🎨 Premium Authentication Experience

As a user, I want the authentication page to feel premium and trustworthy.
The interface should look modern, focused, and easy to scan.
Colors, spacing, and motion should support clarity.
I should feel confident continuing into the product.

👥 Team Perspective

As a team, we want clear sprint delivery and ownership.
We define stories, create issues, and split frontend/backend work.
We track what was planned, completed, and deferred with reasons.
This enables transparent demos and better planning for the next sprint.

📦 Issues Planned for Sprint 1
🎨 Frontend Issues

#1 Build RideX Authentication UI (Login/Sign Up)

#2 Implement Login Form Submission Flow in React

#3 Add Client-Side Validation for Auth Inputs

#6 Integrate Frontend Login with Backend Auth API

#7 Persist User Session in Local Storage and Restore on Reload

#8 Implement Frontend Logout Flow and Session Reset

⚙️ Backend Issues

#4 Create Signup API Endpoint (POST /api/auth/signup)

#5 Create Login API Endpoint (POST /api/auth/login)

#9 Create Authenticated User Endpoint (GET /api/auth/me)

#10 Implement JWT/Auth Middleware for Protected Routes

#11 Add User Persistence with JSON Store (users.json)

#12 Return Frontend-Compatible User Session Payload

✅ Issues Successfully Completed

All planned issues for Sprint 1 were completed.

Key deliverables achieved:

Frontend

Premium authentication UI

Login & signup flows

Real-time validation

Session persistence using local storage

Secure logout flow

Integration with backend APIs

Backend

User registration endpoint

Login endpoint with credential verification

JWT-based authentication

Protected user endpoint

Middleware for route protection

Persistent user storage

Consistent JSON responses

⚠ Issues Not Completed / Deferred

None.

All issues planned for Sprint 1 MVP scope were completed successfully.

Optional enhancements not included in this sprint:

Automated tests

Rate limiting

Password reset functionality

Email verification

These will be considered in future sprints.
