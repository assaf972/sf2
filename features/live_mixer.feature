Feature: Live three-keyboard mixer
  As a gigging keyboardist
  I want per-keyboard sound, volume, pan and key tuning on the Live view
  So that I can balance three controllers at a glance

  Background:
    Given a loaded SoundFont
    And keyboard 1 is enabled with sound "Fender Rhodes"

  Scenario: Adjusting keyboard volume changes the channel CC7
    When I set keyboard 1 volume to 90
    Then the synth receives CC7 value 90 on channel 0

  Scenario: Key tuning shifts incoming notes by semitones
    Given keyboard 1 key tuning is +12 semitones
    When keyboard 1 plays note 60
    Then the synth receives NoteOn key 72 on channel 0

  Scenario: Pan is centre by default and maps to CC10
    When I set keyboard 1 pan to centre
    Then the synth receives CC10 value 64 on channel 0
