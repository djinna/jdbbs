// Sample book using the Protocolized Anthology template
#import "../templates/series-template.typ": *

#show: book.with(
  title: "Sample Anthology",
  subtitle: "A Protocolized Anthology",
)

// =============================================================================
// FRONT MATTER
// =============================================================================

#set page(numbering: "i")
#counter(page).update(1)

#half-title[Sample Anthology]

#title-page("Sample Anthology", subtitle: "A Protocolized Anthology")

#copyright-page[
  Copyright © 2024
  
  All rights reserved.
  
  Published by Example Press
  
  ISBN: 978-0-000-00000-0
  
  First Edition
]

#pagebreak()

#toc-heading

#toc-entry[The First Story][Jane Author][1]
#toc-entry[Another Tale][John Writer][15]
#toc-entry[Final Chapter][Alex Scribe][42]

// =============================================================================
// BODY MATTER
// =============================================================================

#pagebreak()
#set page(numbering: "1")
#counter(page).update(1)

= The First Story

#first-para[
  This is the first paragraph of the story, which should have no indent. Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.
]

Subsequent paragraphs should have a 0.75em indent. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat.

Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident.

#section-break

#first-para[
  After a section break, we start fresh again without an indent. This paragraph begins a new section of the story.
]

The story continues with normal indented paragraphs. Sunt in culpa qui officia deserunt mollit anim id est laborum.

```
> TERMINAL OUTPUT
> Loading system...
> Ready.
```

After the code block, we continue with the narrative. The terminal flickered, casting an eerie glow across the room.

#poem[
  _Roses are red,_
  _Violets are blue,_
  _This is a poem,_
  _In italic too._
]

The poem hung in the air like a whisper.

= Another Tale

#first-para[
  A new chapter begins with a fresh first paragraph. The night was dark and stormy, as all good nights should be.
]

The protagonist walked through the rain, their footsteps echoing on the wet cobblestones. "Where am I?" they wondered aloud.

"You're in the story," replied a voice from the shadows. "Welcome."

#epigraph(
  [All that glitters is not gold;
  Often have you heard that told.],
  attribution: "Shakespeare",
)

The epigraph appeared like a sign, guiding the way forward.

= Final Chapter

#first-para[
  And so we come to the end. Every story must have its conclusion, just as every book must have its final page.
]

The characters gathered one last time, their arcs complete, their journeys ended. It had been a long road, but they had made it.

#sc[The End]
