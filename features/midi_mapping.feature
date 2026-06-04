Feature: Manufacturer MIDI mapping
  As a gigging keyboardist
  I want my controller's knobs to drive volume, pan and effects
  So that I can tweak the rig without touching the screen

  Background:
    Given keyboard 1 uses the "M-Audio" mapping preset
    And keyboard 1 drives layer 0

  Scenario: Volume knob maps to channel volume
    When keyboard 1 sends CC 7 value 90
    Then layer 0 volume becomes 90

  Scenario: Encoder 1 maps to chorus depth
    When keyboard 1 sends CC 74 value 100
    Then layer 0 chorus depth becomes 100

  Scenario: Unmapped CC is ignored by the mixer
    When keyboard 1 sends CC 20 value 5
    Then layer 0 volume is unchanged
