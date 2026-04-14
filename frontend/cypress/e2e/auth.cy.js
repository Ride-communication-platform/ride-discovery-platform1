describe('RideX auth page', () => {
  it('switches from login to signup and shows the signup fields', () => {
    cy.visit('/')

    cy.contains('button', /^sign up$/i).click()

    cy.contains('button', /^sign up$/i).should('have.attr', 'aria-selected', 'true')
    cy.contains('label', /full name/i).should('be.visible')
    cy.contains('button', /create free account/i).should('be.visible')
  })

  it('shows login validation errors when submitted empty', () => {
    cy.visit('/')

    cy.get('button[type="submit"]').contains(/^login$/i).click()

    cy.contains(/email is required/i).should('be.visible')
    cy.contains(/password is required/i).should('be.visible')
  })
})
