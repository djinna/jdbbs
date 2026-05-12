#import "/templates/series-template.typ": *

#show: book.with(
  title: "TITLE",
  author: "AUTHOR",
)

= The Twitter Years
<the-twitter-years>
by Venkatesh Rao

= Preface
<preface>
This book is an attempt to capture the essence of my 15 years as an
active twitter user (I\'m going to use the lowercase spelling except
when referring to named subcultures within twitter), under the handle
\@vgr, in a form that does not entirely murder the spirit of the live
experience of being there, enmeshed in hundreds of live-wire
conversations unfolding over years, through an era when the platform was
the place the narrative of our world unfolded. In the chapters that
follow, you\'ll find a compendium of a few hundred of my best single
tweets (Chapter 1), and 101 of my best threads (Chapters 2-102). That\'s
a small fraction of the 150k+ tweets I posted through the years this
book covers, but hopefully it\'s an interesting distillation. I\'m still
on there, though I mostly only browse the feed. I no longer post
actively except for the rare boost of stuff I, or friends, are up to
elsewhere.

=== Body Text (Normal)
<body-text-normal>
This paragraph uses the Normal style — the default for body text. Font:
Libertinus Serif at 10.0pt with 12.0pt line spacing. First-line indent:
7.5pt. Justified.

Subsequent paragraphs in the same section use Normal style with the
first-line indent. Only the first paragraph after a heading or break
should use \'First Paragraph\' (no indent).

=== First Paragraph
<first-paragraph>
Apply \'First Paragraph\' to the first paragraph after any heading or
section break. It is identical to Normal but has no first-line indent.

=== Block Quote
<block-quote>
Use \'Block Quote\' for extended quotations:

#blockquote[
This is a sample block quotation. It is indented from the left margin and set in italic (per your spec). Use this style for any quotation that runs longer than a few words.
]
=== Code Block
<code-block>
Use \'Code Block\' for terminal output, code listings, or monospaced
content:

\$ echo \'Hello, World!\' \
Hello, World! \
\# Font: JetBrains Mono at 8.0pt

=== Section Break
<section-break>
Use \'Section Break\' to mark scene or section divisions. Your spec uses
the \'breve\' symbol:

˘

The paragraph after a section break should use \'First Paragraph\'.

=== Verse / Poem
<verse-poem>
Use \'Verse\' for poetry or lyrics. Each line is a separate line within
one paragraph (Shift+Enter for soft returns):

Shall I compare thee to a summer\'s day? \
Thou art more lovely and more temperate.

=== Epigraph
<epigraph>
Use \'Epigraph\' for chapter-opening quotations:

In the beginning was the Word. \
— John 1:1

=== Copyright
<copyright>
Use \'Copyright\' for the copyright page:

Copyright © 2025 Venkatesh Rao. All rights reserved. \
Published by Ventatesh Rao. \
No part of this book may be reproduced without permission.

== Custom Styles
<custom-styles>
#strong[tweet-p:] tweet

This paragraph uses the \'tweet-p\' style.

#strong[metadata-c (character style):] sample text — tweet metadata if
run-in

#strong[metadata-p:] tweet metadata if standalone

This paragraph uses the \'metadata-p\' style.

== Style Summary
<style-summary>
#align(center)[#table(
  columns: 3,
  align: (col, row) => (auto,auto,auto,).at(col),
  inset: 6pt,
  [Style Name], [Font / Size], [Usage],
  [Normal],
  [Libertinus Serif, 10.0pt],
  [Default body text with indent],
  [First Paragraph],
  [Libertinus Serif, 10.0pt],
  [After headings/breaks (no indent)],
  [Heading 1],
  [Source Sans 3, 16.7pt],
  [Chapter titles],
  [Heading 2],
  [Source Sans 3, 13.3pt],
  [Sub-sections],
  [Heading 3],
  [Source Sans 3, 10.0pt],
  [Sub-sub-sections],
  [Block Quote],
  [Libertinus Serif, 10.0pt italic],
  [Extended quotations],
  [Code Block],
  [JetBrains Mono, 8.0pt],
  [Code / terminal output],
  [Section Break],
  [Centered, \'breve\'],
  [Scene / section divider],
  [Verse],
  [Libertinus Serif, 7.5pt italic],
  [Poetry / lyrics],
  [Epigraph],
  [Libertinus Serif, 10.0pt italic],
  [Chapter-opening quotation],
  [Copyright],
  [Libertinus Serif, 8.0pt],
  [Copyright page text],
  [tweet-p],
  [paragraph],
  [tweet],
  [metadata-c],
  [character],
  [tweet metadata if run-in],
  [metadata-p],
  [paragraph],
  [tweet metadata if standalone],
)
]

= Preface
<preface-1>
This book is an attempt to capture the essence of my 15 years as an
active twitter user (I\'m going to use the lowercase spelling except
when referring to named subcultures within twitter), under the handle
\@vgr, in a form that does not entirely murder the spirit of the live
experience of being there, enmeshed in hundreds of live-wire
conversations unfolding over years, through an era when the platform was
the place the narrative of our world unfolded. In the chapters that
follow, you\'ll find a compendium of a few hundred of my best single
tweets (Chapter 1), and 101 of my best threads (Chapters 2-102). That\'s
a small fraction of the 150k+ tweets I posted through the years this
book covers, but hopefully it\'s an interesting distillation. I\'m still
on there, though I mostly only browse the feed. I no longer post
actively except for the rare boost of stuff I, or friends, are up to
elsewhere.

Through 2007-22, twitter was neither the biggest social media site, nor
the most representative. But it was the most consequential place, not
just on the internet, but arguably the planet. Entire political regimes
rose and fell, careers were made and destroyed, vast cultural movements
arose and died down. Other platforms may have featured more aggregate
activity, and accumulated orders of magnitude more social dark matter,
but twitter was where events broke into the main currents of history.
The suburbs of reddit may have accumulated deep intelligence, but
twitter was where some were promoted to historic consequentiality.
Facebook ads and groups may have shaped elections, but twitter was where
we collectively decided what it all meant. YouTube might have been where
endless warrens of conspiratorial imaginaries were constructed, but
twitter was where we determined which ones were going to shape the
almighty Discourse. 4chan might have produced many world-changing memes,
but twitter was where that world-changing actually played out.

But twitter was more than a distribution zone for culture manufactured
elsewhere. Increasingly, it became the site of cultural production. As a
blogger who initially signed up to promote my posts on Ribbonfarm, I
initially thought of it as a successor to RSS. Dumb pipes, just
stochastic rather than deterministic. But it quickly became clear that
was an absurdly bad mental model. Twitter was the tail that would come
to wag the rest of the social-media dog. You can read my very early
2007-era understanding of twitter in this old blog post, The Twitter
Zone and Virtual Geography. Now, nearly 20 years later, that mental
model feels, not wrong per se (it was sophisticated for its time), but
charmingly naive. What we thought was a low-stakes global office
watercooler turned out to be the site of future epistemic world wars, in
which the fates of civilizations would be decided.

Speaking of the Discourse, that\'s what we inveterate, degenerate, Very
Online people were there for. Hopelessly addicted to the almighty feed,
and the dopamine hits of likes and retweets from our parasocial boos and
senpais. And atop that milieu, the grand narrative of history could play
out as powerfully as it did because the company\'s management through
those critical years exhibited exactly the right level of public-sector
style benign neglect and managerial ineptitude to allow unfettered and
imaginative social intercourse to flourish. Twitter became the de facto
town square of the world through those years because it was too big to
properly manage except through benign neglect as a commons, even as it
remained too small to turn a profit. In theory a small, haphazardly
managed private corporation. In practice, a large public commons. The
very strategic errors that prevented the company from realizing its
obvious potential as a profit-making entity provided the affordances
that made it such a wonderful space for a true public Discourse to take
root.

My early twitter years, approximately 2007-2013 or so, were rather
unremarkable (and you\'ll note that there isn\'t much content from that
period -- I\'ve included a few of my earliest tweets as a prehistory
section in Chatper 1, but they are clearly only of historical interest).
During those years, the platform was sporadically the site of
world-changing events (\#iranelection being a prominent early example),
but not home to the grand narrative of the world in any sense. Familiar
features like \@ replies and quote tweets were not yet officially
supported. But by 2013 or so, the platform was nearly fully baked. Just
in time for mainstream arrival with the American culture war in 2013,
featuring such notable events as Gamergate, and the events in Ferguson,
which led to movements like BLM and NRx leveling up politically. It was
also during that period that various sectoral \"twitters\" took root. We
had VC Twitter, EconTwitter, Black Twitter, and so forth.

By 2015, the platform had become a variegated landscape of well-defined
subcultures. Importantly, however, the platform remained what my old
buddy Xianhang Zhang dubbed a plaza rather than a warren. The
subncultural fragmentation did not acquire legible organization or
destroy the one-big-public-square feel. This allowed thousands of
illegible subcultures to flourish, such as Weird Sun twitter, and the
mischievously vaguecronymed \"UST\" that grew around my old blogging
comarade-in-arms on Ribbonfarm, Sarah Perry (UST supposedly stood for
Unaffiliated Sloth Twitter, but nobody knows). Sarah was also the
instigating figure around what came to be known as postrat twitter.
Personally, I consciously stayed on the margins of these lively
communities. During these middle years (roughly 2014-18), my blog
Ribbonfarm was something like a insider\'s landmark for some of these
subcultures. Many people I first met on twitter during these years went
on to contribute to ribbonfarm, and participated in the meetups and
annual event (Refactor Camp) I used to run during these years. By the
time the late period began (roughly 2018 or so), however, the
subcultural ferment had largely settled. Ossified, if still illegible,
boundaries emerged. One notable and notorious subculture (partly a
descendant of postrat twitter) that emerged towards the end of the
middle years was tpot (\"this part of twitter\"). Ribbonfarm. primarily
thanks to my long-time co-blogger Sarah Perry, became somewhat awkwardly
retconned into the founding mythos of this subculture. I personally had
(and have) many friends within tpot, but never saw myself as part of it.
Largely because I felt I was a generation removed, as well as
politically a world away. I suppose I was a sort of distant uncle to the
scene.

These middle years also saw the more legible and well-lit parts of
twitter become increasingly prominent in the mainstream culture, as
politicians, journalists, and business leaders realized that twitter was
where the public first response to everything of consequence unfolded,
generally days to weeks before it happened anywhere else. The part I was
most associated with was probably Silicon Valley and VC/startup twitter.
One event that radically altered my experience of twitter was briefly
working with Marc Andreessen (\@pmarca) during 2014-2015 at a16z (a gig
I landed via Chris Dixon, \@cdixon, via twitter). The influence was
twofold. First, there was the direct influence. Being visibly associated
with a huge account, through public banter, rapidly 3x-ed my following.
I used to refer to the phenomenon as a \"gravity assist.\" Second, Marc
was responsible for perhaps the most consequential innovation on twitter
during those years: the notorious tweetstorm,, which was the colorful
term for what later got domesticated as the threading feature. When Marc
first started tweetstorming, it was at once hugely entertaining and
radically annoying at the same time, because his tweets would swamp and
murder everybody\'s feeds. The platform quickly adapted and began to
support a more usable threading UX, but for a while there, it was
anarchy, as hundreds of others jumped into the tweetstorming game,
myself included. Threading was a beautiful thing, allowing people to
workshop complex arguments in real-time with a lively audience, in an
unbundled form suitable for spawning rich comment threads and
quote-and-fork side trails. I quickly became a huge fan of the format,
despite its potential for being annoying. Not only did I start doing a
lot of threads on twitter, I used the thread format in my newsletter for
a while, and adapted it to serialized blogging in a format I called
blogchains. Though Marc and I have since diverged politically, we remain
on cordial terms at a personal level. And regardless of what you might
think of him, history will definitely remember him for at least two
remarkable inventions -- the browser, and the tweetstorm.

No mention of threading culture is complete without a hat-tip to Visakan
Veeraswamy, (\@visakanv) of course. Visa took the basic linear threading
idea pioneered by Marc and turned it into a dizzying artform, turning
his account into a tangled, densely interlinked, quote-linked,
promiscuously forking Lovecraftian monstrosity of a twitter hyperobject.
I came up with a term for it: threadthulhu (my main contribution to
culture through the twitter years was coming up with names for things).
My own threadthulu was only middling crazy. Orderly enough that I was
able to index all my good threads in a meta-thread over the years, and
slaughter it relatively cleanly to create the raw material for this
book. I doubt Visa\'s insane threadthulu can be killed at all, let alone
properly butchered into a book-like echo like this one. I vibecoded the
pipeline that created this book, but it will probably take AGI to
similarly tame Visa\'s threadthulu. I kid of course. AGI is conceptual
nonsense. Visa\'s threadthulu will remain forever untamed. If X dies, it
will ascend into the latent spaces of various LLMs as an eternal
monster.

Threading culture reached its apogee with an event I had a hand in
instigating, called Threadapalooza, in December 2019, where we all egged
each other on to post 100-tweet monster threads. I think I got the idea
because I was tired of seeing people complaining about long threads, and
twitter being the transgressive place it was, the obvious thing to do
was to annoy them more by instigating so many huge threads, you could
drown out their whining. The event was huge fun, and went through a
couple more annual iterations. It was one of the last times I truly had
fun on twitter. Sadly, threading culture was later superseded by the
significantly inferior screenshot essay form (no way to quote bite-sized
fragments for dunking commentary or reply to a specific part of the idea
easily), and eventually killed off entirely in favor of long posts by
paying customers of Elon Musk\'s pay-to-play X. But before we get to
that, perhaps a word or two on the cultural evolution of twitter up the
change of ownership.

The long arc of twitter\'s history as the site of the Discourse, roughly
2014-22, could be divided into 3 chapters:

- 2013-2016: Hypomanic culture-chaos with a side of culture war -- a
  phase that ended with the death of Harambe.

- 2016-2020: Internet of Beefs -- which began with the rise of Trump and
  ended with Covid.

- 2020-2022: Covid era -- which had very special characteristics and
  ended with the acquisition by Musk.

During the Internet of Beefs years, twitter gradually became less fun
and more consequential. More and more people simply retreated, unable to
handle the toxicity. One of my threads, included in this book, Against
Waldenponding, captured my own arguments against retreating. That
argument proved to be quite popular, and I think what happened next
demonstrated why it was in fact correct. When Covid hit, in early 2020
(shortly after the first Threadapalooza iirc), it was twitter\'s time to
absolutely shine. Twitter figured Covid out faster and better than any
other public place on the planet, online or offline. And saved countless
lives I suspect. I myself was able to react weeks ahead of most people
in my non-twitter life. While some individuals, like Balaji Srinivasan
(\@balajis) played individually influential roles in triggering the
effective early response, the truly impressive part was the way the
wisdom of the crowds served as an accelerant. Every aspect of the
pandemic got swarmed and figured out faster and better than anywhere
else, so long as you were plugged into the right subcultures. Of course,
there were also plenty of dumb subcultures getting things wildly wrong.
I look back on twitter through the Covid years as humanity at its best
in many ways. Though there was a great deal of toxicity and awfulness,
there was also a remarkable amount of unprecedented positive power on
display.

At the same time, there was no denying that the large-scale retreat was
very real, and consequential. Whether or not I advocated staying plugged
into the heady drug that was twitter, more and more of the most
interesting people were choosing not to, and going dark. I first flagged
this in a tweet of course:

Eff it, Yolo, I\'m not waiting to see how Gen Z actually shapes up to
tag them. Where\'s the fun in data-driven hindsight. I\'m calling it
early based on early markers and portents If Millennials are Premium
Mediocre... Gen Z is gonna be Domestic Cozy Yeah you heard it here first

// Editorial review: manual list item detected; keep/strip/convert decision required
I theorized and speculated about this trend more throughout 2019, on my
blog and in newsletter, which also spawned one of my last major viral
memes before the era of viral memes basically ended: the cozyweb

// Editorial review: manual list item detected; keep/strip/convert decision required
I decided to quit twitter pretty much the moment it became clear that
the high farce of Elon Musk\'s acquisition bid would actually turn into
reality. I documented my thinking and decision in a two-part essay
series in my newsletter (Part 1, Part 2), so I won\'t bother rehashing
that story here. Suffice it to say that things unfolded exactly as I
expected, and I\'m happy I went dark when I did. Some have since told me
they think I was prescient in making a clean break and moving on so
early, but there was no prescience involved. The writing on the wall was
crystal clear to anyone willing to read it. I think what really made it
an agonizing decision for many was leaving behind large followings and
rich networks. Personally, I suffered zero agonies. Indifference to sunk
costs is one of my core operating principles in life, and I\'ve never
had any regrets about walking away from seemingly priceless stores of
social capital. I\'ve walked away from many milieus, online and offline,
and I\'m sure I\'ll walk away from many more in the future. To the
extent a ghost of twitter still exists, buried beneath the superficially
familiar, but radically different beast that is X, I hang out enough as
a silent reader to keep up with a few things (particularly robotics and
AI news). But I have no desire to become an active part of the
conversation or milieu there anymore, and largely tune it out. To
Musk\'s credit, lately the For You feed appears to have picked up on my
preferences and mostly shows me robotics tweets instead of ragebait.
Maybe there\'s a slim chance of redemption yet, and those of you pining
for the old twitter might yet see your hopes fulfilled. I have, frankly,
moved on.

For some of you, this may be the first you\'re hearing from me in years.
I think at least a few people on twitter have assumed I\'m dead or
retired. When I search for \"\@vgr\"\" or \"ribbonfarm,\" which I do
periodically (I have my little vanities), the results are mildly
hilarious, like being at my own interminably extended funeral. An
agonizingly drawn-out social death that I\'m fortunately not there to
actually suffer. These days, I mostly hang out on Substack Notes and
Farcaster as far as public feeds go (I\'m also on Bluesky, but not
particularly active). But at 51, I find my appetite for the intensity of
the public short-form game is slowly declining. So much of my attention
and social energy now flows through a couple of Discords and Slacks, and
a bunch of group chats. I don\'t think digital public spaces of the sort
twitter represented for a few years are ever coming back. But while it
existed, twitter was a once-in-history zone of pure magic. I\'m glad I
got to be part of it. No book-like artifact can ever capture even a
fraction of that live magic, but I hope this one at least gives you some
sense of what it was like for me, living, blogging and shitposting
through those years.

= 1. Singles
<singles>
// Editorial review: manual list item detected; keep/strip/convert decision required
A compendium of noteworthy standalone tweets, organized chronologically.
While most of this book is threads, this opening chapter contains a
curated selection of stand-alone tweets that were either exceptionally
popular, personal favorites, or both.

== 2007
<section>
Pushing my cat away. When you give him wet food he wants dry, when you
give him dry food he wants wet.

Aug 29, 2007 (prehistory)

Hmm... just finished thorough read of Dan Pink\'s Free Agent Nation.
Mull mull. Seinfeld on. Watch or mull more?

Aug 29, 2007 (prehistory)

Darn, this twitter thing is neat. Only question is, do I want to pay for
texting?

Aug 29, 2007 (prehistory)

\"The compensation for an early success is the conviction that life is a
romantic affair\" -- F. Scott Fitzgerald. Hmm.

Aug 29, 2007 (prehistory)

Woohoo can twitter inside firewall now via cellphone

Aug 29, 2007 (prehistory)

Milk marked Sept 20 expiry spoiled already. Back to soymilk vegan coffee
after dairy break. Cat has been on keyboard again. Damn hairs.

Aug 29, 2007 (prehistory)

Favorite coffeeshop \"Earthtones\" closed due to movie filming :( Hadta
go to second favorite, \"Mona Lisa cafe\"

Aug 29, 2007 (prehistory)

Reading against the gods at starbucks

Aug 29, 2007 (prehistory)

carpe diem, carpe mouse, carpal tunnel. ouch.

Aug 30, 2007 (prehistory)

Caprese lunch

Aug 30, 2007 (prehistory)

== 2014
<section-1>
The future is already here, it\'s just unequally distributed. 1% of the
population lives in 99% of the future. \#OccupyFuture

Jul 10, 2014 (joke)

Reality is that which, when you run out of money, you can no longer
ignore.

Jul 14, 2014 (joke)

The key is to do half-assed things in a full-assed way rather than
full-assed things in a half-assed way.

Oct 9, 2014 (aphorism)

Sufficiently enlightened self-interest is indistinguishable from
altruism. Suff. clueless altruism is indistinguishable from
self-interest.

Nov 7, 2014 (aphorism)

Expecting to be treated \"fairly\" by a friend is reasonable. By a
stranger: idealism. By an enemy: ideology. By nature: insanity.

Dec 17, 2014 (aphorism)

== 2015
<section-2>
Any sufficiently advanced kind of work is indistinguishable from play

Feb 28, 2015 (aphorism)

// Editorial review: manual list item detected; keep/strip/convert decision required
I made up a new, data-driven insult for people who generalize too much
on the basis of just their own experiences: \"Go away, n\=1.\"

May 12, 2015 (joke)

Selfie sticks are an awful technology. Use a proper drone instead.

May 20, 2015 (joke)

Well, looks like software is about to eat the first serious female
presidential campaign.

Aug 12, 2015 (take)

Half of advanced communication skills is learning sophisticated ways to
tell people they\'re wrong while letting them save face gracefully.

Aug 18, 2015 (advice)

Anything is possible if you\'re willing to not take all of humanity
along for the ride.

Aug 19, 2015 (reflection)

Past 30, if you don\'t break script every 3-5y, life becomes 99.9%
maintenance, 0.1% growth. Some actually think this is the goal state. 😱

Sep 5, 2015 (reflection)

We need 3 planets for an agile, continuously integrated civilization:
dev-earth, test-earth, production-earth.

Oct 5, 2015 (joke)

// Editorial review: manual list item detected; keep/strip/convert decision required
I have decided my life\'s work is fighting against the growing evil of
tribalism. You\'re either with me or against me in this righteous cause

Nov 1, 2015 (joke)

Ironically, it is easiest to do business on a handshake in places where
contracts are enforceable and routine

Nov 5, 2015 (reflection)

Dumb idea: \"When the student is ready, the teacher appears.\" And
lesson is too late. Jump in before you\'re ready. Don\'t wait for a
teacher.

Nov 28, 2015 (advice)

== 2016
<section-3>
These days, \"must-read\" books feel like the Windows updates of life.

Jan 9, 2016 (take, joke)

// Editorial review: manual list item detected; keep/strip/convert decision required
I was promised an FTL time-space ship with force fields and artificial
gravity and all I\'m getting is recycled rockets

Jan 23, 2016 (joke)

Occam\'s Razor, now with 6 blades and soothing aloe strip

Jan 26, 2016 (joke)

Paradox: almost nobody likes modernists (I define as \'people who like
today better than yesterday by default\'), but they win 99% of time

Jan 29, 2016 (aphorism, reflection)

First they ignore you, then they ignore you, then they ignore you, then
they ignore you, then you die.

Mar 9, 2016 (joke)

Talent hits the target others can\'t hit. Genius hits the target others
can\'t see. Insanity hits the target that doesn\'t even exist.

Mar 17, 2016 (joke)

The end-state of being eaten by industrialization is commoditization.
The end-state of being eaten by software is artisanization.

Jun 1, 2016 (reflection)

Fun fact: we\'ve learned to reuse rockets instead of dumping them in the
ocean before we\'ve learned that trick for plastic water bottles.

Jun 21, 2016 (reflection)

Moral hazard is when something feels like a video game to you, but is
life and death to people affected by your decisions

Jul 17, 2016 (reflection)

Dumb people become potential problems when they have nothing to lose.
Smart people become potential problems when they have nothing to gain.

Aug 2, 2016 (reflection)

== 2017
<section-4>
Knowledge depreciation and unexpected life turns make us all autodidacts
by 30. So might as well get good at it when young.

May 8, 2017 (reflection)

Freedom is arguably freedom to go deep: into rabbitholes, relationships,
missions, w/o regard to cost, reward, time, productivity, risk

Jun 27, 2017 (reflection)

Analogue to fuck-you money is fuck-you competence. Aspiring to
jack-of-all-trades-master-of-all because you don\'t like dependence on
others

Jul 1, 2017 (reflection)

It still boggles my mind that all life except for one species gets along
fine without money. Where did we go wrong? 🤔What do they know?

Aug 2, 2017 (reflection)

Great minds discuss ideas; mediocre minds discuss events; small minds
discuss people. Premium mediocre minds discuss bitcoin

Aug 14, 2017 (joke)

// Editorial review: manual list item detected; keep/strip/convert decision required
I just want to live a simple, free life, among reasonable, diverse
people, aboard a starship armed with photon torpedoes. Too much to ask?

Aug 16, 2017 (joke)

What is most depressing about 2017 is we don\'t know if it\'s year 1 of
an 8 year aberration, an 80 year macro trend or an 800 year dark age 😳

Aug 19, 2017 (reflection)

Even a shallow understanding to blockchains makes you see them
everywhere. Makes you realize how much of civilization is just
book-keeping.

Aug 19, 2017 (reflection)

Advice to the young: never buy a couch. That\'s when it all starts to go
south on you, when you first buy a couch.

Aug 29, 2017 (advice, joke)

\"Artificial Intelligence\" is a \"horseless-carriage\" grade
anachronism. I hereby declare AI now \= \"Alienized Infrastructure\".
You\'re welcome.

Sep 13, 2017 (joke)

The most common kind of execution failure is one we rarely talk about:
quite simply hating the work too much to do it even passably well.

Sep 21, 2017 (reflection)

Those who ignore history are doomed to repeat it. Those who study
history are doomed to repeat it with elaborate justifications.

Sep 29, 2017 (joke)

// Editorial review: manual list item detected; keep/strip/convert decision required
4 game theory games everybody should know about: prisoner\'s dilemma,
ultimatum, dictator, public goods. Basic literacy on impersonal trust.

Oct 1, 2017 (aphorism, advice)

Most disagreements with people are about dislike of personality traits.
If you like someone, you\'ll mostly deal with their beliefs.

Oct 14, 2017 (aphorism, reflection)

// Editorial review: manual list item detected; keep/strip/convert decision required
I don\'t trust anyone who can\'t amuse themselves indefinitely with
their own thoughts. They invariably cause trouble outside their heads.

Nov 4, 2017 (reflection)

Theory If a military conflict lasts longer than 3 yrs, economic strength
determines outcome. If an economic conflict lasts longer than 30 yrs,
ideological superiority determines outcome. If an ideological conflict
lasts \>300 yrs, technological generativity determines outcome

Nov 11, 2017 (aphorism)

The closer you get to leisure class the more ambitions barbellize:
trivial, like seeking out the best chips, or save-the-world. Nothing in
between. Does not bode well for postscarcity utopian dreams. They will
be unstable.

Nov 16, 2017 (reflection)

// Editorial review: manual list item detected; keep/strip/convert decision required
A tribe is a mechanism for socializing guilt. This is why guilt by
association is a surprisingly effective tactic for undermining them. The
price for the benefits of tribal membership is accepting complicity in
its historic burden of sins.

Nov 19, 2017 (reflection)

Sometimes I get distressing glimpses of how old age is going to suck.
Your future is already here. It\'s just unevenly distributed across your
body.

Nov 22, 2017 (reflection)

Many powerful people think they have a good balance between an inner
circle that provides access, information and intellectual support, and a
bouncer circle that blocks threats, people and noise. Many actually only
have a pure bouncer circle and don\'t realize it.

Nov 23, 2017 (reflection)

History is going through a major cache invalidation period

Nov 25, 2017 (joke, take)

// Editorial review: manual list item detected; keep/strip/convert decision required
I saw the premium mediocre minds of my generation vanish into darkness,
past the event horizons of their own consensual hallucinations, falling
through the twitter feeds towards light, looking for an exit

Dec 10, 2017 (joke)

Fuck-you money is 3 things: a part is just reward for your efforts, a
part is hush money from The System, and the rest is the universe just
paying you to go away because you\'re annoying. Reflect on the
proportions in your case, ye rich

Dec 16, 2017 (reflection)

Most serious and consequential debates are decided by one side running
out of stamina and the other winning by default. When something actually
matters to you, staying in the debate is more important than winning or
losing particular points.

Dec 16, 2017 (reflection)

Conservativism has to reproduce at higher than replacement rate to
persist because it constantly loses people to progressivism, net.
Progressivism is sustainable at lower-than-replacement-rate fertility
because it generally gains converts, net. Memes eat genes

Dec 16, 2017 (reflection)

Many good things in my life have been due to being too lazy to make
active decisions and letting defaults ride. The better part of strategy
is knowing when to do nothing. 9/10 times, the strategic and lazy
choices are the same. Get the 1/10 right though.

Dec 20, 2017 (advice)

It takes many kinds of lives to make a world Most lives aren’t simple,
whether successful or not But the lessons of simple lives, successful or
not, spread more easily So people end up believing all lives are simple
And so end up surprised by the complexities of their own

Dec 23, 2017 (reflection)

The more I learn about pigs, the more I feel bad about using ‘pig’ as a
human insult There are really no good animal comparisons for the worst
things humans do. Except maybe a few of the things really smart ones
like chimps and dolphins do.

Dec 23, 2017 (reflection)

// Editorial review: manual list item detected; keep/strip/convert decision required
I feel bad about reading Wikipedia summaries of vast histories, quick
glosses of big fat books, plot summaries of novels, etc. But
increasingly I find it’s the only way to get anything done. Full
read\=luxury I’d pay for a custom summaries service with my priorities
in mind

Dec 25, 2017 (reflection)

// Editorial review: manual list item detected; keep/strip/convert decision required
I think a century or so from now, people will view national citizenship
and its mobility constraints the way we view serfdom today 🤔

Dec 26, 2017 (reflection)

MyTurnism: If your political ideology is 100% driven by fighting
oppression and exploitation when not in power, your governance will be
100% driven by extraction and profiteering when in power. Purely
palliative political ideologies inevitably turn corrupt at the first
chance

Dec 29, 2017 (reflection)

== 2018
<section-5>
Well, it’s official, crypto is going to be the next big tech China has
tried to ban it Uber drivers are offering tips about it Buffett has
decided to sit it out

Jan 11, 2018 (take)

My evil editor changed absolutely everything in my Ship of Theseus essay
but fortunately the basic idea was preserved

Jan 14, 2018 (joke)

// Editorial review: manual list item detected; keep/strip/convert decision required
A robust truth is one that is strengthened by every bit of added context
If adding context weakens your assertion steadily until it dissolves
into noise, you basically lied

Jan 15, 2018 (advice, reflection)

// Editorial review: manual list item detected; keep/strip/convert decision required
A middle-class trait I can’t shake off is slight guilt about playing
with no-practical-value ideas during “work hours”, like reading a
philosophy book 9-5 on a weekday. Feels like playing hooky. One reason I
blog is to manufacture practical justification. “It’s for a post”

Jan 16, 2018 (reflection)

It seems serious luxury car nuts have the same neural activity in
response to cars as the rest of us do when identifying human faces.
Explains so much about any kind of unique/luxury consumption. They are
substitutes for human connection. One of my big mysteries, solved.

Jan 18, 2018 (take)

// Editorial review: manual list item detected; keep/strip/convert decision required
I read far fewer books these days, but the ones I do read usually end up
being magic bullets, precisely curing a particular mental block. 10
years ago, books were more like broad-spectrum antibiotics. 20 years
ago, they were more like food. 30 years ago, they were placebos.

Jan 21, 2018 (reflection)

Perhaps the best measure of leaders is how their reports interact with
each other and outsiders when they are not in the room. Every leader is
a ghost in a machine they build (often for long after they die). How you
haunt it matters more than how you operate it directly.

Jan 22, 2018 (reflection)

An ounce of physiological insight generally eliminates about a pound of
psychological BS “I have anxiety and depression from work stress; it
must be this bad quarter and that delayed project” #strong[drinks some
water] “Wait scratch that, I was just dehydrated”

Jan 24, 2018 (advice, joke, reflection)

Stages to doing good work 1. Laugh at bullshit 2. Get mad at bullshit 3.
Try to get even by joining the bullshit factory yourself 4. Realize in
the long term it’s actually harder 5. Start to do good work because it’s
actually the laziest way out of the maze

Jan 27, 2018 (reflection, advice)

The opposite of bullshittiness is not authenticity or sincerity but
impatience. You can always be fooled by levels of bullshitting that are
more sophisticated than you have encountered before, but your own “life
is too short for this” reaction will point to the way out.

Jan 27, 2018 (reflection)

// Editorial review: manual list item detected; keep/strip/convert decision required
A city-state world where people are born stateless, live with family
till 14, & can go walkabout wherever they like 14-18 at which point they
auction themselves off to cities based on \# of existing citizens who
want them there. Like NFL draft but for cities and everybody. 🤔

Feb 4, 2018 (reflection)

An interstellar spacefaring civilization would most likely be socialist
inside the spaceships, libertarian-monarchic on space stations, anarchic
between ships/stations, liberal democratic within inner solar systems
(say 1-2 lighthour sphere)

Feb 7, 2018 (reflection)

Capitalism is based on owning something absolutely in a timebox,
regardless of history/future. I suspect postcapitalism will unbundle the
idea of \'asset\' in time. You\'ll own a thread of its history rather
than a timebox cut of its existence. #strong[cough] blockchain
#strong[cough]

Feb 8, 2018 (reflection)

What would happen if parents only gave children temporary code names and
you had to choose your own name at 18? It is really weird that we don\'t
name ourselves, if you think about it.

Feb 8, 2018 (joke, reflection)

The greatest discovery in social psychology in the last century is that
it is funny when a cartoon character runs off a cliff but doesn’t fall
down until they actually look down. That joke explains more about human
group consciousness than most textbooks.

Feb 9, 2018 (joke, reflection)

Social Heisenberg uncertainty principle: you can never be 100% certain
of both your positions and principles. If your principles are crystal
clear, your political positions will be incoherent. If your political
positions are crystal clear, your principles will be incoherent.

Feb 11, 2018 (aphorism)

You probably have a base year for orienting your life even if you aren’t
aware of it. Explanations/valuations for major life events trace back to
it. You likely also periodically rebase to a year other than your birth
year. I’ve rebased thrice: 1999, 2009, 2015

Feb 12, 2018 (reflection)

Success is depressing because it teaches so much less than we expect it
to. Failure is depressing because it dumps more learning potential on us
than we can handle. Blend in the right proportion and you\'ll beat
depression. Non-depression indicates your optimal learning rate.

Feb 14, 2018 (reflection, advice)

There\'s a dozen people I\'d like to get to know better, hang out with,
have deeper chats with, collaborate with over many years. No 2 of them
live in the same city. No 3 in the same country. No 4 on the same
continent. The Internet is like a piquant whiff of an impossible life.

Feb 15, 2018 (reflection)

You have to be kinda happy to market well. At the risk of sounding
sappy, you need a certain lightness in your heart. Which is why
shallowness can help you fake it (but not truly make it). People who
feel the world’s burdens resting heavily on their shoulders cannot
market well.

Feb 16, 2018 (reflection)

Far left hatred for crypto seems entirely out of proportion to severity
of stated concerns (bitcoin energy use, GPU scarcity). I suspect this is
because crypto is charismatic libertarian megafauna. Thin end of a wedge
threatening to undermine all collectivist political action.

Feb 17, 2018 (reflection, take)

Social media platforms basically gerrymander attention. Gerrymandering
flips democracies around so politicians choose their voters rather than
the other way around. Gerrymandering attention allows publishers to
choose their audiences rather than the other way around.

Feb 20, 2018 (take)

// Editorial review: manual list item detected; keep/strip/convert decision required
I just googled “flat earth jet lag” to see what flat earthers have to
say about it. How come there is no term like “shitsearching” to pair
with shitposting?

Mar 3, 2018 (reflection, joke)

Thought of a phrase: alt confidence. The generally higher sense of self
certainty that comes with adopting fringe views. I think it’s a
selection effect. People who can’t tolerate high uncertainty centrifuge
farther out to edge.

Mar 3, 2018 (reflection, take)

Alexa just told me “today you can expect dreary weather” Forget
malicious AI. Start worrying about Marvin-type depressed AI. This kind
of affective language is where it starts. She’s shutting the door to the
possibility that I might enjoy what she calls ‘dreary’ (though I won’t)

Mar 3, 2018 (take)

It’s funny how we use twitter for multiplexed culture-warring and
entertainment. It’s like if soldiers facing off in trenches alternately
shot at each other and yelled knock-knock jokes at each other. This
cyberpunk future has some basic architectural problems.

Mar 5, 2018 (reflection, take)

One reason I’m addicted to Twitter is that most of my curiosities are
extremely shallow. I’m content with the first demystifying perspective
on a question. I don’t need the best, most correct, and deepest answer.
Just a handle on the question with which to hold it correctly.

Mar 12, 2018 (reflection)

\"We\'re a republic not a democracy\" is the \"tomatoes are fruits, not
vegetables\" of political discussions.

Mar 13, 2018 (joke)

Trivial: imagining linearly extrapolated futures (eg: \"extreme
capitalism\") Easy: imagining nonlinearly extrapolated (eg yin-yang
cyclic) futures Hard: imagining evolved futures with mutations Really
hard: imagining futures being invented by actions of imaginative people

Mar 16, 2018 (reflection)

GOMO: Gratitude of Missing Out I have GOMO about not listening to
podcasts and thereby managing to miss what appears to be about 30% of
the zeitgeist conversation.

Mar 16, 2018 (take, joke)

There is a theory most people dismiss because it sounds like sour
grapes, but is mostly true. Most rich people are rich in part because
they wanted to get (or stay) rich more badly than others. This is a
boring life goal. As a result, most rich people are boring.

Mar 18, 2018 (reflection)

The adventurer’s ethic is probably this: you have an obligation to
become as interesting as your surplus resources permit

Mar 18, 2018 (reflection)

I’m beginning to think lack of capital is not really a factor stopping
people from doing the things they feel they were meant to. Most people
don’t have capital-intensive imaginations. If you give most people a
lump sum the best idea they’d likely come up with is buying a house.

Mar 27, 2018 (take)

Buying back your own time with money is like a stock buyback. Unless you
have a clear mission you’re freeing up time for, it’s an unimaginative
retreat, not a deployment of resources. Remaining shareholders
(you/family) get minor boost in existing kind of value.

Mar 27, 2018 (reflection)

Thinking startups are what make nations innovative is the macroeconomic
equivalent of thinking R&D labs are what make companies innovative. Both
are the same mistake. I call it the sequestration fallacy. Startups and
R&D are the consequence, not cause of innovative cultures.

Mar 29, 2018 (take)

The American chopper meme is the opposite pole from harambe. Instead of
perfectly distilled nihilism it’s like the Wikipedia of memes. Packed
with sincere good information and vitamins.

Apr 8, 2018 (take)

Annoying how much of maintaining mental health boils down to recalling
the right trite thought at the right time: This too shall pass Don’t
decide when depressed Drink water Drink Eat Sleep Take a walk Go gymming
Let this one go It’s only money Morning is wiser than evening

Apr 8, 2018 (reflection, advice)

War is negative-sum, not zero-sum. Peace is zero-sum. It’s for
vegetables. We don’t have a true antonym for war. It is not peace, but
some sort of coherent nonzero-sum flourishing. Like “tech boom”, “gold
rush” or “space age” Word proposals now being entertained

Apr 14, 2018 (reflection)

Those who don’t play with language are doomed to be enslaved by it

Apr 15, 2018 (aphorism)

In terms of sheer fraction of life hours transformed from meh to
pleasant, for the most people, TV is the greatest invention ever. Real
greatest good for the greatest number kinda deal.

Apr 17, 2018 (take)

People who grew up with Wikipedia seem far better informed than their
intellectual/personality peers from older gens at same age. Jevons
paradox. They know more #strong[because] they can look up anything. No
curiosity left unsatisfied. I’d estimate a 10y advantage in factual
knowledge

Apr 21, 2018 (take)

Take a moment to appreciate the story of on-demand streaming TV. One of
those rare technologies that everyone fantasized about, that arrived
pretty much on schedule as predicted, and turned out to be exactly as
awesome as we all hoped. Only unknowns were the brand names

Apr 21, 2018 (take)

Myers-Briggs is more useful than IQ. You can tell type by criticism it
evokes. For eg: ISTJ: “there is no statistical evidence for it” INFJ:
“Actually, I Ching and Enneagram are better” INTP: “Socionics is more
consistent” ESTJ: “Only Big 5 is Ministry-approved” ENTP: “BORING!”

Apr 21, 2018 (take)

There is something bizarre about “high-IQ” as an identity core like with
Mensa. Like a trade union of capitalists or a commune of libertarians.
Or that South Park joke about the anti-semitic sect of Judaism.
Shouldn’t you be winning Nobels, rather than performing identity?

Apr 22, 2018 (take)

We need a word for the vague sense of being continuously partially
cheated/taken advantage of because online, automated services are
impossible to reason with and escalating to humans with override
capability is like dragon slaying and almost never worth the effort

Apr 25, 2018 (reflection, take)

Offline we’re still in the fall of Rome. Online we’re already into
Italian city states. There’s an 800y cultural phase lag between online
and offline. Offline I’m some drunk jerk cheering and eating bread at
the Circus Maximus. Online I am hustling the city-state nobility

Apr 27, 2018 (take)

🤔: “Hmm, I’m thinking about \<territory\>” 🤡: “\<map\>, you’ve
discovered \<map\>”

May 1, 2018 (joke)

My ahead-of-curve scorecard, measured in terms of how much earlier than
mainstream I’d heard of various things NRx: 3 years Incels: 3 years
Crypto: 3 years Twitter: 5 years Turmeric in milk: 40 years Things I was
late to: fidget spinners, touchscreens, MP3s, avocados, stocks

May 4, 2018 (reflection)

Human condition pie chart 1%: making dents in universe, not going gentle
into that good night 9%: tormented by existential despair on edge
between meaning and nihilism 90%: mostly satisfied with life while
there’s stuff to watch on Netflix/Prime/Hulu

May 7, 2018 (joke, reflection)

// Editorial review: manual list item detected; keep/strip/convert decision required
A few years ago, my conversational rule of engagement was “joke around
with until proven capable of holding up their end of a serious
conversation” Now it is “engage humorlessly until proven capable of
being joked around with”

May 8, 2018 (reflection)

Before enlightenment: chop wood, carry water After enlightenment: move
back to the city, sign up for electricity and water, join a gym, and
download a meditation app

May 9, 2018 (joke)

Heh I’m in minority on robot/AI self-disclosure. I’d like them to
disclose they’re AIs if asked directly, but not be forced to volunteer
the info. I guess we’re in 3 laws of robotics territory. I for one, will
defend the rights of R. Duplex Olivaw to drop the R.

May 10, 2018 (take)

Every time I go do some birdwatching I\'m struck by how ridiculous it is
that we think dinosaurs are extinct. They just self-disrupted and
leveled up seriously as birds.

May 19, 2018 (reflection)

What if everybody in the world could vote in any country\'s elections,
with an exponentially decaying weight set by shortest distance from
boundary? So Chinese people would have say a 1/32 vote in American
elections? Fourier transform global politics.

May 23, 2018 (reflection)

When you want to let someone finish what they\'re saying but you
disagree with the whole premise of what they\'re saying, it feels like
going into disagreement debt. I often declare bankruptcy and mentally
exit before they finish talking. Body follows as soon as politely
possible.

May 23, 2018 (reflection)

Small brain: use map to navigate territory Regular brain: use territory
to validate map Planet brain: throw away maps, raw territory Galaxy
brain: use territory to navigate map

May 29, 2018 (joke)

// Editorial review: manual list item detected; keep/strip/convert decision required
I have the shortest version of my critique of GDPR It is continent-scale
virtue signaling

May 30, 2018 (joke, take)

The point of a rule is often to consolidate power through
exception-handling

Jun 1, 2018 (aphorism, reflection)

Signaling values and signaling status are both good. Signal values to
make friends/enemies. Signal status to set expectations reflecting real,
dumb-to-ignore asymmetries. It’s when you conflate them that things get
bad. Don’t signal virtue to establish status or vice versa.

Jun 4, 2018 (advice)

One of the healthiest things you can do is simply decide to dislike
certain people, individually, without guilt or rationalization. Simple
dislike is benign. Tribal hatred is the cancerous result of trying
theorize away failures of dumb universal love doctrines.

Jun 5, 2018 (advice)

Sufficiently advanced knowledge is indistinguishable from relationships

Jun 7, 2018 (aphorism)

// Editorial review: manual list item detected; keep/strip/convert decision required
I have overestimated people’s integrity about 2x more often than I have
overestimated their intelligence 🤔 I guess lack of integrity has a
weaker heat signature in things people say than lack of intelligence.
Easier to pretend to be good than smart.

Jun 8, 2018 (reflection)

There is something very boring about American exceptionalism (and
nationalist exceptionalism narratives in general). To truly buy into it,
you kinda have to believe that the rest of the world is a boring
backdrop full of NPCs and procedurally generated filler/world-history
tropes

Jun 11, 2018 (reflection)

In every project, there are people it is important to #emph[not] talk
to. Talking to them merely transforms them from unaware to adversarial,
and puts a target on your back. They are often looking for things to
obstruct, have ‘defender’ self-images, and are addicted to stopping
things

Jun 15, 2018 (advice)

An undocumented benefit of public intellectualizing: the more smart
things you’re perceived to have said, the more dumb questions you’re
allowed to ask with no reputational damage (in fact you get a rep boost
as in “wow, he’s smart but willing to look stupid”)

Jun 16, 2018 (reflection)

I’ve had this idea for a time travel story, “The Unwhatiffable Man”, for
a while. Still working out the fake physics/modal logic, but basically
the main character would be a multiverse constant. No matter which
timeline he traveled to, his own life would be the same shitty one.

Jun 18, 2018 (idea, joke)

The parallels between AI and the 1950s/60s organic chemistry revolution
are quite astounding. Plastics, pharmaceuticals etc really were the
first AI revolution. Just done by programming atoms rather than bits.

Jun 19, 2018 (take)

Hanlon’s twin blade razor: never attribute to malice or incompetence
alone what can be attributed to the combination of the two

Jun 22, 2018 (aphorism, joke)

You usually find out you’ve outgrown something before you figure out
what you’ve grown into

Jun 24, 2018 (aphorism)

It’s not a good idea to teach at the edge of your own knowledge. You’ll
likely mislead people who grok less, and look like an idiot belaboring
the obvious to people who grok more. Either drop the didactic posture or
target level n-3 so you can see 2 steps ahead of students.

Jun 24, 2018 (advice)

If you model classical liberal virtues, deal with events with stoic
grace and honor, and conduct yourself with gentle forbearance (Greek
classics handy) in these trying times, you won’t actually stop the
developing shitshow, but you’ll look good posing in the rubble in your
toga

Jun 26, 2018 (joke)

Interesting how sadness in the environment has to aestheticized to be
made socially acceptable to experience (goth/emo culture, minor chords
in music, wabi sabi, beautiful graveyards). We don’t like banal sadness
like empty airport terminals late night with no services open

Jun 27, 2018 (reflection)

Most people suck at sales/marketing because they like to connect
individually to others. If you’re selling something, even 1:1, it’s
usually best to appeal to a group identity. Tired: people only buy 2
things, happiness and solutions to problems Wired: People only buy
belonging

Jun 27, 2018 (reflection)

The rationalist heuristic to figure out root causes for problems is to
ask “why?” 5 times. Do not use when feeling stupid. The postrationalist
heuristic to figure out root motivations for behaviors is to ask “why
bother?” 5 times. Do not use when feeling depressed.

Jun 29, 2018 (advice)

It is interesting how nature runs without money

Jun 30, 2018 (reflection)

If your ambition doesn’t grow with your wealth, the wealth makes you
stupid by enabling you to delay decisions longer under uncertainty. I’ve
done most of my on-time risky trigger pulling when money was tight,
delayed it too long when it wasn’t

Jul 1, 2018 (reflection, advice)

“Spotting connections across disparate fields” as a default, signature
intellectual style is: — Charming/precocious sign of imagination in your
20s — A workhorse “visionary” playbook in your 30s — The mark of a
dilettante has-been in your 40s

Jul 2, 2018 (reflection)

One of the crueler jokes the universe plays on you is when you start
hitting your own limits and learning of the ways in which you suck, you
think, “maybe it’s time to help the next generation” and then you
realize that’s a much harder problem and you suck even worse at that

Jul 3, 2018 (joke, reflection)

// Editorial review: manual list item detected; keep/strip/convert decision required
I get thie sense powerful criminals have a world-view of “all power is
corrupt, and it’s just arbitrary which gang manages to label the rest
criminals”. Ie a view that the world is run along de facto criminal
lines at the top, and law-abiding conduct is for middle class suckers

Jul 3, 2018 (reflection)

Space opera premise: there is a god, but one who somehow got destroyed
before finishing creating the universe. So we live in an incompletely
built world that is missing obvious things that were meant to exist,
like hyperspace travel and force fields. Title: Cosmic Technical Debt

Jul 4, 2018 (idea)

The main use I have found for stoicism is putting up with stoics

Jul 6, 2018 (joke)

Chesterton’s Trashcan: To figure out what something is for, throw it
away

Jul 6, 2018 (joke, aphorism, advice)

Small dogs are like legacy enterprise branches of ‘wolf’ that nature
stopped supporting several versions ago, and must now be maintained
in-house.

Jul 7, 2018 (joke)

// Editorial review: manual list item detected; keep/strip/convert decision required
I suspect the deepest schism in the world today is not any of the
obvious ones (left/right, race, generations, gender, auth/lib, 1%/99%)
but some non-obvious one induced by technology My candidate is
asynchronous people vs. synchronous people. Text-first vs
phone-interrupt

Jul 8, 2018 (reflection)

Freaky thought: when you talk to someone on the other side of the planet
they’re upside down relative to you It’s ludicrous, I see the appeal of
flat eartherism now..

Jul 11, 2018 (reflection, joke)

Retirement planner product idea: a “where can I afford to retire?”
widget that shows a color-coded world map Red \= in your dreams loser;
try the lottery Orange \= if you work ridiculously hard and get a bit
lucky Yellow \= reachable by 70 Green \= Dump you could retire to right
now

Jul 14, 2018 (joke, idea)

|￣￣￣￣￣￣￣￣￣￣ | | | Best | | Meh | Quadrant | |
\_\_\_\_\_\_\_\_|\_\_\_\_\_\_\_\_\_\_ | | Bad | Meh |
|＿＿＿＿\_\_|＿＿＿＿\_\_\_| (\\\_\_/) || (•ㅅ•) || / 　 づ

Jul 16, 2018 (joke, ascii)

|￣￣￣￣￣￣￣￣￣￣￣| | We demand equal- | | width fonts on Twitter |
|＿＿＿＿＿＿＿＿＿＿＿| (\\\_\_/) || (•ㅅ•) || / 　 づ

Jul 16, 2018 (joke, ascii)

// Editorial review: manual list item detected; keep/strip/convert decision required
A good things about modernity is that you don’t have annoying people
going around starting sentences with “verily I say unto you...” That
must have been really tiresome. You just want to chill after a long day
chopping wood and some dude wants to verily say things unto you

Jul 18, 2018 (joke)

// Editorial review: manual list item detected; keep/strip/convert decision required
A good nights sleep is worth like 50 IQ points for me. A successful
midday nap is worth 20. Failed nap is minus 30.

Jul 19, 2018 (advice, aphorism)

Optimization is a suboptimal primary mental model for decision making
Preference ordering should not be your first preference for analyzing
preferences There are more cons than pros to weighing pros and cons My
priors are that using priors to frame decisions is a lousy idea

Jul 19, 2018 (joke, reflection)

If there’s a single duality that captures the essential difference
between India and the US, it is inconvenience as virtue versus
convenience as virtue. To be Indian is to feel virtuous by navigating
elaborately inconvenient practices for no good reason

Jul 29, 2018 (reflection, take)

Much as I\'m the epitome of the convenience-addicted expat, grousing
about how nothing is easy in India, I do find that every visit recharges
some illegible psyche battery. Reaffirms faith in humanity. If this
billion-person goat rodeo can go on going on, there\'s hope for all.

Jul 30, 2018 (take, reflection)

Damn I just managed to label an illegible new behavior I’ve noticed in
myself in the last few years. “Outsourcing cleverness” Why bother being
clever when it’s cheaper to provoke cleverness in others? Let others
bank the insightcoins. I just want the actual insights.

Jul 30, 2018 (reflection)

Playing by the rules and having roughly the expected outcome is a recipe
for being increasingly bewildered and angry at the world as you age.

Aug 3, 2018 (reflection, advice)

Just hit me that at say \$3/gallon and say 40mpg, it would be cheaper
for a self-driving gasoline car to just circle around than park anywhere
at more than \$3/hr. At say \$0.5/mile fully loaded operating cost it’s
not worth parking above \$20/h. The loitering car cloud is coming.

Aug 4, 2018 (take)

There are four kinds of civilizations: real, utopian, dystopian, and
Japan.

Aug 7, 2018 (joke)

// Editorial review: manual list item detected; keep/strip/convert decision required
I no longer say LMGTFY sarcastically or feel bad about asking googleable
questions. In a noisy, make-your-own-reality world, there’s a real
difference between asking a few known humans live versus querying stored
global memories of questionable provenance.

Aug 8, 2018 (reflection)

Normalcy is simply the majority sect of magical thinking

Aug 9, 2018 (aphorism)

// Editorial review: manual list item detected; keep/strip/convert decision required
A world that’s 100% electrified, with air/sea ships for long-distance
travel, high-quality VR/telepresence, rent-over-own society, high
mobility, hyperloops, rewilded/reforested countrysides, fake meat
without factory farmed farting cows... as exciting a vision as
colonizing Mars

Aug 11, 2018 (reflection, take)

// Editorial review: manual list item detected; keep/strip/convert decision required
A normie is someone who realizes life is much longer than most fads and
plans accordingly

Aug 12, 2018 (aphorism)

// Editorial review: manual list item detected; keep/strip/convert decision required
I have this theory that during periods of intense creative destruction,
old institutions that survive at all turn into de facto banks, defined
primarily by assets they manage to freeze and build moats around. Eg.
Catholic church during reformation turned into a real estate bank

Aug 15, 2018 (take, reflection)

I’m kinda sick of this digital public era with its big, garishly lit
public plazas and broadcast discourse. I would like a dark age marked by
an underground warren of largely hidden spaces. Publics are overrated.
Public figures should feel like the governed are ninjas in shadows

Aug 19, 2018 (take)

You can’t solve with intelligence problems caused by lack of kindness
You can’t solve with kindness problems caused by lack of intelligence A
lot of political failure seems to be rooted in treating the two as
substitutes for one another

Aug 24, 2018 (aphorism, advice, reflection)

// Editorial review: manual list item detected; keep/strip/convert decision required
A system can be more evil than the sum of the evil of its human parts.
If you don’t account for emergent evil, you’ll end up with a useless
morality where you can’t distinguish between people within human range
of good/evil at all. It’s like adding a big constant to your y-axis.

Aug 27, 2018 (reflection)

Humanities people like to lecture STEM people, but damn their clumsy
figurative embrace of sexier math/science ideas is like 10x more
cringeworthy than the ineptitude typically displayed in the other
direction. Not talking everyday innumeracy. That’s understandable &
forgivable.

Sep 1, 2018 (reflection)

The rat race is 10x stronger a reality distortion field than religion.
The equivalent of god is ‘keeping up with the Joneses’. The equivalent
of atheism is disbelieving in the existence of a meaningful comparison
between any two people.

Sep 1, 2018 (reflection)

All societies are built with cognitive guardrails to prevent people from
thinking clearly about mortality. It’s a first-class feature.
Mythologize it, valorize it, spiritualize it, aestheticize it, but don’t
let people think clearly about it. This is stupid. We should change it.

Sep 2, 2018 (reflection, take)

Big dramatic moves to break out of a rut basically never work. Your rut
can make matching big dramatic moves to follow you. You kinda have to
sneak out of the rut when it isn’t looking. Play hooky from rut life or
something.

Sep 7, 2018 (advice)

I’m bored of people on the “this isn’t normal” derpbeat. Normal is over.
Go pro weird or go home.

Sep 7, 2018 (take)

The idea of picking your favorite books from among those stocked in a
small town library or bookstore sounds ludicrous today, now that
everyone basically has access to a big chunk of entire world’s book
catalog. But we’re still kinda expected to pick out friends that way. 🤔

Sep 9, 2018 (reflection, take)

Once I recommend a book to 3+ people, I buy it to read myself. After I
recommend 5 times I crack it open to familiarize myself with it enough
to fake having read it After 7, I feel guilty enough to actually read
it, at which point there’s a 50-50 chance I’ll turn against it

Sep 10, 2018 (joke)

The main purpose of productivity is not to get more things done (that’s
a side effect) but to generate a nice reality distortion field to live
in. Smooth workflows generate a nice bubble reality. Kinda like how
aligned magnetic domains in iron generate a magnetic field.

Sep 11, 2018 (reflection)

Anything artists and creative writers are 100% united against is 100%
sure to happen

Sep 11, 2018 (joke, take)

You know what I would like instead of a watch? An Apple Pen: - Writes on
tablet - Scans text off paper books - Records memos - Directional
microphone for spying - Poison darts - Cap unrolls into tiny
e-notepad/screen - Works as straw - Space for mints - Stabby point

Sep 13, 2018 (joke)

// Editorial review: manual list item detected; keep/strip/convert decision required
I wonder if you could estimate the computational complexity of different
cities to their residents. Like quantify propositions like “London is
50% more demanding than New York.” It’s actually a more meaningful
comparison than cost of living.

Sep 20, 2018 (reflection)

If the universe is computing its own future, all human thought is
speculative execution and branch prediction 🤔 Do we contain
spectre/meltdown type bugs? 🤔🤔

Sep 21, 2018 (reflection)

Things are going to get much worse before they get unbearably awful

Sep 28, 2018 (joke)

Rereading the Hamming You and Your Research essay, I\'m struck by how
simple his formula is. Make your own luck, cultivate drive, make
courageous decisions, work on important\* questions, learn simple
selling skills, avoid bullshit work. \* Important \= \"has a feasible
attack\"

Sep 30, 2018 (advice)

Talent hits the target others can’t hit Genius hits the target others
can’t see Mediocrity sandbags today to live to fight another day Premium
mediocrity sandbags while pretending to aspire to talent and admire
genius

Oct 1, 2018 (joke)

You manage for productivity with incentives (carrots and sticks) You
manage for excellence with esteem and room for self-actualization You
manage for mediocrity with steady evolutionary pressure via automation

Oct 3, 2018 (reflection)

You\'ve heard of homesickness. There is also jobsickness for people
who\'ve been feral free agents too long. When homesickness hits you
make/eat the comfort food you grew up with. When jobsickness hits, you
make slide decks destined for no meeting in particular 🤣

Oct 4, 2018 (reflection)

50% of ‘business skills’ is just knowing how to interrupt and be
interrupted in business meetings.

Oct 4, 2018 (reflection, advice)

// Editorial review: manual list item detected; keep/strip/convert decision required
A mediocre lifegoal is to grow net worth in proportion to your waning
interest in understanding how the world works. The main value of \$ is
buying ignorance privilege. It’s the gentle version of seeking fuck-you
money, for people without the acute trauma that drives FUM-seeking.

Oct 5, 2018 (reflection)

Sometimes it is easier to believe a more complicated theory because the
less complicated theory requires more courage to believe. So smarter
people have to work harder to be courageous. A lot of crackpottery and
consipracy theorizing is simply smart people lacking courage.

Oct 9, 2018 (reflection, take)

When I snap out of procrastination, it is #strong[never] by willing
myself to be a better person. It is always some sort of breakthrough. It
almost invariably turns out that my brain decided to wait for a good
reason.

Oct 10, 2018 (reflection)

Every workflow preference reflects a mild form of a mental illness. My
wild guesses. Agile: ADD (me, early 30s) Waterfall: OCD PMP: Sadism
Flow: psychopath Kanban: Social anxiety GTD/Lean: NPD (me 20s)
Fuguework: psychosis (me post 35) Daily lists: BPD None: Depression

Oct 12, 2018 (take)

Don’t answer questions beginning with “it depends”. It’s tedious and
annoying. Everybody know “it” always “depends”. If you have something
useful to say, frame it as an if/then qualified answer. Like, “I can
speak to 2 cases. If we’re talking X then...and if we’re talking Y...”

Oct 13, 2018 (advice)

Partisanship makes you stupid Centrism makes you clueless Retreat makes
you depressed Privileged exit makes you angry Thoughtful and rigorous
critical analysis makes you frustrated and incomprehensible And you are
expected to choose #strong[wisely] among these options 🤣

Oct 19, 2018 (joke)

Problems expand to occupy the anxiety available

Oct 21, 2018 (aphorism, joke)

The way the 1% run the economy and deploy political power leads me to
conclude that they are all extreme climate believers who’ve concluded
the apocalypse is nigh, cannot be averted, and it’s best to retreat to
Helm’s Deep in New Zealand with 4 generations worth of durable assets

Oct 21, 2018 (reflection, take)

Something I’ve become aware of only recently: effective people have a
bonus-action mentality, where they opportunistically do extra things not
in plans. Sign of surplus energy, orientedness, and nerve. Drives up
average return rate since bonus actions tend to have high margin.

Oct 29, 2018 (reflection)

Language is an underrated source of cognitive leverage. 🤔 I’m at least
3x smarter in English than in Hindi, due to language characteristics
alone, like bigger, more modern vocabulary, non-gendered nouns etc (net
more like 5x since my Hindi is also relatively weaker)

Nov 2, 2018 (reflection)

How to win: 10%: inflict a string of losses on enemy 90%: convince the
enemy that the string of losses is the universe telling them they are
losers An adversary can inflict losses on you by exploiting a temporary
advantage, but only turn you into a loser with your cooperation

Nov 10, 2018 (advice)

These seem solid —- Ask vs guess culture Shame vs guilt culture Honor vs
dignity culture High vs low context Otters vs possums Conflict vs
mistake theory Tight vs loose culture Decoupler vs coupler Combat vs
nurture Debate vs dialogue Wait vs interrupt

Nov 13, 2018 (reflection)

Culture wars feel like nobody has enough of a clue about what’s going on
to be a true Level Boss Villain. It’s mostly idiots scaring themselves
in the dark while a bunch of grifters pick pockets of varying sizes and
psychotic ideologues read epic false narratives into events

Nov 15, 2018 (take)

Going by manosphere tropes, I’m forced to conclude if Genghis Khan were
around today, he’d be hawking vitamins, supplements, and an edutainment
product with “Mastery” in the title. He’d also have a podcast and/or
YouTube fitness video channel.

Nov 15, 2018 (joke)

\"Decorum\", \"civility\", and \"rudeness\" norms are just PC cultures
that spare the sensitivities of the powerful. Most institutions are safe
spaces for the powerful, kept free of triggers and microaggressions for
local top dogs. This is the essential hypocrisy of anti-PCness.

Nov 16, 2018 (reflection, take)

Excellence is really a sign of weakness: you don’t have the strength to
endure the exhausting unproductivity of mediocrity for long. So you go
off and try to do great, perfect things, and cause trouble for the rest
of us.

Nov 20, 2018 (reflection, joke)

You become capable of evil when you discover one other person whom you
consider both weaker than you, and morally inferior to you. Someone you
feel you can righteously hurt without feeling guilt. Innocence ends the
moment you think you’ve met someone you can judge and punish.

Nov 21, 2018 (advice, reflection)

Easily the least compassionate people I’ve ever met are those who’ve
broken out of bad circumstances by individual effort. The depth of their
contempt and hatred towards those still stuck in prisons they’ve broken
out of is astounding. Exit guilt or something.

Nov 21, 2018 (reflection)

You should make decisions when you are 70% certain. I have a whole
backlog of 69.9% certain decisions queued up.

Dec 1, 2018 (advice, joke)

Few things are more depressing than updating a CV. It\'s a case of
seeing yourself like a state.

Dec 7, 2018 (reflection, joke)

You are the sum of all the lessons you refuse to learn.

Dec 9, 2018 (aphorism, joke)

Pyramids are broader at the bottom. Punching down involves making
choices in a target-rich environment. How you punch down always reveals
more about you than who you choose to punch. Your insecurities are
revealed.

Dec 10, 2018 (reflection)

Principle of plausible deniability: Don’t measure what you don’t want to
manage.

Dec 11, 2018 (aphorism)

You have no obligation to be useful or interesting to the world.

Dec 15, 2018 (aphorism)

In general, when people consistently get mad at you, there are 3
possible explanations 1. You\'re a superior Straussian being upsetting
lesser minds 2. You\'re a bold, taboo-shattering thinker 3. You\'re
actually missing something important Don\'t be too quick to diagnose

Dec 18, 2018 (reflection, advice)

There\'s a sort of Dark Tetrad of topics: social identity, intelligence,
religion, and genetics-and-culture. People who only ever talk about
those 4 things... I tend to run from them. For mental health, those
topics should not consume \>20% of your intellectual bandwidth.

Dec 18, 2018 (take, reflection)

You know how on a plane sometimes there’s a cool sight that you can
#strong[just] glimpse if you press face awkwardly against window, but
you don’t get a good look because angle is wrong? And sometimes plane
banks and you get lucky? That’s what chasing down a good idea feels
like.

Dec 23, 2018 (advice)

== 2019
<section-6>
When you\'re used to pursuing excellence, mediocrity feels like freedom

Jan 1, 2019 (joke, aphorism)

The amount of mediocrity in the world can be measured by the extent of
the failure to keep resolutions Excellent people stick to resolutions.
Crappy people never make them in the first place. We mediocre people
make them and then give up easily at the first sign of difficulty.

Jan 6, 2019 (joke)

Wealth is not zero-sum, but power is. If somebody says they want to
empower, ignore what they’re offering to give you. Focus on what they’re
offering to give up.

Jan 10, 2019 (advice, reflection)

Viral outcomes no longer motivate me at any level from tweet to blog to
book. A spike is a very dull kind of attention profile. I like other
kinds better now. The steady burn, the cyclic endemic meme, the incepted
word, the compounding secret language, the long-running gag....

Jan 20, 2019 (reflection)

There’s really only 4 ways to connect with someone of a culture/ideology
you hate at an intellectual level (none guaranteed): 1. Develop a taste
for their food 2. Develop a taste for their music/cultural output 3.
Weather a mortal crisis with them 4. Sleep with them

Jan 23, 2019 (reflection)

Eff it, Yolo, I’m not waiting to see how Gen Z actually shapes up to tag
them. Where’s the fun in data-driven hindsight. I’m calling it early
based on early markers and portents If Millennials are Premium
Mediocre... Gen Z is gonna be Domestic Cozy Yeah you heard it here first

Feb 6, 2019 (take)

Do you want to be part of the problem or part of the other problem?

Feb 12, 2019 (joke)

The annoying thing about computing is that every damn thing has to be
live in an environment of scale. It\'s like if there were only 2 scales
of cooking: cooking a meal for your family at home, and cooking for the
entire planet. No intermediate scales of intermediate difficulty.

Feb 15, 2019 (reflection)

I\'ve often had an idle thought of creating a \"technology
appreciation\" course like art history. It\'s amazing how oblivious
people are to the world they live in. Even when things like factory
tours are open to the public, usually normies only want to go if it\'s a
chocolate factory

Feb 28, 2019 (reflection, idea)

// Editorial review: manual list item detected; keep/strip/convert decision required
A sufficiently mysterious bug is indistinguishable from magic

Feb 28, 2019 (joke, aphorism)

Effective selling of intellectual output rests almost entirely on your
ability to be sincerely impressed by your own thoughts.

Mar 2, 2019 (reflection)

The secret to a meaningful life: Your top interest should be useless,
even to yourself, but not a mere hobby or below-marketable-quality
amateur skill. Your life McGuffin. Anything “useful” is measured by its
ends. The useless thing at the end of all is the measure of everything.

Mar 14, 2019 (reflection, advice)

When you don’t understand algorithms they scare you because they seem
soulless When you do understand them they scare you because they hold up
a mirror to #strong[your] soullessness. “Ah shit, I’m just a
branch-and-bound with pretensions to poetic grace”

Apr 1, 2019 (reflection)

// Editorial review: manual list item detected; keep/strip/convert decision required
I respect some books too much to actually read them

Apr 2, 2019 (joke)

Resumes and job descriptions at their worst are both painstakingly
crafted but highly ineffective clickbait designed to miscast
not-even-wrong candidates into not-even-real-job roles, leading to a
period of entropic loss in the economy.

Apr 5, 2019 (take)

Epistemic obesity: believing too many things relative to the quantity of
reality encountered Epistemic anorexia: believing too few things to
sustain a life Epistemic bullimia: Repeatedly believing too much, then
rejecting it all in a fit of disillusionment till the next bunge

Apr 7, 2019 (reflection)

// Editorial review: manual list item detected; keep/strip/convert decision required
I have a soft spot for people who casually wander back and forth across
the line of crackpottery like they have dual citizenship on the two
sides of it

Apr 19, 2019 (take)

The belief that’s there can only be one stable “reality-based” consensus
reality is our century’s “heavier than air flight is impossible.”

Apr 19, 2019 (take, reflection)

Which word should we misuse and abuse next week? I feel like roughing up
“liminal” a bit. Just ate a liminal cheese and potato puff.

Apr 21, 2019 (joke)

I’m genuinely impressed by people who find a limited but evergreen
schtick that works, hit a lucrative cruising altitude with it, and then
stick with it for decades. It’s like they’ve found their inner Vanna
White. I’d like that. Seems kinda relaxing.

Apr 23, 2019 (take)

An ingroup only coheres properly once all members totally misunderstand
why outgroupers dislike them, and form a consensus around a wildly wrong
statement of the form “they hate us for our X”. Such a statement is the
unacknowledged Axiom 0 of every reality distortion field

Apr 24, 2019 (take)

We’re gonna see a new institutional underground of reimagined secret
societies, lodges, fraternities, sororities etc. Not like medieval ones
associated with universities, religions, or nobility. A cross between
those and hacker collectives, hawala networks etc.

Apr 27, 2019 (take)

“Those who can, do those who can’t teach” vs “The best players rarely
make the best coaches” 🤔

Apr 28, 2019 (reflection)

Your warnings and cautions to others often say more about your
weaknesses and vulnerabilities than about elements in the threat
environment others should be particularly wary of.

May 1, 2019 (reflection)

Poe’s 2nd Law: Every apparently contrarian view is a conformist view for
some group on the Internet

May 10, 2019 (aphorism)

// Editorial review: manual list item detected; keep/strip/convert decision required
A “big” idea is one that: - has megafauna charisma - represents
discontinuous leap wrt its history - arrives as Big Bang in \<5y, not
trickle - redraws boundaries among old ideas - creates new institutional
landscape in \<10y - creates new internal language within that landscape

Jun 1, 2019 (take)

As a macrotrend (not isolated), stock buybacks are the most principled
way for elites to say “as a civilization, we’re collectively out of
ideas and we’re doing this because our next best idea is a big orgy for
senior execs on a private island, and that’s a bad look for 2019.”

Jun 9, 2019 (take)

“Have your API call my API”

Jun 21, 2019 (joke)

// Editorial review: manual list item detected; keep/strip/convert decision required
I often google for the wikipedia entry on things even when I know other
sources are better, because I\'ve become very efficient at parsing
wikipedia pages and extracting what I want to know. Anyone else do this?
Is there a UX word for this? It\'s a kind of imprinting or something

Jul 2, 2019 (advice, reflection)

In a post-scarcity society there will be no leisure because consumption
will acquire all the oppressive characteristics of work. My wife does
most of our our shopping and it definitely looks as demanding as a
make-money job to me. Spending is information work in a complex society

Jul 9, 2019 (take)

My best aphorism is also my worst nightmare: Civilization is the process
of turning the incomprehensible into the arbitrary. Ever since I thought
that thought I’ve been trying to unthink it. It is not a comforting
thought. Turns history into a horror movie.

Jul 12, 2019 (aphorism)

// Editorial review: manual list item detected; keep/strip/convert decision required
A very compact way to explain mediocrity philosophy is this:
non-attachment to finite games (5 words). Unfortunately those who can’t
process the Carse reference will almost certainly misunderstand it.
Carse references are like dependencies on a major library.

Jul 13, 2019 (reflection)

There’s a peculiar sadness to discovering that someone you like is
stupider than you thought. There’s probably a Finnish word for this.
Like discovering a beautiful person is actually a mannequin you happened
to see in bad lighting.

Jul 13, 2019 (reflection)

Arbitrage opp: 2x engineers. By 80-20 rule, they will do 80% of what 10x
engineers can do for typical projects and will be much more readily
available. If you don’t need the last 20% esoteric skills like quantum
blockchain, you’re good. 80-20: The OG mediocrity hack.

Jul 14, 2019 (advice, joke)

If you enjoy Rick and Morty, you’re most likely a Jerry. Think about it.

Jul 16, 2019 (joke, take)

When you’re offered a seat on a non-rocketship with no engines that’s
clearly going nowhere, you must ask as many questions as you can about
the seat.

Jul 21, 2019 (joke)

Humanity’s last message broadcast to the stars will be “the actions of a
few do not reflect our values as a species or who we are as a planet.”

Jul 24, 2019 (joke)

Next time I hear a phrase like “drawing on findings from psychology,
neuroscience, and evolutionary biology...” I’m going to throw something
something I swear. Sick of glib, lumpenpseudosynthesis Big History by
halo-hedgehogs for people rubbernecking at history through TED talks

Jul 27, 2019 (take)

If you decide the world’s problems are too big for you to worry about,
and stop thinking about them, your own little problems will grow in your
imagination until they’re as big as the ones you decided to ignore.
Anxiety finds its level.

Aug 23, 2019 (reflection)

When something is an economic asset but a psychological liability, like
a house you hate but can’t afford to sell, it’s not the asset, you are.
You don’t own it, it owns you. Middle class life is full of such
antiassets. Very unfortunate people also have antirelationships.

Aug 24, 2019 (reflection)

Learning from people disagreeing with you is overrated. Most of my
learning from others actually comes from them thinking thoughts I
wouldn’t be likely to think at all. And most tend to neither confirm or
disconfirm anything I already think, but expand the scope of thoughts.

Aug 31, 2019 (reflection)

“It is amazing what you can accomplish if you do not care who gets the
credit.” — Truman This is true in a radical way. Not caring about credit
\= 10x to 100x leverage. Oddly enough, not true of caring about who gets
the \$. You have to care about that, even if it isn’t you.

Sep 1, 2019 (reflection, advice)

“You divide, I choose” remains the single most basic principle of
voluntary social organization. It really should be an entire academic
subdiscipline by itself, like prisoner’s dilemma

Sep 2, 2019 (take)

Medicine has a good term, “Idiopathic” to mean “we have no clue what’s
going on, but we’re still the experts you should defer to”. It indicates
an expert ignorance All fields need a term like that. How about we use
“cryptodynamic” for matters of expert engineering ignorance?

Sep 4, 2019 (advice, joke)

// Editorial review: manual list item detected; keep/strip/convert decision required
I get the sense the deeper you penetrate into elite circles, the harder
it is to not become complicit in something ugly. To some extent elite
cultures are almost defined by a kind of mutually assured destruction
condition of collective complicity.

Sep 8, 2019 (reflection)

Astrology is nonsense, but what makes it particularly valuable is that
it is #strong[complete] nonsense in a way myers-briggs for eg is not. It
is valuable for personality theorizing the way random number generators
are good for algorithm design

Sep 9, 2019 (reflection)

💩 The best way to design a complex system is to start by designing its
poop. 💩Poop is the most basic boundary condition for a dissipative
system. 💩 When the simulators designed the cosmos, they started with
large-scale poop structure 💩 No I do not need Freudian therapy

Sep 14, 2019 (reflection)

It is really odd that we often forget we don’t actually #strong[know]
anything about the future. That’s kinda what makes it the future. The
future is that which we know nothing about. Everything we think we know
is necessarily some distorted refraction/reflection of the past

Sep 17, 2019 (reflection)

New \#lifegoal — have a unit named after me, like ampere, richter etc 1
rao should ideally measure something sociophysical, like the entropy
increase caused by 1 standard shit post

Sep 20, 2019 (joke)

Goodhart’s Law: When a measure becomes a metric, it ceases to be a good
measure Badhart’s Corollary: But damn it makes for good myth and
ceremony

Sep 27, 2019 (joke)

Today, I explained to a not-Very-Online person why I take social media
personas very seriously, and believe them over irl personas as truer
expressions of self: social media personas are a kind of mask trance (in
Johnstone/Impro sense). Specifically, true-name mask trances

Oct 5, 2019 (reflection)

If you’re not building your own extended universe what are you even
doing

Oct 5, 2019 (reflection)

Never work 100% within an ideology. Once you do something
#strong[successful] that partially defines a party line, the party line
will totally define you, and people will approach you with a “derp or
beef” filter. Success insurance: Always color at least a little outside
of all lines.

Oct 9, 2019 (advice)

Comedy \= tragedy + time Comedy - tragedy \= time No negative durations
so time \> 0 comedy - tragedy \> 0 comedy \> tragedy QED

Oct 22, 2019 (joke)

The giant sucking sound of imagination draining from Silicon Valley 😐
There’s a dreary shift towards stagey tech-for-tech’s-sake charismatic
stunts I don’t like. Like late-style artists making tediously
overwrought self-indulgent works.

Oct 24, 2019 (take)

When I meet true nerds I realize I’m not one

Oct 24, 2019 (reflection)

// Editorial review: manual list item detected; keep/strip/convert decision required
I saw a guy living Not reading must-reads Not listening to must-listens
Not watching must-watches Not knowing must-knows Not suffering
atemporality Just living in his own timeline, like a psychopath

Oct 25, 2019 (joke)

The idea of fixing social media through regulation feels a bit like
fighting forest fires with earthquakes

Oct 31, 2019 (joke)

That feeling that all your peers have figured out something important
you haven’t and moved on in some way, leaving you behind? It’s
universal. It’s the human galactic red shift. Everybody has figured out
something unique and is receding from everybody else. Happy Halloween 💀

Nov 1, 2019 (joke, reflection)

Get 8 hours of sleep a day, drink 8 glasses of water, eat healthy and
exercise, and be slightly mean to everybody That’s the formula, the only
question is what it computes

Nov 1, 2019 (joke)

The more you learn about how the human-built world works the more it
seems like a miracle it works at all. The civilizational stack is a
late-game Jenga tower held together by Narrativium.

Nov 3, 2019 (reflection)

things get monetizable just when they get uninteresting

Nov 4, 2019 (aphorism)

It is a category error to apply the adjective “broken” to a complex
system. Complicated systems break. Complex systems merely enter hostile
configurations relative to you. They’re not fixed by your functional
expectations. They’re still working. Just not for you.

Nov 6, 2019 (reflection)

// Editorial review: manual list item detected; keep/strip/convert decision required
I just realized I genuinely like playing to caricatures of me that
people might hold. Discover the cartoon view people have of you and own
it to absurdity. Usually there is some truth to it that people are
picking up on, but don’t fully realize the implications actions of.

Nov 15, 2019 (reflection)

True or false? The most popular science fiction tends to envision a
high-agency future for the professional class, irrespective of whether
it us otherwise utopian or dystopian. They’re middle class
professionals. The most popular fantasy otoh seems to focus on high or
low born.

Nov 29, 2019 (reflection)

“Moved to bcc” is the finest modern innovation in manners

Dec 5, 2019 (take)

// Editorial review: manual list item detected; keep/strip/convert decision required
I have discovered the ultimate antiaphorism: “this is incorrect.” If
someone offers a glib aphorism as an argument, defense, definition, or
insight, you can use “this is incorrect” to okboomerize it. Don’t offer
a reason. It only works as a fiat rejection if left unjustified.

Dec 6, 2019 (advice)

Lately I\'ve been hanging out with a lot of respectable academics and
realizing just how sketchy I\'ve allowed myself to become 😝 No wonder
at 45 my parents still treat me worriedly like a 25-year-old derailee
who needs to get back on track. Derailee is a good word.

Dec 7, 2019 (reflection)

LA is a terrestrial total perspective vortex, it makes sure you
understand how tiny and insignificant you are in the human story, and
how little of it you will ever experience in a single lifetime. Walk a
few blocks, and you\'ll notice a dozen portals to worlds you\'ll never
enter

Dec 16, 2019 (reflection)

Seated meditation: sit cross legged, head bowed, eyes half-closed, palms
facing up on lap , fingers interlaced lightly (swipeya mudra), phone
cradled lightly in fingers, open to twitter. Breathe gently, scrolling
with left thumb, in a rosary counting motion. Like alternate twetes

Dec 24, 2019 (joke)

The biggest idea I learned this year was probably the power of indices
and indexing Approaching an ideaspace from an index view is attacking a
monster via its underbelly

Dec 24, 2019 (reflection)

Money forms an amnesiac boundary condition. The moment you solve a
problem with money you forget what it’s like to have the problem at all
and consider the idea that others still have it a conspiracy theory.

Dec 28, 2019 (reflection)

Ok Mandalorian is good, 3 episodes in. Better than all the movies. It’s
like cyberpunk Ironman in space from planet of iron men. It’s the art of
gig, in space.

Dec 31, 2019 (take)

== 2020
<section-7>
Most papers are meh Academic research is about producing meh to dampen
the enthusiasm of bloggers for insight porn The result of most research
is “it is complicated and depends on various factors in unhelpfully
nuanced ways” ie, meh.

Jan 6, 2020 (reflection)

I’m increasingly starting to believe this thing has been spreading
globally for at least several weeks longer than we think. All the
containment was closing stable doors after the horse had bolted. The
apparent pandemic is a delayed detection pandemic as testing spreads.

Mar 7, 2020 (unclassified)

If we can’t make cruise ships work there is no hope for space ships

Mar 10, 2020 (take)

The future is already here, it is just unevenly locked down

Mar 28, 2020 (joke)

“Life is a tragedy for those who feel, a comedy for those who think”
Sometimes growth happens when thinkers are forced by circumstances to
feel, and feelers are forced by circumstances to think. For a while,
there is neither tragedy, nor comedy. Only transformation.

Apr 1, 2020 (reflection)

Anthropocene calendar of seasons January: Wildfires February: Locusts
March: Pandemic April: Volcano Can’t wait for the rest.

Apr 11, 2020 (joke)

Growing sense of genuine evil in the air. Not things that could be
interpreted as evil, or have effects that some might judge
unconscionable. Genuine evil as in deeply malicious intentions directed
against large groups and pursuing plans that aim to harm them as primary
effect.

Apr 12, 2020 (reflection)

Money as compensation for work is an absurd mental model to have right
now. Go brrr is changing the fundamental meaning of money. I don’t know
what it means anymore.

Apr 17, 2020 (reflection)

One unpopular opinion I hold relative to people with similar politics to
mine: I could never stand The West Wing. Awful, pretentious,
self-congratulatory show, of/by/for “elites” in the demonized sense used
today.

Apr 18, 2020 (take)

I\'m going to throw something at the next person who uses the phrase
\"now more than ever\"

Apr 25, 2020 (take)

Ok people I’m sick of boring views of your damn sourdough experiments,
cute parenting moments etc. Do something larger than domestic scale in
ambition even if you have to do it at home. Like a video of you
soldering together a small Death Star. Or at least an iron man suit.

May 6, 2020 (take)

You know what would be worse. A weird anticorona virus that forced us to
always be in tight huddles always with 6 people less than 6 feet away.
In crowds of minimum 50. Except to go to the bathroom. Social closening.
That sounds horrifying.

May 6, 2020 (reflection)

90% of “success” seems to be just surviving long enough above a
threshold of “terminally embarrassing” The other 10% is judicious
editing of the story ex-post

May 9, 2020 (reflection)

That awkward mid-game of collapse where it’s too late for clever stock
market trades but too early to start forming your mad max tribe

May 10, 2020 (take)

Cynicism: never being wrong by never being interesting

May 11, 2020 (aphorism)

Being a mirror is so much more interesting than being either a subject
or an object

May 16, 2020 (reflection)

Social darwinism is a terrible thing, but activity darwinism around
projects is a beautiful thing. So many things improve if you\'re willing
to be darwinian about culling weak projects. The trick is killing weak
projects without destroying people or knowledge in the process.

May 22, 2020 (reflection, advice)

Zen Master: What is the sound of one hand clapping? Zen Student: brrrr
Zen Master: Nooooo...you can\'t just answer a mu question. You have to
satori enlightenmenterinoooo! Zen Student: brrrr! brrrr! brrrrrrrrrrr!

May 26, 2020 (joke)

Failure debt: All the failures you’ve avoided, coming due in one big
failure.

Jun 3, 2020 (joke)

Last 4 months since lockdown began have had the feel of an airport
terminal waiting area post-security with a delayed flight that may be
canceled. Can’t leave, nice environs, but limited options and
uncomfortable seats. And can’t really relax till you’re on the plane
taking off.

Jun 6, 2020 (take)

None of us has more than part of a partially correct answer to any
complex question. But if we sincerely and humbly work hard together, we
can at least discover the whole of a completely wrong answer, and ride
it to absolute disaster.

Jun 16, 2020 (joke)

Heisenwoke: when people can’t tell if you’re woke or antiwoke, and you
exist in a state of complex superposition of being canceled and
not-canceled at the same time.

Jun 21, 2020 (joke)

Expertise, good faith, influence Pick 2 of 3

Jul 2, 2020 (aphorism)

Damn what a week. I did more than one thing per day on at least 3 days.
How the hell do you guys do this all the time.

Jul 4, 2020 (advice, joke)

Never trust anyone with a legible aesthetic

Jul 5, 2020 (aphorism, advice)

90% of project management is simply accepting being the person who cares
the most. Being too cool to visibly care is a minor liability for an
individual contributor but a guaranteed project killer in a manager.
Once the PM visibly doesn’t care, everyone else checks out.

Jul 7, 2020 (advice, reflection)

Thought to close out the week: Never mistake access for intelligence.
The smarter the environment, the more the two look alike.

Aug 22, 2020 (aphorism, reflection)

Half of me has regressed to age 12 and the other half skipped ahead to
age 80. Barbell aging strategy.

Sep 7, 2020 (joke)

so how are you de-elitifying yourself?

Sep 12, 2020 (joke)

I’ll never stop being amused by how individualism in America is the most
collectivist larp ever

Sep 28, 2020 (aphorism, take)

Laws of institutions Law 0: They #strong[will] emerge, even if nobody
wants them Law 1: They #strong[will] get captured, regardless of values
Law 2: Once captured they #strong[will] do bad things, even if they only
contain good people The only real question is how soon there are
consequences

Oct 15, 2020 (reflection)

All knowledge eventually turns into a file format. All work eventually
turns into file format translation work. 🤬

Oct 24, 2020 (joke)

I’m not idling, I’m cognitive crop rotating

Oct 24, 2020 (joke)

Serendipity is the only source of infinite game motivation. When you
feel surprisingly lucky, you feel like continuing to play.

Oct 26, 2020 (reflection)

Increasingly convinced manifestos and codes of conduct are mostly for
people who secretly want to avoid the grunt work of actually building
organizations. They’re as bad as free-marketers. Thinking Values can do
all the heavy lifting is the same as thinking Prices can do it all.

Oct 30, 2020 (reflection)

// Editorial review: manual list item detected; keep/strip/convert decision required
I only argue in darmok now

Oct 30, 2020 (joke)

Debate: a mechanism for classical liberals to converge on each other in
the belief that they are converging on the truth.

Nov 11, 2020 (joke)

The paradox of exceptionalism. The more exceptional you believe your
nation is, the more typical you make it by your behaviors.
Exceptionalists are the most universal and predictable subspecies of
nationalist.

Nov 11, 2020 (reflection)

the conceit of this age is the belief that we’ve recognized our problems

Nov 13, 2020 (reflection)

the invention of emojis really expanded my emotional range considerably

Nov 14, 2020 (joke, reflection)

If the biggest risk in your life is bureaucracy risk, you’re not taking
enough risk.

Nov 20, 2020 (aphorism, reflection)

Getting really tired of Age of Explainers. I’d like motivators instead.
Essays explaining why something is worth understanding rather than how
it works. MLI5. Motivate like I’m 5. I’ll look up the explainer if you
succeed in getting me to care at all.

Nov 21, 2020 (reflection)

goals are memories of the future

Nov 29, 2020 (aphorism)

Yet another day when the answer to the question of life, the universe,
and everything turns out to be “you’re dehydrated, drink more water”

Nov 30, 2020 (reflection, advice)

I’m not an artisan, I just suck at scaling. I mass produce one-offs.

Dec 22, 2020 (joke)

“It’s never over, you just let go” good line from The Boys. Context
irrelevant.

Dec 29, 2020 (aphorism, take)

== 2021
<section-8>
What if this is the actual start of the simulation and everything before
was just unit tests?

Jan 7, 2021 (joke)

Of course you have to make up the rules as you go along 🙄 The other way
is called religion

Jan 9, 2021 (joke)

“Everybody leaves behind unfinished business, that’s what dying is.” —
Amos on The Expanse

Jan 10, 2021 (aphorism, take)

robots are the answer the more robots we build and deploy visibly in
more places, the more we\'ll have an appetite for new polities that are
breaks from the past that\'s my political slogan Build More Robots

Jan 10, 2021 (reflection)

Life is a comedy to those who think, and a tragic font choice for those
who feel Face it feelers, the universe is written in Comic Sans

Jan 27, 2021 (aphorism, joke)

// Editorial review: manual list item detected; keep/strip/convert decision required
I can never tell civility apart from passive aggression. This is why I
never join any forum that makes a virtue out of civility.

Jan 30, 2021 (aphorism, reflection)

the basic question of life is how hard you want to work at it once you
answer that honestly everything else follows trivially good to re-ask
periodically or when you feel your energy level shift

Feb 1, 2021 (advice)

Once more in that bad mood caused by looking up at the sky and realizing
I’m never getting off this damn planet.

Feb 6, 2021 (aphorism, reflection)

Sitting is not the new smoking. It’s a lousy analogy. We’re designed for
it. 8 hours is probably bad, but the optimal amount is not 0. We like it
too much and will never stop. Even animals sit around a lot. Ape troops
for eg. If sitting is smoking my cat is a chain smoker.

Feb 21, 2021 (reflection, take)

You can’t truly get interested in the world till you lose interest in
money

Feb 26, 2021 (aphorism, advice, reflection)

Lifestyle design is easy. It’s the debugging that takes time.

Mar 1, 2021 (joke)

The thing about staying above the API is that the API level keeps going
up

Mar 1, 2021 (joke)

Learning the art of not being managed is a prerequisite for learning the
art of not being governed

Mar 7, 2021 (reflection)

Kinda amazing that biology works without measurement. There’s no “what
size bolt does a baby need?” or “do you want a metric or imperial
banana.” No SAE or IEEE standards committees.

Apr 10, 2021 (reflection)

\"Are you enjoying yourself?\" \"Are you enjoying your self?\" Adding a
space can turn a superficial question into a profound one 🤔

May 18, 2021 (reflection)

The more energy you can generate the less philosophy you need

May 29, 2021 (aphorism)

The world can absurdum longer than you can reductio

Jun 16, 2021 (aphorism)

Weird how The Simpsons has vanished from the zeitgeist without a trace.
Even the good seasons.

Aug 28, 2021 (take)

Things not open enough to enjoy the world, things not closed enough to
enjoy home Pandemic is now in annoying purgatory stage

Sep 11, 2021 (take)

geodysphoria: unhappiness living wherever you are and wanting to move
and settle somewhere else indeterminate

Sep 15, 2021 (reflection)

In the biggest economic races, money is how you figure out who won
second and third place. First place usually goes to the player who
redefines the meaning of money. Fourth place onwards is people who
console themselves that they’re working for “human values.”

Sep 26, 2021 (reflection)

// Editorial review: manual list item detected; keep/strip/convert decision required
I rarely think “TMI” but often think “TMP” — Too Much Production. People
overproducing their oversharing like it’s a summer blockbuster 🧐 TMI
can be charming and endearing even when cringe. TMP never is. At least
not on Twitter. Maybe it works on Instagram or tiktok.

Sep 30, 2021 (reflection)

Covid has led me to a new definition of religion. Religion the belief
that the universe is nice to good people. “God” is an optional extra.

Oct 3, 2021 (reflection)

Emerson: “The mind, once stretched by a new idea, never returns to its
original dimensions.” Is the inverse true or false? A mind once
contracted by a major crisis,\* never returns to its original
dimensions. \* ‘crisis’ seems like natural inverse, as in crisis of
faith/meaning

Oct 4, 2021 (reflection)

Open source is apparently when everybody wants to do the project but
nobody wants to do project management and you figure out ways to get by
without it

Oct 5, 2021 (reflection, take)

Automation is the art of replacing one person doing a full-time job with
a thousand people doing 5-minute software admin tasks every few months
(6 minutes if you forgot your password).

Oct 20, 2021 (take)

can\'t believe I can shitpost all day and nobody can stop me if I\'d
tried this before the internet on a street corner, they\'d have had me
taken away

Oct 27, 2021 (reflection)

Wise person once told me: you’re always going to piss off some group if
you try to do anything at all, and the trick is to piss off the right
people. I have since modified to: piss off a unique set of people. If
your hater list matches another person’s too much it’s just a tribe

Nov 18, 2021 (advice)

Meaning is memorability. Dunno why it took me so long to see this.

Nov 19, 2021 (aphorism, reflection)

The costliest signal is simply showing up repeatedly

Dec 1, 2021 (aphorism)

The main task of postmodern social system design is enabling people to
avoid each other in fine-grained ways. “Connecting people” is the clever
false-flag name for this. People seeking connection will usually find a
way. It’s the right degree of avoidance that’s hard to arrange.

Dec 1, 2021 (take)

unnecessary hard work is the root of all evil

Dec 21, 2021 (aphorism)

I\'ve never gotten addicted to a more completely useless concept than
bouba/kiki. There are literally no stable valences. The pair is a
semantic WMD.

Dec 21, 2021 (reflection)

If you suddenly pick up a cat and put it down somewhere else in a
different orientation, they kinda just go with it. No apparent surprise
or being upset at interruption of whatever they were up to. Just start
doing new thing. Instantaneous transients. Zero inertia OODA loops. 🤔

Dec 22, 2021 (reflection)

== 2022
<section-9>
If you promise obviously utopian outcomes don’t complain if critics
evaluate you against absolutely idealistic standards

Feb 4, 2022 (take)

We need some lovecraftian cosmic horror. Not the actual thing or even an
updated pastiche of it, but an equivalent memeplex that does for the
2020s what his cosmic horror world did for the 1920s. We are so busy
being outraged, we have no time to be properly horrified.

Feb 6, 2022 (take)

Just found myself using the phrase “fog of vibes”

Mar 5, 2022 (take)

My biggest disagree-but-commit position is that the AGI people are
deeply and entirely wrong at the foundations about every philosophical
question but will still invent the next era of computing. Like the
alchemists seeking the philosopher’s stone invented chemistry.
tweet#footnote[\@ilyasut —
https://twitter.com/ilyasut/status/1505754945860956160]

Mar 23, 2022 (take)

My energized projects are not well-defined and my well-defined projects
are nor energized 🤬

Apr 18, 2022 (joke)

Weird there’s no proper measure of how lucky someone is, relative to
expectations based on birth circumstances (genes, family situation,
etc). I think I’m about +1 sigma luckier relative to mean of my birth
bell curve

May 14, 2022 (reflection)

Nobody says “behold” unironically any more even though there is so much
to behold

Jun 4, 2022 (aphorism, reflection)

Negative-sum games seem insufficiently theorized. Zero-sum games are way
overtheorized given how rare they are irl.

Jun 9, 2022 (reflection)

Lately I find when I’m stuck, trying out a different base emotion is
often more effective in getting unstuck than trying to reframe from a
different perspective intellectually

Jun 12, 2022 (reflection)

The correct antonym of “strategic” is “exhaustive” not “tactical”

Jun 13, 2022 (aphorism, advice, reflection)

Your life flashing before your eyes is really the simulators retrieving
the log file for analysis

Jun 25, 2022 (joke)

We’re now in inverse hanlon’s razor territory. Never attribute to
incompetence what can safely be attributed to malice.

Jun 30, 2022 (aphorism, take)

Marie Kondo “does it spark joy” but for process rather than things 🤔

Jul 7, 2022 (reflection)

The difference between past and future is that the past doesn’t go away
when you stop believing in it

Jul 23, 2022 (reflection)

If you have fuzzy talents you get fuzzy rewards If you have hard-edged
talents, you get hard-edged rewards This is the talent-reward fuzziness
impedance match theorem

Jul 23, 2022 (aphorism)

Napping is like turning yourself off and on again to fix some annoying
glitch in your thinking you don’t want to actually troubleshoot. Works
90% of the time. Everything should be rebooted often.

Aug 5, 2022 (joke, reflection)

Most tv shows fail the “would you watch this if the people were less
pretty” test 🤔

Sep 1, 2022 (take)

It\'s really dangerous when humans stop pretending they know what
they\'re doing en masse. At least 20% have to pretend or the whole
system falls apart. That\'s the real danger of quiet quitting. It\'s not
the slacking, it\'s the admission of incomprehension.

Sep 12, 2022 (reflection)

Fear of “loss of face” rules the world. There’s a weird American
presumption that it is an Asian psychodynamic. Asians are just more
willing to cop to it and talk about it openly.

Sep 29, 2022 (take)

The arc of the moral universe is basically a random walk. Occasionally a
tech advance makes it slightly easier to be nice and slightly more
pointless to be nasty. We then congratulate ourselves for having
“evolved.”

Oct 20, 2022 (reflection)

People who supposedly live “fuller lives” mostly seem to have lives more
full of chores 🧐 It’s a scam

Oct 25, 2022 (joke)

Product idea: AI dreamcatcher. You wake up in the morning and
immediately describe your dream before you forget and it generates it.

Oct 30, 2022 (idea)

Designate twitter a digital national park

Nov 5, 2022 (take)

The “panic now, it’s all melting down” tribe and “lol nothing’s
happening, chill” tribes are in a mutual gaslighting equilibrium Me, I’m
above these simplistic either/or things, I’m in a complex superposition
of chilling and panicking.

Nov 9, 2022 (joke)

To become properly evil you must start by installing the belief “I am
good” at the foundation of everything else.

Nov 10, 2022 (aphorism)

// Editorial review: manual list item detected; keep/strip/convert decision required
I think the high-gravitas people are going to survive this much better
than the shitposters The air is being sucked out of the rooms mainly
from in the lower shitposter floors.

Nov 17, 2022 (reflection, take)
