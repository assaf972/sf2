Feature: Per-keyboard chorus
  As a gigging keyboardist
  I want a chorus with Rate and Depth on each keyboard
  So that I can thicken a sound independently per controller

  Background:
    Given keyboard 1 has a chorus insert

  Scenario: Enable chorus and set parameters
    When I enable chorus on keyboard 1
    And I set chorus rate to 0.8 and depth to 55
    Then keyboard 1 chorus is enabled
    And keyboard 1 chorus rate is 0.8 and depth is 55

  Scenario: Chorus state is saved with the part
    Given chorus on keyboard 1 is enabled with rate 0.8 depth 55
    When I save the current part and recall it
    Then keyboard 1 chorus is enabled with rate 0.8 and depth 55

  Scenario: Chorus is independent per keyboard
    When I enable chorus on keyboard 1
    Then keyboard 2 chorus remains disabled
