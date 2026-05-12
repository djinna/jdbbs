# Clarification Questions for Book Production Pipeline

*Please edit your responses directly below each question and return this file.*

## 1. Style Naming Conventions

**Question:** Should custom styles follow a specific naming pattern? For example:
- `Custom_Letter` vs `Letter` vs `MS Letter`?
- Any prefix/suffix requirements?

**Response:** 
Let's start a pattern that begins a custom style with a 2- or 3-letter shortcode for the project. I'll favor lowercase conventions. Hyphens over underscores when needed. The tension btw terse and verbose will never go away but let's skew toward terse without being crazy about it. Max length of ~16 characters?
Show me a list of sample style names for special cases a book might have to review and tweak a little if necessary.
example: 
<vgr tweetbook 2026-04-04 0826.docx>


## 2. EPUB Enhancement Policy

**Question:** You mentioned EPUB can preserve more formatting than print. What's your philosophy on:
- Should we actively enhance EPUB (add colors, special fonts) when the print version is plain?
- Or keep both versions similar for brand consistency?

**Response:** 
Good questions: 
- for epub typography we've gone back to no custom type at all. We are committed to small file sizes.
- for images we will want the ability to drop in frontispieces, chapter opener images, and the images throughout the text (with and without captions). 


## 3. Custom Style Limits

**Question:** 
- Is there a maximum number of custom styles per project?
- Any styles that are absolutely forbidden?
- How do you handle conflicting requirements (author wants Comic Sans, house style forbids it)?

**Response:** 
Too many custom styles unlikely to be an issue. We'll move this consideration upstream into the developmental edit. For the three projects we're using to build this out, they each had a reasonable number of custom type styles to accomodate (be sure to review them as our baseline books: Ghosts in the Machine, The Librarians, Terminological Twists)


## 4. Template Versioning

**Question:** 
- If an author needs template adjustments mid-project, how is that handled?
- Do you version templates per project?

**Response:** 
Hmmm I suppose we will want to be able to pull an initial template, then allow for 1 or 2 revision cycles. 
I'm a fan of putting ISO date/time stamps in file names; the correct newest version is always the one with the most current date. Open to other suggestions if there's a nice pattern out there we could use instead.
example: 
<vgr tweetbook 2026-04-03 1426.docx>
<vgr tweetbook 2026-04-04 0826.docx>



## 5. Automated vs Manual Review

**Question:** For edge cases like colored text or manual formatting:
- Should the pipeline automatically strip/convert them?
- Or flag for editorial review?
- Different policy for EPUB vs print?

**Response:** 
Initially let's flag every edge case or possibly spurious local formatting for editorial review. Let's think about an usable simple interface for strip/keep decisions?


## 6. Font Handling

**Question:** I see default fonts are Libertinus Serif, Source Sans 3, and JetBrains Mono.
- Can authors request additional fonts?
- Are these embedded in EPUBs?
- Any web font licensing considerations?

**Response:** 
epub: see 2, no embedded fonts
print: the current defaults were provisional; suggest a small set of popular print typefaces for text and header typefaces that would pair well with each. We'll want to think about how to offer streamlined choices for many clients who won't have any type opinions while also being able to drop in custom choices for any given projects.
cf Vellum app which has been around for a good while and handles this idea of limited choices. This is the app we're hoping the .docx to Typst workflow replaces with a similar level of streamlined guardrails within which there is decent flexibility.
https://vellum.pub/
https://help.vellum.pub/styles/ < this covers how Vellum app handles styles
https://blog.vellum.pub/ < for ideas on how they've improved over time



## 7. Accessibility Requirements

**Question:** 
- Required alt text for images?
- Semantic markup requirements?
- EPUB accessibility standards you follow?

**Response:** 
Yes
Key aspects of the EPUB accessibility standards include:
	•	Accessibility Metadata: EPUB files must include metadata that describes the accessibility features of the publication. This includes information such as ⁠accessibilityFeature, ⁠accessMode, and potential hazards (⁠accessibilityHazard). This metadata helps users understand how accessible the content is and what alternatives are available for different types of disabilities . 	
	•	Content Accessibility: EPUB publications must be structured to ensure that all content is perceivable, operable, understandable, and robust. This means that the reading order must be logical, images must have alternative text, and navigation should be intuitive . 	
	•	Best Practices: To enhance accessibility, it is recommended that EPUB creators follow additional best practices, such as providing a table of contents, ensuring proper heading structure, and including navigation aids like lists of figures and tables . 



## 8. Series Consistency

**Question:** For books in a series:
- Do they share a template?
- How do you ensure typography consistency?
- Central series style management?

**Response:** 
Yes, our ms transmittal form ask if a book is part of a series; if yes, we'll want a way to reuse the epub and print settings.


## Additional Notes

*Space for any other clarifications or context you'd like to add:*
