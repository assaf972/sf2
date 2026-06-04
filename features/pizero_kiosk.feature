Feature: Pi Zero kiosk navigation
  As a gigging keyboardist on a Pi Zero 2 W
  I want a two-row menu with four knobs and 61-key part navigation
  So that I can run a set on a tiny screen and one keyboard

  Background:
    Given the app runs in pizero mode
    And a song with 7 parts, part 3 active

  Scenario: 61-key program buttons step parts
    When the controller sends Program Up
    Then the active part is 4
    When the controller sends Program Down
    Then the active part is 3

  Scenario: FX page banks four encoders to delay
    Given the FX page shows the delay group highlighted
    When encoder 1 changes to 420
    Then keyboard 1 delay time is 420
    When I press FX
    Then the chorus group is highlighted

  Scenario: Knobs stay live across screens
    When the volume knob moves to 70
    Then keyboard 1 volume is 70
