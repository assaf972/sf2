Feature: Song & part persistence (SQLite)
  As a gigging keyboardist
  I want my songs and parts saved to a local database
  So that my set list survives restarts

  Background:
    Given an empty in-memory GigSynth database

  Scenario: Create a song and add ordered parts
    When I create a song "Shine On You Crazy Diamond"
    And I add a part "Intro" to that song
    And I add a part "Solina Pad" to that song
    Then the song has 2 parts in order:
      | order | name       |
      | 1     | Intro      |
      | 2     | Solina Pad |

  Scenario: A part stores per-channel sound, mix and FX
    Given a song "Test" with a part "P1"
    When I set part "P1" channel 1 to sound bank 2 program 1 volume 108 pan 64
    And I set part "P1" channel 1 chorus on rate 0.8 depth 55
    And I set part "P1" channel 1 delay on time 320 feedback 35 mix 30
    And I reload the part from the database
    Then channel 1 sound is bank 2 program 1 with volume 108
    And channel 1 chorus is on with rate 0.8 and depth 55
    And channel 1 delay is on with time 320 feedback 35 mix 30

  Scenario: Deleting a song cascades to its parts
    Given a song "Doomed" with 3 parts
    When I delete the song "Doomed"
    Then no parts remain for "Doomed"
