Feature: MP3 backing track player
  As a gigging keyboardist
  I want to play a backing track with stop/start/pitch/loop
  So that I can perform songs that need backing

  Background:
    Given a loaded MP3 file of known duration

  Scenario: Start then stop
    When I press Start
    Then the player state is "playing"
    When I press Stop
    Then the player state is "stopped"
    And the playhead is at 0

  Scenario: Loop repeats at end of file
    Given looping is enabled
    When playback reaches the end
    Then playback restarts from 0
    And the player state is "playing"

  Scenario: Pitch shift changes resample ratio, not file
    When I set pitch to 2 semitones
    Then the resample ratio is approximately 1.1225
