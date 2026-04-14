describe('Sprint 3 navigation (mocked API)', () => {
  it('opens Find a ride and logo returns Home', () => {
    cy.visit('/', {
      onBeforeLoad(win) {
        win.localStorage.setItem('ridex_token', 'test-token')
      },
    })

    cy.intercept('GET', '**/api/auth/me', {
      statusCode: 200,
      body: {
        user: {
          id: 'u1',
          name: 'Test User',
          email: 'test@example.com',
          avatarData: '',
          interests: [],
          rating: 0,
          ratingCount: 0,
          tripsCompleted: 0,
          emailVerified: true,
          authProvider: 'password',
          createdAt: '2026-04-13T00:00:00Z',
        },
      },
    }).as('me')

    cy.intercept('GET', '**/api/ride-requests', { statusCode: 200, body: { requests: [] } })
    cy.intercept('GET', '**/api/published-rides', { statusCode: 200, body: { rides: [] } })
    cy.intercept('GET', '**/api/trips', { statusCode: 200, body: { trips: [] } })
    cy.intercept('GET', '**/api/notifications', { statusCode: 200, body: { notifications: [] } })
    cy.intercept('GET', '**/api/published-rides/feed*', { statusCode: 200, body: { rides: [] } })

    cy.wait('@me')
    cy.contains('h2', /welcome back/i).should('be.visible')

    cy.contains('h3', /find a ride/i).click()
    cy.contains('h2', /find a ride/i).should('be.visible')

    cy.contains('RideX').click()
    cy.contains('h2', /welcome back/i).should('be.visible')
  })
})

