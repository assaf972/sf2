Feature: Per-keyboard delay
  As a gigging keyboardist
  I want a delay with Time, Feedback and Mix on each keyboard
  So that I can add echoes to a lead independently per controller

  Background:
    Given keyboard 1 has a delay insert

  Scenario: Enable delay and set the three parameters
    When I enable delay on keyboard 1
    And I set delay time 320 ms, feedback 35, mix 30
    Then keyboard 1 delay is enabled
    And keyboard 1 delay time is 320, feedback is 35, mix is 30

  Scenario: Delay state is saved with the part
    Given delay on keyboard 1 is enabled with time 320 feedback 35 mix 30
    When I save the current part and recall it
    Then keyboard 1 delay is enabled with time 320 feedback 35 mix 30

  Scenario: Mapped encoders adjust delay live
    Given keyboard 1 maps encoder 2 to delay mix
    When keyboard 1 sends CC for encoder 2 value 64
    Then keyboard 1 delay mix is approximately 50
