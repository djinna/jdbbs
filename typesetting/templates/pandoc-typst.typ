// Pandoc template for Protocolized Anthology series
// Use with: pandoc -f markdown -t typst --template=pandoc-typst.typ

#import "series-template.typ": *

#show: book.with(
  title: $if(title)$"$title$"$else$"Untitled"$endif$,
  $if(subtitle)$subtitle: "$subtitle$",$endif$
  $if(author)$author: "$author$",$endif$
)

$if(toc)$
// Table of Contents
#set page(numbering: "i")
#counter(page).update(1)

#toc-heading

#outline(
  title: none,
  depth: 1,
)

#pagebreak()
$endif$

// Body
#set page(numbering: "1")
#counter(page).update(1)

$body$
