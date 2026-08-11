• Config-driven —  config  object (lines 76-218) embeds all story data: chapters, map style,
animations, theme
• MapLibre GL — fixed map background with globe projection, inset map, markers
• Scrollama — scroll-driven scroll listener triggers camera flyTo and chapter transitions on
onStepEnter  /  onStepExit
• UI — draggable/resizable cards per chapter, toggle-able nav sidebar, start/end slides

Chapter flow on scroll enter:

1. Camera flies to chapter location
2. Markers reposition
3.  onChapterEnter  layer animations fire
4. Optional rotation starts

To add audio, the cleanest approach: add an  audio  field per chapter in config, and wire playback
into the existing  onStepEnter / onStepExit  Scrollama callbacks.

Here's what I'll implement:

1. Add  audio  (URL) and  autoPlayAudio  (bool) to config + each chapter
2. Create a singleton  HTMLAudioElement , play on chapter enter, pause on exit
3. Optional small audio toggle button in the card bar
4. Update both HTML and JSON with example audio URLs