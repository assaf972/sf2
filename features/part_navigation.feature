Feature: Part navigation shuttle
  As a gigging keyboardist
  I want to step through a song's parts during a performance
  So that all three keyboards reload their sounds in one action

  Background:
    Given a song "Shine On You Crazy Diamond" with parts:
      | order | name       | kb1_sound          |
      | 1     | Intro      | Synth Wind         |
      | 2     | Glass Pad  | Wine Glass Pad     |
      | 3     | Solina Pad | Solina String Ens. |
    And part 1 is active

  Scenario: Next part recalls its sounds
    When I press "Next Part"
    Then the active part is "Glass Pad"
    And keyboard 1 sound is "Wine Glass Pad"
    And a PANIC was issued before the recall

  Scenario: Prev at the first part stays put
    When I press "Prev Part"
    Then the active part is "Intro"

  Scenario: Next at the last part stays on the last part
    Given part 3 is active
    When I press "Next Part"
    Then the active part is "Solina Pad"
